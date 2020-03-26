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
	"net/http"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
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
func (r *KVResource) Post(context *restful.Context) {
	var err error
	project := context.ReadPathParameter(PathParameterProject)
	kvDoc := new(model.KVDoc)
	if err = readRequest(context, kvDoc); err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	if kvDoc.Key == "" {
		WriteErrResponse(context, http.StatusBadRequest, common.MsgKeyMustNotBeEmpty, common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	kvDoc.Domain = domain.(string)
	kvDoc.Project = project
	_, err = checkStatus(kvDoc.Status)
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	kv, err := service.KVService.Create(context.Ctx, kvDoc)
	if err != nil {
		openlogging.Error(fmt.Sprintf("post [%v] err:%s", kvDoc, err.Error()))
		if err == session.ErrKVAlreadyExists {
			WriteErrResponse(context, http.StatusConflict, err.Error(), common.ContentTypeText)
			return
		}
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kvDoc.Key,
		Labels:   kvDoc.Labels,
		Project:  project,
		DomainID: kvDoc.Domain,
		Action:   pubsub.ActionPut,
	})
	if err != nil {
		openlogging.Warn("lost kv change event when post:" + err.Error())
	}
	openlogging.Info(
		fmt.Sprintf("post [%s] success", kvDoc.Key))
	err = writeResponse(context, kv)
	if err != nil {
		openlogging.Error(err.Error())
	}

}

//Put update a kv
func (r *KVResource) Put(context *restful.Context) {
	var err error
	kvID := context.ReadPathParameter(common.QueryParamKVID)
	project := context.ReadPathParameter(PathParameterProject)
	kv := new(model.KVDoc)
	if err = readRequest(context, kv); err != nil {
		WriteErrResponse(context, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	domain := ReadDomain(context)
	kv.ID = kvID
	kv.Domain = domain.(string)
	kv.Project = project
	_, err = checkStatus(kv.Status)
	if err != nil {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	kv, err = service.KVService.Update(context.Ctx, kv)
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
		Action:   pubsub.ActionPut,
	})
	if err != nil {
		openlogging.Warn("lost kv change event when put:" + err.Error())
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
	key := rctx.ReadPathParameter(PathParameterKey)
	if key == "" {
		WriteErrResponse(rctx, http.StatusBadRequest, common.MsgKeyMustNotBeEmpty, common.ContentTypeText)
		return
	}
	project := rctx.ReadPathParameter(PathParameterProject)
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, common.MsgIllegalLabels, common.ContentTypeText)
		return
	}
	domain := ReadDomain(rctx)
	offsetStr := rctx.ReadQueryParameter(common.QueryParamOffset)
	limitStr := rctx.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	sessionID := rctx.ReadHeader(HeaderSessionID)
	statusStr := rctx.ReadQueryParameter(common.QueryParamStatus)
	status, err := checkStatus(statusStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	returnData(rctx, &model.KVDoc{
		Domain:  domain.(string),
		Project: project,
		Key:     key,
		Labels:  labels,
		Status:  status,
	}, offset, limit, sessionID)
}

//List response kv list
func (r *KVResource) List(rctx *restful.Context) {
	var err error
	project := rctx.ReadPathParameter(PathParameterProject)
	domain := ReadDomain(rctx)
	labels, err := getLabels(rctx)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}

	offsetStr := rctx.ReadQueryParameter(common.QueryParamOffset)
	limitStr := rctx.ReadQueryParameter(common.QueryParamLimit)
	offset, limit, err := checkPagination(offsetStr, limitStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	sessionID := rctx.ReadHeader(HeaderSessionID)
	statusStr := rctx.ReadQueryParameter(common.QueryParamStatus)
	status, err := checkStatus(statusStr)
	if err != nil {
		WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
		return
	}
	returnData(rctx, &model.KVDoc{
		Domain:  domain.(string),
		Project: project,
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
			WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
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
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
				return
			}
			WriteErrResponse(rctx, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
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
				WriteErrResponse(rctx, http.StatusBadRequest, err.Error(), common.ContentTypeText)
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

//Delete deletes key by ids
func (r *KVResource) Delete(context *restful.Context) {
	project := context.ReadPathParameter(PathParameterProject)
	domain := ReadDomain(context)
	kvID := context.ReadQueryParameter(common.QueryParamKVID)
	if kvID == "" {
		WriteErrResponse(context, http.StatusBadRequest, common.ErrKvIDMustNotEmpty, common.ContentTypeText)
		return
	}
	kv, err := service.KVService.Get(context.Ctx, domain.(string), project, kvID)
	if err != nil && err != service.ErrKeyNotExists {
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	} else if err == service.ErrKeyNotExists {
		context.WriteHeader(http.StatusNoContent)
		return
	}
	err = service.KVService.Delete(context.Ctx, kvID, domain.(string), project)
	if err != nil {
		openlogging.Error("delete failed ,", openlogging.WithTags(openlogging.Tags{
			"kvID":  kvID,
			"error": err.Error(),
		}))
		WriteErrResponse(context, http.StatusInternalServerError, err.Error(), common.ContentTypeText)
		return
	}
	err = pubsub.Publish(&pubsub.KVChangeEvent{
		Key:      kv.Key,
		Labels:   kv.Labels,
		Project:  project,
		DomainID: domain.(string),
		Action:   pubsub.ActionDelete,
	})
	if err != nil {
		openlogging.Warn("lost kv change event:" + err.Error())
	}
	context.WriteHeader(http.StatusNoContent)
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
				DocPathProject, DocHeaderContentType,
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
				DocPathProject, DocPathKeyID, DocHeaderContentType,
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
			Path:         "/v1/{project}/kie/kv/{key}",
			ResourceFunc: r.GetByKey,
			FuncDesc:     "get key values by key and labels",
			Parameters: []*restful.Parameters{
				DocPathProject, DocPathKey, DocQueryLabelParameters, DocQueryWait, DocQueryMatch, DocQueryRev,
				DocQueryLimitParameters, DocQueryOffsetParameters,
			},
			Returns: []*restful.Returns{
				{
					Code:    http.StatusOK,
					Message: "get key value success",
					Model:   model.DocResponseGetKey{},
					Headers: map[string]goRestful.Header{
						common.HeaderRevision: DocHeaderRevision,
					},
				},
				{
					Code:    http.StatusNotModified,
					Message: "empty body",
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
					Code:    http.StatusInternalServerError,
					Message: "Server error",
				},
			},
		},
	}
}
