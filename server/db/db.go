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

package db

import (
	"crypto/tls"
	"errors"
	"github.com/apache/servicecomb-kie/server/datasource/local/file"
	"time"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/config/tlsutil"
	"github.com/go-chassis/cari/db"
	dconfig "github.com/go-chassis/cari/db/config"
	"github.com/go-chassis/openlog"
)

const (
	DefaultTimeout = 60 * time.Second
	DefaultKind    = "mongo"
)

func Init(c config.DB) error {
	var err error
	if c.Kind == "" {
		c.Kind = DefaultKind
	}
	var timeout time.Duration
	if c.Timeout != "" {
		timeout, err = time.ParseDuration(c.Timeout)
		if err != nil {
			openlog.Fatal(err.Error())
			return errors.New("timeout setting invalid:" + c.Timeout)
		}
	}
	if timeout == 0 {
		timeout = DefaultTimeout
	}
	var tlsConfig *tls.Config
	if c.SSLEnabled {
		var err error
		tlsConfig, err = tlsutil.Config(&c.TLS)
		if err != nil {
			openlog.Fatal(err.Error())
			return errors.New("tls setting invalid:" + err.Error())
		}
	}

	if c.Kind == "etcd_with_localstorage" || c.Kind == "embedded_etcd_with_localstorage" {
		if c.Kind == "embedded_etcd_with_localstorage" {
			c.Kind = "embedded_etcd"
		}
		if c.Kind == "etcd_with_localstorage" {
			c.Kind = "etcd"
		}
		if c.LocalFilePath != "" {
			file.FileRootPath = c.LocalFilePath
		}
	}
	return db.Init(&dconfig.Config{
		Kind:       c.Kind,
		URI:        c.URI,
		PoolSize:   c.PoolSize,
		SSLEnabled: c.SSLEnabled,
		TLSConfig:  tlsConfig,
		Timeout:    timeout,
	})
}
