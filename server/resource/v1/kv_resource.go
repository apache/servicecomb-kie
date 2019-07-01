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
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/dao"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"net/http"
)

//KVResource has API about kv operations
type KVResource struct {
}

//Put create or update kv
func (r *KVResource) Put(context *restful.Context) {
	var err error
	key := context.ReadPathParameter("key")
	kv := new(model.KVDoc)
	decoder := json.NewDecoder(context.ReadRequest().Body)
	if err = decoder.Decode(kv); err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty)
	}
	kv.Key = key
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	kv, err = s.CreateOrUpdate(context.Ctx, domain.(string), kv)
	if err != nil {
		ErrLog("put", kv, err)
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
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
		WriteErrResponse(context, http.StatusForbidden, "key must not be empty")
		return
	}
	values := context.ReadRequest().URL.Query()
	labels := make(map[string]string, len(values))
	for k, v := range values {
		if len(v) != 1 {
			WriteErrResponse(context, http.StatusBadRequest, MsgIllegalLabels)
			return
		}
		labels[k] = v[0]
	}
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty)
		return
	}
	d, err := ReadFindDepth(context)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalDepth)
		return
	}
	kvs, err := s.FindKV(context.Ctx, domain.(string), dao.WithKey(key), dao.WithLabels(labels), dao.WithDepth(d))
	if err != nil {
		if err == dao.ErrKeyNotExists {
			WriteErrResponse(context, http.StatusNotFound, err.Error())
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
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
		WriteErrResponse(context, http.StatusBadRequest, err.Error())
		return
	}
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty)
		return
	}
	var kvs []*model.KVResponse
	for _, labels := range labelCombinations {
		result, err := s.FindKV(context.Ctx, domain.(string), dao.WithLabels(labels))
		if err != nil {
			if err == dao.ErrKeyNotExists {
				continue
			}
			WriteErrResponse(context, http.StatusInternalServerError, err.Error())
			return
		}
		kvs = append(kvs, result...)

	}
	if len(kvs) == 0 {
		WriteErrResponse(context, http.StatusNotFound, "no kv found")
		return
	}

	err = context.WriteHeaderAndJSON(http.StatusOK, kvs, goRestful.MIME_JSON)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//Delete deletes key by ids
func (r *KVResource) Delete(context *restful.Context) {
	domain := ReadDomain(context)
	if domain == nil {
		WriteErrResponse(context, http.StatusInternalServerError, MsgDomainMustNotBeEmpty)
	}
	kvID := context.ReadPathParameter("kvID")
	if kvID == "" {
		WriteErrResponse(context, http.StatusBadRequest, ErrKvIDMustNotEmpty)
		return
	}
	labelID := context.ReadQueryParameter("labelID")
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	err = s.Delete(kvID, labelID, domain.(string))
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete kvID=%s,labelID=%s,error=%s", kvID, labelID, err.Error()))
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	context.WriteHeader(http.StatusNoContent)
}

//URLPatterns defined config operations
func (r *KVResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:           http.MethodPut,
			Path:             "/v1/kv/{key}",
			ResourceFuncName: "Put",
			FuncDesc:         "create or update key value",
			Parameters: []*restful.Parameters{
				DocPathKey, {
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
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/kv/{key}",
			ResourceFuncName: "GetByKey",
			FuncDesc:         "get key values by key and labels",
			Parameters: []*restful.Parameters{
				DocPathKey, DocHeaderDepth,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*model.KVResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/kv",
			ResourceFuncName: "SearchByLabels",
			FuncDesc:         "search key values by labels combination",
			Parameters: []*restful.Parameters{
				DocQueryCombination,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*model.KVResponse{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		}, {
			Method:           http.MethodDelete,
			Path:             "/v1/kv/{kvID}",
			ResourceFuncName: "Delete",
			FuncDesc: "Delete key by kvID and labelID,If the labelID is nil, query the collection kv to get it." +
				"It means if only get kvID, it can also delete normally.But if you want better performance, you need to pass the labelID",
			Parameters: []*restful.Parameters{
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
