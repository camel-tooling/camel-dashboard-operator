//go:build integration
// +build integration

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

package common

import (
	"context"
	"os/exec"
	"testing"

	. "github.com/camel-tooling/camel-dashboard-operator/e2e/support"
	. "github.com/onsi/gomega"
)

func TestVerifyDeployment(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		// Test a simple route which sends 5 messages to log
		t.Run("simple route", func(t *testing.T) {
			ExpectExecSucceed(t, g,
				exec.Command(
					"kubectl",
					"apply",
					"-f",
					"files/sample-it.yaml",
					"-n",
					ns,
				),
			)
			// The name of the selector, "camel.apache.org/app: camel-sample"
			g.Eventually(Pod(t, ctx, ns, "camel-sample"), TestTimeoutMedium).Should(Not(BeNil()))
			g.Eventually(CamelApp(t, ctx, ns, "camel-sample")).Should(Not(BeNil()))
		})
	})
}
