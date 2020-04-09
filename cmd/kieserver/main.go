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

package main

import (
	"os"

	"github.com/apache/servicecomb-kie/pkg/validate"
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/apache/servicecomb-kie/server/pubsub"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	"github.com/apache/servicecomb-kie/server/service"

	"github.com/go-chassis/go-chassis"
	"github.com/go-chassis/go-chassis/core/common"
	"github.com/go-mesh/openlogging"
	"github.com/urfave/cli"

	//custom handlers
	_ "github.com/apache/servicecomb-kie/server/handler"
	_ "github.com/go-chassis/go-chassis/middleware/monitoring"
	_ "github.com/go-chassis/go-chassis/middleware/ratelimiter"
	//storage
	_ "github.com/apache/servicecomb-kie/server/service/mongo"

	_ "github.com/apache/servicecomb-kie/server/plugin/qms"
)

const (
	defaultConfigFile = "/etc/servicecomb-kie/kie-conf.yaml"
)

// parseConfigFromCmd
func parseConfigFromCmd(args []string) (err error) {
	app := cli.NewApp()
	app.HideVersion = true
	app.Usage = "servicecomb-kie server cmd line."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "config file, example: --config=kie-conf.yaml",
			Destination: &config.Configurations.ConfigFile,
			Value:       defaultConfigFile,
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "node name, example: --name=kie0",
			Destination: &config.Configurations.NodeName,
			EnvVar:      "NODE_NAME",
		},
		cli.StringFlag{
			Name:        "peer-addr",
			Usage:       "kie use this ip port to join a kie cluster, example: --peer-addr=10.1.1.10:5000",
			Destination: &config.Configurations.PeerAddr,
			EnvVar:      "PEER_ADDR",
		}, cli.StringFlag{
			Name:        "listen-peer-addr",
			Usage:       "listen on ip port, kie receive events example: --listen-peer-addr=10.1.1.10:5000",
			Destination: &config.Configurations.ListenPeerAddr,
			EnvVar:      "LISTEN_PEER_ADDR",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}

	err = app.Run(args)
	return
}

func main() {
	if err := parseConfigFromCmd(os.Args); err != nil {
		openlogging.Fatal(err.Error())
	}
	chassis.RegisterSchema(common.ProtocolRest, &v1.KVResource{})
	chassis.RegisterSchema(common.ProtocolRest, &v1.HistoryResource{})
	if err := chassis.Init(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := config.Init(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := service.DBInit(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := validate.Init(); err != nil {
		openlogging.Fatal("validate init failed: " + err.Error())
	}
	pubsub.Init()
	pubsub.Start()
	if err := chassis.Run(); err != nil {
		openlogging.Fatal("service exit: " + err.Error())
	}
}
