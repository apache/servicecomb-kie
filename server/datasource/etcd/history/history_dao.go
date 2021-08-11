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

package history

import (
	"context"
	"encoding/json"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/etcd/key"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/go-chassis/openlog"
	"github.com/little-cui/etcdadpt"
)

//Dao is the implementation
type Dao struct {
}

//GetHistory get all history by label id
func (s *Dao) GetHistory(ctx context.Context, kvID, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	opts := datasource.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	kvs, _, err := etcdadpt.List(ctx, key.HisList(domain, project, kvID), etcdadpt.WithOrderByCreate(), etcdadpt.WithDescendOrder())
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	return &model.KVResponse{
		Data:  pagingResult(kvs, opts.Offset, opts.Limit),
		Total: len(kvs),
	}, nil
}

func pagingResult(kvs []*mvccpb.KeyValue, offset, limit int64) []*model.KVDoc {
	total := int64(len(kvs))
	end := offset + limit
	if offset != 0 && limit != 0 {
		if offset >= total {
			return []*model.KVDoc{}
		}
		if end > total {
			end = total
		}
		kvs = kvs[offset:end]
	}
	histories := make([]*model.KVDoc, 0, len(kvs))
	for _, kv := range kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv.Value, &doc)
		if err != nil {
			openlog.Error("decode error: " + err.Error())
			continue
		}
		histories = append(histories, &doc)
	}
	return histories
}

//AddHistory add kv history
func (s *Dao) AddHistory(ctx context.Context, kv *model.KVDoc) error {
	bytes, err := json.Marshal(kv)
	if err != nil {
		openlog.Error("encode error: " + err.Error())
		return err
	}
	err = etcdadpt.PutBytes(ctx, key.His(kv.Domain, kv.Project, kv.ID, kv.UpdateRevision), bytes)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	err = historyRotate(ctx, kv.ID, kv.Project, kv.Domain)
	if err != nil {
		openlog.Error("history rotate err: " + err.Error())
		return err
	}
	return nil
}

//DelayDeletionTime add delete time to all revisions of the kv,
//thus these revisions will be automatically deleted by TTL index.
// TODO support delay deletion
func (s *Dao) DelayDeletionTime(ctx context.Context, kvIDs []string, project, domain string) error {
	var opts []etcdadpt.OpOptions
	for _, kvID := range kvIDs {
		opts = append(opts, etcdadpt.OpDel(etcdadpt.WithStrKey(key.HisList(domain, project, kvID)), etcdadpt.WithPrefix()))
	}
	_, err := etcdadpt.DeleteMany(ctx, opts...)
	if err != nil {
		openlog.Error("delete history error: " + err.Error())
		return err
	}
	return nil
}

//historyRotate delete historical versions for a key that exceeds the limited number
func historyRotate(ctx context.Context, kvID, project, domain string) error {
	kvs, curTotal, err := etcdadpt.List(ctx, key.HisList(domain, project, kvID), etcdadpt.WithKeyOnly(),
		etcdadpt.WithOrderByCreate(), etcdadpt.WithAscendOrder())
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	if curTotal <= datasource.MaxHistoryNum {
		return nil
	}
	kvs = kvs[:curTotal-datasource.MaxHistoryNum]
	return deleteMany(ctx, kvs)
}

func deleteMany(ctx context.Context, kvs []*mvccpb.KeyValue) error {
	var opts []etcdadpt.OpOptions
	for _, kv := range kvs {
		opts = append(opts, etcdadpt.OpDel(etcdadpt.WithKey(kv.Key)))
	}
	_, err := etcdadpt.DeleteMany(ctx, opts...)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	return nil
}
