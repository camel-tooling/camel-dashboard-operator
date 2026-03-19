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
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// addPrometheusPodMonitor will include a Prometeus PodMonitor resource bound to the CamelApp resource.
func addPrometheusPodMonitor(ctx context.Context, c client.Client, target *v1alpha1.CamelApp,
	matchLabelSelector map[string]string) error {
	// Verify the existence of the Prometheus metrics endpoint
	if len(target.Status.Pods) > 0 &&
		target.Status.Pods[0].ObservabilityService != nil &&
		target.Status.Pods[0].ObservabilityService.MetricsEndpoint != "" &&
		target.Status.Pods[0].ObservabilityService.MetricsPort != 0 {
		// We assume all Pods expose the same port and metrics endpoint configuration
		metricsEndpoint := target.Status.Pods[0].ObservabilityService.MetricsEndpoint
		metricsPortNumber := target.Status.Pods[0].ObservabilityService.MetricsPort
		// We must set the ownership in order to get garbage collection for free
		references := []metav1.OwnerReference{
			{
				APIVersion:         target.APIVersion,
				Kind:               target.Kind,
				Name:               target.Name,
				UID:                target.UID,
				Controller:         ptr.To(true),
				BlockOwnerDeletion: ptr.To(true),
			},
		}
		podMonitor := monitoringv1.PodMonitor{
			TypeMeta: metav1.TypeMeta{
				Kind:       "PodMonitor",
				APIVersion: monitoringv1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            target.GetName(),
				Namespace:       target.GetNamespace(),
				OwnerReferences: references,
				Labels:          platform.GetPrometheusLabels(),
			},
			Spec: monitoringv1.PodMonitorSpec{
				Selector: metav1.LabelSelector{
					MatchLabels: matchLabelSelector,
				},
				PodMetricsEndpoints: []monitoringv1.PodMetricsEndpoint{
					{
						PortNumber: ptr.To(int32(metricsPortNumber)),
						Path:       metricsEndpoint,
					},
				},
			},
		}

		err := replacePodMonitor(ctx, c, &podMonitor)
		addCamelAppPrometheusCondition(target, err)

		return err
	}

	return nil
}

func addCamelAppPrometheusCondition(target *v1alpha1.CamelApp, err error) {
	statusCond := metav1.ConditionTrue
	message := "Created a PodMonitor with the same name of this CamelApp"
	if err != nil {
		statusCond = metav1.ConditionFalse
		message = "Some error happened while creating PodMonitor: " + err.Error()
	}
	target.Status.AddCondition(metav1.Condition{
		Type:               "PrometheusPodMonitor",
		Status:             statusCond,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             "PodMonitorAdded",
		Message:            message,
	})
}

func replacePodMonitor(ctx context.Context, c client.Client, pm *monitoringv1.PodMonitor) error {
	existing := &monitoringv1.PodMonitor{}
	err := c.Get(ctx, ctrl.ObjectKey{
		Name:      pm.Name,
		Namespace: pm.Namespace,
	}, existing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return c.Create(ctx, pm)
		}
		return err
	}
	pm.ResourceVersion = existing.ResourceVersion

	return c.Update(ctx, pm)
}

func prometheusCRDExists(ctx context.Context, c client.Client) (bool, error) {
	_, err := c.Discovery().ServerResourcesForGroupVersion("monitoring.coreos.com/v1")
	if err != nil && k8serrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}
