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
	"github.com/apache/servicecomb-kie/server/datasource/local/file"
	"path"
	"strconv"

	"github.com/go-chassis/openlog"
)

const revision = "revision_counter"

// Dao is the implementation
type Dao struct {
}

// GetRevision return current revision number
func (s *Dao) GetRevision(ctx context.Context, domain string) (int64, error) {
	revisionPath := path.Join(file.FileRootPath, domain, "revision")

	revisionByte, err := file.ReadFile(revisionPath)

	if err != nil {
		openlog.Error("get error: " + err.Error())
		return 0, nil
	}
	if revisionByte == nil || string(revisionByte) == "" {
		return 0, nil
	}

	revisionNum, err := strconv.Atoi(string(revisionByte))
	if err != nil {
		return 0, err
	}
	return int64(revisionNum), nil
}

// ApplyRevision increase revision number and return modified value
func (s *Dao) ApplyRevision(ctx context.Context, domain string) (int64, error) {
	currentRevisionNum, err := s.GetRevision(ctx, domain)
	if err != nil {
		return 0, err
	}
	file.CreateOrUpdateFile(path.Join(file.FileRootPath, domain, "revision"), []byte(strconv.Itoa(int(currentRevisionNum+1))), []file.FileDoRecord{})
	return currentRevisionNum + 1, nil
}
