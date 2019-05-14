#!/bin/bash
# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.



root_dir=/opt/servicecomb-kie
net_name=$(ip -o -4 route show to default | awk '{print $5}')
listen_addr=$(ifconfig ${net_name} | grep -E 'inet\W' | grep -o -E [0-9]+.[0-9]+.[0-9]+.[0-9]+ | head -n 1)


writeConfig(){
echo "write template config..."
cat <<EOM > ${root_dir}/conf/chassis.yaml
cse:
  service:
    registry:
      disabled: true
  protocols:
    rest:
      listenAddress: ${listen_addr}:30110
    rest-consul: #consul compatible API
      listenAddress: ${listen_addr}:8500
  handler:
    chain:
      Provider:
        default: auth-handler,ratelimiter-provider
EOM

cat <<EOM > /etc/servicecomb-kie/kie-conf.yaml
db:
  uri: mongodb://${MONGODB_USER}:${MONGODB_PWD}@${MONGODB_ADDR}/kie
  type: mongodb
  poolSize: 10
  ssl: false
  sslCA:
  sslCert:
EOM
}


echo "prepare config file...."
writeConfig

echo "start kie server"
/opt/servicecomb-kie/kie --config /etc/servicecomb-kie/kie-conf.yaml