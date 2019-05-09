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

//v1 package hold http rest v1 API
package v1

import (
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/dao"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"net/http"
)

type KVResource struct {
}

func (r *KVResource) Put(context *restful.Context) {
	var err error
	key := context.ReadPathParameter("key")
	kv := new(model.KV)
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
	kv.Domain = domain.(string)
	s, err := dao.NewKVService()
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	kv, err = s.CreateOrUpdate(kv)
	if err != nil {
		ErrLog("put", kv, err)
		WriteErrResponse(context, http.StatusInternalServerError, err.Error())
		return
	}
	InfoLog("put", kv)
	context.WriteHeader(http.StatusOK)
	context.WriteHeaderAndJSON(http.StatusOK, kv, goRestful.MIME_JSON)

}
func (r *KVResource) Find(context *restful.Context) {
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
	policy := ReadFindPolicy(context)
	var kvs []*model.KV
	switch policy {
	case FindMany:
		kvs, err = s.Find(domain.(string), dao.WithKey(key), dao.WithLabels(labels))
	case FindExact:
		kvs, err = s.Find(domain.(string), dao.WithKey(key), dao.WithLabels(labels),
			dao.WithExactLabels())
	default:
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalFindPolicy)
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
	policy := ReadFindPolicy(context)
	var kvs []*model.KV
	switch policy {
	case FindMany:
		kvs, err = s.Find(domain.(string), dao.WithLabels(labels))
	case FindExact:
		kvs, err = s.Find(domain.(string), dao.WithLabels(labels),
			dao.WithExactLabels())
	default:
		WriteErrResponse(context, http.StatusBadRequest, MsgIllegalFindPolicy)
		return
	}
	err = context.WriteHeaderAndJSON(http.StatusOK, kvs, goRestful.MIME_JSON)
	if err != nil {
		openlogging.Error(err.Error())
	}

}
func (r *KVResource) Delete(context *restful.Context) {

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
				{
					DataType:  "string",
					Name:      "key",
					ParamType: goRestful.PathParameterKind,
				}, {
					DataType:  "string",
					Name:      "X-Domain-Name",
					ParamType: goRestful.HeaderParameterKind,
					Desc:      "set kv to other tenant",
				}, {
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
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/kv/{key}",
			ResourceFuncName: "Find",
			FuncDesc:         "get key values by key and labels",
			Parameters: []*restful.Parameters{
				{
					DataType:  "string",
					Name:      "key",
					ParamType: goRestful.PathParameterKind,
				}, {
					DataType:  "string",
					Name:      "X-Domain-Name",
					ParamType: goRestful.HeaderParameterKind,
				}, {
					DataType:  "string",
					Name:      "X-Find",
					ParamType: goRestful.HeaderParameterKind,
					Desc:      "greedy or exact",
				},
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*KVBody{},
				},
			},
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Read:     &KVBody{},
		}, {
			Method:           http.MethodGet,
			Path:             "/v1/kv",
			ResourceFuncName: "FindByLabels",
			FuncDesc:         "find key values only by labels",
			Parameters: []*restful.Parameters{
				{
					DataType:  "string",
					Name:      "X-Domain-Name",
					ParamType: goRestful.HeaderParameterKind,
				}, {
					DataType:  "string",
					Name:      "X-Find",
					ParamType: goRestful.HeaderParameterKind,
					Desc:      "greedy or exact",
				},
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   []*KVBody{},
				},
			},
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
		},
	}
}
