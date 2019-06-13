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
	"github.com/apache/servicecomb-kie/pkg/common"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
)

//swagger doc elements
var (
	DocHeaderDepth = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderDepth,
		ParamType: goRestful.HeaderParameterKind,
		Desc:      "integer, default is 1, if you set match policy, you can set,depth to decide label number",
	}
	DocPathKey = &restful.Parameters{
		DataType:  "string",
		Name:      "key",
		ParamType: goRestful.PathParameterKind,
	}
	DocHeaderMath = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderMatch,
		ParamType: goRestful.HeaderParameterKind,
		Desc:      "greedy or exact",
	}
)

//KVBody is open api doc
type KVBody struct {
	Labels    map[string]string `json:"labels"`
	ValueType string            `json:"valueType"`
	Value     string            `json:"value"`
}

//ErrorMsg is open api doc
type ErrorMsg struct {
	Msg string `json:"msg"`
}
