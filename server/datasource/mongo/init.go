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

package mongo

import (
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/counter"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/history"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/kv"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/track"
	"github.com/go-chassis/openlog"
)

type Broker struct {
}

func NewFrom(c *datasource.Config) (datasource.Broker, error) {
	openlog.Info("use mongodb as storage")
	return &Broker{}, session.Init(c)
}
func (*Broker) GetRevisionDao() datasource.RevisionDao {
	return &counter.Service{}
}
func (*Broker) GetKVDao() datasource.KVDao {
	return &kv.Service{}
}
func (*Broker) GetHistoryDao() datasource.HistoryDao {
	return &history.Service{}
}
func (*Broker) GetTrackDao() datasource.TrackDao {
	return &track.Service{}
}
func init() {
	datasource.RegisterPlugin("mongo", NewFrom)
}
