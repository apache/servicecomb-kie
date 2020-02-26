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
	"github.com/apache/servicecomb-kie/server/service/mongo/record"
	"net/http"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
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
	pageNumStr := context.ReadQueryParameter("pageNum")
	pageSizeStr := context.ReadQueryParameter("pageSize")
	pageNum, pageSize, err := checkPagination(pageNumStr, pageSizeStr)
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
		service.WithPageSize(pageSize),
		service.WithPageNum(pageNum))
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

//GetPollingData get the record of the get or list history
func (r *HistoryResource) GetPollingData(context *restful.Context) {
	query := &model.PollingDetail{}
	sessionID := context.ReadQueryParameter("sessionId")
	if sessionID != "" {
		query.SessionID = sessionID
	}
	ip := context.ReadQueryParameter("ip")
	if ip != "" {
		query.IP = ip
	}
	urlPath := context.ReadQueryParameter("urlPath")
	if urlPath != "" {
		query.URLPath = urlPath
	}
	userAgent := context.ReadQueryParameter("userAgent")
	if userAgent != "" {
		query.UserAgent = userAgent
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	query.Domain = domain.(string)
	records, err := record.Get(context.Ctx, query)
	if err != nil {
		if err == service.ErrRecordNotExists {
			WriteErrResponse(context, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	if len(records) == 0 {
		WriteErrResponse(context, http.StatusNotFound, "no polling data found", common.ContentTypeText)
		return
	}
	err = writeResponse(context, records)
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
					Code:    http.StatusOK,
					Message: "true",
					Model:   []model.KVDoc{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/polling_data",
			ResourceFunc: r.GetPollingData,
			FuncDesc:     "get all history record of get and list",
			Parameters: []*restful.Parameters{
				DocPathProject,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "true",
					Model:   []model.PollingDetail{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		},
	}
}
