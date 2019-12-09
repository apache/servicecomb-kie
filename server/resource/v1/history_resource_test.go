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
	"encoding/json"
	"github.com/apache/servicecomb-kie/server/service"
	"io/ioutil"

	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-chassis/go-chassis/core/handler"

	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	_ "github.com/apache/servicecomb-kie/server/service/mongo"
	"github.com/go-chassis/go-chassis/server/restful/restfultest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("v1 history resource", func() {

	config.Configurations = &config.Config{
		DB: config.DB{},
	}

	Describe("get history revisions", func() {
		config.Configurations.DB.URI = "mongodb://kie:123@127.0.0.1:27017"
		err := service.DBInit()
		It("should not return err", func() {
			Expect(err).Should(BeNil())
		})
		Context("valid param", func() {
			kv := &model.KVDoc{
				Key:   "test",
				Value: "revisions",
				Labels: map[string]string{
					"test": "revisions",
				},
				Domain:  "default",
				Project: "test",
			}
			kv, err = service.KVService.CreateOrUpdate(context.Background(), kv)
			It("should not return err or nil", func() {
				Expect(err).Should(BeNil())
			})

			path := fmt.Sprintf("/v1/%s/kie/revision/%s", "test", kv.LabelID)
			r, _ := http.NewRequest("GET", path, nil)
			revision := &v1.HistoryResource{}
			chain, _ := handler.GetChain(common.Provider, "")
			c, err := restfultest.New(revision, chain)
			It("should not return err or nil", func() {
				Expect(err).Should(BeNil())
			})
			resp := httptest.NewRecorder()
			c.ServeHTTP(resp, r)

			body, err := ioutil.ReadAll(resp.Body)
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			data := []*model.LabelRevisionDoc{}
			err = json.Unmarshal(body, &data)
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			It("should return all revisions with the same label ID", func() {
				Expect(len(data) > 0).Should(Equal(true))
				Expect((data[0]).LabelID).Should(Equal(kv.LabelID))
			})
		})
	})
})
