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

//Package v1 hold http rest v1 API
package v1

import (
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/iputil"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/record"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

//KVResource has API about kv operations
type KVResource struct {
}

//Put create or update kv
func (r *KVResource) Put(context *restful.Context) {
	var err error
	key := context.ReadPathParameter("key")
	project := context.ReadPathParameter("project")
	kv := new(model.KVDoc)
	if err = readRequest(context, kv); err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	kv.Key = key
	kv.Domain = domain.(string)
	kv.Project = project
	_, err = checkStatus(kv.Status)
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	kv, err = service.KVService.CreateOrUpdate(context.Ctx, kv)
	if err != nil {
		openlogging.Error(fmt.Sprintf("put [%v] err:%s", kv, err.Error()))
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kv.Key,
		Labels:   kv.Labels,
		Project:  project,
		DomainID: kv.Domain,
		Action:   "put",
	})
	if err != nil {
		openlogging.Warn("lost kv change event:" + err.Error())
	}
	openlogging.Info(
		fmt.Sprintf("put [%s] success", kv.Key))
	err = writeResponse(context, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//GetByKey search key by label and key
func (r *KVResource) GetByKey(rctx *restful.Context) {
	var err error
	key := rctx.ReadPathParameter("key")
	if key == "" {
		WriteErrResponse(rctx, http.StatusBadRequest, "key must not be empty", common.ContentTypeText)
		return
	}
	project := rctx.ReadPathParameter("project")
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, common.MsgIllegalLabels, common.ContentTypeText)
		return
	}
	domain := ReadDomain(rctx)
	if domain == nil {
		WriteErrResponse(rctx, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	pageNumStr := rctx.ReadQueryParameter("pageNum")
	pageSizeStr := rctx.ReadQueryParameter("pageSize")
	pageNum, pageSize, err := checkPagination(pageNumStr, pageSizeStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	insID := rctx.ReadHeader("sessionID")
	statusStr := rctx.ReadQueryParameter("status")
	status, err := checkStatus(statusStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	returnData(rctx, domain, project, labels, pageNum, pageSize, status, insID)
}

//List response kv list
func (r *KVResource) List(rctx *restful.Context) {
	var err error
	project := rctx.ReadPathParameter("project")
	domain := ReadDomain(rctx)
	if domain == nil {
		WriteErrResponse(rctx, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	pageNumStr := rctx.ReadQueryParameter("pageNum")
	pageSizeStr := rctx.ReadQueryParameter("pageSize")
	pageNum, pageSize, err := checkPagination(pageNumStr, pageSizeStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	sessionID := rctx.ReadHeader("sessionID")
	statusStr := rctx.ReadQueryParameter("status")
	status, err := checkStatus(statusStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	returnData(rctx, domain, project, labels, pageNum, pageSize, status, sessionID)
}

func returnData(rctx *restful.Context, domain interface{}, project string, labels map[string]string, pageNum, pageSize int64, status, sessionID string) {
	revStr := rctx.ReadQueryParameter(common.QueryParamRev)
	wait := rctx.ReadQueryParameter(common.QueryParamWait)
	if sessionID != "" {
		defer RecordPollingDetail(rctx, revStr, wait, domain.(string), project, labels, pageNum, pageSize, sessionID)
	}
	if revStr == "" {
		if wait == "" {
			queryAndResponse(rctx, domain, project, "", labels, pageNum, pageSize, status)
			return
		}
		changed, err := eventHappened(rctx, wait, &pubsub.Topic{
			Labels:    labels,
			Project:   project,
			MatchType: getMatchPattern(rctx),
			DomainID:  domain.(string),
		})
		if err != nil {
			WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
			return
		}
		if changed {
			queryAndResponse(rctx, domain, project, "", labels, pageNum, pageSize, status)
			return
		}
		rctx.WriteHeader(http.StatusNotModified)
	} else {
		revised, err := isRevised(rctx.Ctx, revStr, domain.(string))
		if err != nil {
			if err == ErrInvalidRev {
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
				return
			}
			WriteErrResponse(rctx, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
			return
		}
		if revised {
			queryAndResponse(rctx, domain, project, "", labels, pageNum, pageSize, status)
			return
		} else if wait != "" {
			changed, err := eventHappened(rctx, wait, &pubsub.Topic{
				Labels:    labels,
				Project:   project,
				MatchType: getMatchPattern(rctx),
				DomainID:  domain.(string),
			})
			if err != nil {
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
				return
			}
			if changed {
				queryAndResponse(rctx, domain, project, "", labels, pageNum, pageSize, status)
				return
			}
			rctx.WriteHeader(http.StatusNotModified)
			return
		} else {
			rctx.WriteHeader(http.StatusNotModified)
		}
	}
}

//RecordPollingDetail to record data after get or list
func RecordPollingDetail(context *restful.Context, revStr, wait, domain, project string, labels map[string]string, limit, offset int64, sessionID string) {
	data := &model.PollingDetail{}
	data.ID = uuid.NewV4().String()
	data.SessionID = sessionID
	data.Domain = domain
	data.IP = iputil.ClientIP(context.Req.Request)
	dataMap := map[string]interface{}{
		"revStr":  revStr,
		"wait":    wait,
		"domain":  domain,
		"project": project,
		"labels":  labels,
		"limit":   limit,
		"offset":  offset,
	}
	data.PollingData = dataMap
	data.UserAgent = context.Req.HeaderParameter("User-Agent")
	data.URLPath = context.ReadRequest().Method + " " + context.ReadRequest().URL.Path
	data.ResponseHeader = context.Resp.Header()
	data.ResponseCode = context.Resp.StatusCode()
	data.ResponseBody = context.Ctx.Value(common.RespBodyContextKey)
	_, err := record.CreateOrUpdate(context.Ctx, data)
	if err != nil {
		openlogging.Warn("record polling detail failed" + err.Error())
		return
	}
}

//Search search key only by label
func (r *KVResource) Search(context *restful.Context) {
	var err error
	labelCombinations, err := ReadLabelCombinations(context.ReadRestfulRequest())
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	project := context.ReadPathParameter("project")
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	var kvs []*model.KVResponse
	pageNumStr := context.ReadQueryParameter("pageNum")
	pageSizeStr := context.ReadQueryParameter("pageSize")
	pageNum, pageSize, err := checkPagination(pageNumStr, pageSizeStr)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	if labelCombinations == nil {
		result, err := service.KVService.FindKV(context.Ctx, domain.(string), project,
			service.WithPageNum(pageNum),
			service.WithPageSize(pageSize))
		if err != nil {
			openlogging.Error("can not find by labels", openlogging.WithTags(openlogging.Tags{
				"err": err.Error(),
			}))
			WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
			return
		}
		kvs = append(kvs, result...)
	}
	for _, labels := range labelCombinations {
		openlogging.Debug("find by combination", openlogging.WithTags(openlogging.Tags{
			"q": labels,
		}))
		result, err := service.KVService.FindKV(context.Ctx, domain.(string), project,
			service.WithLabels(labels),
			service.WithPageNum(pageNum),
			service.WithPageSize(pageSize))
		if err != nil {
			if err == service.ErrKeyNotExists {
				continue
			} else {
				openlogging.Error("can not find by labels", openlogging.WithTags(openlogging.Tags{
					"err": err.Error(),
				}))
				WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
				return
			}
		}
		kvs = append(kvs, result...)

	}
	if len(kvs) == 0 {
		WriteErrResponse(context, http.StatusNotFound, "no kv found", common.ContentTypeText)
		return
	}

	err = writeResponse(context, kvs)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//Delete deletes key by ids
func (r *KVResource) Delete(context *restful.Context) {
	project := context.ReadPathParameter("project")
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, common.MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	kvID := context.ReadQueryParameter(common.QueryParamKeyID)
	if kvID == "" {
		WriteErrResponse(context, http.StatusBadRequest, common.ErrKvIDMustNotEmpty, common.ContentTypeText)
		return
	}
	err := service.KVService.Delete(context.Ctx, kvID, domain.(string), project)
	if err != nil {
		openlogging.Error("delete failed ,", openlogging.WithTags(openlogging.Tags{
			"kvID":  kvID,
			"error": err.Error(),
		}))
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	context.WriteHeader(http.StatusNoContent)
}

//URLPatterns defined config operations
func (r *KVResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodPut,
			Path:         "/v1/{project}/kie/kv/{key}",
			ResourceFunc: r.Put,
			FuncDesc:     "create or update key value",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKey,
			},
			Read: KVBody{},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: KVBody{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv/{key}",
			ResourceFunc: r.GetByKey,
			FuncDesc:     "get key values by key and labels",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKey, DocQueryLabelParameters, DocQueryWait, DocQueryMatch, DocQueryRev,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []model.KVResponse{},
				},
				{
					Code:    http.StatusNotModified,
					Message: "empty body",
				},
			},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/summary",
			ResourceFunc: r.Search,
			FuncDesc:     "search key values by labels combination, it returns multiple labels group",
			Parameters: []*restful.Parameters{
				DocPathProject, DocQueryCombination,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []model.KVResponse{},
				},
			},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.List,
			FuncDesc:     "list key values by labels and key",
			Parameters: []*restful.Parameters{
				DocPathProject, DocQueryLabelParameters, DocQueryWait, DocQueryMatch, DocQueryRev,
			},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.KVResponse{},
				}, {
					Code:    http.StatusNotModified,
					Message: "empty body",
				},
			},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodDelete,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.Delete,
			FuncDesc:     "delete key by kv ID.",
			Parameters: []*restful.Parameters{
				DocPathProject,
				DocQueryKeyIDParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusNoContent,
					Message: "Delete success",
				},
				{
					Code:    http.StatusBadRequest,
					Message: "Failed,check url",
				},
				{
					Code:    http.StatusInternalServerError,
					Message: "Server error",
				},
			},
		},
	}
}
