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

package handler

import (
	"github.com/go-chassis/go-chassis/core/handler"
	"github.com/go-chassis/go-chassis/core/invocation"
	"github.com/go-mesh/openlogging"
)

//NoopAuthHandler not need implement any logic
//developer can extend authenticate and authorization by set new handler in chassis.yaml
type NoopAuthHandler struct{}

//Handle set local attribute to http request
func (bk *NoopAuthHandler) Handle(chain *handler.Chain, inv *invocation.Invocation, cb invocation.ResponseCallBack) {
	inv.SetMetadata("domain", "default")
	chain.Next(inv, cb)
}

func newDomainResolver() handler.Handler {
	return &NoopAuthHandler{}
}

//Name is handler name
func (bk *NoopAuthHandler) Name() string {
	return "auth-handler"
}
func init() {
	if err := handler.RegisterHandler("auth-handler", newDomainResolver); err != nil {
		openlogging.Fatal("register auth-handler failed: " + err.Error())
	}
}
