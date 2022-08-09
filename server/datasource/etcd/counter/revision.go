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

package counter

import (
	"context"

	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
)

const revision = "revision_counter"

// Dao is the implementation
type Dao struct {
}

// GetRevision return current revision number
func (s *Dao) GetRevision(ctx context.Context, domain string) (int64, error) {
	kv, err := etcdadpt.Get(ctx, key.Counter(revision, domain))
	if err != nil {
		openlog.Error("get error: " + err.Error())
		return 0, err
	}
	if kv == nil {
		return 0, nil
	}
	return kv.Version, nil
}

// ApplyRevision increase revision number and return modified value
func (s *Dao) ApplyRevision(ctx context.Context, domain string) (int64, error) {
	resp, err := etcdadpt.PutBytesAndGet(ctx, key.Counter(revision, domain), nil)
	if err != nil {
		openlog.Error("put bytes error: " + err.Error())
		return 0, err
	}
	if resp.Count == 0 {
		return 0, datasource.ErrRevisionNotExist
	}
	return resp.Kvs[0].Version, nil
}
