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

package auth

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestFilterKVs(t *testing.T) {
	permResourceLabel := []map[string]string{
		{"environment": "production", "appId": "default"},
		{"appId": "default"},
		{"environment": "production", "serviceName": "service-center"},
		{"serviceName": "service-center", "version": "1.0.0"},
		{"serviceName": "service-center"},
		{"environment": "testing"},
	}

	var kvs []*model.KVDoc

	kv1 := new(model.KVDoc)
	kv1.Key = "k1"
	kv1.Value = "v1"
	kv1.Labels = map[string]string{"environment": "production", "appId": "default"}
	kvs = append(kvs, kv1)

	kv2 := new(model.KVDoc)
	kv2.Key = "k2"
	kv2.Value = "v2"
	kv2.Labels = map[string]string{"environment": "production", "appId": "default"}
	kvs = append(kvs, kv2)

	kv3 := new(model.KVDoc)
	kv3.Key = "k3"
	kv3.Value = "v3"
	kv3.Labels = map[string]string{"environment": "xxx", "appId": "xxx", "serviceName": "xxx", "version": "xxx"}
	kvs = append(kvs, kv3)

	r := FilterKVs(kvs, permResourceLabel)
	assert.Equal(t, 2, len(r))
}
