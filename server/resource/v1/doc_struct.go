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
	"github.com/apache/servicecomb-kie/pkg/model"

	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

// swagger doc request header params
var (
	DocHeaderDepth = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderDepth,
		ParamType: goRestful.HeaderParameterKind,
		Desc:      "integer, default is 1, if you set match policy, you can set,depth to decide label number",
	}
	DocHeaderContentTypeJSONAndYaml = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderContentType,
		ParamType: goRestful.HeaderParameterKind,
		Required:  true,
		Desc:      "used to indicate the media type of the resource, the value can be application/json or text/yaml",
	}
	DocHeaderContentTypeJSON = &restful.Parameters{
		DataType:  "string",
		Name:      common.HeaderContentType,
		ParamType: goRestful.HeaderParameterKind,
		Required:  true,
		Desc:      "used to indicate the media type of the resource, the value only can be application/json",
	}
)

// swagger doc response header params
var (
	DocHeaderRevision = goRestful.Header{
		Items: &goRestful.Items{
			Type: "integer",
		},
		Description: "cluster latest revision number, if key value is changed, it will increase.",
	}
)

// swagger doc query params
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
			"for example wait=5s, server will not response until 5 seconds, during that time window, " +
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
	DocQueryKeyParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamKey,
		ParamType: goRestful.QueryParameterKind,
		Desc: "key support prefix matching syntax, e.g. beginWith(servicecomb.) means to filter KV staring with 'servicecomb.'. " +
			"And support wildcard matching syntax, e.g. wildcard(*consumer*) means to filter KV include 'consumer'. " +
			"In addition to the above syntax means to filter KV full matching the input 'key'",
	}
	DocQueryLabelParameters = &restful.Parameters{
		DataType:  "string",
		Name:      "label",
		ParamType: goRestful.QueryParameterKind,
		Desc:      "label pairs,for example &label=service:order&label=version:1.0.0",
	}
	DocQueryStatusParameters = &restful.Parameters{
		DataType:  "string",
		Name:      common.QueryParamStatus,
		ParamType: goRestful.QueryParameterKind,
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

// swagger doc path params
var (
	DocPathProject = &restful.Parameters{
		DataType:  "string",
		Name:      common.PathParameterProject,
		ParamType: goRestful.PathParameterKind,
		Required:  true,
	}
	DocPathKeyID = &restful.Parameters{
		DataType:  "string",
		Name:      common.PathParamKVID,
		ParamType: goRestful.PathParameterKind,
		Required:  true,
	}
)

// KVCreateBody is open api doc
type KVCreateBody struct {
	Key       string            `json:"key"`
	Labels    map[string]string `json:"labels"`
	Status    string            `json:"status"`
	Value     string            `json:"value"`
	ValueType string            `json:"value_type"`
}

// KVUploadBody is open api doc
type KVUploadBody struct {
	MetaData MetaData       `json:"metadata"`
	Data     []*model.KVDoc `json:"data"`
}

// MetaData is extra info
type MetaData struct {
	Version     string      `json:"version"`
	Annotations Annotations `json:"annotations"`
}

type Annotations struct {
}

// KVUpdateBody is open api doc
type KVUpdateBody struct {
	Status string `json:"status"`
	Value  string `json:"value"`
}

// DeleteBody is the request body struct of delete multiple kvs interface
type DeleteBody struct {
	IDs []string `json:"ids"`
}

// ErrorMsg is open api doc
type ErrorMsg struct {
	Msg string `json:"error_msg"`
}
