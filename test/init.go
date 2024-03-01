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
	"fmt"
	"math/rand"

	"github.com/apache/servicecomb-kie/server/db"
	_ "github.com/go-chassis/cari/db/bootstrap"

	_ "github.com/apache/servicecomb-kie/server/datasource/etcd"
	_ "github.com/apache/servicecomb-kie/server/datasource/local"
	_ "github.com/apache/servicecomb-kie/server/datasource/mongo"
	_ "github.com/apache/servicecomb-kie/server/plugin/qms"
	_ "github.com/apache/servicecomb-kie/server/pubsub/notifier"
	_ "github.com/apache/servicecomb-service-center/eventbase/bootstrap"
	_ "github.com/go-chassis/go-chassis/v2/security/cipher/plugins/plain"

	"github.com/apache/servicecomb-kie/pkg/validator"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/pubsub"
	edatasource "github.com/apache/servicecomb-service-center/eventbase/datasource"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/pkg/backends/quota"
	"github.com/go-chassis/go-chassis/v2/security/cipher"
)

var (
	uri           string
	kind          string
	localFilePath string
)

func init() {
	err := archaius.Init(archaius.WithENVSource(),
		archaius.WithMemorySource())
	if err != nil {
		panic(err)
	}
	kind = archaius.GetString("TEST_DB_KIND", "etcd")
	uri = archaius.GetString("TEST_DB_URI", "http://127.0.0.1:2379")
	localFilePath = archaius.GetString("TEST_KVS_ROOT_PATH", "")

	err = archaius.Init(archaius.WithMemorySource())
	if err != nil {
		panic(err)
	}
	err = archaius.Set("servicecomb.cipher.plugin", "default")
	if err != nil {
		panic(err)
	}
	err = cipher.Init()
	if err != nil {
		panic(err)
	}
	err = validator.Init()
	if err != nil {
		panic(err)
	}
	err = db.Init(config.DB{
		URI:           uri,
		Timeout:       "10s",
		Kind:          kind,
		LocalFilePath: localFilePath,
	})
	if err != nil {
		panic(err)
	}
	err = datasource.Init(kind)
	if err != nil {
		panic(err)
	}

	edatasourceKind := kind
	if kind == "etcd_with_localstorage" {
		edatasourceKind = "etcd"
	}
	if kind == "embedded_etcd_with_localstorage" {
		edatasourceKind = "embedded_etcd"
	}
	err = edatasource.Init(edatasourceKind)
	if err != nil {
		panic(err)
	}

	//for UT
	addr := randomListenAddress()
	config.Configurations = &config.Config{
		DB:             config.DB{},
		ListenPeerAddr: addr,
		AdvertiseAddr:  addr,
	}

	pubsub.Init()
	pubsub.Start()

	err = quota.Init(quota.Options{
		Plugin: "build-in",
	})
	if err != nil {
		panic(err)
	}
}

func randomListenAddress() string {
	min := 4000
	step := 1000
	port := min + rand.Intn(step)
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	return addr
}

func IsEmbeddedetcdMode() bool {
	return kind == "embedded_etcd" || kind == "embedded_etcd_with_localstorage"
}
