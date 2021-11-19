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

package kv

import (
	"context"
	"fmt"

	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/openlog"

	"github.com/apache/servicecomb-kie/pkg/model"
)

func init() {
	RegisterStrategy("force", &Force{})
}

type Force struct {
}

func (f *Force) Execute(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, *errsvc.Error) {
	input := kv
	kv, err := Create(ctx, kv)
	if err == nil {
		return kv, nil
	}
	if err.Code != config.ErrRecordAlreadyExists {
		return input, err
	}

	request := &model.ListKVRequest{
		Project: input.Project,
		Domain:  input.Domain,
		Key:     input.Key,
		Labels:  input.Labels,
	}
	_, getKvsByOpts, getKvErr := ListKV(ctx, request)
	if getKvErr != nil {
		openlog.Info(fmt.Sprintf("get record [key: %s, labels: %s] failed", input.Key, input.Labels))
		return input, getKvErr
	}
	kvReq := &model.UpdateKVRequest{
		ID:      getKvsByOpts.Data[0].ID,
		Value:   input.Value,
		Status:  input.Status,
		Project: input.Project,
		Domain:  input.Domain,
	}
	kv, updateErr := Update(ctx, kvReq)
	if updateErr != nil {
		openlog.Error(fmt.Sprintf("update record [key: %s, labels: %s] failed", input.Key, input.Labels))
		return input, config.NewError(config.ErrInternal, updateErr.Error())
	}
	return kv, nil
}
