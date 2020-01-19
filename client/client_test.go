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
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	. "github.com/apache/servicecomb-kie/client"
	"github.com/apache/servicecomb-kie/pkg/model"
)

func TestClient_Put(t *testing.T) {
	os.Setenv("HTTP_DEBUG", "1")
	c, _ := New(Config{
		Endpoint: "http://127.0.0.1:30110",
	})
	kv := model.KVRequest{
		Key:    "app.properties",
		Labels: map[string]string{"service": "client"},
		Value:  "timeout: 1s",
	}
	_, err := c.Put(context.TODO(), kv, WithProject("test"))
	assert.NoError(t, err)

	kvs, _ := c.Get(context.TODO(),
		WithKey("app.properties"),
		WithGetProject("test"),
		WithLabels(map[string]string{"service": "client"}))
	assert.Equal(t, 1, len(kvs.Data))

	_, err = c.Get(context.TODO(),
		WithGetProject("test"),
		WithLabels(map[string]string{"service": "client"}))
	assert.Equal(t, ErrNoChanges, err)

	_, err = c.Get(context.TODO(),
		WithLabels(map[string]string{"service": "client"}))
	assert.Error(t, err)
}
func TestClient_Delete(t *testing.T) {
	c, err := New(Config{
		Endpoint: "http://127.0.0.1:30110",
	})

	kvBody := model.KVRequest{}
	kvBody.Key = "time"
	kvBody.Value = "100s"
	kvBody.ValueType = "text"
	kvBody.Labels = make(map[string]string)
	kvBody.Labels["env"] = "test"
	kv, err := c.Put(context.TODO(), kvBody, WithProject("test"))
	assert.NoError(t, err)
	kvs, err := c.Get(context.TODO(),
		WithKey("time"),
		WithGetProject("test"),
		WithLabels(map[string]string{"env": "test"}))
	assert.NoError(t, err)
	assert.NotNil(t, kvs)
	err = c.Delete(context.TODO(), kv.ID, "", WithProject("test"))
	assert.NoError(t, err)
}
