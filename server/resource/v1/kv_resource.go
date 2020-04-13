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
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chassis/go-chassis/pkg/backends/quota"
	"net/http"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/validate"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
)

//KVResource has API about kv operations
type KVResource struct {
}

//Post create a kv
func (r *KVResource) Post(rctx *restful.Context) {
	var err error
	project := rctx.ReadPathParameter(common.PathParameterProject)
	kv := new(model.KVDoc)
	if err = readRequest(rctx, kv); err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	domain := ReadDomain(rctx)
	kv.Domain = domain.(string)
	kv.Project = project
	err = validate.Validate(kv)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	err = quota.PreCreate("", kv.Domain, "", 1)
	if err != nil {
		if err == quota.ErrReached {
			openlogging.Info("can not create kv, due to quota violation")
			WriteErrResponse(rctx, http.StatusUnprocessableEntity, err.Error())
			return
		}
		openlogging.Error(err.Error())
		WriteErrResponse(rctx, http.StatusInternalServerError, "quota check failed")
		return
	}
	kv, err = service.KVService.Create(rctx.Ctx, kv)
	if err != nil {
		openlogging.Error(fmt.Sprintf("post err:%s", err.Error()))
		if err == session.ErrKVAlreadyExists {
			WriteErrResponse(rctx, http.StatusConflict, err.Error())
			return
		}
		WriteErrResponse(rctx, http.StatusInternalServerError, "create kv failed")
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
		openlogging.Warn("lost kv change event when post:" + err.Error())
	}
	openlogging.Info(
		fmt.Sprintf("post [%s] success", kv.ID))
	err = writeResponse(rctx, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//Put update a kv
func (r *KVResource) Put(rctx *restful.Context) {
	var err error
	kvID := rctx.ReadPathParameter(common.PathParamKVID)
	project := rctx.ReadPathParameter(common.PathParameterProject)
	kv := new(model.KVDoc)
	if err = readRequest(rctx, kv); err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	domain := ReadDomain(rctx)
	kv.ID = kvID
	kv.Domain = domain.(string)
	kv.Project = project
	err = validatePut(kv)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	kv, err = service.KVService.Update(rctx.Ctx, kv)
	if err != nil {
		openlogging.Error(fmt.Sprintf("put [%s] err:%s", kvID, err.Error()))
		WriteErrResponse(rctx, http.StatusInternalServerError, "update kv failed")
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
		openlogging.Warn("lost kv change event when put:" + err.Error())
	}
	openlogging.Info(
		fmt.Sprintf("put [%s] success", kvID))
	err = writeResponse(rctx, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//Get search key by kv id
func (r *KVResource) Get(rctx *restful.Context) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx).(string)
	kvID := rctx.ReadPathParameter(common.PathParamKVID)
	err := validateGet(domain, project, kvID)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	kv, err := service.KVService.Get(rctx.Ctx, domain, project, kvID)
	if err != nil {
		openlogging.Error("kv_resource: " + err.Error())
		if err == service.ErrKeyNotExists {
			WriteErrResponse(rctx, http.StatusNotFound, err.Error())
			return
		}
		WriteErrResponse(rctx, http.StatusInternalServerError, "get kv failed")
		return
	}
	kv.Domain = ""
	kv.Project = ""
	err = writeResponse(rctx, kv)
	rctx.Ctx = context.WithValue(rctx.Ctx, common.RespBodyContextKey, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}
}

//List response kv list
func (r *KVResource) List(rctx *restful.Context) {
	var err error
	key := rctx.ReadQueryParameter(common.QueryParamKey)
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx).(string)
	err = validateList(domain, project)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, common.MsgIllegalLabels)
		return
	}

	offsetStr := rctx.ReadQueryParameter(common.QueryParamOffset)
	limitStr := rctx.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	sessionID := rctx.ReadHeader(HeaderSessionID)
	statusStr := rctx.ReadQueryParameter(common.QueryParamStatus)
	status, err := checkStatus(statusStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	returnData(rctx, &model.KVDoc{
		Domain:  domain,
		Project: project,
		Key:     key,
		Labels:  labels,
		Status:  status,
	}, offset, limit, sessionID)
}

