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

import "github.com/apache/servicecomb-kie/pkg/model"

func FilterKV(kvs []*model.KVDoc, labelsList []map[string]string) []*model.KVDoc {
	var permKVs []*model.KVDoc
	for _, kv := range kvs {
		for _, labels := range labelsList {
			if !matchOne(kv, labels) {
				continue
			}
			permKVs = append(permKVs, kv)
			break
		}
	}
	return permKVs
}

func matchOne(kv *model.KVDoc, labels map[string]string) bool {
	for lk, lv := range labels {
		if v, ok := kv.Labels[lk]; ok && v != lv {
			return false
		}
	}
	return true
}

func MatchLabelsList(kv *model.KVDoc, labelsList []map[string]string) bool {
	for _, labels := range labelsList {
		if !matchOne(kv, labels) {
			continue
		}
		return true
	}
	return false
}
