/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package app

import (
	"context"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
)

// Action --.
type Action interface {
	client.Injectable
	log.Injectable

	// a user friendly name for the action
	Name() string

	// returns true if the action can handle the Pipe
	CanHandle(app *v1alpha1.CamelApp) bool

	// executes the handling function
	Handle(ctx context.Context, kamelet *v1alpha1.CamelApp) (*v1alpha1.CamelApp, error)
}

type baseAction struct {
	client client.Client
	L      log.Logger
}

func (action *baseAction) InjectClient(client client.Client) {
	action.client = client
}

func (action *baseAction) InjectLogger(log log.Logger) {
	action.L = log
}
