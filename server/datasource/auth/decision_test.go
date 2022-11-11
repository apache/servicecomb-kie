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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLabelMatched(t *testing.T) {
	targetResourceLabel := map[string]string{
		"environment": "production",
		"appId":       "default",
	}
	t.Run("value not match, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"environment": "testing",
		}
		assert.False(t, LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("key not match, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"serviceName": "default",
		}
		assert.False(t, LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches no permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"version": "1.0.0",
		}
		assert.False(t, LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches part permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"version":     "1.0.0",
			"environment": "production",
		}
		assert.False(t, LabelMatched(targetResourceLabel, permResourceLabel))
	})
	t.Run("target resource label matches  permission resource label, should not match", func(t *testing.T) {
		permResourceLabel := map[string]string{
			"environment": "production",
		}
		assert.True(t, LabelMatched(targetResourceLabel, permResourceLabel))
	})
}
func TestFilterLabel(t *testing.T) {
	targetResourceLabel := []map[string]string{
		{"environment": "production", "appId": "default"},
		{"serviceName": "service-center"},
	}
	permResourceLabel := []map[string]string{
		{"environment": "production", "appId": "default"},
		{"appId": "default"},
		{"environment": "production", "serviceName": "service-center"},
		{"serviceName": "service-center", "version": "1.0.0"},
		{"serviceName": "service-center"},
		{"environment": "testing"},
	}
	l := FilterLabel(targetResourceLabel, permResourceLabel)
	assert.Equal(t, 3, len(l))
}
