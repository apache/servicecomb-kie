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
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"net/http"

	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
)

//HistoryResource TODO
type HistoryResource struct {
}

//GetRevisions search key only by label
func (r *HistoryResource) GetRevisions(context *restful.Context) {
	var err error
	kvID := context.ReadPathParameter(common.PathParamKVID)
	offsetStr := context.ReadQueryParameter(common.QueryParamOffset)
	limitStr := context.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error())
		return
	}
	if kvID == "" {
		openlog.Error("kv id is nil")
		WriteErrResponse(context, http.StatusForbidden, "kv_id must not be empty")
		return
	}
	revisions, err := service.HistoryService.GetHistory(context.Ctx, kvID,
		service.WithOffset(offset),
		service.WithLimit(limit))
	if err != nil {
		if err == service.ErrRevisionNotExist {
			WriteErrResponse(context, http.StatusNotFound, err.Error())
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	err = writeResponse(context, revisions)
	if err != nil {
		openlog.Error(err.Error())
	}
}

//GetPollingData get the record of the get or list history
func (r *HistoryResource) GetPollingData(context *restful.Context) {
	query := &model.PollingDetail{}
	sessionID := context.ReadQueryParameter(common.QueryParamSessionID)
	if sessionID != "" {
		query.SessionID = sessionID
	}
	ip := context.ReadQueryParameter(common.QueryParamIP)
	if ip != "" {
		query.IP = ip
	}
	urlPath := context.ReadQueryParameter(common.QueryParamURLPath)
	if urlPath != "" {
		query.URLPath = urlPath
	}
	userAgent := context.ReadQueryParameter(common.QueryParamUserAgent)
	if userAgent != "" {
		query.UserAgent = userAgent
	}
	domain := ReadDomain(context.Ctx)
	if domain == "" {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty)
		return
	}
	query.Domain = domain
	records, err := service.TrackService.GetPollingDetail(context.Ctx, query)
	if err != nil {
		if err == service.ErrRecordNotExists {
			WriteErrResponse(context, http.StatusNotFound, err.Error())
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	resp := &model.PollingDataResponse{}
	resp.Data = records
	resp.Total = len(records)
	err = writeResponse(context, resp)
	if err != nil {
		openlog.Error(err.Error())
	}
}

//URLPatterns defined config operations
func (r *HistoryResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/revision/{kv_id}",
			ResourceFunc: r.GetRevisions,
			FuncDesc:     "get all revisions by key id",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKeyID,
			},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocResponseGetKey{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		},

		{
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/track",
			ResourceFunc: r.GetPollingData,
			FuncDesc:     "get polling tracks of clients of kie server",
			Parameters: []*restful.Parameters{
				DocPathProject, DocQuerySessionIDParameters, DocQueryIPParameters, DocQueryURLPathParameters, DocQueryUserAgentParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "true",
					Model:   []model.PollingDataResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		},
	}
}
