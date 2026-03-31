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
	"testing"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/internal"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func TestAddPrometheusRule_Success(t *testing.T) {
	target := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
			UID:       "12345",
		},
		Status: v1alpha1.CamelAppStatus{
			Pods: []v1alpha1.PodInfo{
				{
					ObservabilityService: &v1alpha1.ObservabilityServiceInfo{
						MetricsEndpoint: "/metrics",
						MetricsPort:     8080,
					},
				},
			},
		},
	}

	fakeClient, err := internal.NewFakeClient()
	require.NoError(t, err)

	err = addPrometheusRuleAlerts(context.TODO(), fakeClient, target)
	require.NoError(t, err)

	pr := &monitoringv1.PrometheusRule{}
	err = fakeClient.Get(context.TODO(), ctrl.ObjectKey{
		Name:      "test-app",
		Namespace: "default",
	}, pr)

	require.NoError(t, err)
	assert.Len(t, pr.Spec.Groups, 1)
	assert.Len(t, pr.Spec.Groups[0].Rules, 1)
	assert.Equal(t, "CamelHighFailureRateCritical", pr.Spec.Groups[0].Rules[0].Alert)
	assert.Contains(t, pr.Spec.Groups[0].Rules[0].Expr.StrVal, "camel_exchanges_failed_total")
	assert.Contains(t, pr.Spec.Groups[0].Rules[0].Expr.StrVal, "camel_exchanges_total")
}

func TestAddPrometheusRule_NoPods(t *testing.T) {
	target := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Status: v1alpha1.CamelAppStatus{
			Pods: []v1alpha1.PodInfo{},
		},
	}

	fakeClient, err := internal.NewFakeClient()
	require.NoError(t, err)

	err = addPrometheusRuleAlerts(context.TODO(), fakeClient, target)
	require.NoError(t, err)

	pr := &monitoringv1.PrometheusRule{}
	err = fakeClient.Get(context.TODO(), ctrl.ObjectKey{
		Name:      "test-app",
		Namespace: "default",
	}, pr)

	require.Error(t, err)
	assert.Equal(t, "prometheusrules.monitoring.coreos.com \"test-app\" not found", err.Error())
}

func TestAddPrometheusPrometheusRule_NoObservabilityServices(t *testing.T) {
	target := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Status: v1alpha1.CamelAppStatus{
			Pods: []v1alpha1.PodInfo{
				{},
			},
		},
	}

	fakeClient, err := internal.NewFakeClient()
	require.NoError(t, err)

	err = addPrometheusRuleAlerts(context.TODO(), fakeClient, target)
	require.NoError(t, err)

	pr := &monitoringv1.PrometheusRule{}
	err = fakeClient.Get(context.TODO(), ctrl.ObjectKey{
		Name:      "test-app",
		Namespace: "default",
	}, pr)

	require.Error(t, err)
	assert.Equal(t, "prometheusrules.monitoring.coreos.com \"test-app\" not found", err.Error())
}

func TestAddPrometheusPrometheusRule_NoMetrics(t *testing.T) {
	target := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
		Status: v1alpha1.CamelAppStatus{
			Pods: []v1alpha1.PodInfo{
				{ObservabilityService: &v1alpha1.ObservabilityServiceInfo{HealthEndpoint: "/health"}},
			},
		},
	}

	fakeClient, err := internal.NewFakeClient()
	require.NoError(t, err)

	err = addPrometheusRuleAlerts(context.TODO(), fakeClient, target)
	require.NoError(t, err)

	pr := &monitoringv1.PrometheusRule{}
	err = fakeClient.Get(context.TODO(), ctrl.ObjectKey{
		Name:      "test-app",
		Namespace: "default",
	}, pr)

	require.Error(t, err)
	assert.Equal(t, "prometheusrules.monitoring.coreos.com \"test-app\" not found", err.Error())
}

func TestAddPrometheusRule_UpdateExisting(t *testing.T) {
	target := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
			UID:       "12345",
		},
		Status: v1alpha1.CamelAppStatus{
			Pods: []v1alpha1.PodInfo{
				{
					ObservabilityService: &v1alpha1.ObservabilityServiceInfo{
						MetricsEndpoint: "/metrics-new",
						MetricsPort:     9090,
					},
				},
			},
		},
	}

	// Pre-existing PrometheusRule
	existing := &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
	}

	fakeClient, err := internal.NewFakeClient(existing)
	require.NoError(t, err)

	err = addPrometheusRuleAlerts(context.TODO(), fakeClient, target)
	require.NoError(t, err)

	pr := &monitoringv1.PrometheusRule{}
	err = fakeClient.Get(context.TODO(), ctrl.ObjectKey{
		Name:      "test-app",
		Namespace: "default",
	}, pr)

	require.NoError(t, err)
	assert.Len(t, pr.Spec.Groups, 1)
	assert.Len(t, pr.Spec.Groups[0].Rules, 1)
	assert.Equal(t, "CamelHighFailureRateCritical", pr.Spec.Groups[0].Rules[0].Alert)
	assert.Contains(t, pr.Spec.Groups[0].Rules[0].Expr.StrVal, "camel_exchanges_failed_total")
	assert.Contains(t, pr.Spec.Groups[0].Rules[0].Expr.StrVal, "camel_exchanges_total")
}
