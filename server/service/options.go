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
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"time"
)

//NewDefaultFindOpts return default options
func NewDefaultFindOpts() FindOptions {
	return FindOptions{
		Timeout: session.DefaultTimeout,
	}
}

//FindOptions is option to find key value
type FindOptions struct {
	ExactLabels bool
	Status      string
	Depth       int
	ID          string
	Key         string
	Labels      map[string]string
	LabelID     string
	ClearLabel  bool
	Timeout     time.Duration
	PageNum     int64
	PageSize    int64
}

//FindOption is functional option to find key value
type FindOption func(*FindOptions)

//WithExactLabels tell model service to return only one kv matches the labels
func WithExactLabels() FindOption {
	return func(o *FindOptions) {
		o.ExactLabels = true
	}
}

//WithID find by kvID
func WithID(id string) FindOption {
	return func(o *FindOptions) {
		o.ID = id
	}
}

//WithKey find by key
func WithKey(key string) FindOption {
	return func(o *FindOptions) {
		o.Key = key
	}
}

//WithStatus enabled/disabled
func WithStatus(status string) FindOption {
	return func(o *FindOptions) {
		o.Status = status
	}
}

//WithTimeout will return err if execution take too long
func WithTimeout(d time.Duration) FindOption {
	return func(o *FindOptions) {
		o.Timeout = d
	}
}

//WithLabels find kv by labels
func WithLabels(labels map[string]string) FindOption {
	return func(o *FindOptions) {
		o.Labels = labels
	}
}

//WithLabelID find kv by labelID
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

//WithPageNum tells service paging limit
func WithPageNum(l int64) FindOption {
	return func(o *FindOptions) {
		o.PageNum = l
	}
}

//WithPageSize tells service paging offset
func WithPageSize(os int64) FindOption {
	return func(o *FindOptions) {
		o.PageSize = os
	}
}
