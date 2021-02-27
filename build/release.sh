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

export BUILD_DIR=$(cd "$(dirname "$0")"; pwd)
export PROJECT_DIR=$(dirname ${BUILD_DIR})
sign(){
  #asc
  gpg --armor --output "$1".asc --detach-sig "$1"
  #sha512
  sha512sum "$1" > "$1".sha512
}
component="apache-servicecomb-kie"
x86_pkg_name="$component-$VERSION-linux-amd64.tar.gz"
arm_pkg_name="$component-$VERSION-linux-arm64.tar.gz"
darwin_pkg_name="$component-$VERSION-darwin-amd64.tar.gz"
windows_pkg_name="$component-$VERSION-windows-amd64.tar.gz"
cd $PROJECT_DIR/release/kie
sign "${x86_pkg_name}"
sign "${arm_pkg_name}"
sign "${darwin_pkg_name}"
sign "${windows_pkg_name}"
#src
wget "https://github.com/apache/servicecomb-kie/archive/v${VERSION}.tar.gz"
tar xzf "v${VERSION}.tar.gz"
mv servicecomb-kie-${VERSION} apache-servicecomb-kie-${VERSION}
src_name="${component}-${VERSION}-src.tar.gz"
tar czf ${src_name} apache-servicecomb-kie-${VERSION}
rm -rf "v${VERSION}.tar.gz" apache-servicecomb-kie-${VERSION}

gpg --armor --output "$src_name.asc" --detach-sig "${src_name}"

sha512sum  "${src_name}" > "${src_name}".sha512