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

package crypto

//Cipher interface declares two function for encryption and decryption
//See https://github.com/go-chassis/foundation/blob/master/security/cipher.go
type Cipher interface {
	Encrypt(src string) (string, error)

	Decrypt(src string) (string, error)
}

var ciphers map[string]Cipher

func Register(name string, c Cipher) {
	if ciphers == nil {
		ciphers = make(map[string]Cipher)
	}
	ciphers[name] = c
}

func Lookup(name string) Cipher {
	cipher, ok := ciphers[name]
	if !ok {
		return &namedNoop{Name: name}
	}

	return cipher
}
