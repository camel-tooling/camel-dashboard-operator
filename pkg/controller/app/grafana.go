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
	"fmt"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	integreatlyv1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func grafanaCRDExists(ctx context.Context, c client.Client) (bool, error) {
	_, err := c.Discovery().ServerResourcesForGroupVersion("grafana.integreatly.org/v1beta1")
	if err != nil && k8serrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

// addGrafanaDashboard will include a GrafanaDashboard resource bound to the CamelApp resource.
func addGrafanaDashboard(ctx context.Context, c client.Client, target *v1alpha1.CamelApp) error {
	// Verify the existence of the Prometheus metrics endpoint
	if len(target.Status.Pods) > 0 &&
		target.Status.Pods[0].ObservabilityService != nil &&
		target.Status.Pods[0].ObservabilityService.MetricsEndpoint != "" &&
		target.Status.Pods[0].ObservabilityService.MetricsPort != 0 {
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
		dashboard := &integreatlyv1beta1.GrafanaDashboard{
			TypeMeta: metav1.TypeMeta{
				Kind:       "GrafanaDashboard",
				APIVersion: integreatlyv1beta1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:            target.GetName(),
				Namespace:       target.GetNamespace(),
				OwnerReferences: references,
			},
			Spec: integreatlyv1beta1.GrafanaDashboardSpec{
				AllowCrossNamespaceImport: ptr.To(true),
				FolderTitle:               "camel-dashboard",
				InstanceSelector:          &metav1.LabelSelector{MatchLabels: platform.GetGrafanaLabels()},
				Json:                      buildGrafanaDashboardJSON(target),
			},
		}

		err := replaceGrafanaDashboard(ctx, c, dashboard)
		addCamelAppGrafanaCondition(target, err)

		return err
	}

	return nil
}

func replaceGrafanaDashboard(ctx context.Context, c client.Client, dashboard *integreatlyv1beta1.GrafanaDashboard) error {
	existing := &integreatlyv1beta1.GrafanaDashboard{}
	err := c.Get(ctx, ctrl.ObjectKey{
		Name:      dashboard.Name,
		Namespace: dashboard.Namespace,
	}, existing)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return c.Create(ctx, dashboard)
		}
		return err
	}
	dashboard.ResourceVersion = existing.ResourceVersion

	return c.Update(ctx, dashboard)
}

func addCamelAppGrafanaCondition(target *v1alpha1.CamelApp, err error) {
	statusCond := metav1.ConditionTrue
	message := "Created a GrafanaDashboard with the same name of this CamelApp"
	if err != nil {
		statusCond = metav1.ConditionFalse
		message = "Some error happened while creating GrafanaDashboard: " + err.Error()
	}
	target.Status.AddCondition(metav1.Condition{
		Type:               "GrafanaDashboard",
		Status:             statusCond,
		LastTransitionTime: metav1.NewTime(time.Now()),
		Reason:             "GrafanaDashboardAdded",
		Message:            message,
	})
}

// buildGrafanaDashboardJSON is in charge to generate a JSON configuration of the dashboard.
func buildGrafanaDashboardJSON(target *v1alpha1.CamelApp) string {
	return fmt.Sprintf(`{
	  "title": "Camel exchange metrics: %s",
      "panels": [
        {
          "datasource": "prometheus",
          "type": "timeseries",
          "title": "Camel total exchanges per second",
          "targets": [
            {
              "expr": "sum(rate(camel_exchanges_total{job=\"%s/%s\", eventType=\"route\"}[5m]))"
            }
          ]
        }
      ],
      "schemaVersion": 36,
      "version": 1
    }`, target.GetName(), target.GetNamespace(), target.GetName())
}
