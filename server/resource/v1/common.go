/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package v1

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chassis/cari/pkg/errsvc"

	"github.com/apache/servicecomb-kie/server/cache"
	"github.com/gofrs/uuid"

	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	"github.com/go-chassis/cari/rbac"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/pubsub"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
	"gopkg.in/yaml.v2"
)

// const of server
const (
	HeaderUserAgent    = "User-Agent"
	HeaderSessionID    = "X-Session-Id"
	HeaderSessionGroup = "X-Session-Group"
	AttributeDomainKey = "domain"

	FmtReadRequestError = "decode request body failed: %v"
)

func NewObserver() (*pubsub.Observer, error) {
	id, err := uuid.NewV4()
	if err != nil {
		openlog.Error("can not gen uuid")
		return nil, err
	}
	return &pubsub.Observer{
		UUID:  id.String(),
		Event: make(chan *pubsub.KVChangeEvent, 1),
	}, nil
}

// err
var (
	ErrInvalidRev = errors.New(common.MsgInvalidRev)

	ErrMissingDomain  = errors.New("domain info missing, illegal access")
	ErrMissingProject = errors.New("project info missing, illegal access")
	ErrIDIsNil        = errors.New("id is empty")
)

// ReadClaims get auth info
func ReadClaims(ctx context.Context) map[string]interface{} {
	c, err := rbac.FromContext(ctx)
	if err != nil {
		return nil
	}
	return c
}

// ReadDomain get domain info
func ReadDomain(ctx context.Context) string {
	c := ReadClaims(ctx)
	if c == nil {
		return "default"
	}
	if d, ok := c["domain"].(string); ok {
		return d
	}
	return "default"
}

// ReadLabelCombinations get query combination from url
// q=app:default+service:payment&q=app:default
func ReadLabelCombinations(req *goRestful.Request) ([]map[string]string, error) {
	queryCombinations := req.QueryParameters(common.QueryParamQ)
	labelCombinations := make([]map[string]string, 0)
	for _, queryStr := range queryCombinations {
		labelStr := strings.Split(queryStr, " ")
		labels := make(map[string]string, len(labelStr))
		for _, label := range labelStr {
			l := strings.Split(label, ":")
			if len(l) != 2 {
				return nil, errors.New("wrong query syntax:" + label)
			}
			labels[l[0]] = l[1]
		}
		if len(labels) == 0 {
			continue
		}
		labelCombinations = append(labelCombinations, labels)
	}
	if len(labelCombinations) == 0 {
		return []map[string]string{{"default": "default"}}, nil
	}

	return labelCombinations, nil
}

// WriteErrResponse write error message to client
func WriteErrResponse(context *restful.Context, code int32, msg string) {
	configErr := config.NewError(code, msg)
	context.Resp.Header().Set(goRestful.HEADER_ContentType, goRestful.MIME_JSON)
	context.WriteHeader(configErr.StatusCode())
	b, err := json.MarshalIndent(configErr, "", " ")
	if err != nil {
		openlog.Error("can not marshal:" + err.Error())
		return
	}
	err = context.Write(b)
	if err != nil {
		openlog.Error("can not marshal:" + err.Error())
	}
}

func WriteError(context *restful.Context, err error) {
	svcErr, ok := err.(*errsvc.Error)
	if !ok {
		svcErr = config.NewError(config.ErrInternal, err.Error())
	}
	WriteErrResponse(context, svcErr.Code, svcErr.Message)
}

func readRequest(ctx *restful.Context, v interface{}) error {
	if ctx.ReadHeader(common.HeaderContentType) == common.ContentTypeYaml {
		return yaml.NewDecoder(ctx.ReadRequest().Body).Decode(v)
	}
	return json.NewDecoder(ctx.ReadRequest().Body).Decode(v) // json is default
}

func writeYaml(resp *goRestful.Response, v interface{}) error {
	if v == nil {
		resp.WriteHeader(http.StatusOK)
		return nil
	}
	resp.Header().Set(common.HeaderContentType, common.ContentTypeYaml)
	resp.WriteHeader(http.StatusOK)
	return yaml.NewEncoder(resp).Encode(v)
}

