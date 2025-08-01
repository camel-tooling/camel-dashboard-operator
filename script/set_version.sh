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

set -e

if [ "$#" -lt 1 ] || [ "$#" -gt 2 ]; then
    echo "usage: $0 version [image_name]"
    exit 1
fi

location=$(dirname $0)
version=$1
image_name=${2:-quay.io\/camel-tooling\/camel-dashboard-operator}
sanitized_image_name=${image_name//\//\\\/}
k8s_version_label="app.kubernetes.io\/version"

for f in $(find $location/../pkg/resources/config/manager -type f -name "*.yaml");
do
  sed -i -r "s/image: .*/image: ${sanitized_image_name}:${version}/" $f
  sed -i -r "s/${k8s_version_label}: .*/${k8s_version_label}: \""${version}"\"/" $f
done

echo "Camel Dashboard version set to: $version and image name to: $image_name"
