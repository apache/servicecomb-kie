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
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/util"
	"regexp"
	"strings"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	MsgFindKvFailed    = "find kv failed, deadline exceeded"
	FmtErrFindKvFailed = "can not find kv in %s"
)

//Dao operate data in mongodb
type Dao struct {
}

func (s *Dao) Create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	var err error
	collection := session.GetDB().Collection(session.CollectionKV)
	_, err = collection.InsertOne(ctx, kv)
	if err != nil {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
			"kv":  kv,
		}))
		return nil, err
	}

	return kv, nil
}

//Update update key value
func (s *Dao) Update(ctx context.Context, kv *model.KVDoc) error {
	collection := session.GetDB().Collection(session.CollectionKV)
	_, err := collection.UpdateOne(ctx, bson.M{"key": kv.Key, "label_format": kv.LabelFormat}, bson.D{
		{Key: "$set", Value: bson.D{
			{Key: "value", Value: kv.Value},
			{Key: "status", Value: kv.Status},
			{Key: "checker", Value: kv.Checker},
			{Key: "update_time", Value: kv.UpdateTime},
			{Key: "update_revision", Value: kv.UpdateRevision},
		}},
	})
	if err != nil {
		return err
	}
	return nil

}

//Extract key values
func getValue(str string) string {
	rex := regexp.MustCompile(`\(([^)]+)\)`)
	res := rex.FindStringSubmatch(str)
	return res[len(res)-1]
}

func findKV(ctx context.Context, domain string, project string, opts datasource.FindOptions) (*mongo.Cursor, int, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	ctx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	filter := bson.M{"domain": domain, "project": project}
	if opts.Key != "" {
		filter["key"] = opts.Key
		switch {
		case strings.HasPrefix(opts.Key, "beginWith("):
			value := strings.ReplaceAll(getValue(opts.Key), ".", "\\.")
			filter["key"] = bson.M{"$regex": value, "$options": "$i"}
		case strings.HasPrefix(opts.Key, "wildcard("):
			value := strings.ReplaceAll(getValue(opts.Key), ".", "\\.")
			value = strings.ReplaceAll(value, "*", ".*")
			filter["key"] = bson.M{"$regex": value, "$options": "$i"}
		}
	}
	if len(opts.Labels) != 0 {
		for k, v := range opts.Labels {
			filter["labels."+k] = v
		}
	}
	opt := options.Find()
	if opts.Offset != 0 && opts.Limit != 0 {
		opt = opt.SetLimit(opts.Limit)
		opt = opt.SetSkip(opts.Offset)
	}
	curTotal, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlog.Error(MsgFindKvFailed, openlog.WithTags(openlog.Tags{
				"timeout": opts.Timeout,
			}))
			return nil, 0, fmt.Errorf(FmtErrFindKvFailed, opts.Timeout)
		}
		return nil, 0, err
	}
	if opts.Status != "" {
		filter["status"] = opts.Status
	}
	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlog.Error(MsgFindKvFailed, openlog.WithTags(openlog.Tags{
				"timeout": opts.Timeout,
			}))
			return nil, 0, fmt.Errorf(FmtErrFindKvFailed, opts.Timeout)
		}
		return nil, 0, err
	}
	return cur, int(curTotal), err
}
func findOneKey(ctx context.Context, filter bson.M) ([]*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	sr := collection.FindOne(ctx, filter)
	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return nil, datasource.ErrKeyNotExists
		}
		return nil, sr.Err()
	}
	curKV := &model.KVDoc{}
	err := sr.Decode(curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return []*model.KVDoc{curKV}, nil
}

//Exist supports you query a key value by label map or labels id
func (s *Dao) Exist(ctx context.Context, key, project, domain string, options ...datasource.FindOption) (bool, error) {
	opts := datasource.FindOptions{}
	for _, o := range options {
		o(&opts)
	}
	if opts.LabelFormat != "" {
		_, err := findKVByLabel(ctx, domain, opts.LabelFormat, key, project)
		if err != nil {
			if err == datasource.ErrKeyNotExists {
				return false, nil
			}
			openlog.Error(err.Error())
			return false, err
		}
		return true, nil
	}
	kvs, err := s.List(ctx, domain, project,
		datasource.WithExactLabels(),
		datasource.WithLabels(opts.Labels),
		datasource.WithKey(key))
	if err != nil {
		openlog.Error("check kv exist: " + err.Error())
		return false, err
	}
	if len(kvs.Data) != 1 {
		return false, datasource.ErrTooMany
	}

	return true, nil

}