func writeResponse(ctx *restful.Context, v interface{}) error {
	if ctx.ReadHeader(common.HeaderAccept) == common.ContentTypeYaml {
		return writeYaml(ctx.Resp, v)
	}
	return ctx.WriteJSON(v, goRestful.MIME_JSON) // json is default
}
func getLabels(rctx *restful.Context) (map[string]string, error) {
	labelSlice := rctx.Req.QueryParameters(common.QueryParamLabel)
	if len(labelSlice) == 0 {
		return nil, nil
	}
	labels := make(map[string]string, len(labelSlice))
	for _, v := range labelSlice {
		v := strings.Split(v, ":")
		if len(v) != 2 {
			return nil, errors.New(common.MsgIllegalLabels)
		}
		labels[v[0]] = v[1]
	}
	return labels, nil
}
func revNotMatch(ctx context.Context, revStr, domain string) (bool, error) {
	rev, err := strconv.ParseInt(revStr, 10, 64)
	if err != nil {
		return false, ErrInvalidRev
	}
	latest, err := datasource.GetBroker().GetRevisionDao().GetRevision(ctx, domain)
	if err != nil {
		return false, err
	}
	if latest == rev {
		return false, nil
	}
	if latest < rev {
		openlog.Warn("the rev param is larger than db rev: db may be restored")
	}
	return true, nil
}
func getMatchPattern(rctx *restful.Context) string {
	m := rctx.ReadQueryParameter(common.QueryParamMatch)
	if m != "" && m != common.PatternExact {
		return ""
	}
	return m
}
func eventHappened(waitStr string, topic *pubsub.Topic, ctx context.Context) (bool, string, error) {
	d, err := time.ParseDuration(waitStr)
	if err != nil || d > common.MaxWait {
		return false, "", errors.New(common.MsgInvalidWait)
	}
	happened := true
	o, err := NewObserver()
	if err != nil {
		openlog.Error(err.Error())
		return false, "", err
	}
	topicName, err := pubsub.AddObserver(o, topic)
	if err != nil {
		return false, "", errors.New("observe once failed: " + err.Error())
	}
	select {
	case <-time.After(d):
		happened = false
		pubsub.RemoveObserver(o.UUID, topic)
	case <-o.Event:
		prepareCache(topicName, topic, ctx)
	}
	return happened, topicName, nil
}

// size from 1 to start
func checkPagination(offsetStr, limitStr string) (int64, int64, error) {
	var err error
	var offset, limit int64
	if offsetStr != "" {
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if offset < 0 {
			return 0, 0, errors.New("invalid offset number")
		}
	}

	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil || (limit < 1 || limit > 100) {
			return 0, 0, errors.New("invalid limit number")
		}
	}
	return offset, limit, err
}

func validateGet(domain, project, kvID string) error {
	if kvID == "" {
		return ErrIDIsNil
	}
	return checkDomainAndProject(domain, project)
}

func validateDelete(domain, project, kvID string) error {
	return validateGet(domain, project, kvID)
}

func validateDeleteList(domain, project string) error {
	return checkDomainAndProject(domain, project)
}

func checkDomainAndProject(domain, project string) error {
	if domain == "" {
		return ErrMissingDomain
	}
	if project == "" {
		return ErrMissingProject
	}
	return nil
}
func queryFromCache(rctx *restful.Context, topic string) {
	rev, kv, queryErr := cache.CachedKV().Read(topic)
	if queryErr != nil {
		WriteErrResponse(rctx, queryErr.Code, queryErr.Message)
		return
	}
	rctx.ReadResponseWriter().Header().Set(common.HeaderRevision, strconv.FormatInt(rev, 10))
	err := writeResponse(rctx, kv)
	rctx.ReadRestfulRequest().SetAttribute(common.RespBodyContextKey, kv.Data)
	if err != nil {
		openlog.Error(err.Error())
	}
}
func queryAndResponse(rctx *restful.Context, request *model.ListKVRequest) {
	rev, kv, queryErr := kvsvc.ListKV(rctx.Ctx, request)
	if queryErr != nil {
		WriteErrResponse(rctx, queryErr.Code, queryErr.Message)
		return
	}
	rctx.ReadResponseWriter().Header().Set(common.HeaderRevision, strconv.FormatInt(rev, 10))
	err := writeResponse(rctx, kv)
	rctx.ReadRestfulRequest().SetAttribute(common.RespBodyContextKey, kv.Data)
	if err != nil {
		openlog.Error(err.Error())
	}
}

func prepareCache(topicName string, topic *pubsub.Topic, ctx context.Context) {
	rev, kvs, err := kvsvc.ListKV(ctx, &model.ListKVRequest{
		Domain:  topic.DomainID,
		Project: topic.Project,
		Labels:  topic.Labels,
		Match:   topic.MatchType,
	})
	if err != nil {
		openlog.Error("can not query kvs:" + err.Error())
	}
	cache.CachedKV().Write(topicName, &cache.DBResult{
		KVs: kvs,
		Rev: rev,
		Err: err,
	})
}
