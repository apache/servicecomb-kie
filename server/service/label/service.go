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

package label

import (
	"context"
	"fmt"

	"github.com/apache/servicecomb-kie/server/db"
	"github.com/go-mesh/openlogging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

//Exist check label exists or not and return label ID
func Exist(ctx context.Context, domain string, labels map[string]string) (primitive.ObjectID, error) {
	l, err := FindLabels(ctx, domain, labels)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find label failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": db.Timeout,
			}))
			return primitive.NilObjectID, fmt.Errorf("operation timout %s", db.Timeout)
		}
		return primitive.NilObjectID, err
	}

	return l.ID, nil

}

//ProjectHasLabel check whether the project has certain label or not, if so return label ID
func ProjectHasLabel(ctx context.Context, domain string, project string, labels map[string]string) (primitive.ObjectID, error) {
	l, err := projectHasLabels(ctx, domain, project, labels)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlogging.Error("find project's label failed, dead line exceeded", openlogging.WithTags(openlogging.Tags{
				"timeout": db.Timeout,
			}))
			return primitive.NilObjectID, fmt.Errorf("operation timout %s", db.Timeout)
		}
		return primitive.NilObjectID, err
	}
	return l.ID, nil
}
