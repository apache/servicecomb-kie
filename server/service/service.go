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

package service

import (
	"context"
	"errors"
	"github.com/apache/servicecomb-kie/pkg/model"
)

//services
var (
	KVService      KV
	HistoryService History
	DBInit         Init
)

//db errors
var (
	ErrKeyNotExists     = errors.New("key with labels does not exits")
	ErrRevisionNotExist = errors.New("label revision not exist")
)

//KV provide api of KV entity
type KV interface {
	//below 3 methods is usually for admin console
	CreateOrUpdate(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error)
	List(ctx context.Context, domain, project, key string, labels map[string]string, limit, offset int) (*model.KVResponse, error)
	Delete(kvID string, labelID string, domain, project string) error
	//FindKV is usually for service to pull configs
	FindKV(ctx context.Context, domain, project string, options ...FindOption) ([]*model.KVResponse, error)
}

//History provide api of History entity
type History interface {
	GetHistory(ctx context.Context, labelID string, options ...FindOption) ([]*model.LabelRevisionDoc, error)
}

//Init init db session
type Init func() error
