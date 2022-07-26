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
	"bytes"
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	_ "github.com/apache/servicecomb-kie/server/datasource/mongo"
	_ "github.com/apache/servicecomb-kie/server/plugin/qms"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	_ "github.com/apache/servicecomb-kie/test"
	"github.com/go-chassis/go-chassis/v2/server/restful/restfultest"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProjectResource_Get(t *testing.T) {
	t.Run("get project list", func(t *testing.T) {
		kv := &model.ProjectDoc{}
		j, _ := json.Marshal(kv)
		r, _ := http.NewRequest("GET", "/v1/project", bytes.NewBuffer(j))
		r.Header.Set("Content-Type", "application/json")
		pr := &v1.ProjectResource{}
		c, _ := restfultest.New(pr, nil)
		resp := httptest.NewRecorder()
		c.ServeHTTP(resp, r)
		body, err := ioutil.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.Code, string(body))
		result := &model.ProjectResponse{}
		err = json.Unmarshal(body, result)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result.Data))
	})
}
