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

package strategy

import (
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
)

func init() {
	service.Register("force", &Force{})
}

type Force struct {
}

func (f *Force) Execute(kv *model.KVDoc, rctx *restful.Context) (*model.KVDoc, *errsvc.Error) {
	project := rctx.ReadPathParameter(common.PathParameterProject)
	domain := v1.ReadDomain(rctx.Ctx)
	inputKV := kv
	kv, err := v1.PostOneKv(rctx, kv)
	if err == nil {
		return kv, nil
	}
	if err.Code == config.ErrRecordAlreadyExists {
		getKvsByOpts, getKvErr := v1.GetKvByOptions(rctx, inputKV)
		if getKvErr != nil {
			openlog.Info(fmt.Sprintf("get record [key: %s, labels: %s] failed", inputKV.Key, inputKV.Labels))
			return inputKV, getKvErr
		}
		kvReq := &model.UpdateKVRequest{
			ID:      getKvsByOpts[0].ID,
			Value:   inputKV.Value,
			Status:  inputKV.Status,
			Domain:  domain,
			Project: project,
		}
		kv, updateErr := service.KVService.Update(rctx.Ctx, kvReq)
		if updateErr != nil {
			openlog.Error(fmt.Sprintf("update record [key: %s, labels: %s] failed", inputKV.Key, inputKV.Labels))
			return inputKV, config.NewError(config.ErrInternal, updateErr.Error())
		}
		return kv, nil
	}
	return inputKV, err
}
