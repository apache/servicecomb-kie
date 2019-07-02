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
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/server/restful"
	"github.com/go-mesh/openlogging"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

//const of server
const (
	MsgDomainMustNotBeEmpty = "domain must not be empty"
	MsgIllegalLabels        = "label's value can not be empty, " +
		"label can not be duplicated, please check your query parameters"
	MsgIllegalDepth     = "X-Depth must be number"
	ErrKvIDMustNotEmpty = "must supply kv id if you want to remove key"
)

//ReadDomain get domain info from attribute
func ReadDomain(context *restful.Context) interface{} {
	return context.ReadRestfulRequest().Attribute("domain")
}

//ReadFindDepth get find depth
func ReadFindDepth(context *restful.Context) (int, error) {
	d := context.ReadRestfulRequest().HeaderParameter(common.HeaderDepth)
	if d == "" {
		return 1, nil
	}
	depth, err := strconv.Atoi(d)
	if err != nil {
		return 0, err
	}
	return depth, nil
}

//ReadLabelCombinations get query combination from url
//q=app:default+service:payment&q=app:default
func ReadLabelCombinations(req *goRestful.Request) ([]map[string]string, error) {
	queryCombinations := req.QueryParameters(common.QueryParamQ)
	labelCombinations := make([]map[string]string, 0)
	for _, queryStr := range queryCombinations {
		labelStr := strings.Split(queryStr, " ")
		labels := make(map[string]string, len(labelStr))
		for _, label := range labelStr {
			l := strings.Split(label, ":")
			if len(l) != 2 {
				return nil, errors.New("wrong query syntax:" + label)
			}
			labels[l[0]] = l[1]
		}
		if len(labels) == 0 {
			continue
		}
		labelCombinations = append(labelCombinations, labels)
	}
	if len(labelCombinations) == 0 {
		return []map[string]string{{"default": "default"}}, nil
	}
	return labelCombinations, nil
}

//WriteErrResponse write error message to client
func WriteErrResponse(context *restful.Context, status int, msg string) {
	context.WriteHeader(status)
	b, _ := json.MarshalIndent(&ErrorMsg{Msg: msg}, "", " ")
	context.Write(b)
}

//ErrLog record error
func ErrLog(action string, kv *model.KVDoc, err error) {
	openlogging.Error(fmt.Sprintf("[%s] [%v] err:%s", action, kv, err.Error()))
}

//InfoLog record info
func InfoLog(action string, kv *model.KVDoc) {
	openlogging.Info(
		fmt.Sprintf("[%s] [%s:%s] in [%s] success", action, kv.Key, kv.Value, kv.Domain))
}
