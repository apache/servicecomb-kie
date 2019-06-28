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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"bytes"
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"
	"net/http"
)

var _ = Describe("v1 kv resource", func() {
	//for UT
	config.Configurations = &config.Config{
		DB: config.DB{},
	}
	Describe("put kv", func() {
		config.Configurations.DB.URI = "mongodb://kie:123@127.0.0.1:27017"
		Context("valid param", func() {
			kv := &model.KVDoc{
				Value:  "1s",
				Labels: map[string]string{"service": "tester"},
			}
			j, _ := json.Marshal(kv)
			r, _ := http.NewRequest("PUT", "/v1/kv/timeout", bytes.NewBuffer(j))
			rctx := restfultest.NewRestfulContext(context.Background(), r)
			rctx.ReadRestfulRequest().SetAttribute("domain", "default")
			kvr := &v1.KVResource{}
			kvr.Put(rctx)
			It("should be 200 ", func() {
				Expect(rctx.Resp.StatusCode()).Should(Equal(http.StatusOK))
			})

		})
	})
})
