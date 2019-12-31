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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/apache/servicecomb-kie/server/service"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"

	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"gopkg.in/yaml.v2"
)

//const of server
const (
	PatternExact            = "exact"
	MsgDomainMustNotBeEmpty = "domain must not be empty"
	MsgIllegalLabels        = "label value can not be empty, " +
		"label can not be duplicated, please check query parameters"
	MsgIllegalDepth     = "X-Depth must be number"
	MsgInvalidWait      = "wait param should be formed with number and time unit like 5s,100ms, and less than 5m"
	ErrKvIDMustNotEmpty = "must supply kv id if you want to remove key"
)

//ReadDomain get domain info from attribute
func ReadDomain(context *restful.Context) interface{} {
	return context.ReadRestfulRequest().Attribute("domain")
}

//ReadFindDepth get find depth
func ReadFindDepth(context *restful.Context) (int, error) {
	d := context.ReadRestfulRequest().HeaderParameter(common.HeaderDepth)
	if d == "" {
		return 1, nil
	}
	depth, err := strconv.Atoi(d)
	if err != nil {
		return 0, err
	}
	return depth, nil
}

//ReadLabelCombinations get query combination from url
//q=app:default+service:payment&q=app:default
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

//WriteErrResponse write error message to client
func WriteErrResponse(context *restful.Context, status int, msg, contentType string) {
	context.WriteHeader(status)
	b, _ := json.MarshalIndent(&ErrorMsg{Msg: msg}, "", " ")
	context.ReadRestfulResponse().AddHeader(goRestful.HEADER_ContentType, contentType)
	context.Write(b)

}

//ErrLog record error
func ErrLog(action string, kv *model.KVDoc, err error) {
	openlogging.Error(fmt.Sprintf("[%s] [%v] err:%s", action, kv, err.Error()))
}

//InfoLog record info
func InfoLog(action string, kv *model.KVDoc) {
	openlogging.Info(
		fmt.Sprintf("[%s] [%s] success", action, kv.Key))
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
func getLabels(labelStr string) (map[string]string, error) {
	labelsSlice := strings.Split(labelStr, ",")
	labels := make(map[string]string, len(labelsSlice))
	for _, v := range labelsSlice {
		v := strings.Split(v, ":")
		if len(v) != 2 {
			return nil, errors.New(MsgIllegalLabels)
		}
		labels[v[0]] = v[1]
	}
	return labels, nil
}
func getWaitDuration(rctx *restful.Context) string {
	wait := rctx.ReadQueryParameter(common.QueryParamWait)
	if wait == "" {
		wait = "0s"
	}
	return wait
}
func getMatchPattern(rctx *restful.Context) string {
	m := rctx.ReadQueryParameter(common.QueryParamMatch)

	if m != "" && m != PatternExact {
		return ""
	}
	return m
}
func wait(d time.Duration, rctx *restful.Context, topic *pubsub.Topic) bool {
	changed := true
	if d != 0 {
		o := &pubsub.Observer{
			UUID:      uuid.NewV4().String(),
			RemoteIP:  rctx.ReadRequest().RemoteAddr, //TODO x forward ip
			UserAgent: rctx.ReadHeader("User-Agent"),
			Event:     make(chan *pubsub.KVChangeEvent, 1),
		}
		pubsub.ObserveOnce(o, topic)
		select {
		case <-time.After(d):
			changed = false
		case <-o.Event:
		}
	}
	return changed
}
func checkPagination(limitStr, offsetStr string) (int64, int64, error) {
	var err error
	var limit, offset int64
	if limitStr != "" {
		limit, err = strconv.ParseInt(limitStr, 10, 64)
		if err != nil {
			return 0, 0, err
		}
		if limit < 1 || limit > 50 {
			return 0, 0, errors.New("invalid limit number")
		}
	}

	if offsetStr != "" {
		offset, err = strconv.ParseInt(offsetStr, 10, 64)
		if err != nil {
			return 0, 0, errors.New("invalid offset number")
		}
		if offset < 1 {
			return 0, 0, errors.New("invalid offset number")
		}
	}
	return limit, offset, err
}
func queryAndResponse(rctx *restful.Context,
	domain interface{}, project string, key string, labels map[string]string, limit, offset int) {
	m := getMatchPattern(rctx)
	opts := []service.FindOption{service.WithKey(key), service.WithLabels(labels)}
	if m == PatternExact {
		opts = append(opts, service.WithExactLabels())
	}
	kv, err := service.KVService.List(rctx.Ctx, domain.(string), project,
		limit, offset, opts...)
	if err != nil {
		if err == service.ErrKeyNotExists {
			WriteErrResponse(rctx, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(rctx, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = writeResponse(rctx, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}
}
