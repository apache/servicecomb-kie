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

//swagger doc header params
var (
	DocHeaderDepth = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderDepth,
		ParamType: goRestful.HeaderParameterKind,
		Desc:      "integer, default is 1, if you set match policy, you can set,depth to decide label number",
	}
)

//swagger doc query params
var (
	DocQueryCombination = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamQ,
		ParamType: goRestful.QueryParameterKind,
		Desc: "the combination format is {label_key}:{label_value}+{label_key}:{label_value} " +
			"for example: /v1/test/kie/kv?q=app:mall&q=app:mall+service:cart, " +
			"that will query key values from 2 kinds of labels",
	}
	DocQueryKVIDParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "kvID",
		ParamType: goRestful.QueryParameterKind,
		Required:  true,
	}
	DocQueryKeyParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "key",
		ParamType: goRestful.QueryParameterKind,
		Desc:      "only return history about a specific key",
	}
	DocQueryLabelParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "any",
		ParamType: goRestful.QueryParameterKind,
		Desc:      "label pairs",
	}
	DocQueryLabelIDParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "labelID",
		ParamType: goRestful.QueryParameterKind,
	}
)

//swagger doc path params
var (
	DocPathKey = &restful.Parameters{
		DataType:  "string",
		Name:      "key",
		ParamType: goRestful.PathParameterKind,
		Required:  true,
	}
	DocPathProject = &restful.Parameters{
		DataType:  "string",
		Name:      "project",
		ParamType: goRestful.PathParameterKind,
		Required:  true,
	}
	DocPathLabelID = &restful.Parameters{
		DataType:  "string",
		Name:      "label_id",
		ParamType: goRestful.PathParameterKind,
		Required:  true,
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
