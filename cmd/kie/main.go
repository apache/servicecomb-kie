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

package main

import (
	_ "github.com/apache/servicecomb-kie/server/handler"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/go-chassis/go-chassis"
	"github.com/go-mesh/openlogging"
	"os"
)

func main() {
	if err := Init(); err != nil {
		openlogging.Fatal(err.Error())
	}
	chassis.RegisterSchema("rest", &v1.KVResource{})
	if err := chassis.Init(); err != nil {
		openlogging.Error(err.Error())
		os.Exit(1)
	}
	if err := config.Init(Configs.ConfigFile); err != nil {
		openlogging.Error(err.Error())
		os.Exit(1)
	}
	if err := chassis.Run(); err != nil {
		openlogging.Error("service exit: " + err.Error())
		os.Exit(1)
	}
}
