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

//Package session manage db connection
package session

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/cipherutil"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/openlog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2"
	"io/ioutil"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/apache/servicecomb-kie/server/config"
	"go.mongodb.org/mongo-driver/mongo"
)

//const for db name and collection name
const (
	DBName = "kie"

	CollectionLabel         = "label"
	CollectionKV            = "kv"
	CollectionKVRevision    = "kv_revision"
	CollectionPollingDetail = "polling_detail"
	CollectionCounter       = "counter"
	CollectionView          = "view"
	DefaultTimeout          = 5 * time.Second
	DefaultValueType        = "text"
)

//db errors
var (
	ErrMissingDomain   = errors.New("domain info missing, illegal access")
	ErrMissingProject  = errors.New("project info missing, illegal access")
	ErrLabelNotExists  = errors.New("labels does not exits")
	ErrTooMany         = errors.New("key with labels should be only one")
	ErrKeyMustNotEmpty = errors.New("must supply key if you want to get exact one result")

	ErrIDIsNil         = errors.New("id is empty")
	ErrKeyIsNil        = errors.New("key must not be empty")
	ErrRootCAMissing   = errors.New("rootCAFile is empty in config file")
	ErrKVAlreadyExists = errors.New("kv already exists")

	ErrViewCreation = errors.New("can not create view")
	ErrViewUpdate   = errors.New("can not update view")
	ErrViewDelete   = errors.New("can not delete view")
	ErrViewNotExist = errors.New("view not exists")
	ErrViewFinding  = errors.New("view search error")
	ErrGetPipeline  = errors.New("can not get criteria")
)

const (
	MsgDBExists  = "already exists"
	MsgDuplicate = "duplicate key error collection"
)

var client *mongo.Client
var once sync.Once
var db *mongo.Database

//Timeout db operation time out
var Timeout time.Duration

//Init prepare params
func Init() error {
	var err error
	if config.GetDB().Timeout != "" {
		Timeout, err = time.ParseDuration(config.GetDB().Timeout)
		if err != nil {
			return errors.New("timeout setting invalid:" + config.GetDB().Timeout)
		}
	}
	if Timeout == 0 {
		Timeout = DefaultTimeout
	}
	once.Do(func() {
		sc, _ := bsoncodec.NewStructCodec(bsoncodec.DefaultStructTagParser)
		reg := bson.NewRegistryBuilder().
			RegisterTypeEncoder(reflect.TypeOf(model.LabelDoc{}), sc).
			RegisterTypeEncoder(reflect.TypeOf(model.KVDoc{}), sc).
			Build()
		uri := cipherutil.TryDecrypt(config.GetDB().URI)
		clientOps := []*options.ClientOptions{options.Client().ApplyURI(uri)}
		if config.GetDB().SSLEnabled {
			if config.GetDB().RootCA == "" {
				openlog.Error(ErrRootCAMissing.Error())
				return
			}
			pool := x509.NewCertPool()
			caCert, err := ioutil.ReadFile(config.GetDB().RootCA)
			if err != nil {
				openlog.Error(fmt.Sprintf("read ca cert file %s failed", caCert))
				return
			}
			pool.AppendCertsFromPEM(caCert)
			// #nosec
			tc := &tls.Config{
				RootCAs:            pool,
				InsecureSkipVerify: !config.GetDB().VerifyPeer,
			}
			clientOps = append(clientOps, options.Client().SetTLSConfig(tc))
			openlog.Info("enabled ssl communication to mongodb")
		}
		client, err = mongo.NewClient(clientOps...)
		if err != nil {
			return
		}
		openlog.Info("DB connecting")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err = client.Connect(ctx)
		if err != nil {
			return
		}
		openlog.Info("DB connected")
		db = client.Database(DBName, &options.DatabaseOptions{
			Registry: reg,
		})

	})
	EnsureDB()
	return nil
}

//GetDB get mongo db client
func GetDB() *mongo.Database {
	return db
}

//CreateView run mongo db command to create view
func CreateView(ctx context.Context, view, source string, pipeline mongo.Pipeline) error {
	sr := GetDB().RunCommand(ctx,
		bson.D{
			{Key: "create", Value: view},
			{Key: "viewOn", Value: source},
			{Key: "pipeline", Value: pipeline},
		})
	if sr.Err() != nil {
		openlog.Error("can not create view: " + sr.Err().Error())
		return ErrViewCreation
	}
	return nil
}

//DropView deletes view
func DropView(ctx context.Context, view string) error {
	err := GetDB().Collection(view).Drop(ctx)
	if err != nil {
		openlog.Error(err.Error())
		return err
	}
	return nil
}

