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
	"net/http"

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
	if project == "" {
		WriteErrResponse(context, http.StatusForbidden, "project must not be empty", common.ContentTypeText)
		return
	}
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
	InfoLog("put", kv)
	err = writeResponse(context, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//GetByKey search key by label and key
func (r *KVResource) GetByKey(context *restful.Context) {
	var err error
	key := context.ReadPathParameter("key")
	if key == "" {
		WriteErrResponse(context, http.StatusBadRequest, "key must not be empty", common.ContentTypeText)
		return
	}
	project := context.ReadPathParameter("project")
	if project == "" {
		WriteErrResponse(context, http.StatusForbidden, "project must not be empty", common.ContentTypeText)
		return
	}
	values := context.ReadRequest().URL.Query()
	labels := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) != 1 {
			WriteErrResponse(context, http.StatusBadRequest, MsgIllegalLabels, common.ContentTypeText)
			return
		}
		labels[k] = v[0]
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	d, err := ReadFindDepth(context)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalDepth, common.ContentTypeText)
		return
	}
	kvs, err := service.KVService.FindKV(context.Ctx, domain.(string), project,
		service.WithKey(key), service.WithLabels(labels), service.WithDepth(d))
	if err != nil {
		if err == service.ErrKeyNotExists {
			WriteErrResponse(context, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = writeResponse(context, kvs)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//SearchByLabels search key only by label
func (r *KVResource) SearchByLabels(context *restful.Context) {
	var err error
	labelCombinations, err := ReadLabelCombinations(context.ReadRestfulRequest())
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	project := context.ReadPathParameter("project")
	if project == "" {
		WriteErrResponse(context, http.StatusForbidden, "project must not be empty", common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
		return
	}
	var kvs []*model.KVResponse
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
	if project == "" {
		WriteErrResponse(context, http.StatusForbidden, "project must not be empty", common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
	}
	kvID := context.ReadQueryParameter("kvID")
	if kvID == "" {
		WriteErrResponse(context, http.StatusBadRequest, ErrKvIDMustNotEmpty, common.ContentTypeText)
		return
	}
	labelID := context.ReadQueryParameter("labelID")
	err := service.KVService.Delete(kvID, labelID, domain.(string), project)
	if err != nil {
		openlogging.Error("delete failed ,", openlogging.WithTags(openlogging.Tags{
			"kvID":    kvID,
			"labelID": labelID,
			"error":   err.Error(),
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
					Code:    http.StatusOK,
					Message: "true",
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
				DocPathProject, DocPathKey,
				DocHeaderDepth,
				DocQueryLabelParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []model.KVResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.SearchByLabels,
			FuncDesc:     "search key values by labels combination",
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
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodDelete,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.Delete,
			FuncDesc:     "delete key by kvID and labelID. if you want better performance, you need to give labelID",
			Parameters: []*restful.Parameters{
				DocPathProject,
				DocQueryKVIDParameters,
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
