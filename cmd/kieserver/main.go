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
	"github.com/apache/servicecomb-kie/server/service"
	"os"

	"github.com/apache/servicecomb-kie/server/config"
	_ "github.com/apache/servicecomb-kie/server/handler"
	v1 "github.com/apache/servicecomb-kie/server/resource/v1"
	_ "github.com/apache/servicecomb-kie/server/service/mongo"
	"github.com/go-chassis/go-chassis"
	"github.com/go-mesh/openlogging"
	"github.com/urfave/cli"
)

const (
	defaultConfigFile = "/etc/servicecomb-kie/kie-conf.yaml"
)

//ConfigFromCmd store cmd params
type ConfigFromCmd struct {
	ConfigFile string
}

//Configs is a pointer of struct ConfigFromCmd
var Configs *ConfigFromCmd

// parseConfigFromCmd
func parseConfigFromCmd(args []string) (err error) {
	app := cli.NewApp()
	app.HideVersion = true
	app.Usage = "servicecomb-kie server cmd line."
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "config",
			Usage:       "config file, example: --config=kie-conf.yaml",
			Destination: &Configs.ConfigFile,
			Value:       defaultConfigFile,
		},
		cli.StringFlag{
			Name:        "name",
			Usage:       "node name, example: --name=kie0",
			Destination: &Configs.ConfigFile,
			EnvVar:      "NODE_NAME",
		},
		cli.StringFlag{
			Name:        "peer-addr",
			Usage:       "peer address any node address in a cluster, example: --peer-addr=10.1.1.10:5000",
			Destination: &Configs.ConfigFile,
			EnvVar:      "PEER_ADDR",
		},
		cli.StringFlag{
			Name:        "listen-peer-addr",
			Usage:       "peer address, example: --listen-peer-addr=0.0.0.0:5000",
			Destination: &Configs.ConfigFile,
			EnvVar:      "LISTEN_PEER_ADDR",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}

	err = app.Run(args)
	return
}

//Init get config and parses those command
func Init() error {
	Configs = &ConfigFromCmd{}
	return parseConfigFromCmd(os.Args)
}
func main() {
	if err := Init(); err != nil {
		openlogging.Fatal(err.Error())
	}
	chassis.RegisterSchema("rest", &v1.KVResource{})
	chassis.RegisterSchema("rest", &v1.HistoryResource{})
	if err := chassis.Init(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := config.Init(Configs.ConfigFile); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := service.DBInit(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := service.CryptoInit(); err != nil {
		openlogging.Fatal(err.Error())
	}
	if err := chassis.Run(); err != nil {
		openlogging.Fatal("service exit: " + err.Error())
	}
}
