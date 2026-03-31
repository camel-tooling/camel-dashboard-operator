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
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// addPrometheusRule will include a PrometeusRule resource setting certain alerts.
func addPrometheusRuleAlerts(ctx context.Context, c client.Client, target *v1alpha1.CamelApp) error {
	if target.Status.DoesExposeMetrics() {
		references := target.GetOwnerReferences()
		prometheusRule := monitoringv1.PrometheusRule{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PrometheusRule",
				APIVersion: monitoringv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            target.GetName(),
				Namespace:       target.GetNamespace(),
				OwnerReferences: references,
				Labels:          platform.GetPrometheusRuleLabels(),
			},
			Spec: monitoringv1.PrometheusRuleSpec{
				Groups: []monitoringv1.RuleGroup{
					{
						Name:  "camel-exchanges-failure-rate",
						Rules: getExchangeFailureAlerts(),
					},
				},
			},
		}

		err := replacePrometheusRule(ctx, c, &prometheusRule)
		addCamelAppPrometheusRuleCondition(target, err)

		return err
	}

	return nil
}

func getExchangeFailureAlerts() []monitoringv1.Rule {
	return []monitoringv1.Rule{getCamelHighFailureRateCritical()}
}

// TODO: provide proper parameters and make it generic.
func getCamelHighFailureRateCritical() monitoringv1.Rule {
	return monitoringv1.Rule{
		Alert: "CamelHighFailureRateCritical",
		Expr: intstr.FromString(`sum by(job) (increase(camel_exchanges_failed_total[5m]))
            /
            clamp_min(sum by(job) (increase(camel_exchanges_total[5m])), 1)
            > 0.10`),
		For: ptr.To(monitoringv1.Duration("2m")),
		Labels: map[string]string{
			"severity": "critical",
		},
		Annotations: map[string]string{
			"summary":     "camel exchange failed total rate > 10%",
			"description": "Job {{ $labels.job }} has a failure rate above 10% in the last 5 minutes.",
		},
	}
}

func addCamelAppPrometheusRuleCondition(target *v1alpha1.CamelApp, err error) {
	statusCond := metav1.ConditionTrue
	message := "Created a PrometheusRule with the same name of this CamelApp"
	if err != nil {
		statusCond = metav1.ConditionFalse
		message = "Some error happened while creating PrometheusRule: " + err.Error()
	}
	target.Status.AddCondition(metav1.Condition{
		Type:               "PrometheusRule",
		Status:             statusCond,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             "PrometheusRuleAdded",
		Message:            message,
	})
}

func replacePrometheusRule(ctx context.Context, c client.Client, pr *monitoringv1.PrometheusRule) error {
	existing := &monitoringv1.PrometheusRule{}
	err := c.Get(ctx, ctrl.ObjectKey{
		Name:      pr.Name,
		Namespace: pr.Namespace,
	}, existing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return c.Create(ctx, pr)
		}
		return err
	}
	pr.ResourceVersion = existing.ResourceVersion

	return c.Update(ctx, pr)
}
