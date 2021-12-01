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

package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	goRestful "github.com/emicklei/go-restful"
	"github.com/go-chassis/go-chassis/v2/server/restful"
	"github.com/stretchr/testify/assert"
)

func TestGetLabels(t *testing.T) {
	r, err := http.NewRequest("GET",
		"/kv?q=app:mall+service:payment&q=app:mall+service:payment+version:1.0.0",
		nil)
	assert.NoError(t, err)
	c, err := v1.ReadLabelCombinations(goRestful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(c))

	r, err = http.NewRequest("GET",
		"/kv",
		nil)
	assert.NoError(t, err)
	c, err = v1.ReadLabelCombinations(goRestful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c))

}

func TestQueryFromCache(t *testing.T) {
	r, _ := http.NewRequest("GET",
		"/v1/kv_test/kie/kv?label=match:test&wait=10s&revision=100&match=exact",
		nil)
	ctx := &restful.Context{
		Ctx:  context.TODO(),
		Resp: goRestful.NewResponse(httptest.NewRecorder()),
		Req:  goRestful.NewRequest(r),
	}
	topic := "service:utService"
	v1.QueryFromCache(ctx, topic)
	response := ctx.ReadRestfulRequest().Attribute(common.RespBodyContextKey).([]*model.KVDoc)
	assert.Equal(t, 0, len(response))
}
