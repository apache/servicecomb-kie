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

package stringutil_test

import (
	"github.com/apache/servicecomb-kie/pkg/stringutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormat(t *testing.T) {
	s := stringutil.FormatMap(map[string]string{
		"service": "a",
		"version": "1",
	})
	s2 := stringutil.FormatMap(map[string]string{
		"version": "1",
		"service": "a",
	})
	t.Log(s)
	assert.Equal(t, s, s2)
	s3 := stringutil.FormatMap(nil)
	assert.Equal(t, "none", s3)
}
