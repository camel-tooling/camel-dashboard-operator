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

location=$(dirname $0)

GO111MODULE=on

# Entering the client module
cd $location/../pkg/client/

rm -rf camel
rm -rf knative

echo "Generating Go client code..."

$(go env GOPATH)/bin/applyconfiguration-gen \
	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1" \
	--go-header-file=../../script/headers/default.txt \
	--output-dir=./camel/applyconfiguration/ \
	--output-pkg=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/applyconfiguration

$(go env GOPATH)/bin/client-gen \
	--input camel/v1alpha1 \
	--go-header-file=../../script/headers/default.txt \
	--clientset-name "versioned"  \
	--input-base=github.com/camel-tooling/camel-dashboard-operator/pkg/apis \
	--apply-configuration-package=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/applyconfiguration \
	--output-dir=./camel/clientset/ \
	--output-pkg=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/clientset

$(go env GOPATH)/bin/client-gen \
	--input knative/v1 \
	--go-header-file=../../script/headers/default.txt \
	--input-base=github.com/camel-tooling/camel-dashboard-operator/pkg/apis/duck \
	--output-dir=./knative/clientset/ \
	--output-pkg=github.com/camel-tooling/camel-dashboard-operator/pkg/client/knative/clientset

$(go env GOPATH)/bin/lister-gen \
	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1" \
	--go-header-file=../../script/headers/default.txt \
	--output-dir=./camel/listers/ \
	--output-pkg=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/listers

$(go env GOPATH)/bin/informer-gen \
    "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1" \
	--versioned-clientset-package=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/clientset/versioned \
	--listers-package=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/listers \
	--go-header-file=../../script/headers/default.txt \
	--output-dir=./camel/informers/ \
	--output-pkg=github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/informers
