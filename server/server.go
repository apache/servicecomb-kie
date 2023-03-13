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

package server

import (
	"github.com/apache/servicecomb-kie/pkg/validator"
	"github.com/apache/servicecomb-kie/server/cache"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/db"
	"github.com/apache/servicecomb-kie/server/pubsub"
	"github.com/apache/servicecomb-kie/server/rbac"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-chassis/v2"
	"github.com/go-chassis/go-chassis/v2/core/common"
	"github.com/go-chassis/openlog"
)

func Run() {
	chassis.RegisterSchema(common.ProtocolRest, &v1.KVResource{})
	chassis.RegisterSchema(common.ProtocolRest, &v1.HistoryResource{})
	chassis.RegisterSchema(common.ProtocolRest, &v1.AdminResource{})
	if err := chassis.Init(); err != nil {
		openlog.Fatal(err.Error())
	}
	if err := config.Init(); err != nil {
		openlog.Fatal(err.Error())
	}
	if err := db.Init(config.GetDB()); err != nil {
		openlog.Fatal(err.Error())
	}
	if err := datasource.Init(config.GetDB().Kind); err != nil {
		openlog.Fatal(err.Error())
	}
	if err := validator.Init(); err != nil {
		openlog.Fatal("validate init failed: " + err.Error())
	}
	rbac.Init()
	pubsub.Init()
	pubsub.Start()
	cache.Init()
	if err := chassis.Run(); err != nil {
		openlog.Fatal("service exit: " + err.Error())
	}
}
