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

package label_test

import (
	"context"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service/mongo/label"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestService_CreateOrUpdate(t *testing.T) {
	var err error
	config.Configurations = &config.Config{DB: config.DB{URI: "mongodb://kie:123@127.0.0.1:27017/kie"}}
	err = session.Init()
	assert.NoError(t, err)
	d, err := label.CreateLabel(context.TODO(), "default",
		map[string]string{
			"cluster":   "a",
			"role":      "b",
			"component": "c",
		}, "default")
	assert.NoError(t, err)
	assert.NotEmpty(t, d.ID)
	assert.Equal(t, "cluster=a::component=c::role=b", d.Format)
	t.Log(d)

	id, err := label.Exist(context.TODO(), "default", "default", map[string]string{
		"cluster":   "a",
		"role":      "b",
		"component": "c",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
}
