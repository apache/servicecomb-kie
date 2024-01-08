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

package track

import (
	"context"
	"encoding/json"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/local/file"
	"github.com/go-chassis/openlog"
	"os"
	"path"
)

// Dao is the implementation
type Dao struct {
}

// CreateOrUpdate create a record or update exist record
// If revision and session_id exists then update else insert
func (s *Dao) CreateOrUpdate(ctx context.Context, detail *model.PollingDetail) (*model.PollingDetail, error) {
	bytes, err := json.Marshal(detail)
	if err != nil {
		openlog.Error("encode polling detail error: " + err.Error())
		return nil, err
	}

	revision := "default"
	if detail.Revision != "" {
		revision = detail.Revision
	}
	trackPath := path.Join(file.FileRootPath, "track", detail.Domain, detail.Project, revision, detail.SessionID+".json")

	err = file.CreateOrUpdateFile(trackPath, bytes, &[]file.FileDoRecord{}, false)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	return detail, nil
}

// GetPollingDetail is to get a track data
func (s *Dao) GetPollingDetail(ctx context.Context, detail *model.PollingDetail) ([]*model.PollingDetail, error) {
	trackFolderPath := path.Join(file.FileRootPath, "track", detail.Domain, detail.Project)
	_, kvs, err := file.ReadAllFiles(trackFolderPath)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]*model.PollingDetail, 0, 0), nil
		}
		openlog.Error(err.Error())
		return nil, err
	}

	records := make([]*model.PollingDetail, 0, len(kvs))
	for _, kv := range kvs {
		var doc model.PollingDetail
		err := json.Unmarshal(kv, &doc)
		if err != nil {
			openlog.Error("decode polling detail error: " + err.Error())
			continue
		}
		if detail.SessionID != "" && doc.SessionID != detail.SessionID {
			continue
		}
		if detail.IP != "" && doc.IP != detail.IP {
			continue
		}
		if detail.UserAgent != "" && doc.UserAgent != detail.UserAgent {
			continue
		}
		if detail.URLPath != "" && doc.URLPath != detail.URLPath {
			continue
		}
		if detail.Revision != "" && doc.Revision != detail.Revision {
			continue
		}
		records = append(records, &doc)
	}
	if len(records) == 0 {
		return nil, datasource.ErrRecordNotExists
	}
	return records, nil
}
