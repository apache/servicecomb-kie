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
	"net/http"

	"github.com/apache/servicecomb-kie/server/datasource"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/pubsub"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/foundation/validator"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
)

//KVResource has API about kv operations
type KVResource struct {
}

//Upload upload kvs
func (r *KVResource) Upload(rctx *restful.Context) {
	if rctx.ReadQueryParameter(common.QueryParamOverride) == "" {
		WriteErrResponse(rctx, config.ErrInvalidParams, "Query parameter 'override' is required")
		return
	}
	var err error
	inputUpload := new(KVUploadBody)
	if err = readRequest(rctx, &inputUpload); err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	result := kvsvc.Upload(rctx.Ctx, &model.UploadKVRequest{
		Domain:   ReadDomain(rctx.Ctx),
		Project:  rctx.ReadPathParameter(common.PathParameterProject),
		KVs:      inputUpload.Data,
		Override: rctx.ReadQueryParameter(common.QueryParamOverride),
	})
	err = writeResponse(rctx, result)
	if err != nil {
		openlog.Error(err.Error())
	}
}

//Post create a kv
func (r *KVResource) Post(rctx *restful.Context) {
	var err error
	kv := new(model.KVDoc)
	if err = readRequest(rctx, kv); err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	kv.Domain = ReadDomain(rctx.Ctx)
	kv.Project = rctx.ReadPathParameter(common.PathParameterProject)
	kv, postErr := kvsvc.Create(rctx.Ctx, kv)
	if postErr != nil {
		WriteErrResponse(rctx, postErr.Code, postErr.Message)
		return
	}
	kvsvc.Publish(kv)
	err = writeResponse(rctx, kv)
	if err != nil {
		openlog.Error(err.Error())
	}
}

//Put update a kv
func (r *KVResource) Put(rctx *restful.Context) {
	var err error
	kvID := rctx.ReadPathParameter(common.PathParamKVID)
	project := rctx.ReadPathParameter(common.PathParameterProject)
	kvReq := new(model.UpdateKVRequest)
	if err = readRequest(rctx, kvReq); err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	domain := ReadDomain(rctx.Ctx)
	kvReq.ID = kvID
	kvReq.Domain = domain
	kvReq.Project = project
	err = validator.Validate(kvReq)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	kv, err := kvsvc.Update(rctx.Ctx, kvReq)
	if err != nil {
		openlog.Error(fmt.Sprintf("put [%s] err:%s", kvID, err.Error()))
		WriteErrResponse(rctx, config.ErrInternal, "update kv failed")
		return
	}
	err = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kv.Key,
		Labels:   kv.Labels,
		Project:  project,
		DomainID: kv.Domain,
		Action:   pubsub.ActionPut,
	})
	if err != nil {
		openlog.Warn("lost kv change event when put:" + err.Error())
	}
	openlog.Info(
		fmt.Sprintf("put [%s] success", kvID))
	err = writeResponse(rctx, kv)
	if err != nil {
		openlog.Error(err.Error())
	}

}

//Get search key by kv id
func (r *KVResource) Get(rctx *restful.Context) {
	request := &model.GetKVRequest{
		Project: rctx.ReadPathParameter(common.PathParameterProject),
		Domain:  ReadDomain(rctx.Ctx),
		ID:      rctx.ReadPathParameter(common.PathParamKVID),
	}
	err := validator.Validate(request)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	kv, err := kvsvc.Get(rctx.Ctx, request)
	if err != nil {
		openlog.Error("kv_resource: " + err.Error())
		if err == datasource.ErrKeyNotExists {
			WriteErrResponse(rctx, config.ErrRecordNotExists, err.Error())
			return
		}
		WriteErrResponse(rctx, config.ErrInternal, "get kv failed")
		return
	}
	kv.Domain = ""
	kv.Project = ""
	err = writeResponse(rctx, kv)
	if err != nil {
		openlog.Error(err.Error())
	}
}

