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
	"encoding/json"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
)

const (
	FindExact               = "exact"
	FindMany                = "greedy"
	MsgDomainMustNotBeEmpty = "domain must not be empty"
	MsgIllegalFindPolicy    = "value of header " + common.HeaderMatch + " can be greedy or exact"
	MsgIllegalLabels        = "label's value can not be empty, " +
		"label can not be duplicated, please check your query parameters"
	ErrIDMustNotEmpty = "must supply id if you want to remove key"
)

func ReadDomain(context *restful.Context) interface{} {
	return context.ReadRestfulRequest().Attribute("domain")
}
func ReadMatchPolicy(context *restful.Context) string {
	policy := context.ReadRestfulRequest().HeaderParameter(common.HeaderMatch)
	if policy == "" {
		//default is exact to reduce network traffic
		return common.MatchExact
	}
	return policy
}
func WriteErrResponse(context *restful.Context, status int, msg string) {
	context.WriteHeader(status)
	b, _ := json.MarshalIndent(&ErrorMsg{Msg: msg}, "", " ")
	context.Write(b)
}

func ErrLog(action string, kv *model.KV, err error) {
	openlogging.Error(fmt.Sprintf("[%s] [%v] err:%s", action, kv, err.Error()))
}

func InfoLog(action string, kv *model.KV) {
	openlogging.Info(
		fmt.Sprintf("[%s] [%s:%s] in [%s] success", action, kv.Key, kv.Value, kv.Domain))
}
