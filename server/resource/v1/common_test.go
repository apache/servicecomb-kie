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

package v1_test

import (
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/emicklei/go-restful"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetLabels(t *testing.T) {
	r, err := http.NewRequest("GET",
		"/kv?q=app:mall+service:payment&q=app:mall+service:payment+version:1.0.0",
		nil)
	assert.NoError(t, err)
	c, err := v1.ReadLabelCombinations(restful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(c))

	r, err = http.NewRequest("GET",
		"/kv",
		nil)
	assert.NoError(t, err)
	c, err = v1.ReadLabelCombinations(restful.NewRequest(r))
	assert.NoError(t, err)
	assert.Equal(t, 1, len(c))

	r, err = http.NewRequest("GET",
		"/kv?label=app:mall&label=service:payment",
		nil)
	assert.NoError(t, err)
	req := restful.NewRequest(r)
	m, err := v1.GetLabels(req.QueryParameters("label"))
	assert.NoError(t, err)
	assert.Equal(t, 2, len(m))
}
