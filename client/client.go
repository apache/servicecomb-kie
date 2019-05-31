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

package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/apache/servicecomb-kie/pkg/common"
	"github.com/apache/servicecomb-kie/pkg/model"
	"github.com/go-chassis/foundation/httpclient"
	"github.com/go-chassis/foundation/security"
	"github.com/go-chassis/go-chassis/pkg/util/httputil"
	"github.com/go-mesh/openlogging"
	"net/http"
	"net/url"
)

const (
	APIPathKV = "v1/kv"
)

var (
	ErrKeyNotExist = errors.New("can not find value")
)

type Client struct {
	opts   Config
	cipher security.Cipher
	c      *httpclient.URLClient
}
type Config struct {
	Endpoint      string
	DefaultLabels map[string]string
	VerifyPeer    bool //TODO make it works, now just keep it false
}

func New(config Config) (*Client, error) {
	u, err := url.Parse(config.Endpoint)
	if err != nil {
		return nil, err
	}
	httpOpts := &httpclient.URLClientOption{}
	if u.Scheme == "https" {
		httpOpts.TLSConfig = &tls.Config{
			InsecureSkipVerify: !config.VerifyPeer,
		}
	}
	c, err := httpclient.GetURLClient(httpOpts)
	if err != nil {
		return nil, err
	}
	return &Client{
		opts: config,
		c:    c,
	}, nil
}

//GetValue get value of a key
func (c *Client) Get(ctx context.Context, key string, opts ...GetOption) ([]*model.KVDoc, error) {
	options := GetOptions{}
	for _, o := range opts {
		o(&options)
	}
	url := fmt.Sprintf("%s/%s/%s", c.opts.Endpoint, APIPathKV, key)
	h := http.Header{}
	if options.MatchMode != "" {
		h.Set(common.HeaderMatch, options.MatchMode)
	}
	resp, err := c.c.HTTPDoWithContext(ctx, "GET", url, h, nil)
	if err != nil {
		return nil, err
	}
	b := httputil.ReadBody(resp)
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrKeyNotExist
		}
		return nil, fmt.Errorf("get %s failed,http status [%s], body [%s]", key, resp.Status, b)
	}

	kvs := make([]*model.KVDoc, 0)
	err = json.Unmarshal(b, kvs)
	if err != nil {
		openlogging.Error("unmarshal kv failed:" + err.Error())
		return nil, err
	}
	return kvs, nil
}
