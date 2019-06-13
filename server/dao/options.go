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

package dao

//FindOption is option to find key value
type FindOptions struct {
	ExactLabels bool
	Depth       int
	Key         string
	Labels      map[string]string
	LabelID     string
	ClearLabel  bool
}

//FindOption is functional option to find key value
type FindOption func(*FindOptions)

//WithExactLabels tell model service to return only one kv matches the labels
func WithExactLabels() FindOption {
	return func(o *FindOptions) {
		o.ExactLabels = true
	}
}

//WithKey find by key
func WithKey(key string) FindOption {
	return func(o *FindOptions) {
		o.Key = key
	}
}

//WithLabels find kv by labels
func WithLabels(labels map[string]string) FindOption {
	return func(o *FindOptions) {
		o.Labels = labels
	}
}

//WithLabels find kv by labelID
func WithLabelID(label string) FindOption {
	return func(o *FindOptions) {
		o.LabelID = label
	}
}

//WithDepth if you use greedy match this can specify the match depth
func WithDepth(d int) FindOption {
	return func(o *FindOptions) {
		o.Depth = d
	}
}

//WithOutLabelField will clear all labels attributes in kv doc
func WithOutLabelField() FindOption {
	return func(o *FindOptions) {
		o.ClearLabel = true
	}
}
