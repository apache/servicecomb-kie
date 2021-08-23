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

package util_test

import (
	"testing"

	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestIsEquivalentLabel(t *testing.T) {
	var m1 map[string]string
	m2 := make(map[string]string)
	m3 := map[string]string{
		"foo": "bar",
	}
	m4 := map[string]string{
		"foo": "bar",
	}
	m5 := map[string]string{
		"bar": "foo",
	}
	assert.Equal(t, util.IsEquivalentLabel(m1, m1), true)
	assert.Equal(t, util.IsEquivalentLabel(m1, m2), true)
	assert.Equal(t, util.IsEquivalentLabel(m2, m3), false)
	assert.Equal(t, util.IsEquivalentLabel(m3, m4), true)
	assert.Equal(t, util.IsEquivalentLabel(m3, m5), false)
}
func BenchmarkIsEquivalentLabel(b *testing.B) {
	m1 := map[string]string{
		"foo": "bar",
		"a":   "b",
	}
	m2 := map[string]string{
		"foo": "bar",
		"c":   "d",
		"s":   "d",
	}
	for i := 0; i < b.N; i++ {
		util.IsEquivalentLabel(m1, m2)
	}
}
