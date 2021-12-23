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
	"regexp"
	"strings"

	"github.com/go-chassis/cari/sync"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/apache/servicecomb-kie/pkg/util"
	"github.com/apache/servicecomb-kie/server/datasource"
	"github.com/apache/servicecomb-kie/server/datasource/mongo/session"
)

const (
	MsgFindKvFailed    = "find kv failed, deadline exceeded"
	FmtErrFindKvFailed = "can not find kv in %s"
)

//Dao operate data in mongodb
type Dao struct {
}

func (s *Dao) Create(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) (*model.KVDoc, error) {
	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if syncEnable is true, will create kv with task
		return txnCreate(ctx, kv)
	}
	return create(ctx, kv)
}

func create(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	_, err := collection.InsertOne(ctx, kv)
	if err != nil {
		openlog.Error("create error", openlog.WithTags(openlog.Tags{
			"err": err.Error(),
			"kv":  kv,
		}))
		return nil, err
	}
	return kv, nil
}

// txnCreate is to start transaction when creating KV, will create task in a transaction operation
func txnCreate(ctx context.Context, kv *model.KVDoc) (*model.KVDoc, error) {
	taskSession, err := session.GetDB().Client().StartSession()
	if err != nil {
		return nil, err
	}
	if err = taskSession.StartTransaction(); err != nil {
		return nil, err
	}
	defer taskSession.EndSession(ctx)
	if err = mongo.WithSession(ctx, taskSession, func(sessionContext mongo.SessionContext) error {
		collection := session.GetDB().Collection(session.CollectionKV)
		_, err = collection.InsertOne(sessionContext, kv)
		if err != nil {
			openlog.Error("create error", openlog.WithTags(openlog.Tags{
				"err": err.Error(),
				"kv":  kv,
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("fail to abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
					"kv":  kv,
				}))
			}
			return err
		}
		task, err := sync.NewTask(kv.Domain, kv.Project, sync.CreateAction, datasource.ConfigResource)
		if err != nil {
			openlog.Error("fail to create task" + err.Error())
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("fail to abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
					"kv":  kv,
				}))
			}
			return err
		}
		task.Data = kv
		collection = session.GetDB().Collection(session.CollectionTask)
		_, err = collection.InsertOne(sessionContext, task)
		if err != nil {
			openlog.Error("create task error", openlog.WithTags(openlog.Tags{
				"err":  err.Error(),
				"task": task,
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err":  errAbort.Error(),
					"task": task,
				}))
			}
			return err
		}
		if err = taskSession.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	}); err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	return kv, nil
}

//Update update key value
func (s *Dao) Update(ctx context.Context, kv *model.KVDoc, options ...datasource.WriteOption) error {
	opts := datasource.NewWriteOptions(options...)
	// if syncEnable is true, will create kv with task
	if opts.SyncEnable {
		return txnUpdate(ctx, kv)
	}
	return update(ctx, kv)
}

