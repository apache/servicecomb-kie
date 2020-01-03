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
		Context("only by key with default project", func() {
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

		Context("by project", func() {
			_, err := c1.Get(context.TODO(), "app.properties", WithGetProject("test"))
			It("should be 404 error", func() {
				Expect(err).Should(Equal(ErrKeyNotExist))
			})
		})
	})

	Describe("put /v1/{project}/kie/kv/{key}", func() {
		Context("create or update key value", func() {
			c1, _ = New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})
			kv := model.KVRequest{
				Key:    "app.properties",
				Labels: map[string]string{"service": "tester"},
				Value:  "1s",
			}
			res, err := c1.Put(context.TODO(), kv, WithProject("test"))
			It("should not be error", func() {
				Expect(err).Should(BeNil())
			})
			It("should return the exact content passed", func() {
				Expect(res.Key).Should(Equal(kv.Key))
				Expect(res.Labels).Should(Equal(kv.Labels))
				Expect(res.Value).Should(Equal(kv.Value))
				Expect(res.Project).Should(Equal(""))
				Expect(res.Domain).Should(Equal(""))
			})
			kvs, _ := c1.Get(context.TODO(), "app.properties",
				WithGetProject("test"), WithLabels(map[string]string{"service": "tester"}))
			It("should exactly 1 kv", func() {
				Expect(kvs).Should(Not(BeNil()))
			})
		})
	})

	Describe("DELETE /v1/{project}/kie/kv/", func() {
		Context("by kvID", func() {
			client2, err := New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})

			kvBody := model.KVRequest{}
			kvBody.Key = "time"
			kvBody.Value = "100s"
			kvBody.ValueType = "string"
			kvBody.Labels = make(map[string]string)
			kvBody.Labels["env"] = "test"
			kv, err := client2.Put(context.TODO(), kvBody, WithProject("test"))
			It("should not be error", func() {
				Ω(err).ShouldNot(HaveOccurred())
				Expect(kv.Key).To(Equal(kvBody.Key))
				Expect(kv.Value).To(Equal(kvBody.Value))
				Expect(kv.Labels).To(Equal(kvBody.Labels))
				Expect(kv.Project).To(Equal(""))
				Expect(kv.Domain).To(Equal(""))
			})
			kvs, err := client2.Get(context.TODO(), "time",
				WithGetProject("test"), WithLabels(map[string]string{"env": "test"}))
			It("should return exactly 1 kv", func() {
				Expect(kvs).Should(Not(BeNil()))
				Expect(err).Should(BeNil())
			})
			client3, err := New(Config{
				Endpoint: "http://127.0.0.1:30110",
			})
			It("should be 204", func() {
				err := client3.Delete(context.TODO(), kv.ID, "", WithProject("test"))
				Ω(err).ShouldNot(HaveOccurred())
			})
		})
	})

})
