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
	"os"

	"github.com/apache/servicecomb-kie/server"
	"github.com/apache/servicecomb-kie/server/command"
	"github.com/go-mesh/openlogging"
	//custom handlers
	_ "github.com/apache/servicecomb-kie/server/handler"
	_ "github.com/go-chassis/go-chassis/middleware/jwt"
	_ "github.com/go-chassis/go-chassis/middleware/monitoring"
	_ "github.com/go-chassis/go-chassis/middleware/ratelimiter"
	//storage
	_ "github.com/apache/servicecomb-kie/server/service/mongo"
	//quota management
	_ "github.com/apache/servicecomb-kie/server/plugin/qms"
)

func main() {
	if err := command.ParseConfig(os.Args); err != nil {
		openlogging.Fatal(err.Error())
	}

	server.Run()
}
