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

// package dao supply pure persistence layer access
package datasource

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-chassis/openlog"

	"github.com/apache/servicecomb-kie/pkg/model"
)

var (
	b       Broker
	plugins = make(map[string]New)
)

var (
	ErrKeyNotExists     = errors.New("can not find any key value")
	ErrRecordNotExists  = errors.New("can not find any polling data")
	ErrRevisionNotExist = errors.New("revision does not exist")
	ErrKVAlreadyExists  = errors.New("kv already exists")
	ErrTooMany          = errors.New("key with labels should be only one")
)

const (
	DefaultValueType = "text"
	MaxHistoryNum    = 100

	ConfigResource = "config"
)

// New init db session
type New func(c *Config) (Broker, error)

func RegisterPlugin(name string, f New) {
	plugins[name] = f
}

// Broker avoid directly depend on one kind of persistence solution
type Broker interface {
	GetRevisionDao() RevisionDao
	GetHistoryDao() HistoryDao
	GetTrackDao() TrackDao
	GetKVDao() KVDao
}

func GetBroker() Broker {
	return b
}

// KVDao provide api of KV entity
type KVDao interface {
	// Create Update List are usually for admin console
	Create(ctx context.Context, kv *model.KVDoc, options ...WriteOption) (*model.KVDoc, error)
	Update(ctx context.Context, kv *model.KVDoc, options ...WriteOption) error
	List(ctx context.Context, project, domain string, options ...FindOption) (*model.KVResponse, error)
	//FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
	FindOneAndDelete(ctx context.Context, kvID string, project, domain string, options ...WriteOption) (*model.KVDoc, error)
	//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
	FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string, options ...WriteOption) ([]*model.KVDoc, int64, error)

	//Get return kv by id
	Get(ctx context.Context, req *model.GetKVRequest) (*model.KVDoc, error)
	Exist(ctx context.Context, key, project, domain string, options ...FindOption) (bool, error)
	// Total should return kv resource number by domain id and project id
	Total(ctx context.Context, project, domain string) (int64, error)
}

// HistoryDao provide api of History entity
type HistoryDao interface {
	AddHistory(ctx context.Context, kv *model.KVDoc) error
	GetHistory(ctx context.Context, keyID, project, domain string, options ...FindOption) (*model.KVResponse, error)
	DelayDeletionTime(ctx context.Context, kvIDs []string, project, domain string) error
}

// TrackDao provide api of Track entity
type TrackDao interface {
	CreateOrUpdate(ctx context.Context, detail *model.PollingDetail) (*model.PollingDetail, error)
	GetPollingDetail(ctx context.Context, detail *model.PollingDetail) ([]*model.PollingDetail, error)
}

// RevisionDao is global revision number management
type RevisionDao interface {
	GetRevision(ctx context.Context, domain string) (int64, error)
	ApplyRevision(ctx context.Context, domain string) (int64, error)
}

// ViewDao create update and get view data
type ViewDao interface {
	Create(ctx context.Context, viewDoc *model.ViewDoc, options ...FindOption) error
	Update(ctx context.Context, viewDoc *model.ViewDoc) error
	//TODO
	List(ctx context.Context, domain, project string, options ...FindOption) ([]*model.ViewDoc, error)
	GetCriteria(ctx context.Context, viewName, domain, project string) (map[string]map[string]string, error)
	GetContent(ctx context.Context, id, domain, project string, options ...FindOption) ([]*model.KVResponse, error)
}

func Init(kind string) error {
	var err error
	f, ok := plugins[kind]
	if !ok {
		return fmt.Errorf("do not support '%s'", kind)
	}
	dbc := &Config{}
	if b, err = f(dbc); err != nil {
		return err
	}
	openlog.Info(fmt.Sprintf("use %s as storage", kind))
	return nil
}

// ClearPart remove domain and project of kv
func ClearPart(kv *model.KVDoc) {
	kv.Domain = ""
	kv.Project = ""
	kv.LabelFormat = ""
}

// TombstoneID return tombstone's resourceID, using key and labelFormat as resourceID
func TombstoneID(kv *model.KVDoc) string {
	return kv.Key + "/" + kv.LabelFormat
}
