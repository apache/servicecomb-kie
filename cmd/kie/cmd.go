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

package kie

import (
	"os"

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
