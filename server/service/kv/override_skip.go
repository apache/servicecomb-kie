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
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/openlog"
)

func init() {
	RegisterStrategy("skip", &Skip{})
}

type Skip struct {
}

func (s *Skip) Execute(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, *errsvc.Error) {
	inputKV := kv
	kv, err := Post(ctx, kv)
	if err == nil {
		return kv, nil
	}
	if err.Code == config.ErrRecordAlreadyExists {
		openlog.Info(fmt.Sprintf("skip overriding duplicate [key: %s, labels: %s]", inputKV.Key, inputKV.Labels))
		return inputKV, config.NewError(config.ErrSkipDuplicateKV, "skip overriding duplicate kvs")
	}
	return inputKV, err
}
