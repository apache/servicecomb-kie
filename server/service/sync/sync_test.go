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

package sync_test

import (
	"context"
	"testing"

	"github.com/apache/servicecomb-kie/server/service/sync"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	t.Run("parent is nil should new one", func(t *testing.T) {
		ctx := sync.NewContext(nil, true)
		assert.NotNil(t, ctx)
		assert.True(t, sync.FromContext(ctx))
	})
	t.Run("parent is not nil should be ok", func(t *testing.T) {
		ctx := sync.NewContext(context.Background(), true)
		assert.NotNil(t, ctx)
		assert.True(t, sync.FromContext(ctx))

		ctx = sync.NewContext(context.Background(), false)
		assert.False(t, sync.FromContext(ctx))
	})
}

func TestFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"no ctx should return false", args{nil}, false},
		{"ctx without sync should return false", args{context.Background()}, false},
		{"ctx with sync=true should return true", args{sync.NewContext(context.Background(), true)}, true},
		{"ctx with sync=false should return true", args{sync.NewContext(context.Background(), false)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sync.FromContext(tt.args.ctx); got != tt.want {
				t.Errorf("FromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
