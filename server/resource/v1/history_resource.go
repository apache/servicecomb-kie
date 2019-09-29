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
	"github.com/apache/servicecomb-kie/server/service"
	"net/http"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
)

//HistoryResource TODO
type HistoryResource struct {
}

//GetRevisionsByLabelID search key only by label
func (r *HistoryResource) GetRevisionsByLabelID(context *restful.Context) {
	var err error
	labelID := context.ReadPathParameter("label_id")
	if labelID == "" {
		openlogging.Debug("label id is null")
		WriteErrResponse(context, http.StatusForbidden, "label_id must not be empty", common.ContentTypeText)
		return
	}
	revisions, err := service.HistoryService.GetHistoryByLabelID(context.Ctx, labelID)
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
	err = context.WriteHeaderAndJSON(http.StatusOK, revisions, goRestful.MIME_JSON)
	if err != nil {
		openlogging.Error(err.Error())
	}
}

//URLPatterns defined config operations
func (r *HistoryResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:           http.MethodGet,
			Path:             "/v1/{project}/kie/revision/{label_id}",
			ResourceFuncName: "GetRevisionsByLabelID",
			FuncDesc:         "get all revisions by label id",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathLabelID,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "true",
					Model:   []model.LabelHistoryResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		},
	}
}