func returnData(rctx *restful.Context, doc *model.KVDoc, offset, limit int64, sessionID string) {
	revStr := rctx.ReadQueryParameter(common.QueryParamRev)
	wait := rctx.ReadQueryParameter(common.QueryParamWait)
	if revStr == "" {
		if wait == "" {
			queryAndResponse(rctx, doc, offset, limit)
			return
		}
		changed, err := eventHappened(rctx, wait, &pubsub.Topic{
			Labels:    doc.Labels,
			Project:   doc.Project,
			MatchType: getMatchPattern(rctx),
			DomainID:  doc.Domain,
		})
		if err != nil {
			WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
			return
		}
		if changed {
			queryAndResponse(rctx, doc, offset, limit)
			return
		}
		rctx.WriteHeader(http.StatusNotModified)
	} else {
		revised, err := isRevised(rctx.Ctx, revStr, doc.Domain)
		if err != nil {
			if err == ErrInvalidRev {
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
				return
			}
			WriteErrResponse(rctx, http.StatusInternalServerError, err.Error())
			return
		}
		if revised {
			queryAndResponse(rctx, doc, offset, limit)
			return
		} else if wait != "" {
			changed, err := eventHappened(rctx, wait, &pubsub.Topic{
				Labels:    doc.Labels,
				Project:   doc.Project,
				MatchType: getMatchPattern(rctx),
				DomainID:  doc.Domain,
			})
			if err != nil {
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
				return
			}
			if changed {
				queryAndResponse(rctx, doc, offset, limit)
				return
			}
			rctx.WriteHeader(http.StatusNotModified)
			return
		} else {
			rctx.WriteHeader(http.StatusNotModified)
		}
	}
}

//Delete deletes one kv by id
func (r *KVResource) Delete(rctx *restful.Context) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx).(string)
	kvID := rctx.ReadPathParameter(common.PathParamKVID)
	err := validateDelete(domain, project, kvID)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	kv, err := service.KVService.FindOneAndDelete(rctx.Ctx, kvID, domain, project)
	if err != nil {
		openlogging.Error("delete failed, ", openlogging.WithTags(openlogging.Tags{
			"kvID":  kvID,
			"error": err.Error(),
		}))
		if err == service.ErrKeyNotExists {
			WriteErrResponse(rctx, http.StatusNotFound, err.Error())
			return
		}
		WriteErrResponse(rctx, http.StatusInternalServerError, common.MsgDeleteKVFailed)
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
		openlogging.Warn("lost kv change event:" + err.Error())
	}
	rctx.WriteHeader(http.StatusNoContent)
}

//DeleteList deletes multiple kvs by ids
func (r *KVResource) DeleteList(rctx *restful.Context) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx).(string)
	b := new(DeleteBody)
	if err := json.NewDecoder(rctx.ReadRequest().Body).Decode(b); err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, fmt.Sprintf(FmtReadRequestError, err))
		return
	}
	err := validateDeleteList(domain, project)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error())
		return
	}
	kvs, err := service.KVService.FindManyAndDelete(rctx.Ctx, b.IDs, domain, project)
	if err != nil {
		if err == service.ErrKeyNotExists {
			rctx.WriteHeader(http.StatusNoContent)
			return
		}
		openlogging.Error("delete list failed, ", openlogging.WithTags(openlogging.Tags{
			"kvIDs": b.IDs,
			"error": err.Error(),
		}))
		WriteErrResponse(rctx, http.StatusInternalServerError, common.MsgDeleteKVFailed)
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
			openlogging.Warn("lost kv change event:" + err.Error())
		}
	}
	rctx.WriteHeader(http.StatusNoContent)
}

//URLPatterns defined config operations
func (r *KVResource) URLPatterns() []restful.Route {
	return []restful.Route{
		{
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
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
		}, {
			Method:       http.MethodGet,
			Path:         "/v1/{project}/kie/kv",
			ResourceFunc: r.List,
			FuncDesc:     "list key values by labels and key",
			Parameters: []*restful.Parameters{
				DocPathProject, DocQueryKeyParameters, DocQueryLabelParameters, DocQueryWait, DocQueryMatch, DocQueryRev,
				DocQueryLimitParameters, DocQueryOffsetParameters,
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
			Produces: []string{goRestful.MIME_JSON, common.ContentTypeYaml},
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
