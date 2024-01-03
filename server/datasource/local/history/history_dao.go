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
	"github.com/apache/servicecomb-kie/server/datasource/local/file"
	"path"
	"strconv"

	"github.com/apache/servicecomb-kie/server/datasource/auth"
	"github.com/go-chassis/openlog"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
)

// Dao is the implementation
type Dao struct {
}

// GetHistory get all history by label id
func (s *Dao) GetHistory(ctx context.Context, kvID, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	kvreq := &model.GetKVRequest{
		Domain:  domain,
		Project: project,
		ID:      kvID,
	}
	kvdoc, err := datasource.GetBroker().GetKVDao().Get(ctx, kvreq)
	if err != nil {
		return nil, err
	}
	if err := auth.CheckGetKV(ctx, kvdoc); err != nil {
		return nil, err
	}

	opts := datasource.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	kvFolderPath := path.Join(file.FileRootPath, domain, project, kvID)
	_, kvs, err := file.ReadAllFiles(kvFolderPath)

	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	histories := make([]*model.KVDoc, 0, len(kvs))
	for _, kv := range kvs {
		var doc model.KVDoc
		err := json.Unmarshal(kv, &doc)
		if err != nil {
			openlog.Error("decode error: " + err.Error())
			continue
		}
		histories = append(histories, &doc)
	}
	return &model.KVResponse{
		Data:  pagingResult(histories, opts.Offset, opts.Limit),
		Total: len(kvs),
	}, nil
}

func pagingResult(histories []*model.KVDoc, offset, limit int64) []*model.KVDoc {
	total := int64(len(histories))
	if limit != 0 && offset >= total {
		return []*model.KVDoc{}
	}

	datasource.ReverseByPriorityAndUpdateRev(histories)

	if limit == 0 {
		return histories
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return histories[offset:end]
}

// AddHistory add kv history
func (s *Dao) AddHistory(ctx context.Context, kv *model.KVDoc) error {
	err := s.historyRotate(ctx, kv.ID, kv.Project, kv.Domain)
	return err
}

// DelayDeletionTime add delete time to all revisions of the kv,
// thus these revisions will be automatically deleted by TTL index.
// TODO support delay deletion
func (s *Dao) DelayDeletionTime(ctx context.Context, kvIDs []string, project, domain string) error {
	// history have been deleted in function <FindManyAndDelete> and <FindOneAndDelete> in kv_dao.go
	return nil
}

// historyRotate delete historical versions for a key that exceeds the limited number
func (s *Dao) historyRotate(ctx context.Context, kvID, project, domain string) error {
	resp, err := s.GetHistory(ctx, kvID, project, domain)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	if resp.Total <= datasource.MaxHistoryNum {
		return nil
	}
	kvs := resp.Data
	kvs = kvs[datasource.MaxHistoryNum:]

	for _, kv := range kvs {
		revision := kv.UpdateRevision
		revisionFilePath := path.Join(file.FileRootPath, domain, project, kvID, strconv.FormatInt(revision, 10)+".json")
		err = file.DeleteFile(revisionFilePath, &[]file.FileDoRecord{})
		if err != nil {
			openlog.Error(err.Error())
			return err
		}
	}
	return nil
}
