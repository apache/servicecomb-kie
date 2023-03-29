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

package config

// Config is yaml file struct
type Config struct {
	DB    DB    `yaml:"db"`
	RBAC  RBAC  `yaml:"rbac"`
	Sync  Sync  `yaml:"sync"`
	Cache Cache `yaml:"cache"`
	//config from cli
	ConfigFile     string
	NodeName       string
	ListenPeerAddr string
	PeerAddr       string
	AdvertiseAddr  string
}

type TLS struct {
	SSLEnabled  bool   `yaml:"sslEnabled"`
	RootCA      string `yaml:"rootCAFile"`
	CertFile    string `yaml:"certFile"`
	KeyFile     string `yaml:"keyFile"`
	CertPwdFile string `yaml:"certPwdFile"`
	VerifyPeer  bool   `yaml:"verifyPeer"`
}

// DB is yaml file struct to set persistent config
type DB struct {
	TLS      `yaml:",inline" json:",inline"`
	URI      string `yaml:"uri" json:"uri,omitempty"`
	Kind     string `yaml:"kind" json:"kind,omitempty"`
	PoolSize int    `yaml:"poolSize" json:"pool_size,omitempty"`
	Timeout  string `yaml:"timeout" json:"timeout,omitempty"`
}

// RBAC is rbac config
type RBAC struct {
	Enabled        bool   `yaml:"enabled"`
	AllowMissToken bool   `yaml:"allowMissToken"`
	PubKeyFile     string `yaml:"rsaPublicKeyFile"`
}

// Sync is sync config
type Sync struct {
	Enabled bool `yaml:"enabled"`
}

type Cache struct {
	Labels []string `yaml:"labels"`
}
