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

package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	MsgFindKvFailed    = "find kv failed, deadline exceeded"
	FmtErrFindKvFailed = "can not find kv in %s"
)

//Dao operate data in mongodb
type Dao struct {
}

//Total get projects total counts by domain
func (s *Dao) Total(ctx context.Context, domain string) (int64, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	filter := bson.M{"domain": domain}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		openlog.Error("find total number: " + err.Error())
		return 0, err
	}
	return total, err
}

//List get projects list by domain and name
func (s *Dao) List(ctx context.Context, domain string, options ...datasource.FindOption) (*model.ProjectResponse, error) {
	opts := datasource.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	data, total, err := findProjects(ctx, domain, opts)
	if err != nil {
		return nil, err
	}
	result := &model.ProjectResponse{
		Data: []*model.ProjectDoc{},
	}
	for _, value := range data {
		curKV := &model.ProjectDoc{Project: value.(string)}
		result.Data = append(result.Data, curKV)
	}
	result.Total = total
	return result, nil
}

func findProjects(ctx context.Context, domain string, opts datasource.FindOptions) ([]interface{}, int, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	filter := bson.M{"domain": domain}
	if len(strings.TrimSpace(opts.Project)) > 0 {
		filter["project"] = opts.Project
	}
	result, err := collection.Distinct(ctx, "project", filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlog.Error(MsgFindKvFailed, openlog.WithTags(openlog.Tags{
				"timeout": opts.Timeout,
			}))
			return nil, 0, fmt.Errorf(FmtErrFindKvFailed, opts.Timeout)
		}
		return nil, 0, err
	}
	curTotal := len(result)
	offset := opts.Offset
	limit := offset + opts.Limit
	if offset > int64(curTotal) {
		offset = int64(curTotal)
		limit = int64(curTotal)
	} else if limit > int64(curTotal) {
		limit = int64(curTotal)
	}
	result = result[offset:limit]
	return result, len(result), err
}
