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

//Config is yaml file struct
type Config struct {
	DB   DB   `yaml:"db"`
	RBAC RBAC `yaml:"rbac"`
	//config from cli
	ConfigFile     string
	NodeName       string
	ListenPeerAddr string
	PeerAddr       string
	AdvertiseAddr  string
}

//DB is yaml file struct to set persistent config
type DB struct {
	URI         string `yaml:"uri"`
	Kind        string `yaml:"kind"`
	PoolSize    int    `yaml:"poolSize"`
	SSLEnabled  bool   `yaml:"sslEnabled"`
	RootCA      string `yaml:"rootCAFile"`
	CertFile    string `yaml:"certFile"`
	KeyFile     string `yaml:"keyFile"`
	CertPwdFile string `yaml:"certPwdFile"`
	Timeout     string `yaml:"timeout"`
	VerifyPeer  bool   `yaml:"verifyPeer"`
	SyncEnable  bool   `yaml:"syncEnabled"`
}

//RBAC is rbac config
type RBAC struct {
	Enabled    bool   `yaml:"enabled"`
	PubKeyFile string `yaml:"rsaPublicKeyFile"`
}
