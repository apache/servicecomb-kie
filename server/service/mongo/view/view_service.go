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

package view

import (
	"context"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/service"
	"github.com/apache/servicecomb-kie/server/service/mongo/session"
	"github.com/go-mesh/openlogging"
	uuid "github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

//Service operate data in mongodb
type Service struct {
	timeout time.Duration
}

//Create insert a view data and create a mongo db view
func (s *Service) Create(ctx context.Context, viewDoc *model.ViewDoc, options ...service.FindOption) (*model.ViewDoc, error) {
	if viewDoc.Domain == "" {
		return nil, session.ErrMissingDomain
	}
	var pipeline mongo.Pipeline = []bson.D{
		{{"$match", bson.D{{"domain", viewDoc.Domain}}}},
		{{"$match", bson.D{{"project", viewDoc.Project}}}},
	}
	opts := service.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.Key != "" {
		pipeline = append(pipeline, bson.D{{"$match", bson.D{{"key", opts.Key}}}})
	}
	if len(opts.Labels) != 0 {
		for k, v := range opts.Labels {
			pipeline = append(pipeline, bson.D{{"$match", bson.D{{"labels." + k, v}}}})
		}
	}
	viewDoc.ID = uuid.NewV4().String()
	viewDoc.Criteria = "" //TODO parse pipe line to sql-like lang
	err := create(ctx, viewDoc)
	if err != nil {
		openlogging.Error("can not insert view collection: " + err.Error())
		return nil, session.ErrViewCreation
	}
	err = session.CreateView(ctx, generateViewName(viewDoc.ID, viewDoc.Domain, viewDoc.Project), session.CollectionKV, pipeline)
	if err != nil {
		openlogging.Error("can not create view: " + err.Error())
		return nil, session.ErrViewCreation
	}
	return viewDoc, nil
}

func (s *Service) Update(ctx context.Context, viewDoc *model.ViewDoc) error {
	if viewDoc.Domain == "" {
		return session.ErrMissingDomain
	}
	if viewDoc.Project == "" {
		return session.ErrMissingProject
	}
	if viewDoc.ID == "" {
		return session.ErrIDIsNil
	}
	oldView, err := findOne(ctx, viewDoc.ID, viewDoc.Domain, viewDoc.Project)
	if err != nil {
		return err
	}
	if viewDoc.Display != "" {
		oldView.Display = viewDoc.Display
	}
	if viewDoc.Criteria != "" {
		oldView.Criteria = viewDoc.Criteria
	}
	_, err = session.GetDB().Collection(session.CollectionView).UpdateOne(ctx, bson.M{"domain": oldView.Domain,
		"project": oldView.Project,
		"id":      oldView.ID},
		bson.D{
			{"$set", bson.D{
				{"name", oldView.Display},
				{"criteria", oldView.Criteria},
			}},
		})
	if err != nil {
		openlogging.Error("can not update view: " + err.Error())
		return session.ErrViewUpdate
	}
	return nil
}
func (s *Service) List(ctx context.Context, domain, project string, opts ...service.FindOption) (*model.ViewResponse, error) {
	option := service.FindOptions{}
	for _, o := range opts {
		o(&option)
	}
	collection := session.GetDB().Collection(session.CollectionView)
	filter := bson.M{"domain": domain, "project": project}
	mOpt := options.Find()
	if option.Limit != 0 {
		mOpt = mOpt.SetLimit(option.Limit)
	}
	if option.Offset != 0 {
		mOpt = mOpt.SetSkip(option.Offset)
	}
	cur, err := collection.Find(ctx, filter, mOpt)
	if err != nil {
		openlogging.Error("can not find view: " + err.Error())
		return nil, session.ErrViewFinding
	}
	result := &model.ViewResponse{}
	for cur.Next(ctx) {
		v := &model.ViewDoc{}
		if err := cur.Decode(v); err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		result.Data = append(result.Data, v)
	}
	return result, nil
}
func (s *Service) GetContent(ctx context.Context, id, domain, project string, opts ...service.FindOption) (*model.ViewResponse, error) {
	option := service.FindOptions{}
	for _, o := range opts {
		o(&option)
	}
	mOpt := options.Find()
	if option.Limit != 0 {
		mOpt = mOpt.SetLimit(option.Limit)
	}
	if option.Offset != 0 {
		mOpt = mOpt.SetSkip(option.Offset)
	}
	collection := session.GetDB().Collection(generateViewName(id, domain, project))
	cur, err := collection.Find(ctx, bson.D{}, mOpt)
	if err != nil {
		openlogging.Error("can not find view content: " + err.Error())
		return nil, session.ErrViewFinding
	}
	result := &model.ViewResponse{}
	for cur.Next(ctx) {
		v := &model.ViewDoc{}
		if err := cur.Decode(v); err != nil {
			openlogging.Error("decode error: " + err.Error())
			return nil, err
		}
		result.Data = append(result.Data, v)
	}
	return result, nil
}

func generateViewName(id, domain, project string) string {
	return domain + "::" + project + "::" + id
}
