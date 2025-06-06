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

location=$(dirname $0)
rootdir=$(realpath $location/..)
version=$1
targetdir=$rootdir/docs/charts/$version
helm_index=$rootdir/docs/charts/index.yaml

mkdir -p $targetdir

cd $rootdir/helm

helm package ./camel-dashboard --version $version
mv camel-dashboard-*.tgz $targetdir/
# TODO: create https://github.com/camel-tooling/camel-dashboard-charts similar to https://github.com/hawtio/hawtio-charts
helm repo index $targetdir --url https://camel-tooling.github.io/camel-dashboard-charts/ --merge $helm_index
# Required to prevent https://github.com/helm/helm/issues/7363
mv $targetdir/* $targetdir/../.
rm -rf $targetdir
