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
	"strings"
	"testing"
	"time"

	. "github.com/camel-tooling/camel-dashboard-operator/e2e/support"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
)

func TestVerifyDeployment(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		t.Run("simple Deployment", func(t *testing.T) {
			ExpectExecSucceed(t, g,
				exec.Command(
					"kubectl",
					strings.Split("create deployment camel-app-main --image=docker.io/squakez/db-app-main:1.0 -n "+ns, " ")...,
				),
			)
			g.Eventually(PodStatusPhase(t, ctx, ns, "app=camel-app-main"), TestTimeoutMedium).Should(Equal(corev1.PodRunning))
			// As there is no label, there is not yet any CamelApp CR
			g.Consistently(CamelApps(t, ctx, ns), TestTimeoutShort, 10*time.Second).Should(BeEmpty())

			// Add the labels to discover it
			ExpectExecSucceed(t, g,
				exec.Command(
					"kubectl",
					strings.Split("label deployment camel-app-main camel.apache.org/app=camel-sample -n "+ns, " ")...,
				),
			)
			// The name of the selector, "camel.apache.org/app: camel-sample"
			g.Eventually(CamelApp(t, ctx, ns, "camel-sample")).Should(Not(BeNil()))
			g.Eventually(
				CamelAppStatus(t, ctx, ns, "camel-sample"),
				TestTimeoutMedium,
			).Should(
				MatchFields(IgnoreExtras, Fields{
					"Phase":    Equal(v1alpha1.CamelAppPhaseRunning),
					"Replicas": PointTo(Equal(int32(1))),
				}),
			)
		})
	})
}
