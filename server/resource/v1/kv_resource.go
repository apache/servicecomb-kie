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
	"github.com/apache/servicecomb-kie/server/pubsub"
	"net/http"
	"time"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
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
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
	}
	kv.Key = key
	kv.Domain = domain.(string)
	kv.Project = project
	kv, err = service.KVService.CreateOrUpdate(context.Ctx, kv)
	if err != nil {
		ErrLog("put", kv, err)
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
	InfoLog("put", kv)
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
	labelStr := rctx.ReadQueryParameter("label")
	var labels map[string]string
	if labelStr != "" {
		labels, err = getLabels(labelStr)
		if err != nil {
			WriteErrResponse(rctx, http.StatusBadRequest, MsgIllegalLabels, common.ContentTypeText)
			return
		}
	}
	domain := ReadDomain(rctx)
	if domain == nil {
		WriteErrResponse(rctx, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	waitStr := getWaitDuration(rctx)
	d, err := time.ParseDuration(waitStr)
	if err != nil || d > 5*time.Minute {
		WriteErrResponse(rctx, http.StatusBadRequest, MsgInvalidWait, common.ContentTypeText)
		return
	}
	if d == 0 {
		queryAndResponse(rctx, domain, project, key, labels, 0, 0)
		return
	}
	changed := wait(d, rctx, &pubsub.Topic{
		Key:      key,
		Labels:   labels,
		Project:  project,
		DomainID: domain.(string),
	})
	if changed {
		queryAndResponse(rctx, domain, project, key, labels, 0, 0)
		return
	}
	rctx.WriteHeader(http.StatusNotModified)
}

//List TODO pagination
func (r *KVResource) List(rctx *restful.Context) {
	var err error
	project := rctx.ReadPathParameter("project")
	domain := ReadDomain(rctx)
	if domain == nil {
		WriteErrResponse(rctx, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	labelStr := rctx.ReadQueryParameter("label")
	var labels map[string]string
	if labelStr != "" {
		labels, err = getLabels(labelStr)
		if err != nil {
			WriteErrResponse(rctx, http.StatusBadRequest, MsgIllegalLabels, common.ContentTypeText)
			return
		}
	}
	limitStr := rctx.ReadPathParameter("limit")
	offsetStr := rctx.ReadPathParameter("offset")
	limit, offset, err := checkPagination(limitStr, offsetStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	waitStr := getWaitDuration(rctx)
	d, err := time.ParseDuration(waitStr)
	if err != nil || d > 5*time.Minute {
		WriteErrResponse(rctx, http.StatusBadRequest, MsgInvalidWait, common.ContentTypeText)
		return
	}
	if d == 0 {
		queryAndResponse(rctx, domain, project, "", labels, int(limit), int(offset))
		return
	}
	changed := wait(d, rctx, &pubsub.Topic{
		Labels:   labels,
		Project:  project,
		DomainID: domain.(string),
	})
	if changed {
		queryAndResponse(rctx, domain, project, "", labels, int(limit), int(offset))
		return
	}
	rctx.WriteHeader(http.StatusNotModified)

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
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	var kvs []*model.KVResponse
	if labelCombinations == nil {
		result, err := service.KVService.FindKV(context.Ctx, domain.(string), project)
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
		result, err := service.KVService.FindKV(context.Ctx, domain.(string), project, service.WithLabels(labels))
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
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
	}
	kvID := context.ReadQueryParameter(common.QueryParamKeyID)
	if kvID == "" {
		WriteErrResponse(context, http.StatusBadRequest, ErrKvIDMustNotEmpty, common.ContentTypeText)
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
				DocPathProject, DocPathKey, DocQueryLabelParameters, DocQueryMatch,
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
				DocPathProject, DocQueryLabelParameters, DocQueryWait, DocQueryMatch,
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
			FuncDesc:     "delete key by kvID and labelID. Want better performance, give labelID",
			Parameters: []*restful.Parameters{
				DocPathProject,
				DocQueryKeyIDParameters,
				DocQueryLabelIDParameters,
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
