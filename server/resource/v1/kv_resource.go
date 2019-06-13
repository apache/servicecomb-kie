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
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/dao"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"net/http"
	"strings"
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

//FindWithKey search key by label and key
func (r *KVResource) FindWithKey(context *restful.Context) {
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
	policy := ReadMatchPolicy(context)
	d, err := ReadFindDepth(context)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalDepth)
		return
	}
	var kvs []*model.KVResponse
	switch policy {
	case common.MatchGreedy:
		kvs, err = s.FindKV(context.Ctx, domain.(string), dao.WithKey(key), dao.WithLabels(labels), dao.WithDepth(d))
	case common.MatchExact:
		kvs, err = s.FindKV(context.Ctx, domain.(string), dao.WithKey(key), dao.WithLabels(labels),
			dao.WithExactLabels())
	default:
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalFindPolicy)
		return
	}
	if err == dao.ErrKeyNotExists {
		WriteErrResponse(context, http.StatusNotFound, err.Error())
		return
	}
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	err = context.WriteHeaderAndJSON(http.StatusOK, kvs, goRestful.MIME_JSON)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//FindByLabels search key only by label
func (r *KVResource) FindByLabels(context *restful.Context) {
	var err error
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
	policy := ReadMatchPolicy(context)
	d, err := ReadFindDepth(context)
	if err != nil {
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalDepth)
		return
	}
	var kvs []*model.KVResponse
	switch policy {
	case common.MatchGreedy:
		kvs, err = s.FindKV(context.Ctx, domain.(string), dao.WithLabels(labels), dao.WithDepth(d))
	case common.MatchExact:
		kvs, err = s.FindKV(context.Ctx, domain.(string), dao.WithLabels(labels),
			dao.WithExactLabels())
	default:
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalFindPolicy)
		return
	}
	if err == dao.ErrKeyNotExists {
		WriteErrResponse(context, http.StatusNotFound, err.Error())
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
	ids := context.ReadPathParameter("ids")
	if ids == "" {
		WriteErrResponse(context, http.StatusBadRequest, ErrIDMustNotEmpty)
		return
	}
	idArray := strings.Split(ids, ",")
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	err = s.Delete(idArray, domain.(string))
	if err != nil {
		openlogging.Error(fmt.Sprintf("delete ids=%s,err=%s", ids, err.Error()))
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
					Desc:      "set kv to heterogeneous config server",
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
			ResourceFuncName: "FindWithKey",
			FuncDesc:         "get key values by key and labels",
			Parameters: []*restful.Parameters{
				DocPathKey, DocHeaderMath, DocHeaderDepth,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*KVBody{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/kv",
			ResourceFuncName: "FindByLabels",
			FuncDesc:         "find key values only by labels",
			Parameters: []*restful.Parameters{
				DocHeaderMath, DocHeaderDepth,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*KVBody{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		}, {
			Method:           http.MethodDelete,
			Path:             "/v1/kv/{ids}",
			ResourceFuncName: "Delete",
			FuncDesc:         "delete key by id,separated by ','",
			Parameters: []*restful.Parameters{{
				DataType:  "string",
				Name:      "ids",
				ParamType: goRestful.PathParameterKind,
				Desc: "The id strings to be removed are separated by ',',If the actual number of deletions " +
					"and the number of parameters are not equal, no error will be returned and only warn log will be printed.",
			},
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