//List response kv list
func (r *KVResource) List(rctx *restful.Context) {
	var err error
	request := &model.ListKVRequest{
		Project: rctx.ReadPathParameter(common.PathParameterProject),
		Domain:  ReadDomain(rctx.Ctx),
		Key:     rctx.ReadQueryParameter(common.QueryParamKey),
		Status:  rctx.ReadQueryParameter(common.QueryParamStatus),
		Match:   getMatchPattern(rctx),
	}
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, common.MsgIllegalLabels)
		return
	}
	request.Labels = labels
	offsetStr := rctx.ReadQueryParameter(common.QueryParamOffset)
	limitStr := rctx.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	request.Offset = offset
	request.Limit = limit
	err = validator.Validate(request)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	returnData(rctx, request)
}

func returnData(rctx *restful.Context, request *model.ListKVRequest) {
	revStr := rctx.ReadQueryParameter(common.QueryParamRev)
	wait := rctx.ReadQueryParameter(common.QueryParamWait)
	if revStr == "" {
		if wait == "" {
			queryAndResponse(rctx, request)
			return
		}
		if !isLegalWaitRequest(rctx, request) {
			return
		}
		if watch(rctx, request, wait) {
			return
		}
		rctx.WriteHeader(http.StatusNotModified)
		return
	} else {
		revised, err := isRevised(rctx.Ctx, revStr, request.Domain)
		if err != nil {
			if err == ErrInvalidRev {
				WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
				return
			}
			WriteErrResponse(rctx, config.ErrInternal, err.Error())
			return
		}
		if revised {
			queryAndResponse(rctx, request)
			return
		} else if wait != "" {
			if !isLegalWaitRequest(rctx, request) {
				return
			}
			if watch(rctx, request, wait) {
				return
			}
			rctx.WriteHeader(http.StatusNotModified)
			return
		} else {
			rctx.WriteHeader(http.StatusNotModified)
		}
	}
}
func isLegalWaitRequest(rctx *restful.Context, request *model.ListKVRequest) bool {
	if request.Key != "" {
		WriteErrResponse(rctx, config.ErrInvalidParams, "can not accept key params, when using wait")
		return false
	}
	return true
}
func watch(rctx *restful.Context, request *model.ListKVRequest, wait string) bool {
	changed, err := eventHappened(wait, &pubsub.Topic{
		Labels:    request.Labels,
		Project:   request.Project,
		MatchType: request.Match,
		DomainID:  request.Domain,
	})
	if err != nil {
		WriteErrResponse(rctx, config.ErrObserveEvent, err.Error())
		return true
	}
	if changed {
		queryAndResponse(rctx, request)
		return true
	}
	return false
}

//Delete deletes one kv by id
func (r *KVResource) Delete(rctx *restful.Context) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx.Ctx)
	kvID := rctx.ReadPathParameter(common.PathParamKVID)
	err := validateDelete(domain, project, kvID)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	kv, err := kvsvc.FindOneAndDelete(rctx.Ctx, kvID, project, domain)
	if err != nil {
		openlog.Error("delete failed, ", openlog.WithTags(openlog.Tags{
			"kvID":  kvID,
			"error": err.Error(),
		}))
		if err == datasource.ErrKeyNotExists {
			WriteErrResponse(rctx, config.ErrRecordNotExists, err.Error())
			return
		}
		WriteErrResponse(rctx, config.ErrInternal, common.MsgDeleteKVFailed)
		return
	}
	err = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kv.Key,
		Labels:   kv.Labels,
		Project:  project,
		DomainID: domain,
		Action:   pubsub.ActionDelete,
	})
	if err != nil {
		openlog.Warn("lost kv change event:" + err.Error())
	}
	rctx.WriteHeader(http.StatusNoContent)
}

