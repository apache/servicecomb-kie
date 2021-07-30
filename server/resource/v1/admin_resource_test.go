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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/apache/servicecomb-kie/pkg/model"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-chassis/v2/server/restful/restfultest"
	"github.com/stretchr/testify/assert"
)

func Test_HeathCheck(t *testing.T) {
	path := fmt.Sprintf("/v1/health")
	r, _ := http.NewRequest("GET", path, nil)
	revision := &v1.AdminResource{}
	c, err := restfultest.New(revision, nil)
	assert.NoError(t, err)
	resp := httptest.NewRecorder()
	c.ServeHTTP(resp, r)
	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	data := &model.DocHealthCheck{}
	err = json.Unmarshal(body, &data)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)
}
