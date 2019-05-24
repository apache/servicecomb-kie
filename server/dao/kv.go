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

//package dao is a persis layer of kie
package dao

import (
	"crypto/tls"
	"errors"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"time"
)

var ErrMissingDomain = errors.New("domain info missing, illegal access")
var ErrNotExists = errors.New("key with labels does not exits")
var ErrTooMany = errors.New("key with labels should be only one")
var ErrKeyMustNotEmpty = errors.New("must supply key if you want to get exact one result")

type KV interface {
	CreateOrUpdate(kv *model.KV) (*model.KV, error)
	//do not use primitive.ObjectID as return to decouple with mongodb, we can afford perf lost
	Exist(key, domain string, labels model.Labels) (string, error)
	Delete(ids []string, domain string) error
	Find(domain string, options ...FindOption) ([]*model.KV, error)
	AddHistory(kv *model.KV) error
	//RollBack(kv *KV, version string) error
}

type Options struct {
	URI      string
	PoolSize int
	SSL      bool
	TLS      *tls.Config
	Timeout  time.Duration
}

func NewKVService() (KV, error) {
	opts := Options{
		URI:      config.GetDB().URI,
		PoolSize: config.GetDB().PoolSize,
		SSL:      config.GetDB().SSL,
	}
	if opts.SSL {

	}
	return NewMongoService(opts)
}
