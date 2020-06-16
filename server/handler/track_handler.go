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

package handler

import (
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/iputil"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service/mongo/track"
	"github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/core/invocation"
	"github.com/go-mesh/openlogging"
	"net/http"
	"strings"
)

//const of noop auth handler
const (
	TrackHandlerName = "track-handler"
)

//TrackHandler tracks polling data
type TrackHandler struct{}

//Handle set local attribute to http request
func (h *TrackHandler) Handle(chain *handler.Chain, inv *invocation.Invocation, cb invocation.ResponseCallBack) {
	req, ok := inv.Args.(*restful.Request)
	if !ok {
		chain.Next(inv, cb)
		return
	}
	if req.Request.Method != http.MethodGet {
		chain.Next(inv, cb)
		return
	}
	if !strings.Contains(req.Request.URL.Path, "kie/kv") {
		chain.Next(inv, cb)
		return
	}
	sessionID := req.HeaderParameter(v1.HeaderSessionID)
	if sessionID == "" {
		chain.Next(inv, cb)
		return
	}
	chain.Next(inv, func(ir *invocation.Response) error {
		resp, _ := ir.Result.(*restful.Response)
		revStr := req.QueryParameter(common.QueryParamRev)
		wait := req.QueryParameter(common.QueryParamWait)
		data := &model.PollingDetail{}
		data.URLPath = req.Request.Method + " " + req.Request.URL.Path
		data.SessionID = sessionID
		data.UserAgent = req.HeaderParameter(v1.HeaderUserAgent)
		data.Domain = inv.Metadata[v1.AttributeDomainKey].(string)
		data.IP = iputil.ClientIP(req.Request)
		data.ResponseBody = req.Attribute(common.RespBodyContextKey).([]*model.KVDoc)
		data.ResponseCode = ir.Status
		if resp != nil {
			data.ResponseHeader = resp.Header()
		}
		data.PollingData = map[string]interface{}{
			"revision": revStr,
			"wait":     wait,
			"labels":   req.QueryParameter("label"),
		}
		_, err := track.CreateOrUpdate(inv.Ctx, data)
		if err != nil {
			openlogging.Warn("record polling detail failed:" + err.Error())
			err := cb(ir)
			if err != nil {
				return err
			}
			return nil
		}
		err = cb(ir)
		if err != nil {
			return err
		}
		return nil
	})

}

func newTrackHandler() handler.Handler {
	return &TrackHandler{}
}

//Name is handler name
func (h *TrackHandler) Name() string {
	return TrackHandlerName
}
func init() {
	if err := handler.RegisterHandler(TrackHandlerName, newTrackHandler); err != nil {
		openlogging.Fatal("register handler failed: " + err.Error())
	}
}