//FindOneAndDelete deletes one kv by id and return the deleted kv as these appeared before deletion
//domain=tenant
func (s *Dao) FindOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	sr := collection.FindOneAndDelete(ctx, bson.M{"id": kvID, "project": project, "domain": domain})
	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return nil, datasource.ErrKeyNotExists
		}
		return nil, sr.Err()
	}
	curKV := &model.KVDoc{}
	err := sr.Decode(curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return curKV, nil
}

//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Dao) FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
	filter := bson.D{
		{Key: "id", Value: bson.M{"$in": kvIDs}},
		{Key: "project", Value: project},
		{Key: "domain", Value: domain}}
	kvs, err := findKeys(ctx, filter, false)
	if err != nil {
		if err != datasource.ErrKeyNotExists {
			openlog.Error("find Keys error: " + err.Error())
		}
		return nil, 0, err
	}
	collection := session.GetDB().Collection(session.CollectionKV)
	dr, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		openlog.Error(fmt.Sprintf("delete kvs [%v] failed : [%v]", kvIDs, err))
		return nil, 0, err
	}

	return kvs, dr.DeletedCount, nil
}

func findKeys(ctx context.Context, filter interface{}, withoutLabel bool) ([]*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	cur, err := collection.Find(ctx, filter)
	if err != nil {
		if err.Error() == context.DeadlineExceeded.Error() {
			openlog.Error("find kvs failed: " + err.Error())
			return nil, fmt.Errorf("can not find keys due to timout")
		}
		return nil, err
	}
	defer cur.Close(ctx)
	if cur.Err() != nil {
		return nil, err
	}
	kvs := make([]*model.KVDoc, 0)
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
		if err := cur.Decode(curKV); err != nil {
			openlog.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if withoutLabel {
			curKV.Labels = nil
		}
		kvs = append(kvs, curKV)

	}
	if len(kvs) == 0 {
		return nil, datasource.ErrKeyNotExists
	}
	return kvs, nil
}

//findKVByLabel get kvs by key and label
//key can be empty, then it will return all key values
//if key is given, will return 0-1 key value
func findKVByLabel(ctx context.Context, domain, labelFormat, key string, project string) ([]*model.KVDoc, error) {
	filter := bson.M{"label_format": labelFormat, "domain": domain, "project": project}
	if key != "" {
		filter["key"] = key
		return findOneKey(ctx, filter)
	}
	return findKeys(ctx, filter, true)

}

//Get get kv by kv id
func (s *Dao) Get(ctx context.Context, req *model.GetKVRequest) (*model.KVDoc, error) {
	filter := bson.M{"id": req.ID, "domain": req.Domain, "project": req.Project}
	kvs, err := findOneKey(ctx, filter)
	if err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	return kvs[0], nil
}

func (s *Dao) Total(ctx context.Context, domain string) (int64, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	filter := bson.M{"domain": domain}
	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		openlog.Error("find total number: " + err.Error())
		return 0, err
	}
	return total, err
}

//List get kv list by key and criteria
func (s *Dao) List(ctx context.Context, project, domain string, options ...datasource.FindOption) (*model.KVResponse, error) {
	opts := datasource.NewDefaultFindOpts()
	for _, o := range options {
		o(&opts)
	}
	cur, total, err := findKV(ctx, domain, project, opts)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	result := &model.KVResponse{
		Data: []*model.KVDoc{},
	}
	for cur.Next(ctx) {
		curKV := &model.KVDoc{}
		if err := cur.Decode(curKV); err != nil {
			openlog.Error("decode to KVs error: " + err.Error())
			return nil, err
		}
		if opts.ExactLabels {
			if !util.IsEquivalentLabel(opts.Labels, curKV.Labels) {
				continue
			}
		}
		datasource.ClearPart(curKV)
		result.Data = append(result.Data, curKV)
	}
	result.Total = total
	return result, nil
}
