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

package model_test

import (
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKV_UnmarshalJSON(t *testing.T) {
	kv := &model.KV{
		Value: "test",
		Labels: map[string]string{
			"test": "env",
		},
	}
	b, _ := json.Marshal(kv)
	t.Log(string(b))

	var kv2 model.KV
	err := json.Unmarshal([]byte(` 
        {"value": "1","labels":{"test":"env"}}
    `), &kv2)
	assert.NoError(t, err)
	assert.Equal(t, "env", kv2.Labels["test"])
	assert.Equal(t, "1", kv2.Value)

}
