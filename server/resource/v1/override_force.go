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

package v1

import (
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
)

type Force struct {
}

func (f *Force) Execute(kv *model.KVDoc, rctx *restful.Context, _ bool) (*model.KVDoc, errsvc.Error) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := ReadDomain(rctx.Ctx)
	inputKV := kv
	kv, err := postOneKv(rctx, kv)
	if err.Code == config.ErrRecordAlreadyExists {
		getKvsByOpts, err0 := getKvByOptions(rctx, inputKV)
		if err0.Message != "" {
			openlog.Info(fmt.Sprintf("get record [key: %s, labels: %s] failed", inputKV.Key, inputKV.Labels))
			return inputKV, err0
		}
		kvReq := &model.UpdateKVRequest{
			ID:      getKvsByOpts[0].ID,
			Value:   getKvsByOpts[0].Value,
			Status:  getKvsByOpts[0].Status,
			Domain:  domain,
			Project: project,
		}
		kv, err := service.KVService.Update(rctx.Ctx, kvReq)
		if err != nil {
			openlog.Error(fmt.Sprintf("update record [key: %s, labels: %s] failed", inputKV.Key, inputKV.Labels))
			return inputKV, errsvc.Error{
				Code:    config.ErrInternal,
				Message: err.Error(),
			}
		}
		return kv, errsvc.Error{
			Code:    0,
			Message: "",
		}
	}
	if err.Message != "" {
		return inputKV, err
	}
	return kv, errsvc.Error{
		Code:    0,
		Message: "",
	}
}
