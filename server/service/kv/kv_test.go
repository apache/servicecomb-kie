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

package kvsvc_test

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/db"
	kvsvc "github.com/apache/servicecomb-kie/server/service/kv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kv mongodb service", func() {
	var err error
	Describe("connecting db", func() {
		config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017"}}
		err = db.Init()
		It("should not return err", func() {
			Expect(err).Should(BeNil())
		})
	})

	Describe("put kv timeout", func() {
		Context("with labels app and service", func() {
			kv, err := kvsvc.CreateOrUpdate(context.TODO(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "2s",
				Labels: map[string]string{
					"app":     "mall",
					"service": "cart",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has ID", func() {
				Expect(kv.ID.Hex()).ShouldNot(BeEmpty())
			})

		})
		Context("with labels app, service and version", func() {
			kv, err := kvsvc.CreateOrUpdate(context.TODO(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "2s",
				Labels: map[string]string{
					"app":     "mall",
					"service": "cart",
					"version": "1.0.0",
				},
			})
			oid, err := kvsvc.KVExist(context.TODO(), "default", "timeout", kvsvc.WithLabels(map[string]string{
				"app":     "mall",
				"service": "cart",
				"version": "1.0.0",
			}))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has ID", func() {
				Expect(kv.ID.Hex()).ShouldNot(BeEmpty())
			})
			It("should exist", func() {
				Expect(oid).ShouldNot(BeEmpty())
			})
		})
		Context("with labels app,and update value", func() {
			beforeKV, err := kvsvc.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:    "timeout",
				Value:  "1s",
				Domain: "default",
				Labels: map[string]string{
					"app": "mall",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			kvs1, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithKey("timeout"), kvsvc.WithLabels(map[string]string{
				"app": "mall",
			}), kvsvc.WithExactLabels())
			It("should be 1s", func() {
				Expect(kvs1[0].Data[0].Value).Should(Equal(beforeKV.Value))
			})
			afterKV, err := kvsvc.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:    "timeout",
				Value:  "3s",
				Domain: "default",
				Labels: map[string]string{
					"app": "mall",
				},
			})
			It("should has same id", func() {
				Expect(afterKV.ID.Hex()).Should(Equal(beforeKV.ID.Hex()))
			})
			oid, err := kvsvc.KVExist(context.Background(), "default", "timeout", kvsvc.WithLabels(map[string]string{
				"app": "mall",
			}))
			It("should exists", func() {
				Expect(oid.Hex()).Should(Equal(beforeKV.ID.Hex()))
			})
			kvs, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithKey("timeout"), kvsvc.WithLabels(map[string]string{
				"app": "mall",
			}), kvsvc.WithExactLabels())
			It("should be 3s", func() {
				Expect(kvs[0].Data[0].Value).Should(Equal(afterKV.Value))
			})
		})
	})

	Describe("greedy find by kv and labels", func() {
		Context("with labels app,depth is 1 ", func() {
			kvs, err := kvsvc.FindKV(context.Background(), "default",
				kvsvc.WithKey("timeout"), kvsvc.WithLabels(map[string]string{
					"app": "mall",
				}))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 1 records", func() {
				Expect(len(kvs)).Should(Equal(1))
			})

		})
		Context("with labels app,depth is 2 ", func() {
			kvs, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithKey("timeout"),
				kvsvc.WithLabels(map[string]string{
					"app": "mall",
				}),
				kvsvc.WithDepth(2))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 3 records", func() {
				Expect(len(kvs)).Should(Equal(3))
			})

		})
	})
	Describe("exact find by kv and labels", func() {
		Context("with labels app ", func() {
			kvs, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithKey("timeout"), kvsvc.WithLabels(map[string]string{
				"app": "mall",
			}), kvsvc.WithExactLabels())
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 1 records", func() {
				Expect(len(kvs)).Should(Equal(1))
			})

		})
	})
	Describe("exact find by labels", func() {
		Context("with labels app ", func() {
			kvs, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithLabels(map[string]string{
				"app": "mall",
			}), kvsvc.WithExactLabels())
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 1 records", func() {
				Expect(len(kvs)).Should(Equal(1))
			})

		})
	})
	Describe("greedy find by labels", func() {
		Context("with labels app ans service ", func() {
			kvs, err := kvsvc.FindKV(context.Background(), "default", kvsvc.WithLabels(map[string]string{
				"app":     "mall",
				"service": "cart",
			}))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 1 records", func() {
				Expect(len(kvs)).Should(Equal(1))
			})

		})
	})

	Describe("delete key", func() {
		Context("delete key by kvID", func() {
			kv1, err := kvsvc.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "20s",
				Labels: map[string]string{
					"env": "test",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			err = kvsvc.Delete(kv1.ID.Hex(), "", "default")
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

		})
		Context("delete key by kvID and labelID", func() {
			kv1, err := kvsvc.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "20s",
				Labels: map[string]string{
					"env": "test",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			err = kvsvc.Delete(kv1.ID.Hex(), kv1.LabelID, "default")
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

		})
		Context("test miss kvID, no panic", func() {
			err := kvsvc.Delete("", "", "default")
			It("should not return err", func() {
				Expect(err).Should(HaveOccurred())
			})
		})
		Context("Test encode error ", func() {
			err := kvsvc.Delete("12312312321", "", "default")
			It("should return err", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Test miss domain error ", func() {
			err := kvsvc.Delete("12312312321", "", "")
			It("should return err", func() {
				Expect(err).Should(Equal(db.ErrMissingDomain))
			})
		})
	})
})
