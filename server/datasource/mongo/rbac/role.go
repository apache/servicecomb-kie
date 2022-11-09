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

package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chassis/openlog"

	crbac "github.com/go-chassis/cari/rbac"
	"github.com/little-cui/etcdadpt"
)

type RBAC_Mongo struct {
}

func (rm *RBAC_Mongo) GetRole(ctx context.Context, name string) (*crbac.Role, error) {
	kv, err := etcdadpt.Get(ctx, rm.GenerateRBACRoleKey(name))
	if err != nil {
		return nil, err
	}
	if kv == nil {
		return nil, errors.New("role not exist")
	}
	role := &crbac.Role{}
	err = json.Unmarshal(kv.Value, role)
	if err != nil {
		openlog.Error("role info format invalid", openlog.WithErr(err))
		return nil, err
	}
	return role, nil
}

func (rm *RBAC_Mongo) GenerateRBACRoleKey(name string) string {
	return "/cse-sr/roles/" + name
}

func (rm *RBAC_Mongo) AccountExist(ctx context.Context, name string) (bool, error) {
	return etcdadpt.Exist(ctx, rm.GenerateRBACAccountKey(name))
}

func (rm *RBAC_Mongo) GenerateRBACAccountKey(name string) string {
	return "/cse-sr/accounts/" + name
}
