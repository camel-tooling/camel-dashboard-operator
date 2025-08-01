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

location=$(dirname "$0")
apidir=$location/../pkg/apis/camel

cd "$apidir"
$(go env GOPATH)/bin/controller-gen crd \
  paths=./... \
  output:crd:artifacts:config=../../../pkg/resources/config/crd/bases \
  output:crd:dir=../../../pkg/resources/config/crd/bases

# cleanup working directory in $apidir
rm -rf ./config

# to root
cd ../../../

# Importing helm CRDs
cat ./script/headers/yaml.txt > ./helm/camel-dashboard/crds/camel-dashboard-crds.yaml
kustomize build ./pkg/resources/config/crd/. >> ./helm/camel-dashboard/crds/camel-dashboard-crds.yaml

deploy_crd_file() {
  source=$1

  # Make a copy to serve as the base for post-processing
  cp "$source" "${source}.orig"

  # Post-process source
  cat ./script/headers/yaml.txt > "$source"
  echo "" >> "$source"
  cat ${source}.orig >> "$source"

  for dest in "${@:2}"; do
    cp "$source" "$dest"
  done

  # Remove the copy as no longer required
  rm -f "${source}.orig"
}

deploy_crd() {
  name=$1
  plural=$2

  deploy_crd_file ./pkg/resources/config/crd/bases/camel.apache.org_"$plural".yaml
}

deploy_crd camelapp camelapps
