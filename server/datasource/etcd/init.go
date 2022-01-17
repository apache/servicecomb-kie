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

package etcd

import (
	"crypto/tls"
	"fmt"

	// support embedded etcd
	_ "github.com/little-cui/etcdadpt/embedded"
	_ "github.com/little-cui/etcdadpt/remote"

	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/counter"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/history"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/kv"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/track"
	"github.com/apache/servicecomb-kie/server/datasource/tlsutil"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
)

type Broker struct {
}

func NewFrom(c *datasource.Config) (datasource.Broker, error) {
	kind := config.GetDB().Kind
	openlog.Info(fmt.Sprintf("use %s as storage", kind))
	var tlsConfig *tls.Config
	if c.SSLEnabled {
		var err error
		tlsConfig, err = tlsutil.Config(c)
		if err != nil {
			return nil, err
		}
	}
	return &Broker{}, etcdadpt.Init(etcdadpt.Config{
		Kind:             kind,
		ClusterAddresses: c.URI,
		SslEnabled:       c.SSLEnabled,
		TLSConfig:        tlsConfig,
	})
}
func (*Broker) GetRevisionDao() datasource.RevisionDao {
	return &counter.Dao{}
}
func (*Broker) GetKVDao() datasource.KVDao {
	return &kv.Dao{}
}
func (*Broker) GetHistoryDao() datasource.HistoryDao {
	return &history.Dao{}
}
func (*Broker) GetTrackDao() datasource.TrackDao {
	return &track.Dao{}
}

func init() {
	datasource.RegisterPlugin("etcd", NewFrom)
	datasource.RegisterPlugin("embedded_etcd", NewFrom)
}
