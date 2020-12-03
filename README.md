# Apache-ServiceComb-Kie 

[![Build Status](https://travis-ci.org/apache/servicecomb-kie.svg?branch=master)](https://travis-ci.org/apache/servicecomb-kie?branch=master) 
[![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![Coverage Status](https://coveralls.io/repos/github/apache/servicecomb-kie/badge.svg?branch=master)](https://coveralls.io/github/apache/servicecomb-kie?branch=master)
A service for configuration management in distributed system.

## Conceptions

### Key
Key could indicate a configuration like "timeout",
then the value could be "3s"
or indicates a file name "app.properties", 
then the value could be content of app.properties

### Labels
Each key could has labels. labels indicates a unique key.
A key "log_level" with labels "env=production" 
may saves the value "INFO" for all application log level in production environment.
A key "log_level" with labels "env=production, component=payment" 
may saves the value "DEBUG" for payment service in production environment.

It means all payment service print debug log, but for other service print info log.

So you can control your application runtime behaviors 
by setting different labels to a key.


## Why use kie
kie is a highly flexible config server. Nowadays, an operation team is facing different "x-centralized" system.
For example a classic application-centralized system. A operator wants to change config based on application name and version, then the label could be "app,version" for locating a app's configurations.
Meanwhile some teams manage app in a data center, each application instance will be deployed in a VM machine. then label could be "farm,role,server,component" to locate a app's configurations.
kie fit different senario for configuration management which benifit from label design.


## Components
It includes 1 components

- server: rest api service to manage kv

## Features
- kv management: you can manage config item by key and label
- kv revision mangement: you can mange all kv change history
- kv change event: use long polling to watch kv changes, highly decreased network cost
- polling detail track: if any client poll config from server, the detail will be tracked
## Quick Start

### Run locally with Docker compose

```bash
git clone git@github.com:apache/servicecomb-kie.git
cd servicecomb-kie/deployments/docker
sudo docker-compose up
```
It will launch 3 components 
- mongodb: 127.0.0.1:27017
- mongodb UI: http://127.0.0.1:8081
- servicecomb-kie: http://127.0.0.1:30110


## Development
To see how to build a local dev environment, check [here](examples/dev)

### Build
This will build your own service image and binary in local
```bash
cd build
export VERSION=0.0.1 #optional, it is latest by default
./build_docker.sh
```

This will generate a "servicecomb-kie-0.0.1-linux-amd64.tar" in "release" folder,
and a docker image "servicecomb/kie:0.0.1"

# API Doc
After you launch kie server, you can browse API doc in http://127.0.0.1:30110/apidocs.json, 
copy this doc to http://editor.swagger.io/
# Documentations
https://kie.readthedocs.io/en/latest/

or follow [here](docs/README.md) to generate it in local

## Clients
- go https://github.com/go-chassis/kie-client

## Contact

Bugs: [issues](https://issues.apache.org/jira/browse/SCB)

## Contributing

See [Contribution guide](http://servicecomb.apache.org/developers/contributing) for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.
