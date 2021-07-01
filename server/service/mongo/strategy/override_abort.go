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
	"github.com/apache/servicecomb-kie/pkg/model"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/go-chassis/openlog"
)

func init() {
	service.Register("abort", &Abort{})
}

type Abort struct {
}

func (a *Abort) Execute(kv *model.KVDoc, rctx *restful.Context, isDuplicate bool) (*model.KVDoc, *errsvc.Error) {
	inputKV := kv
	if isDuplicate {
		openlog.Info(fmt.Sprintf("stop overriding kvs after reaching the duplicate [key: %s, labels: %s]", kv.Key, kv.Labels))
		return inputKV, config.NewError(config.ErrStopUpload, "stop overriding kvs after reaching the duplicate kv")
	}
	kv, err := v1.PostOneKv(rctx, kv)
	if err == nil {
		return kv, nil
	}
	if err.Code == config.ErrRecordAlreadyExists {
		openlog.Info(fmt.Sprintf("stop overriding duplicate [key: %s, labels: %s]", inputKV.Key, inputKV.Labels))
		return inputKV, config.NewError(config.ErrRecordAlreadyExists, "stop overriding duplicate kv")
	}
	return inputKV, err
}