func update(ctx context.Context, kv *model.KVDoc) error {
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

// txnUpdate is to start transaction when updating kV, will create task in a transaction operation
func txnUpdate(ctx context.Context, kv *model.KVDoc) error {
	taskSession, err := session.GetDB().Client().StartSession()
	if err != nil {
		return err
	}
	if err = taskSession.StartTransaction(); err != nil {
		return err
	}
	defer taskSession.EndSession(ctx)
	if err = mongo.WithSession(ctx, taskSession, func(sessionContext mongo.SessionContext) error {
		collection := session.GetDB().Collection(session.CollectionKV)
		result := collection.FindOneAndUpdate(sessionContext, bson.M{"key": kv.Key, "label_format": kv.LabelFormat}, bson.D{
			{Key: "$set", Value: bson.D{
				{Key: "value", Value: kv.Value},
				{Key: "status", Value: kv.Status},
				{Key: "checker", Value: kv.Checker},
				{Key: "update_time", Value: kv.UpdateTime},
				{Key: "update_revision", Value: kv.UpdateRevision},
			}},
		})
		if result.Err() != nil {
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
					"kv":  kv,
				}))
			}
			if result.Err() == mongo.ErrNoDocuments {
				return datasource.ErrKeyNotExists
			}
			return result.Err()
		}
		curKV := &model.KVDoc{}
		err := result.Decode(curKV)
		if err != nil {
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
					"kv":  kv,
				}))
			}
			openlog.Error("decode error: " + err.Error())
			return err
		}
		task, err := sync.NewTask(kv.Domain, kv.Project, sync.UpdateAction, datasource.ConfigResource)
		if err != nil {
			openlog.Error("fail to create task" + err.Error())
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
					"kv":  kv,
				}))
			}
			return err
		}
		task.Data = curKV
		collection = session.GetDB().Collection(session.CollectionTask)
		_, err = collection.InsertOne(sessionContext, task)
		if err != nil {
			openlog.Error("create task error", openlog.WithTags(openlog.Tags{
				"err":  err.Error(),
				"task": task,
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err":  errAbort.Error(),
					"task": task,
				}))
			}
			return err
		}
		if err = taskSession.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	}); err != nil {
		openlog.Error(err.Error())
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
			filter["key"] = bson.M{"$regex": "^" + value + ".*$", "$options": "$i"}
		case strings.HasPrefix(opts.Key, "wildcard("):
			value := strings.ReplaceAll(getValue(opts.Key), ".", "\\.")
			value = strings.ReplaceAll(value, "*", ".*")
			filter["key"] = bson.M{"$regex": "^" + value + "$", "$options": "$i"}
		}
	}
	if len(opts.Labels) != 0 {
		for k, v := range opts.Labels {
			filter["labels."+k] = v
		}
	}
	opt := options.Find().SetSort(map[string]interface{}{
		"update_revision": -1,
	})
	if opts.Limit > 0 {
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
func (s *Dao) FindOneAndDelete(ctx context.Context, kvID, project, domain string, options ...datasource.WriteOption) (*model.KVDoc, error) {
	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if syncEnable is ture, will delete kv, create task and create tombstone
		return txnFindOneAndDelete(ctx, kvID, project, domain)
	}
	return findOneAndDelete(ctx, kvID, project, domain)
}

func findOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	curKV := &model.KVDoc{}
	collection := session.GetDB().Collection(session.CollectionKV)
	sr := collection.FindOneAndDelete(ctx, bson.M{"id": kvID, "project": project, "domain": domain})
	if sr.Err() != nil {
		if sr.Err() == mongo.ErrNoDocuments {
			return nil, datasource.ErrKeyNotExists
		}
		return nil, sr.Err()
	}
	err := sr.Decode(curKV)
	if err != nil {
		openlog.Error("decode error: " + err.Error())
		return nil, err
	}
	return curKV, nil
}

// txnFindOneAndDelete is to start transaction when delete KV, will create task and tombstone in a transaction operation
func txnFindOneAndDelete(ctx context.Context, kvID, project, domain string) (*model.KVDoc, error) {
	curKV := &model.KVDoc{}
	taskSession, err := session.GetDB().Client().StartSession()
	if err != nil {
		openlog.Error("fail to start session" + err.Error())
		return nil, err
	}
	if err = taskSession.StartTransaction(); err != nil {
		openlog.Error("fail to start transaction" + err.Error())
		return nil, err
	}
	defer taskSession.EndSession(ctx)
	if err = mongo.WithSession(ctx, taskSession, func(sessionContext mongo.SessionContext) error {
		collection := session.GetDB().Collection(session.CollectionKV)
		sr := collection.FindOneAndDelete(sessionContext, bson.M{"id": kvID, "project": project, "domain": domain})
		if sr.Err() != nil {
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
				return errAbort
			}
			if sr.Err() == mongo.ErrNoDocuments {
				openlog.Error(datasource.ErrKeyNotExists.Error())
				return datasource.ErrKeyNotExists
			}
			return sr.Err()
		}
		err := sr.Decode(curKV)
		if err != nil {
			openlog.Error("decode error: " + err.Error())
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
				return errAbort
			}
			return err
		}
		task, err := sync.NewTask(domain, project, sync.DeleteAction, datasource.ConfigResource)
		if err != nil {
			openlog.Error("fail to create task" + err.Error())
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
				return errAbort
			}
			return err
		}
		task.Data = curKV
		collection = session.GetDB().Collection(session.CollectionTask)
		_, err = collection.InsertOne(sessionContext, task)
		if err != nil {
			openlog.Error("create task error", openlog.WithTags(openlog.Tags{
				"err":  err.Error(),
				"task": task,
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err":  errAbort.Error(),
					"task": task,
				}))
			}
			return err
		}
		tombstone := sync.NewTombstone(domain, project, datasource.ConfigResource, datasource.TombstoneID(curKV))
		collection = session.GetDB().Collection(session.CollectionTombstone)
		_, err = collection.InsertOne(sessionContext, tombstone)
		if err != nil {
			openlog.Error("create tombstone error", openlog.WithTags(openlog.Tags{
				"err":       err.Error(),
				"tombstone": tombstone,
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err":       errAbort.Error(),
					"tombstone": tombstone,
				}))
			}
			return err
		}
		if err = taskSession.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	}); err != nil {
		openlog.Error(err.Error())
		return nil, err
	}
	return curKV, nil
}

