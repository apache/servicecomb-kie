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
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/ctxsvc"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/go-chassis/v2/pkg/runtime"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
	"net/http"
	"strconv"
	"time"
)

type AdminResource struct {
}

//URLPatterns defined config operations
func (r *AdminResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/health",
			ResourceFunc: r.HealthCheck,
			FuncDesc:     "health check return version and revision",
			Parameters:   []*restful.Parameters{},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocHealthCheck{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		},
	}
}

//HealthCheck provider version info and time info
func (r *AdminResource) HealthCheck(context *restful.Context) {
	domain := ctxsvc.ReadDomain(context.Ctx)
	resp := &model.DocHealthCheck{}
	latest, err := service.RevisionService.GetRevision(context.Ctx, domain)
	if err != nil {
		WriteErrResponse(context, config.ErrInternal, err.Error())
		return
	}
	resp.Revision = strconv.FormatInt(latest, 10)
	resp.Version = runtime.Version
	resp.Timestamp = time.Now().Unix()
	total, err := service.KVService.Total(context.Ctx, domain)
	if err != nil {
		WriteErrResponse(context, config.ErrInternal, err.Error())
		return
	}
	resp.Total = total
	err = writeResponse(context, resp)
	if err != nil {
		openlog.Error(err.Error())
	}
}
