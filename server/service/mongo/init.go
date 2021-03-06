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
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/counter"
	"github.com/apache/servicecomb-kie/server/service/mongo/history"
	"github.com/apache/servicecomb-kie/server/service/mongo/kv"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/apache/servicecomb-kie/server/service/mongo/track"
	"github.com/go-chassis/openlog"
)

func init() {
	openlog.Info("use mongodb as storage")
	service.DBInit = session.Init
	service.KVService = &kv.Service{}
	service.HistoryService = &history.Service{}
	service.TrackService = &track.Service{}
	service.RevisionService = &counter.Service{}
}
