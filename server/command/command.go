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

package command

import (
	"github.com/apache/servicecomb-kie/server/config"
	"github.com/urfave/cli"
)

const (
	defaultConfigFile = "/etc/servicecomb-kie/kie-conf.yaml"
)

// parseConfigFromCmd
func ParseConfig(args []string) (err error) {
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
		},
		cli.StringFlag{
			Name:        "listen-peer-addr",
			Usage:       "listen on ip port, kie receive events example: --listen-peer-addr=10.1.1.10:5000",
			Destination: &config.Configurations.ListenPeerAddr,
			EnvVar:      "LISTEN_PEER_ADDR",
		},
		cli.StringFlag{
			Name:        "advertise-addr",
			Usage:       "kie advertise addr, kie advertise addr example: --advertise-addr=10.1.1.10:5000",
			Destination: &config.Configurations.AdvertiseAddr,
			EnvVar:      "ADVERTISE_ADDR",
		},
	}
	app.Action = func(c *cli.Context) error {
		return nil
	}

	err = app.Run(args)
	return
}
