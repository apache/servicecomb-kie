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
	"errors"
	"fmt"

	"github.com/apache/servicecomb-kie/server/datasource"
	rbacmodel "github.com/go-chassis/cari/rbac"
	"github.com/go-chassis/openlog"
)

const (
	RootName = "root"
)

var ErrNoRoles = errors.New("no role found in token")

func Identify(ctx context.Context) (*rbacmodel.Account, error) {
	claims, err := rbacmodel.FromContext(ctx)
	if err != nil {
		openlog.Error("get account from token failed", openlog.WithErr(err))
		return nil, err
	}
	account, err := rbacmodel.GetAccount(claims)
	if err != nil {
		openlog.Error("get account from claims failed", openlog.WithErr(err))
		return nil, err
	}
	if len(account.Roles) == 0 {
		openlog.Error("no role found in token")
		return nil, rbacmodel.NewError(rbacmodel.ErrNoPermission, "no role found in token")
	}
	err = accountExist(ctx, account.Name)
	if err != nil {
		return nil, err
	}
	return account, nil
}

func accountExist(ctx context.Context, user string) error {
	// if root should pass, cause of root initialization
	if user == RootName {
		return nil
	}
	exist, err := datasource.GetBroker().GetRbacDao().AccountExist(ctx, user)
	if err != nil {
		return err
	}
	if !exist {
		msg := fmt.Sprintf("account [%s] is deleted", user)
		return rbacmodel.NewError(rbacmodel.ErrTokenOwnedAccountDeleted, msg)
	}
	return nil
}
