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
	"context"

	"github.com/apache/servicecomb-kie/server/config"
	rbacmodel "github.com/go-chassis/cari/rbac"
)

// CheckPerm return the resource scope ...
func CheckPerm(ctx context.Context, targetResource *ResourceScope) ([]map[string]string, error) {
	account, err := Identify(ctx)
	if err != nil {
		return nil, err
	}
	hasAdmin, normalRoles := filterRoles(account.Roles)
	if hasAdmin {
		return nil, nil
	}
	return Allow(ctx, normalRoles, targetResource)
}

func filterRoles(roleList []string) (hasAdmin bool, normalRoles []string) {
	for _, r := range roleList {
		if r == rbacmodel.RoleAdmin {
			hasAdmin = true
			return
		}
		normalRoles = append(normalRoles, r)
	}
	return
}

func CheckEnable(ctx context.Context) bool {
	if !config.GetRBAC().Enabled {
		return false
	}
	if !config.GetRBAC().AllowMissToken {
		return true
	}
	claims, _ := rbacmodel.FromContext(ctx)
	return claims != nil
}
