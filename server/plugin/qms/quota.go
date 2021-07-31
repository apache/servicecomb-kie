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

package qms

import (
	"context"

	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/go-chassis/go-archaius"
	"github.com/go-chassis/go-chassis/v2/pkg/backends/quota"
	"github.com/go-chassis/openlog"
)

//const
const (
	DefaultQuota   = 10000
	QuotaConfigKey = "QUOTA_CONFIG"
)

//BuildInManager read env config to max config item number, and db total usage
// it is not a centralized QMS.
type BuildInManager struct {
}

//GetQuotas get usage and quota
func (m *BuildInManager) GetQuotas(serviceName, domain string) ([]*quota.Quota, error) {
	max := archaius.GetInt64(QuotaConfigKey, DefaultQuota)
	total, err := datasource.GetBroker().GetKVDao().Total(context.TODO(), domain)
	if err != nil {
		openlog.Error("find quotas failed: " + err.Error())
		return nil, err
	}
	return []*quota.Quota{{
		Limit: max,
		Used:  total,
	}}, nil
}

//IncreaseUsed no use
func (m *BuildInManager) IncreaseUsed(service, domain, resource string, used int64) error {
	return nil
}

//DecreaseUsed no use
func (m *BuildInManager) DecreaseUsed(service, domain, resource string, used int64) error {
	return nil
}
func newQMS(opts quota.Options) (quota.Manager, error) {
	return &BuildInManager{}, nil
}
func init() {
	quota.Install("build-in", newQMS)
}
