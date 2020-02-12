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

package record

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
)

//Service is db service
type Service struct {
}

//RecordSuccess record success
func (s *Service) RecordSuccess(ctx context.Context, detail *model.PollingDetail, respStatus int, respData, respHeader string) (*model.PollingDetail, error) {
	detail.ResponseCode = respStatus
	detail.ResponseBody = respData
	detail.ResponseHeader = respHeader
	return CreateOrUpdateRecord(ctx, detail)
}

//RecordFailed record failed
func (s *Service) RecordFailed(ctx context.Context, detail *model.PollingDetail, respStatus int, respData string) (*model.PollingDetail, error) {
	detail.ResponseCode = respStatus
	detail.ResponseBody = respData
	return CreateOrUpdateRecord(ctx, detail)
}
