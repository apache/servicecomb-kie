#!/bin/sh
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

set -e

root_dir=/opt/servicecomb-kie
net_name=$(ip -o -4 route show to default | awk '{print $5}')
listen_addr=$(ifconfig ${net_name} | grep -E 'inet\W' | grep -o -E [0-9]+.[0-9]+.[0-9]+.[0-9]+ | head -n 1)
if [ -z "${LOG_LEVEL}" ]; then
  export LOG_LEVEL="DEBUG"
fi

writeConfig() {
  echo "write template config..."
  cat <<EOM >${root_dir}/conf/chassis.yaml
servicecomb:
  registry:
    disabled: true
  protocols:
    rest:
      listenAddress: ${listen_addr}:30110
  match:
    rateLimitPolicy: |
      matches:
        - apiPath:
            contains: "/kie"
  rateLimiting:
    limiterPolicy1: |
      match: rateLimitPolicy
      rate: 200
      burst: 200
  handler:
    chain:
      Provider:
        default: access-log,monitoring,jwt,track-handler,traffic-marker,rate-limiter
EOM

  cat <<EOM >${root_dir}/conf/lager.yaml
logWriters: file
logLevel: ${LOG_LEVEL}
logFile: log/chassis.log
logFormatText: true
LogRotateCompress: true
logRotateSize: 30
logBackupCount: 20
logRotateAge: 0
accessLogFile: log/access.log
EOM

  db_type=${DB_KIND:-"mongo"}
  uri=${DB_URI}
  if [ -z "${uri}" ] && [ "${db_type}" == "mongo" ]; then
    uri="mongodb://${MONGODB_ADDR}/kie"
    if [ -n "${MONGODB_USER}" ]; then
      uri="mongodb://${MONGODB_USER}:${MONGODB_PWD}@${MONGODB_ADDR}/kie"
    fi
  fi

  cat <<EOM >${root_dir}/conf/kie-conf.yaml
db:
  kind: ${db_type}
  uri: ${uri}
  localFilePath: ${KVS_ROOT_PATH}
EOM
}

echo "prepare config file...."
writeConfig

echo "start kie server"
/opt/servicecomb-kie/kie --config ${root_dir}/conf/kie-conf.yaml
