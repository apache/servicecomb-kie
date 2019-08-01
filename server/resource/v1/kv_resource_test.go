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
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/db"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
)

var _ = Describe("v1 kv resource", func() {
	//for UT
	config.Configurations = &config.Config{
		DB: config.DB{},
	}
	config.Configurations.DB.URI = "mongodb://kie:123@127.0.0.1:27017"
	err := db.Init()
	It("should not return err", func() {
		Expect(err).Should(BeNil())
	})
	Describe("put kv", func() {
		Context("valid param", func() {
			kv := &model.KVDoc{
				Value:  "1s",
				Labels: map[string]string{"service": "tester"},
			}
			j, _ := json.Marshal(kv)
			r, _ := http.NewRequest("PUT", "/v1/kv/timeout", bytes.NewBuffer(j))
			handler.RegisterHandler("fake", newFakeHandler)
			chain, _ := handler.CreateChain(common.Provider, "testchain1", "fake")
			r.Header.Set("Content-Type", "application/json")
			kvr := &v1.KVResource{}
			c, err := restfultest.New(kvr, chain)
			It("should not return error", func() {
				Expect(err).Should(BeNil())
			})
			resp := httptest.NewRecorder()
			c.ServeHTTP(resp, r)

			body, err := ioutil.ReadAll(resp.Body)
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			data := &model.KVDoc{}
			err = json.Unmarshal(body, data)
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			It("should return created or updated kv", func() {
				Expect(data.Value).Should(Equal(kv.Value))
				Expect(data.Labels).Should(Equal(kv.Labels))
			})
		})
	})
})
