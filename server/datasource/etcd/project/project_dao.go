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

package project

import (
	"context"
	"errors"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
)

//Dao operate data in etcd
type Dao struct {
}

//Total get projects total counts by domain
func (s *Dao) Total(ctx context.Context, domain string) (int64, error) {
	// TODO etcd needs to be done
	return 0, errors.New("can not list project,etcd not support yet")
}

//List get projects list by domain and name
func (s *Dao) List(ctx context.Context, domain string, options ...datasource.FindOption) (*model.ProjectResponse, error) {
	// TODO etcd needs to be done
	return nil, errors.New("can not list project,etcd not support yet")
}