//DeleteList deletes multiple kvs by ids
func (r *KVResource) DeleteList(rctx *restful.Context) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx.Ctx)
	b := new(DeleteBody)
	if err := json.NewDecoder(rctx.ReadRequest().Body).Decode(b); err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	err := validateDeleteList(domain, project)
	if err != nil {
		WriteErrResponse(rctx, config.ErrInvalidParams, err.Error())
		return
	}
	kvs, err := kvsvc.FindManyAndDelete(rctx.Ctx, b.IDs, project, domain)
	if err != nil {
		if err == datasource.ErrKeyNotExists {
			rctx.WriteHeader(http.StatusNoContent)
			return
		}
		openlog.Error("delete list failed, ", openlog.WithTags(openlog.Tags{
			"kvIDs": b.IDs,
			"error": err.Error(),
		}))
		WriteErrResponse(rctx, config.ErrInternal, common.MsgDeleteKVFailed)
		return
	}
	for _, kv := range kvs {
		err = pubsub.Publish(&pubsub.KVChangeEvent{
			Key:      kv.Key,
			Labels:   kv.Labels,
			Project:  project,
			DomainID: domain,
			Action:   pubsub.ActionDelete,
		})
		if err != nil {
			openlog.Warn("lost kv change event:" + err.Error())
		}
	}
	rctx.WriteHeader(http.StatusNoContent)
}

//URLPatterns defined config operations
func (r *KVResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
			Method:       http.MethodPost,
			Path:         "/v1/{project}/kie/file",
			ResourceFunc: r.Upload,
			FuncDesc:     "upload key values",
			Parameters: []*restful.Parameters{
				DocPathProject,
				DocHeaderContentTypeJSONAndYaml,
			},
			Read: KVUploadBody{},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocRespOfUpload{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodPost,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.Post,
			FuncDesc:     "create a key value",
			Parameters: []*restful.Parameters{
				DocPathProject, DocHeaderContentTypeJSONAndYaml,
			},
			Read: KVCreateBody{},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocResponseSingleKey{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodPut,
			Path:         "/v1/{project}/kie/kv/{kv_id}",
			ResourceFunc: r.Put,
			FuncDesc:     "update a key value",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKeyID, DocHeaderContentTypeJSONAndYaml,
			},
			Read: KVUpdateBody{},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocResponseSingleKey{},
				},
			},
			Consumes: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv/{kv_id}",
			ResourceFunc: r.Get,
			FuncDesc:     "get key values by kv_id",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKeyID,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   model.DocResponseSingleKey{},
					Headers: map[string]goRestful.Header{
						common.HeaderRevision: DocHeaderRevision,
					},
				},
				{
					Code:    http.StatusNotFound,
					Message: "key value not found",
				},
			},
			Produces: []string{goRestful.MIME_JSON},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.List,
			FuncDesc:     "list key values by labels and key",
			Parameters: []*restful.Parameters{
				DocPathProject, DocQueryKeyParameters, DocQueryStatusParameters, DocQueryLabelParameters,
				DocQueryWait, DocQueryMatch, DocQueryRev, DocQueryLimitParameters, DocQueryOffsetParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:  http.StatusOK,
					Model: model.DocResponseGetKey{},
					Headers: map[string]goRestful.Header{
						common.HeaderRevision: DocHeaderRevision,
					},
				}, {
					Code:    http.StatusNotModified,
					Message: "empty body",
				},
			},
			Produces: []string{goRestful.MIME_JSON},
		}, {
			Method:       http.MethodDelete,
			Path:         "/v1/{project}/kie/kv/{kv_id}",
			ResourceFunc: r.Delete,
			FuncDesc:     "delete key by kv ID.",
			Parameters: []*restful.Parameters{
				DocPathProject,
				DocPathKeyID,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusNoContent,
					Message: "delete success",
				},
				{
					Code:    http.StatusNotFound,
					Message: "no key value found for deletion",
				},
				{
					Code:    http.StatusInternalServerError,
					Message: "server error",
				},
			},
		}, {
			Method:       http.MethodDelete,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.DeleteList,
			FuncDesc:     "delete keys.",
			Parameters: []*restful.Parameters{
				DocPathProject, DocHeaderContentTypeJSON,
			},
			Read: DeleteBody{},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusNoContent,
					Message: "delete success",
				},
				{
					Code:    http.StatusInternalServerError,
					Message: "server error",
				},
			},
			Consumes: []string{goRestful.MIME_JSON},
			Produces: []string{goRestful.MIME_JSON},
		},
	}
}
