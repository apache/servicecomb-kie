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

package kv

import (
	"context"

	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/go-chassis/cari/config"
	"github.com/go-chassis/cari/pkg/errsvc"
	"github.com/go-chassis/openlog"
)

func ListProjects(ctx context.Context, request *model.ListProjectRequest) (int64, *model.ProjectResponse, *errsvc.Error) {
	opts := []datasource.FindOption{
		datasource.WithOffset(request.Offset),
		datasource.WithLimit(request.Limit),
	}
	if len(request.Project) > 0 {
		opts = append(opts, datasource.WithProject(request.Project))
	}
	rev, err := datasource.GetBroker().GetRevisionDao().GetRevision(ctx, request.Domain)
	if err != nil {
		return rev, nil, config.NewError(config.ErrInternal, err.Error())
	}
	kv, err := listProjectsHandle(ctx, request.Domain, opts...)
	if err != nil {
		openlog.Error("common: " + err.Error())
		return rev, nil, config.NewError(config.ErrInternal, common.MsgDBError)
	}
	return rev, kv, nil
}

func listProjectsHandle(ctx context.Context, domain string, options ...datasource.FindOption) (*model.ProjectResponse, error) {
	listSema.Acquire()
	defer listSema.Release()
	return datasource.GetBroker().GetProjectDao().List(ctx, domain, options...)
}