//GetColInfo get collection info
func GetColInfo(ctx context.Context, name string) (*CollectionInfo, error) {
	cur, err := GetDB().ListCollections(ctx, bson.M{"name": name, "type": "view"})
	if err != nil {
		openlog.Error(err.Error())
		return nil, ErrGetPipeline
	}
	defer cur.Close(ctx)
	if !cur.Next(ctx) {
		return nil, ErrGetPipeline
	}
	openlog.Debug(cur.Current.String())
	c := &CollectionInfo{}
	if err := cur.Decode(c); err != nil {
		openlog.Error(err.Error())
		return nil, ErrGetPipeline
	}
	return c, nil
}

//EnsureDB build mongo db schema
func EnsureDB() {
	session := OpenSession()
	defer session.Close()
	session.SetMode(mgo.Primary, true)

	ensureRevisionCounter(session)

	ensureKV(session)

	ensureKVRevision(session)

	ensureView(session)

	ensureKVLongPolling(session)
}

func OpenSession() *mgo.Session {
	var timeout time.Duration
	var uri string
	var err error
	if config.GetDB().Timeout != "" {
		timeout, err = time.ParseDuration(config.GetDB().Timeout)
		if err != nil {
			openlog.Fatal("invalid timeout :" + err.Error())
		}
	}
	uri = cipherutil.TryDecrypt(config.GetDB().URI)
	session, err := mgo.DialWithTimeout(uri, timeout)
	if err != nil {
		openlog.Warn("can not dial db, retry once:" + err.Error())
		session, err = mgo.DialWithTimeout(uri, timeout)
		if err != nil {
			openlog.Fatal("can not dial db:" + err.Error())
		}
	}
	return session
}

func ensureKVLongPolling(session *mgo.Session) {
	c := session.DB(DBName).C(CollectionPollingDetail)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		"id":         bson.M{"$exists": true},
		"revision":   bson.M{"$exists": true},
		"session_id": bson.M{"$exists": true},
		"url_path":   bson.M{"$exists": true},
	}})
	wrapError(err, MsgDBExists)
	err = c.EnsureIndex(mgo.Index{
		Key:         []string{"timestamp"},
		ExpireAfter: 7 * 24 * time.Hour,
	})
	wrapError(err)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"revision", "domain", "session_id"},
		Unique: true,
	})
	wrapError(err)
}

func ensureView(session *mgo.Session) {
	c := session.DB(DBName).C(CollectionView)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		"id":      bson.M{"$exists": true},
		"domain":  bson.M{"$exists": true},
		"project": bson.M{"$exists": true},
		"display": bson.M{"$exists": true},
	}})
	wrapError(err, MsgDBExists)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"id"},
		Unique: true,
	})
	wrapError(err)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"display", "domain", "project"},
		Unique: true,
	})
	wrapError(err)
}

func ensureKVRevision(session *mgo.Session) {
	c := session.DB(DBName).C(CollectionKVRevision)
	err := c.EnsureIndex(mgo.Index{
		Key:         []string{"delete_time"},
		ExpireAfter: 7 * 24 * time.Hour,
	})
	wrapError(err, MsgDBExists)
}

func ensureKV(session *mgo.Session) {
	c := session.DB(DBName).C(CollectionKV)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		"key":     bson.M{"$exists": true},
		"domain":  bson.M{"$exists": true},
		"project": bson.M{"$exists": true},
		"id":      bson.M{"$exists": true},
	}})
	wrapError(err, MsgDBExists)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"id"},
		Unique: true,
	})
	wrapError(err)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"key", "label_format", "domain", "project"},
		Unique: true,
	})
	wrapError(err)
}

func ensureRevisionCounter(session *mgo.Session) {
	c := session.DB(DBName).C(CollectionCounter)
	err := c.Create(&mgo.CollectionInfo{Validator: bson.M{
		"name":   bson.M{"$exists": true},
		"domain": bson.M{"$exists": true},
		"count":  bson.M{"$exists": true},
	}})
	wrapError(err, MsgDBExists)
	err = c.EnsureIndex(mgo.Index{
		Key:    []string{"name", "domain"},
		Unique: true,
	})
	wrapError(err)
	docs := map[string]interface{}{"name": "revision_counter", "count": 1, "domain": "default"}
	err = c.Insert(docs)
	wrapError(err, MsgDuplicate)
}

func wrapError(err error, skipMsg ...string) {
	if err != nil {
		for _, str := range skipMsg {
			if strings.Contains(err.Error(), str) {
				openlog.Debug(err.Error())
				return
			}
		}
		openlog.Fatal(err.Error())
	}
}
