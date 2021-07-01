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

package ctxsvc

import (
	"context"
	"github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/go-chassis/v2/core/common"
)

//ReadClaims get auth info
func ReadClaims(ctx context.Context) map[string]interface{} {
	c, err := rbac.FromContext(ctx)
	if err != nil {
		return nil
	}
	return c
}

//ReadDomain get domain info
func ReadDomain(ctx context.Context) string {
	c := ReadClaims(ctx)
	if c != nil {
		return c["domain"].(string)
	}
	return "default"
}

func SetProject(ctx context.Context, project string) context.Context {
	return common.WithContext(ctx, "project", project)
}

//ReadProject get project info
func ReadProject(ctx context.Context) string {
	c := ReadClaims(ctx)
	if c != nil {
		return c["project"].(string)
	}
	return "default"
}