//FindManyAndDelete deletes multiple kvs and return the deleted kv list as these appeared before deletion
func (s *Dao) FindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string, options ...datasource.WriteOption) ([]*model.KVDoc, int64, error) {
	opts := datasource.NewWriteOptions(options...)
	if opts.SyncEnable {
		// if sync enable is true, will delete kvs, create tasks and tombstones
		return txnFindManyAndDelete(ctx, kvIDs, project, domain)
	}
	return findManyAndDelete(ctx, kvIDs, project, domain)
}

func findManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
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

// txnFindManyAndDelete is to start transaction when delete KVs, will create tasks and tombstones in a transaction operation
func txnFindManyAndDelete(ctx context.Context, kvIDs []string, project, domain string) ([]*model.KVDoc, int64, error) {
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
	var deletedCount int64
	taskSession, err := session.GetDB().Client().StartSession()
	if err != nil {
		openlog.Error("fail to start session" + err.Error())
		return nil, 0, err
	}
	if err = taskSession.StartTransaction(); err != nil {
		openlog.Error("fail to start transaction" + err.Error())
		return nil, 0, err
	}
	defer taskSession.EndSession(ctx)

	if err = mongo.WithSession(ctx, taskSession, func(sessionContext mongo.SessionContext) error {
		collection := session.GetDB().Collection(session.CollectionKV)
		filter := bson.D{
			{Key: "id", Value: bson.M{"$in": kvIDs}},
			{Key: "project", Value: project},
			{Key: "domain", Value: domain}}
		dr, err := collection.DeleteMany(sessionContext, filter)
		if err != nil {
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
				return errAbort
			}
			openlog.Error(fmt.Sprintf("delete kvs [%v] failed : [%v]", kvIDs, err))
			return err
		}
		deletedCount = dr.DeletedCount
		tasksDoc := make([]interface{}, deletedCount)
		tombstonesDoc := make([]interface{}, deletedCount)
		for i := 0; i < int(deletedCount); i++ {
			kv := kvs[i]
			task, _ := sync.NewTask(domain, project, sync.DeleteAction, datasource.ConfigResource)
			task.Data = kv
			tombstone := sync.NewTombstone(domain, project, datasource.ConfigResource, datasource.TombstoneID(kv))
			tasksDoc[i] = task
			tombstonesDoc[i] = tombstone
		}
		collection = session.GetDB().Collection(session.CollectionTask)
		_, err = collection.InsertMany(sessionContext, tasksDoc)
		if err != nil {
			openlog.Error("create tasks error", openlog.WithTags(openlog.Tags{
				"err": err.Error(),
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
			}
			return err
		}
		collection = session.GetDB().Collection(session.CollectionTombstone)
		_, err = collection.InsertMany(sessionContext, tombstonesDoc)
		if err != nil {
			openlog.Error("create tombstone error", openlog.WithTags(openlog.Tags{
				"err": err.Error(),
			}))
			errAbort := taskSession.AbortTransaction(sessionContext)
			if errAbort != nil {
				openlog.Error("abort transaction", openlog.WithTags(openlog.Tags{
					"err": errAbort.Error(),
				}))
			}
			return err
		}
		if err = taskSession.CommitTransaction(sessionContext); err != nil {
			return err
		}
		return nil
	}); err != nil {
		openlog.Error(err.Error())
		return nil, 0, err
	}
	return kvs, deletedCount, nil
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

func (s *Dao) Total(ctx context.Context, project, domain string) (int64, error) {
	collection := session.GetDB().Collection(session.CollectionKV)
	filter := bson.M{"domain": domain, "project": project}
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
