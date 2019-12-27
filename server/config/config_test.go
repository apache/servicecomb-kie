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

package config_test

import (
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/go-chassis/go-archaius"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	err := archaius.Init()
	assert.NoError(t, err)
	b := []byte(`
db:
  uri: mongodb://admin:123@127.0.0.1:27017/kie
  type: mongodb
  poolSize: 10
  ssl: false
  sslCA:
  sslCert:

`)
	defer os.Remove("test.yaml")
	f1, err := os.Create("test.yaml")
	assert.NoError(t, err)
	_, err = io.WriteString(f1, string(b))
	assert.NoError(t, err)
	config.Configurations.ConfigFile = "test.yaml"
	err = config.Init()
	assert.NoError(t, err)
	assert.Equal(t, 10, config.GetDB().PoolSize)
	assert.Equal(t, "mongodb://admin:123@127.0.0.1:27017/kie", config.GetDB().URI)
}
