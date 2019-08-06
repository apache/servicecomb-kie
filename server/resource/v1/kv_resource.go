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
	"encoding/json"
	"net/http"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/db"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
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
	decoder := json.NewDecoder(context.ReadRequest().Body)
	if err = decoder.Decode(kv); err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty, common.ContentTypeText)
	}
	kv.Key = key
	kv, err = kvsvc.CreateOrUpdate(context.Ctx, domain.(string), kv, project)
	if err != nil {
		ErrLog("put", kv, err)
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	InfoLog("put", kv)
	context.WriteHeader(http.StatusOK)
	context.WriteHeaderAndJSON(http.StatusOK, kv, goRestful.MIME_JSON)

}

//GetByKey search key by label and key
func (r *KVResource) GetByKey(context *restful.Context) {
	var err error
	key := context.ReadPathParameter("key")
	if key == "" {
		WriteErrResponse(context, http.StatusForbidden, "key must not be empty", common.ContentTypeText)
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
	kvs, err := kvsvc.FindKV(context.Ctx, domain.(string), project,
		kvsvc.WithKey(key), kvsvc.WithLabels(labels), kvsvc.WithDepth(d))
	if err != nil {
		if err == db.ErrKeyNotExists {
			WriteErrResponse(context, http.StatusNotFound, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = context.WriteHeaderAndJSON(http.StatusOK, kvs, goRestful.MIME_JSON)
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
		result, err := kvsvc.FindKV(context.Ctx, domain.(string), project, kvsvc.WithLabels(labels))
		if err != nil {
			if err == db.ErrKeyNotExists {
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

	err = context.WriteHeaderAndJSON(http.StatusOK, kvs, goRestful.MIME_JSON)
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
	err := kvsvc.Delete(kvID, labelID, domain.(string), project)
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
			Method:           http.MethodPut,
			Path:             "/v1/{project}/kie/kv/{key}",
			ResourceFuncName: "Put",
			FuncDesc:         "create or update key value",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKey, {
					DataType:  "string",
					Name:      "X-Realm",
					ParamType: goRestful.HeaderParameterKind,
					Desc:      "set kv to heterogeneous config server, not implement yet",
				},
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "true",
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
			Read:     KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/{project}/kie/kv/{key}",
			ResourceFuncName: "GetByKey",
			FuncDesc:         "get key values by key and labels",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKey, DocHeaderDepth,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []model.KVResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/{project}/kie/kv",
			ResourceFuncName: "SearchByLabels",
			FuncDesc:         "search key values by labels combination",
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
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		}, {
			Method:           http.MethodDelete,
			Path:             "/v1/{project}/kie/kv/",
			ResourceFuncName: "Delete",
			FuncDesc: "Delete key by kvID and labelID,If the labelID is nil, query the collection kv to get it." +
				"It means if only get kvID, it can also delete normally.But if you want better performance, you need to pass the labelID",
			Parameters: []*restful.Parameters{
				DocPathProject,
				kvIDParameters,
				labelIDParameters,
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
