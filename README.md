# Apache-ServiceComb-Kie [![Build Status](https://travis-ci.org/apache/servicecomb-kie.svg?branch=master)](https://travis-ci.org/apache/servicecomb-kie?branch=master) [![License](https://img.shields.io/badge/license-Apache%202-4EB1BA.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

A service for configuration management in distributed system.

## Conceptions

### Key
key could indicate a configuration like "timeout",
then the value could be "3s"
or indicates a file name "app.properties", 
then the value could be content of app.properties

### labels
Each key could has labels. labels indicates a unique key.
A key "log_level" with labels "env=production" 
may saves the value "INFO" for all application log level in production environment.
A key "log_level" with labels "env=production, component=payment" 
may saves the value "DEBUG" for payment service in production environment.

it means all payment service print debug log, but for other service print info log.

so you can control your application runtime behaviors 
by setting different labels to a key.

## Components
it includes 5 components

- server: rest api service to manage kv
- client: restful client for go
- kie-template: agent can be deployed in your k8s pod 
or VM, it connects to server and writes kv into config file 
based on template language
- kiectl: CLI tool for kie
- frontend: web console for kie

## Features
- simple key name with rich labels: user can define labels for a key, 
that distinguish from key to another key.  

TODO

- a key will not be stringed by fixed schema. 
labels for a key is like "env=test, service=cart, version=1.0" or "cluster=xxx"  
or "env=test, service=cart, version=1.0, ip=x.x.x.x"
- validator: value can be checked by user defined python script, 
so in runtime if someone want to change this value, 
the script will check if this value is appropriate.
- encryption web hook: value can by encrypt 
by your custom encryption service like vault.
- Long polling: client can get key value changes by long polling
- config view: by setting labels criteria, servicecomb-kie 
is able to aggregate a view to return all key values which match those labels, 
so that operator can mange key in their own understanding 
to a distributed system in separated views.
- rich value type: not only plain text, but support to be aware of ini, json,yaml,xml and java properties
- heterogeneous config server: able to fetch configuration in k8s and consul 
 even more, you can update, delete, 
 and use config view for those systems, 
 and you can integrate with your own config system to MetaConfig by 
 following standardized API and model
- consul compatible: partially compatible with consul kv management API
- kv change history: all kv changes is recorded and can be easily roll back by UI
## Quick Start

### Run locally with Docker compose

```bash
git clone git@github.com:apache/servicecomb-kie.git
cd servicecomb-kie/deployments/docker
sudo docker-compose up
```
it will launch 3 components 
- mongodb: 127.0.0.1:27017
- mongodb UI:http://127.0.0.1:8081
- servicecomb-kie: http://127.0.0.1:30110


## Development
To see how to build a local dev environment, check [here](examples/dev)

### Build
this will build your own service image and binary in local
```bash
cd build
export VERSION=0.0.1 #optional, it is latest by default
./build_docker.sh
```

this will generate a "servicecomb-kie-0.0.1-linux-amd64.tar" in "release" folder,
and a docker image "servicecomb/kie:0.0.1"

# API Doc
After you launch kie server, you can browse API doc in http://127.0.0.1:30110/apidocs, 
copy this doc to http://editor.swagger.io/
# Documentations
https://servicecomb.apache.org/docs/users/

or follow [here](docs/README.md) to generate it in local

## Contact

Bugs: [issues](https://issues.apache.org/jira/browse/SCB)

## Contributing

See [Contribution guide](http://servicecomb.apache.org/developers/contributing) for details on submitting patches and the contribution workflow.

## Reporting Issues

See reporting bugs for details about reporting any issues.
