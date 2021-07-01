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
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/go-chassis/v2/server/restful"
)

//services
var (
	KVService       KV
	HistoryService  History
	TrackService    Track
	RevisionService Revision
	DBInit          Init
	StrategyService OverrideStrategy
)

//db errors
var (
	ErrKeyNotExists     = errors.New("can not find any key value")
	ErrRecordNotExists  = errors.New("can not find any polling data")
	ErrRevisionNotExist = errors.New("revision does not exist")
	ErrAliasNotGiven    = errors.New("label alias not given")
)

//KV provide api of KV entity
type KV interface {
	//below 3 methods is usually for admin console
	Create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error)
	Update(ctx context.Context, kv *model.UpdateKVRequest) (*model.KVDoc, error)
	List(ctx context.Context, domain, project string, options ...FindOption) (*model.KVResponse, error)
	//FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
	FindOneAndDelete(ctx context.Context, kvID string, domain, project string) (*model.KVDoc, error)
	//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
	FindManyAndDelete(ctx context.Context, kvIDs []string, domain, project string) ([]*model.KVDoc, error)
	//Get return kv by id
	Get(ctx context.Context, request *model.GetKVRequest) (*model.KVDoc, error)
	//KV is a resource of kie, this api should return kv resource number by domain id
	Total(ctx context.Context, domain string) (int64, error)
}

//History provide api of History entity
type History interface {
	GetHistory(ctx context.Context, keyID string, options ...FindOption) (*model.KVResponse, error)
}

//History provide api of History entity
type Track interface {
	CreateOrUpdate(ctx context.Context, detail *model.PollingDetail) (*model.PollingDetail, error)
	GetPollingDetail(ctx context.Context, detail *model.PollingDetail) ([]*model.PollingDetail, error)
}

//Revision is global revision number management
type Revision interface {
	GetRevision(ctx context.Context, domain string) (int64, error)
}

//View create update and get view data
type View interface {
	Create(ctx context.Context, viewDoc *model.ViewDoc, options ...FindOption) error
	Update(ctx context.Context, viewDoc *model.ViewDoc) error
	//TODO
	List(ctx context.Context, domain, project string, options ...FindOption) ([]*model.ViewDoc, error)
	GetCriteria(ctx context.Context, viewName, domain, project string) (map[string]map[string]string, error)
	GetContent(ctx context.Context, id, domain, project string, options ...FindOption) ([]*model.KVResponse, error)
}

var overrideMap = make(map[string]OverrideStrategy)

type OverrideStrategy interface {
	Execute(input *model.KVDoc, rctx *restful.Context) (*model.KVDoc, *errsvc.Error)
}

func Register(override string, overrideStrategy OverrideStrategy) {
	overrideMap[override] = overrideStrategy
}

func GetType(override string) OverrideStrategy {
	return overrideMap[override]
}

//Init init db session
type Init func() error
