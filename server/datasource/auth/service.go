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

package auth

import (
	"context"

	"github.com/apache/servicecomb-kie/pkg/model"
)

const verbGet, verbCreate, verbUpdate, verbDelete = "get", "create", "update", "delete"

func configPerms(verb string, labels map[string]string) *ResourceScope {
	var labelsList []map[string]string
	if labels != nil {
		labelsList = append(labelsList, labels)
	}
	return &ResourceScope{
		Type:   "config",
		Verb:   verb,
		Labels: labelsList,
	}
}

func FilterKVList(ctx context.Context, kvs []*model.KVDoc) ([]*model.KVDoc, error) {
	if !CheckEnable(ctx) {
		return kvs, nil
	}
	// TODO error
	labels, err := CheckPerm(ctx, configPerms(verbGet, nil))
	if err != nil {
		return []*model.KVDoc{}, nil
	}
	if len(labels) == 0 {
		// allow all
		return kvs, nil
	}
	return FilterKVs(kvs, labels), nil
}

func CheckGetKV(ctx context.Context, kv *model.KVDoc) error {
	if !CheckEnable(ctx) {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbGet, kv.Labels))
	return err
}

func CheckCreateKV(ctx context.Context, kv *model.KVDoc) error {
	if !CheckEnable(ctx) {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbCreate, kv.Labels))
	return err
}

func CheckDeleteKV(ctx context.Context, kv *model.KVDoc) error {
	if !CheckEnable(ctx) {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbDelete, kv.Labels))
	return err
}

func CheckUpdateKV(ctx context.Context, kv *model.KVDoc) error {
	if !CheckEnable(ctx) {
		return nil
	}
	_, err := CheckPerm(ctx, configPerms(verbUpdate, kv.Labels))
	return err
}
