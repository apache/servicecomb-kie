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

package util

import (
	"context"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

type StringContext struct {
	parentCtx context.Context
	kv        *ConcurrentMap
}

type CtxKey string

func (c *StringContext) Deadline() (deadline time.Time, ok bool) {
	return c.parentCtx.Deadline()
}

func (c *StringContext) Done() <-chan struct{} {
	return c.parentCtx.Done()
}

func (c *StringContext) Err() error {
	return c.parentCtx.Err()
}

func (c *StringContext) Value(key interface{}) interface{} {
	k, ok := key.(CtxKey)
	if !ok {
		return c.parentCtx.Value(key)
	}
	v, ok := c.kv.Get(k)
	if !ok {
		return FromContext(c.parentCtx, k)
	}
	return v
}

func (c *StringContext) SetKV(key CtxKey, val interface{}) {
	c.kv.Put(key, val)
}

func NewStringContext(ctx context.Context) *StringContext {
	strCtx, ok := ctx.(*StringContext)
	if !ok {
		strCtx = &StringContext{
			parentCtx: ctx,
			kv:        NewConcurrentMap(0),
		}
	}
	return strCtx
}

func SetContext(ctx context.Context, key CtxKey, val interface{}) context.Context {
	strCtx := NewStringContext(ctx)
	strCtx.SetKV(key, val)
	return strCtx
}

// FromContext return the value from ctx, return empty STRING if not found
func FromContext(ctx context.Context, key CtxKey) interface{} {
	if v := ctx.Value(key); v != nil {
		return v
	}
	return FromMetadata(ctx, key)
}

func SetRequestContext(r *http.Request, key CtxKey, val interface{}) *http.Request {
	ctx := r.Context()
	ctx = SetContext(ctx, key, val)
	if ctx != r.Context() {
		nr := r.WithContext(ctx)
		*r = *nr
	}
	return r
}

func FromMetadata(ctx context.Context, key CtxKey) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	if values, ok := md[string(key)]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}
