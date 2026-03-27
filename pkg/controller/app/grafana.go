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
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/controller/synthetic"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	integreatlyv1beta1 "github.com/grafana-operator/grafana-operator/v5/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
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
func addGrafanaDashboard(ctx context.Context, c client.Client, target *v1alpha1.CamelApp, app synthetic.NonManagedCamelApplicationAdapter) error {
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
		dashboardJson, err := buildGrafanaDashboardJSON(target, app)
		if err != nil {
			return err
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
				Json:                      dashboardJson,
			},
		}

		err = replaceGrafanaDashboard(ctx, c, dashboard)
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
func buildGrafanaDashboardJSON(target *v1alpha1.CamelApp, app synthetic.NonManagedCamelApplicationAdapter) (string, error) {
	dashboard := v1alpha1.Dashboard{
		Title: "Camel exchange metrics: " + target.GetName(),
		Panels: []v1alpha1.Panel{
			getTimeSeriesPanel(v1alpha1.Metric_camel_exchanges_total, target.GetNamespace(), target.GetName(), "route", "5m"),
			getTimeSeriesPanel(v1alpha1.Metric_camel_exchanges_failed_total, target.GetNamespace(), target.GetName(), "route", "5m"),
			getLastExchangeGaugePanel(target.GetNamespace(), target.GetName()),
			getCPUUsagePanel(target.GetNamespace(), target.GetName(), float64(app.GetResourcesLimitSize(corev1.ResourceCPU))),
			getJVMMemoryUsagePanel(target.GetNamespace(), target.GetName(), float64(app.GetResourcesLimitSize(corev1.ResourceMemory))),
		},
		SchemaVersion: 36,
		Version:       1,
	}

	bytes, err := json.Marshal(dashboard)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func getTimeSeriesPanel(metric, jobNamespace, jobName, eventType, sample string) v1alpha1.Panel {
	panelTitle, panelExpression := getRateExpression(metric, jobNamespace, jobName, eventType, sample)
	return v1alpha1.Panel{
		Datasource: platform.GetGrafanaDatasource(),
		Type:       "timeseries",
		Title:      panelTitle,
		Targets: []v1alpha1.Target{
			{
				Expr: panelExpression,
			},
		},
	}
}

func getLastExchangeGaugePanel(jobNamespace, jobName string) v1alpha1.Panel {
	return v1alpha1.Panel{
		Datasource: platform.GetGrafanaDatasource(),
		Type:       "gauge",
		Title:      "Last exchange delay (in seconds)",
		Targets: []v1alpha1.Target{
			{
				Expr: fmt.Sprintf(`time() - (%s{job="%s/%s"} / 1000)`,
					v1alpha1.Metric_camel_exchanges_last_timestamp, jobNamespace, jobName),
			},
		},
		FieldConfig: v1alpha1.FieldConfig{
			Defaults: v1alpha1.FieldDefaults{
				Unit: "seconds",
				Min:  0,
				// TODO parametrize
				Max: 60,
				Thresholds: &v1alpha1.Thresholds{
					Mode: "absolute",
					Steps: []v1alpha1.ThresholdStep{
						{Color: "green", Value: nil},
						{Color: "yellow", Value: ptr.To(float64(50))},
						{Color: "red", Value: ptr.To(float64(55))},
					},
				},
			},
		},
	}
}

func getCPUUsagePanel(jobNamespace, jobName string, maxValue float64) v1alpha1.Panel {
	panel := v1alpha1.Panel{
		Datasource: platform.GetGrafanaDatasource(),
		Type:       "timeseries",
		Title:      "CPU usage (in core)",
		Targets: []v1alpha1.Target{
			{
				Expr: fmt.Sprintf(`avg(system_cpu_usage{job="%s/%s"})`, jobNamespace, jobName),
			},
		},
	}
	if maxValue > 0 {
		panel.FieldConfig = getFieldConfigWithThresholds(maxValue, "core")
	}

	return panel
}

func getFieldConfigWithThresholds(maxValue float64, unit string) v1alpha1.FieldConfig {
	warnThreshold := maxValue * .8
	errThreshold := maxValue * .9
	return v1alpha1.FieldConfig{
		Defaults: v1alpha1.FieldDefaults{
			Unit: unit,
			Min:  0,
			Max:  maxValue,
			Thresholds: &v1alpha1.Thresholds{
				Mode: "absolute",
				Steps: []v1alpha1.ThresholdStep{
					{Color: "green", Value: nil},
					{Color: "yellow", Value: ptr.To(warnThreshold)},
					{Color: "red", Value: ptr.To(errThreshold)},
				},
			},
			Custom: &v1alpha1.CustomOptions{
				ThresholdsStyle: &v1alpha1.ThresholdsStyle{
					Mode: "dashed+area",
				},
			},
		},
	}
}

func getJVMMemoryUsagePanel(jobNamespace, jobName string, maxValue float64) v1alpha1.Panel {
	panel := v1alpha1.Panel{
		Datasource: platform.GetGrafanaDatasource(),
		Type:       "timeseries",
		Title:      "JVM Heap memory (in Mi)",
		Targets: []v1alpha1.Target{
			{
				Expr: fmt.Sprintf(`avg(jvm_memory_used_bytes{area="heap", job="%s/%s"} / 1024 / 1024)`, jobNamespace, jobName),
			},
		},
	}

	if maxValue > 0 {
		panel.FieldConfig = getFieldConfigWithThresholds(maxValue, "Mi")
	}

	return panel
}

// getRateExpression return an expression with the format expected for a rate count.
func getRateExpression(metric, jobNamespace, jobName, eventType, sample string) (string, string) {
	metricTitle := strings.ReplaceAll(metric, "_", " ") + " per second"
	return metricTitle, fmt.Sprintf("sum(rate(%s{job=\"%s/%s\", eventType=\"%s\"}[%s]))", metric, jobNamespace, jobName, eventType, sample)
}
