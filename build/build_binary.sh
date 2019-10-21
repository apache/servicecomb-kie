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

if [ -d ${release_dir} ]; then
    rm -rf ${release_dir}
fi

mkdir -p ${release_dir}/conf


export GIT_COMMIT=`git rev-parse HEAD | cut -b 1-7`
echo "build from ${GIT_COMMIT}"




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
cp ${PROJECT_DIR}/licenses/LICENSE ${PROJECT_DIR}/licenses/NOTICE ${release_dir}
cp -r ${PROJECT_DIR}/licenses ${release_dir}
rm -f ${release_dir}/licenses/LICENSE ${release_dir}/licenses/NOTICE
cd ${release_dir}
component="apache-servicecomb-kie"

buildAndPackage(){
  GOOS=$1
  GOARCH=$2
  echo "building & packaging ${GOOS} ${GOARCH}..."
  GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${release_dir}/kie github.com/apache/servicecomb-kie/cmd/kieserver
  if [ $? -eq 0 ]; then
    tar zcf "$component-$VERSION-${GOOS}-${GOARCH}.tar.gz" conf kie LICENSE NOTICE licenses
  else
    echo -e "\033[31m build ${GOOS}-${GOARCH} fail !! \033[0m"
  fi
}

for GOOS in 'windows' 'darwin' 'linux'
do
  buildAndPackage $GOOS "amd64"
done

buildAndPackage "linux" "arm64"

