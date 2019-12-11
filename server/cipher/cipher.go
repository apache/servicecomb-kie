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

package cipher

import (
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/foundation/security"
)

//See https://github.com/go-chassis/foundation/blob/master/security/cipher.go
var ciphers map[string]security.Cipher

// Register is register crypto
func Register(name string, c security.Cipher) {
	if ciphers == nil {
		ciphers = make(map[string]security.Cipher)
	}
	ciphers[name] = c
}

// Lookup is lookup crypto
func Lookup(name string) security.Cipher {
	cipher, ok := ciphers[name]
	if !ok {
		cipher = &namedNoop{Name: name}
		ciphers[name] = cipher
	}

	return cipher
}

// Init init crypto config
func Init() error {
	if config.GetCrypto().Name == "" {
		return nil
	}
	if service.KVService != nil {
		service.KVService = newCipherKV(service.KVService)
	}
	if service.HistoryService != nil {
		service.HistoryService = newCipherHistory(service.HistoryService)
	}

	return nil
}
