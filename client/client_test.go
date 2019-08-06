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

package client_test

import (
	"context"
	"os"

	. "github.com/apache/servicecomb-kie/client"
	"github.com/apache/servicecomb-kie/pkg/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client", func() {
	var c1 *Client
	os.Setenv("HTTP_DEBUG", "1")
	Describe("new client ", func() {
		Context("with http protocol", func() {
			var err error
			c1, err = New(Config{
				Endpoint: "http://127.0.0.1:8081",
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should return client", func() {
				Expect(c1).ShouldNot(BeNil())
			})

		})
	})

	Describe("get ", func() {
		Context("only by key", func() {
			_, err := c1.Get(context.TODO(), "app.properties")
			It("should be 404 error", func() {
				Expect(err).Should(Equal(ErrKeyNotExist))
			})

		})
		Context("by key and labels", func() {
			_, err := c1.Get(context.TODO(), "app.properties", WithLabels(map[string]string{
				"app": "mall",
			}))
			It("should be 404 error", func() {
				Expect(err).Should(Equal(ErrKeyNotExist))
			})

		})
	})

	Describe("put /v1/test/kie/kv/{key}", func() {
		Context("create or update key value", func() {
			c1, _ = New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})
			kv := model.KVDoc{
				Key:     "app.properties",
				Labels:  map[string]string{"service": "tester"},
				Value:   "1s",
				Project: "test",
			}
			res, err := c1.Put(context.TODO(), kv)
			It("should not be error", func() {
				Expect(err).Should(BeNil())
			})
			It("should return the exact content passed", func() {
				Expect(res.Key).Should(Equal(kv.Key))
				Expect(res.Labels).Should(Equal(kv.Labels))
				Expect(res.Value).Should(Equal(kv.Value))
				Expect(res.Project).Should(Equal(kv.Project))
			})
		})
	})

	Describe("DELETE /v1/test/kie/kv/", func() {
		Context("by kvID", func() {
			client2, err := New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})

			kvBody := model.KVDoc{}
			kvBody.Key = "time"
			kvBody.Value = "100s"
			kvBody.ValueType = "string"
			kvBody.Project = "test"
			kvBody.Labels = make(map[string]string)
			kvBody.Labels["evn"] = "test"
			kv, err := client2.Put(context.TODO(), kvBody)
			It("should not be error", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Expect(kv.Key).To(Equal(kvBody.Key))
			})
			client3, err := New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})
			It("should be 204", func() {
				err := client3.Delete(context.TODO(), kv.ID.Hex(), "")
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})

})
