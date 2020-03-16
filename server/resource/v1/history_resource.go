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
	"github.com/go-chassis/go-chassis/pkg/runtime"
	"net/http"
	"strconv"
	"time"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/server/service"

	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
)

//HistoryResource TODO
type HistoryResource struct {
}

//GetRevisions search key only by label
func (r *HistoryResource) GetRevisions(context *restful.Context) {
	var err error
	keyID := context.ReadPathParameter("key_id")
	offsetStr := context.ReadQueryParameter(common.QueryParamOffset)
	limitStr := context.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	if keyID == "" {
		openlogging.Error("key id is nil")
		WriteErrResponse(context, http.StatusForbidden, "key_id must not be empty", common.ContentTypeText)
		return
	}
	key := context.ReadQueryParameter("key")
	revisions, _, err := service.HistoryService.GetHistory(context.Ctx, keyID,
		service.WithKey(key),
		service.WithOffset(offset),
		service.WithLimit(limit))
	if err != nil {
		if err == service.ErrRevisionNotExist {
			WriteErrResponse(context, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	if len(revisions) == 0 {
		WriteErrResponse(context, http.StatusNotFound, "no revisions found", common.ContentTypeText)
		return
	}
	err = writeResponse(context, revisions)
	if err != nil {
		openlogging.Error(err.Error())
	}
}

//HealthCheck provider version info and time info
func (r *HistoryResource) HealthCheck(context *restful.Context) {
	domain := ReadDomain(context)
	resp := &model.DocHealthCheck{}
	latest, err := service.RevisionService.GetRevision(context.Ctx, domain.(string))
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	resp.Revision = strconv.FormatInt(latest, 10)
	resp.Version = runtime.Version
	resp.Timestamp = time.Now().Unix()
	err = writeResponse(context, resp)
	if err != nil {
		openlogging.Error(err.Error())
	}
}

//URLPatterns defined config operations
func (r *HistoryResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/revision/{key_id}",
			ResourceFunc: r.GetRevisions,
			FuncDesc:     "get all revisions by key id",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKeyID,
			},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: []model.DocResponseSingleKey{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		},
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
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		},
	}
}
