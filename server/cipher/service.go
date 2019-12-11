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

package cipher

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/go-chassis/foundation/security"
)

func newCipherHistory(service service.History) service.History {
	return &cryptoHistory{service: service}
}

func newCipherKV(service service.KV) service.KV {
	return &cryptoKV{service: service}
}

// lookup Cipher
func lookupCrypto(unused *model.KVDoc) security.Cipher {
	return Lookup(config.GetCrypto().Name)
}

// History service security proxy
type cryptoHistory struct {
	service service.History
}

func (history *cryptoHistory) GetHistory(ctx context.Context, labelID string, options ...service.FindOption) ([]*model.LabelRevisionDoc, error) {
	res, err := history.service.GetHistory(ctx, labelID, options...)

	cipher := lookupCrypto(nil)

	for i := 0; i < len(res); i++ {
		doc := res[i]
		for j := 0; j < len(doc.KVs); j++ {
			kv := doc.KVs[j]
			val, err := cipher.Decrypt(kv.Value)
			if err != nil {
				return nil, err
			}
			kv.Value = val
		}
	}

	return res, err
}

// KV service security proxy
type cryptoKV struct {
	service service.KV
}

func (ckv *cryptoKV) CreateOrUpdate(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	cipher := lookupCrypto(nil)
	val, err := cipher.Encrypt(kv.Value)
	if err != nil {
		return nil, err
	}
	kv.Value = val

	res, err := ckv.service.CreateOrUpdate(ctx, kv)
	if res == nil {
		return res, err
	}

	val, err = cipher.Decrypt(kv.Value)
	if err != nil {
		return nil, err
	}
	kv.Value = val

	return res, err
}

func (ckv *cryptoKV) List(ctx context.Context, domain, project, key string, labels map[string]string, limit, offset int) (*model.KVResponse, error) {
	res, err := ckv.service.List(ctx, domain, project, key, labels, limit, offset)
	cipher := lookupCrypto(nil)
	if res == nil {
		return res, err
	}

	for j := 0; j < len(res.Data); j++ {
		kv := res.Data[j]
		val, err := cipher.Decrypt(kv.Value)
		if err != nil {
			return nil, err
		}
		kv.Value = val
	}

	return res, err
}

func (ckv *cryptoKV) Delete(kvID string, labelID string, domain, project string) error {
	return ckv.Delete(kvID, labelID, domain, project)
}

func (ckv *cryptoKV) FindKV(ctx context.Context, domain, project string, options ...service.FindOption) ([]*model.KVResponse, error) {
	res, err := ckv.service.FindKV(ctx, domain, project, options...)
	cipher := lookupCrypto(nil)
	if res == nil {
		return res, err
	}

	for i := 0; i < len(res); i++ {
		doc := res[i]
		for j := 0; j < len(doc.Data); j++ {
			kv := doc.Data[j]
			val, err := cipher.Decrypt(kv.Value)
			if err != nil {
				return nil, err
			}
			kv.Value = val
		}
	}

	return res, err
}
