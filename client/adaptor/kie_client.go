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

package adaptor

import (
	"context"
	"errors"

	"github.com/apache/servicecomb-kie/client"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/go-chassis-config"
	"github.com/go-mesh/openlogging"
)

// Client contains the implementation of Client
type Client struct {
	KieClient *client.Client
	opts      config.Options
}

const (
	//Name of the Plugin
	Name = "servicecomb-kie"
)

// NewClient init the necessary objects needed for seamless communication to Kie Server
func NewClient(options config.Options) (config.Client, error) {
	kieClient := &Client{
		opts: options,
	}
	configInfo := client.Config{Endpoint: kieClient.opts.ServerURI, DefaultLabels: options.Labels, VerifyPeer: kieClient.opts.EnableSSL}
	var err error
	kieClient.KieClient, err = client.New(configInfo)
	if err != nil {
		openlogging.Error("KieClient Initialization Failed: " + err.Error())
	}
	openlogging.Debug("KieClient Initialized successfully")
	return kieClient, err
}

// PullConfigs is used for pull config from servicecomb-kie
func (c *Client) PullConfigs(labels ...map[string]string) (map[string]interface{}, error) {
	openlogging.Debug("KieClient begin PullConfigs")
	configsInfo := make(map[string]interface{})
	var err error
	var configurationsValue []*model.KVResponse
	if len(labels) != 0 {
		configurationsValue, err = c.KieClient.SearchByLabels(context.TODO(), client.WithGetProject("default"), client.WithLabels(labels...))
	} else {
		configurationsValue, err = c.KieClient.SearchByLabels(context.TODO(), client.WithGetProject("default"), client.WithLabels(c.opts.Labels))
	}
	if err != nil {
		openlogging.GetLogger().Errorf("Error in Querying the Response from Kie %s %#v", err.Error(), labels)
		return nil, err
	}
	openlogging.GetLogger().Debugf("KieClient SearchByLabels. %#v", labels)
	//Parse config result.
	for _, docRes := range configurationsValue {
		for _, docInfo := range docRes.Data {
			configsInfo[docInfo.Key] = docInfo.Value
		}
	}
	return configsInfo, nil
}

// PullConfig get config by key and labels.
func (c *Client) PullConfig(key, contentType string, labels map[string]string) (interface{}, error) {
	configurationsValue, err := c.KieClient.Get(context.TODO(), key, client.WithGetProject("default"), client.WithLabels(labels))
	if err != nil {
		openlogging.GetLogger().Error("Error in Querying the Response from Kie: " + err.Error())
		return nil, err
	}
	for _, doc := range configurationsValue {
		for _, kvDoc := range doc.Data {
			if key == kvDoc.Key {
				openlogging.GetLogger().Debugf("The Key Value of : ", kvDoc.Value)
				return doc, nil
			}
		}
	}
	return nil, errors.New("can not find value")
}

//PushConfigs put config in kie by key and labels.
func (c *Client) PushConfigs(data map[string]interface{}, labels map[string]string) (map[string]interface{}, error) {
	var configReq model.KVDoc
	configResult := make(map[string]interface{})
	for key, configValue := range data {
		configReq.Key = key
		configReq.Value = configValue.(string)
		configReq.Labels = labels
		configurationsValue, err := c.KieClient.Put(context.TODO(), configReq, client.WithProject("default"))
		if err != nil {
			openlogging.Error("Error in PushConfigs to Kie: " + err.Error())
			return nil, err
		}
		openlogging.Debug("The Key Value of : " + configurationsValue.Value)
		configResult[configurationsValue.Key] = configurationsValue.Value
	}
	return configResult, nil
}

//DeleteConfigsByKeys use keyId for delete
func (c *Client) DeleteConfigsByKeys(keys []string, labels map[string]string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for _, keyId := range keys {
		err := c.KieClient.Delete(context.TODO(), keyId, "", client.WithProject("default"))
		if err != nil {
			openlogging.Error("Error in Delete from Kie. " + err.Error())
			return nil, err
		}
		openlogging.GetLogger().Debugf("Delete The KeyId:%s", keyId)
	}
	return result, nil
}

//Watch not implemented because kie not support.
func (c *Client) Watch(f func(map[string]interface{}), errHandler func(err error), labels map[string]string) error {
	// TODO watch change events
	return errors.New("not implemented")
}

//Options.
func (c *Client) Options() config.Options {
	return c.opts
}

func init() {
	config.InstallConfigClientPlugin(Name, NewClient)
}
