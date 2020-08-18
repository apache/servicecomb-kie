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

package rbac

import (
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-service-center/pkg/rbacframe"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/middleware/jwt"
	"github.com/go-chassis/go-chassis/security/secret"
	"github.com/go-chassis/go-chassis/security/token"
	"github.com/go-mesh/openlogging"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	pubContentKey = "rbac.publicKey"
)

//Init initialize the rbac module
func Init() {
	if !config.GetRBAC().Enabled {
		openlogging.Info("rbac is disabled")
		return
	}

	jwt.Use(&jwt.Auth{
		MustAuth: func(req *http.Request) bool {
			if !config.GetRBAC().Enabled {
				return false
			}
			if strings.Contains(req.URL.Path, "/v1/health") {
				return false
			}
			return true
		},
		Realm: "servicecomb-kie-realm",
		SecretFunc: func(claims interface{}, method token.SigningMethod) (interface{}, error) {
			p, err := secret.ParseRSAPPublicKey(PublicKey())
			if err != nil {
				openlogging.Error("can not parse public key:" + err.Error())
				return nil, err
			}
			return p, nil
		},
		Authorize: func(payload map[string]interface{}, req *http.Request) error {
			payload["domain"] = "default" //TODO eliminate dead code
			newReq := req.WithContext(rbacframe.NewContext(req.Context(), payload))
			*req = *newReq
			//TODO role perm check
			return nil
		},
	})
	loadPublicKey()
	openlogging.Info("rbac is enabled")
}

//loadPublicKey read key to memory
func loadPublicKey() {
	pf := config.GetRBAC().PubKeyFile
	content, err := ioutil.ReadFile(filepath.Clean(pf))
	if err != nil {
		openlogging.Fatal(err.Error())
		return
	}
	err = archaius.Set(pubContentKey, string(content))
	if err != nil {
		openlogging.Fatal(err.Error())
	}
}

//PublicKey get public key to verify a token
func PublicKey() string {
	return archaius.GetString(pubContentKey, "")
}
