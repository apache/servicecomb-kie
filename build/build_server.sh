#!/usr/bin/env bash
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
if [ -z "${GOPATH}" ]; then
 echo "missing GOPATH env, can not build"
 exit 1
fi
echo "GOPATH is "${GOPATH}
export BUILD_DIR=$(cd "$(dirname "$0")"; pwd)
export PROJECT_DIR=$(dirname ${BUILD_DIR})
echo "downloading dependencies"
cd ${PROJECT_DIR}
GO111MODULE=on go mod vendor
version="latest"
release_dir=${PROJECT_DIR}/release/kie

if [ -z "${VERSION}" ]; then
 echo "missing VERSION env, use ${version} as release version"
else
 version=${VERSION}
fi

CURRENT_OS=`go env GOHOSTOS`
if [ -z "${GO_HOST_OS}" ]; then
 echo "missing GO_HOST_OS env, use current OS ${GOOS} as host OS"
 GOOS=${CURRENT_OS}
else
 GOOS=${GO_HOST_OS}
fi

if [ -d ${release_dir} ]; then
    rm -rf ${release_dir}
fi

mkdir -p ${release_dir}/conf


export GIT_COMMIT=`git rev-parse HEAD | cut -b 1-7`
echo "build from ${GIT_COMMIT}"


echo "building x86..."
GOOS=${GOOS} go build -o ${release_dir}/kie github.com/apache/servicecomb-kie/cmd/kieserver

writeConfig(){
echo "write chassis config..."
cat <<EOM > ${release_dir}/conf/chassis.yaml
cse:
  service:
    registry:
      disabled: true
      address: http://127.0.0.1:30100
  protocols:
    rest:
      listenAddress: 127.0.0.1:30108
    rest-consul: #consul compatible API
      listenAddress: 127.0.0.1:8500
  handler:
    chain:
      Provider:
        default: auth-handler,ratelimiter-provider
EOM
echo "write miroservice config..."
cat <<EOM > ${release_dir}/conf/microservice.yaml
service_description:
  name: servicecomb-kie
  version: ${version}
EOM

cat <<EOM > ${release_dir}/conf/kie-conf.yaml
db:
  uri: mongodb://root:root@127.0.0.1:27017/kie
  type: mongodb
  poolSize: 10
  ssl: false
  sslCA:
  sslCert:
EOM
}

writeConfig

component="apache-servicecomb-kie"
x86_pkg_name="$component-$version-linux-amd64.tar.gz"
arm_pkg_name="$component-$version-linux-arm64.tar.gz"

echo "packaging x86 tar.gz..."
cp ${PROJECT_DIR}/LICENSE ${PROJECT_DIR}/NOTICE ${release_dir}
cd ${release_dir}
tar zcf ${x86_pkg_name} conf kie LICENSE NOTICE

echo "building docker..."
cp ${PROJECT_DIR}/scripts/start.sh ./
cp ${PROJECT_DIR}/build/docker/server/Dockerfile ./
sudo docker version
sudo docker build -t servicecomb/kie:${version} .


echo "building arm64"
GOOS=${GO_HOST_OS} GOARCH=arm64 go build -o ${release_dir}/kie github.com/apache/servicecomb-kie/cmd/kieserver
echo "packaging arm64 tar.gz..."
cp ${PROJECT_DIR}/LICENSE ${PROJECT_DIR}/NOTICE ${release_dir}
cd ${release_dir}
tar zcf ${arm_pkg_name} conf kie LICENSE NOTICE

