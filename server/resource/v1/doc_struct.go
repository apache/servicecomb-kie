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

//swagger doc request header params
var (
	DocHeaderDepth = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderDepth,
		ParamType: goRestful.HeaderParameterKind,
		Desc:      "integer, default is 1, if you set match policy, you can set,depth to decide label number",
	}
	DocHeaderContentType = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderContentType,
		ParamType: goRestful.HeaderParameterKind,
		Required:  true,
		Desc:      "used to indicate the media type of the resource, the value can be application/json or text/yaml",
	}
)

//swagger doc response header params
var (
	DocHeaderRevision = goRestful.Header{
		Items: &goRestful.Items{
			Type: "integer",
		},
		Description: "cluster latest revision number, if key value is changed, it will increase.",
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
	DocQueryWait = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamWait,
		ParamType: goRestful.QueryParameterKind,
		Required:  false,
		Desc: "wait until any kv changed. " +
			"for example wait=5s, server will not response until 5 seconds during that time window, " +
			"if any kv changed, server will return 200 and kv list, otherwise return 304 and empty body",
	}
	DocQueryRev = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamRev,
		ParamType: goRestful.QueryParameterKind,
		Required:  false,
		Desc: "each time you query,server will return a number in header X-Kie-Revision. " +
			"you can record it in client side, use this number as param value. " +
			"if current revision is greater than it, server will return data",
	}
	DocQueryMatch = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamMatch,
		ParamType: goRestful.QueryParameterKind,
		Required:  false,
		Desc: "match works with label query param, it specifies label match pattern. " +
			"if it is empty, server will return kv which's labels partial match the label query param. " +
			"uf it is exact, server will return kv which's labels exact match the label query param",
	}
	DocQueryKeyIDParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamKeyID,
		ParamType: goRestful.QueryParameterKind,
		Required:  true,
	}
	DocQueryLabelParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "label",
		ParamType: goRestful.QueryParameterKind,
		Desc:      "label pairs,for example &label=service:order&label=version:1.0.0",
	}
	DocQueryLimitParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamLimit,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "pagination",
	}
	DocQueryOffsetParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamOffset,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "pagination",
	}
	//polling data
	DocQuerySessionIDParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamSessionID,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "sessionId is the Unique identification of the client",
	}
	DocQueryIPParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamIP,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "client ip",
	}
	DocQueryURLPathParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamURLPath,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "address of the call",
	}
	DocQueryUserAgentParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamUserAgent,
		ParamType: goRestful.QueryParameterKind,
		Desc:      "user agent of the call",
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
	DocPathKeyID = &restful.Parameters{
		DataType:  "string",
		Name:      "key_id",
		ParamType: goRestful.PathParameterKind,
		Required:  true,
	}
)

//KVBody is open api doc
type KVBody struct {
	Labels    map[string]string `json:"labels"`
	ValueType string            `json:"value_type"`
	Value     string            `json:"value"`
}

//ErrorMsg is open api doc
type ErrorMsg struct {
	Msg string `json:"msg"`
}
