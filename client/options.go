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

package client

import "github.com/apache/servicecomb-kie/pkg/model"

//GetOption is the functional option of client func
type GetOption func(*GetOptions)

//GetOptions is the options of client func
type GetOptions struct {
	Labels map[string]string
	Depth  int
}

//WithLabels query kv by labels
func WithLabels(l map[string]string) GetOption {
	return func(options *GetOptions) {
		options.Labels = l
	}
}

//WithDepth query keys with partial match query labels
func WithDepth(d int) GetOption {
	return func(options *GetOptions) {
		options.Depth = d
	}
}

//PutOption is the functional option of client func
type PutOption func(*PutOptions)

//PutOptions is the options of client func
type PutOptions struct {
	Labels    map[string]string
	Value     string
	ValueType string
}

//SetKeyValue update or create key value
func SetKeyValue(kv model.KVDoc) PutOption {
	return func(options *PutOptions) {
		options.Labels = kv.Labels
		options.Value = kv.Value
		options.ValueType = kv.ValueType
	}
}
