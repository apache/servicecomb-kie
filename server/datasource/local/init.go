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

package local

import (
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/local/counter"
	"github.com/apache/servicecomb-kie/server/datasource/local/history"
	"github.com/apache/servicecomb-kie/server/datasource/local/kv"
	"github.com/apache/servicecomb-kie/server/datasource/local/rbac"
	"github.com/apache/servicecomb-kie/server/datasource/local/track"
	rbacdao "github.com/apache/servicecomb-kie/server/datasource/rbac"
)

type Broker struct {
}

func NewFrom(c *datasource.Config) (datasource.Broker, error) {
	return &Broker{}, nil
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
func (*Broker) GetRbacDao() rbacdao.Dao {
	return &rbac.Dao{}
}

func init() {
	datasource.RegisterPlugin("etcd_with_localstorage", NewFrom)
	datasource.RegisterPlugin("embedded_etcd_with_localstorage", NewFrom)
}
