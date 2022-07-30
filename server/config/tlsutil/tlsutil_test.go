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

package tlsutil_test

import (
	"testing"

	_ "github.com/go-chassis/go-chassis/v2/security/cipher/plugins/plain"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/config/tlsutil"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/security/cipher"
	"github.com/stretchr/testify/assert"
)

const sslRoot = "./../../../examples/dev/ssl/"

func init() {
	err := archaius.Init()
	if err != nil {
		panic(err)
	}
	err = cipher.Init()
	if err != nil {
		panic(err)
	}
}

func TestConfig(t *testing.T) {
	t.Run("normal scene, should return ok", func(t *testing.T) {
		cfg, err := tlsutil.Config(&config.TLS{
			RootCA:      sslRoot + "trust.cer",
			CertFile:    sslRoot + "server.cer",
			KeyFile:     sslRoot + "server_key.pem",
			CertPwdFile: sslRoot + "cert_pwd",
			VerifyPeer:  false,
		})
		assert.NoError(t, err)
		assert.NotNil(t, cfg)
	})
	t.Run("without ca file, should return false", func(t *testing.T) {
		cfg, err := tlsutil.Config(&config.TLS{})
		assert.ErrorIs(t, tlsutil.ErrRootCAMissing, err)
		assert.Nil(t, cfg)
	})
	t.Run("set not exist pwd file, should return false", func(t *testing.T) {
		cfg, err := tlsutil.Config(&config.TLS{
			RootCA:      sslRoot + "trust.cer",
			CertFile:    sslRoot + "server.cer",
			KeyFile:     sslRoot + "server_key.pem",
			CertPwdFile: sslRoot + "xxx",
			VerifyPeer:  false,
		})
		assert.Error(t, err)
		assert.Nil(t, cfg)
	})
}
