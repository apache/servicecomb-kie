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
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

var overrideMap = make(map[string]strategyOverride)

func init() {
	Register()
}

type OverrideType struct {
}

type strategyOverride interface {
	Execute(input *model.KVDoc, rctx *restful.Context, isDuplicate bool) (*model.KVDoc, errsvc.Error)
}

func Register() {
	overrideMap["force"] = &Force{}
	overrideMap["abort"] = &Abort{}
	overrideMap["skip"] = &Skip{}
}

func (o *OverrideType) getType(override string) strategyOverride {
	return overrideMap[override]
}
