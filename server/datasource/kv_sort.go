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

package datasource

import (
	"sort"

	"github.com/apache/servicecomb-kie/pkg/model"
)

type KVDocSorter struct {
	KVs []*model.KVDoc
}

func (k *KVDocSorter) Len() int {
	return len(k.KVs)
}

func (k *KVDocSorter) Less(i, j int) bool {
	return k.KVs[i].UpdateRevision > k.KVs[j].UpdateRevision
}

func (k *KVDocSorter) Swap(i, j int) {
	k.KVs[i], k.KVs[j] = k.KVs[j], k.KVs[i]
}

func ReverseByUpdateRev(kvs []*model.KVDoc) {
	sorter := &KVDocSorter{KVs: kvs}
	sort.Sort(sorter)
}
