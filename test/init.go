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

package test

import (
	"github.com/apache/servicecomb-kie/pkg/validator"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/cipher"

	_ "github.com/apache/servicecomb-kie/server/datasource/etcd"
	_ "github.com/apache/servicecomb-kie/server/datasource/mongo"
	_ "github.com/apache/servicecomb-kie/server/pubsub/notifier"
	_ "github.com/go-chassis/go-chassis/v2/security/cipher/plugins/plain"
	_ "github.com/little-cui/etcdadpt/embedded"
	_ "github.com/little-cui/etcdadpt/remote"
)

var (
	uri  string
	kind string
)

func init() {
	err := archaius.Init(archaius.WithENVSource(),
		archaius.WithMemorySource())
	if err != nil {
		panic(err)
	}
	kind := archaius.GetString("TEST_DB_KIND", "etcd")
	uri := archaius.GetString("TEST_DB_URI", "http://127.0.0.1:2379")
	archaius.Init(archaius.WithMemorySource())
	archaius.Set("servicecomb.cipher.plugin", "default")
	cipher.Init()
	validator.Init()
	config.Configurations.DB.Kind = kind
	err = datasource.Init(config.DB{
		URI:     uri,
		Timeout: "10s",
		Kind:    kind,
	})
	if err != nil {
		panic(err)
	}
}
