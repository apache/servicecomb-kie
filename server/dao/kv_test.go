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

package dao_test

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/dao"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Kv mongodb service", func() {
	var s *dao.MongodbService
	var err error
	Describe("connecting db", func() {
		s, err = dao.NewMongoService(dao.Options{
			URI: "mongodb://kie:123@127.0.0.1:27017",
		})
		It("should not return err", func() {
			Expect(err).Should(BeNil())
		})
	})

	Describe("put kv timeout", func() {
		Context("with labels app and service", func() {
			kv, err := s.CreateOrUpdate(context.TODO(), "default", &model.KVDoc{
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
			kv, err := s.CreateOrUpdate(context.TODO(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "2s",
				Labels: map[string]string{
					"app":     "mall",
					"service": "cart",
					"version": "1.0.0",
				},
			})
			oid, err := s.KVExist(context.TODO(), "default", "timeout", dao.WithLabels(map[string]string{
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
			beforeKV, err := s.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
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
			kvs1, err := s.FindKV(context.Background(), "default", dao.WithKey("timeout"), dao.WithLabels(map[string]string{
				"app": "mall",
			}), dao.WithExactLabels())
			It("should be 1s", func() {
				Expect(kvs1[0].Data[0].Value).Should(Equal(beforeKV.Value))
			})
			afterKV, err := s.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
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
			oid, err := s.KVExist(context.Background(), "default", "timeout", dao.WithLabels(map[string]string{
				"app": "mall",
			}))
			It("should exists", func() {
				Expect(oid.Hex()).Should(Equal(beforeKV.ID.Hex()))
			})
			kvs, err := s.FindKV(context.Background(), "default", dao.WithKey("timeout"), dao.WithLabels(map[string]string{
				"app": "mall",
			}), dao.WithExactLabels())
			It("should be 3s", func() {
				Expect(kvs[0].Data[0].Value).Should(Equal(afterKV.Value))
			})
		})
	})

	Describe("greedy find by kv and labels", func() {
		Context("with labels app,depth is 1 ", func() {
			kvs, err := s.FindKV(context.Background(), "default", dao.WithKey("timeout"), dao.WithLabels(map[string]string{
				"app": "mall",
			}))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 2 records", func() {
				Expect(len(kvs)).Should(Equal(2))
			})

		})
		Context("with labels app,depth is 2 ", func() {
			kvs, err := s.FindKV(context.Background(), "default", dao.WithKey("timeout"),
				dao.WithLabels(map[string]string{
					"app": "mall",
				}),
				dao.WithDepth(2))
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
			kvs, err := s.FindKV(context.Background(), "default", dao.WithKey("timeout"), dao.WithLabels(map[string]string{
				"app": "mall",
			}), dao.WithExactLabels())
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
			kvs, err := s.FindKV(context.Background(), "default", dao.WithLabels(map[string]string{
				"app": "mall",
			}), dao.WithExactLabels())
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
			kvs, err := s.FindKV(context.Background(), "default", dao.WithLabels(map[string]string{
				"app":     "mall",
				"service": "cart",
			}))
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
			It("should has 2 records", func() {
				Expect(len(kvs)).Should(Equal(2))
			})

		})
	})

	Describe("delete key", func() {
		Context("delete key by id,seperated by ',' ", func() {
			kv1, err := s.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:   "timeout",
				Value: "20s",
				Labels: map[string]string{
					"env": "test",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			kv2, err := s.CreateOrUpdate(context.Background(), "default", &model.KVDoc{
				Key:    "times",
				Value:  "3",
				Domain: "default",
				Labels: map[string]string{
					"env": "test",
				},
			})
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

			ids := []string{kv1.ID.Hex(), kv2.ID.Hex()}
			err = s.Delete(ids, "default")
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})

		})
		Context("test miss ids, no panic", func() {
			err := s.Delete(nil, "default")
			It("should not return err", func() {
				Expect(err).Should(BeNil())
			})
		})
		Context("Test encode error ", func() {
			err := s.Delete([]string{"12312312321"}, "default")
			It("should return err", func() {
				Expect(err).To(HaveOccurred())
			})
		})
		Context("Test miss domain error ", func() {
			err := s.Delete([]string{"5ce3602381fc6e33708b9621"}, "")
			It("should return err", func() {
				Expect(err).Should(Equal(dao.ErrMissingDomain))
			})
		})
	})
})
