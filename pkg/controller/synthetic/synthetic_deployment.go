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

package synthetic

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/kubernetes"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"

	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"
)

// nonManagedCamelDeployment represents a regular Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelDeployment struct {
	deploy *appsv1.Deployment
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelDeployment) CamelApp(ctx context.Context, c client.Client) *v1alpha1.CamelApp {
	newApp := v1alpha1.NewApp(app.deploy.Namespace, app.deploy.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.deploy.Name,
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	newApp.ImportCamelAnnotations(app.deploy.Annotations)
	references := []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       app.deploy.Name,
			UID:        app.deploy.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)

	return &newApp
}

// GetAppPhase returns the phase of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppPhase() v1alpha1.CamelAppPhase {
	if app.deploy.Status.AvailableReplicas == app.deploy.Status.Replicas {
		if app.deploy.Status.Replicas == 0 {
			return v1alpha1.CamelAppPhasePaused
		}
		return v1alpha1.CamelAppPhaseRunning
	}

	return v1alpha1.CamelAppPhaseError
}

// GetAppImage returns the container image of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppImage() string {
	return app.deploy.Spec.Template.Spec.Containers[0].Image
}

// GetReplicas returns the number of desired replicas for the backing Camel application.
func (app *nonManagedCamelDeployment) GetReplicas() *int32 {
	return app.deploy.Spec.Replicas
}

// GetAnnotations returns the backing deployment object annotations.
func (app *nonManagedCamelDeployment) GetAnnotations() map[string]string {
	return app.deploy.Annotations
}

// GetPods returns the pods backing the Camel application.
func (app *nonManagedCamelDeployment) GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error) {
	var podsInfo []v1alpha1.PodInfo
	pods := &corev1.PodList{}
	err := c.List(ctx, pods,
		ctrl.InNamespace(app.deploy.GetNamespace()),
		ctrl.MatchingLabels(app.deploy.Spec.Selector.MatchLabels),
	)
	if err != nil {
		return nil, err
	}
	observabilityPort := app.getObservabilityPort()
	for _, pod := range pods.Items {
		podIp := pod.Status.PodIP
		podInfo := v1alpha1.PodInfo{
			Name:       pod.GetName(),
			Status:     string(pod.Status.Phase),
			InternalIP: podIp,
		}

		// Check the services only if the Pod is ready
		if ready := kubernetes.GetPodCondition(pod, corev1.PodReady); ready != nil && ready.Status == corev1.ConditionTrue {
			podInfo.UptimeTimestamp = &metav1.Time{Time: ready.LastTransitionTime.Time}
			ready := true
			podInfo.ObservabilityService = &v1alpha1.ObservabilityServiceInfo{}
			if err := setHealth(&podInfo, podIp, observabilityPort); err != nil {
				ready = false
				reason := fmt.Sprintf("Could not scrape health endpoint: %s", err.Error())
				log.Infof("Deployment %s/%s: %s", app.deploy.GetNamespace(), app.deploy.GetName(), reason)
				podInfo.Reason = reason
			}
			if err := setMetrics(&podInfo, podIp, observabilityPort); err != nil {
				ready = false
				reason := fmt.Sprintf("Could not scrape metrics endpoint: %s", err.Error())
				log.Infof("Deployment %s/%s: %s", app.deploy.GetNamespace(), app.deploy.GetName(), reason)
				if podInfo.Reason != "" {
					podInfo.Reason += ". "
				}
				podInfo.Reason += reason
			}
			podInfo.Ready = ready
		}

		podsInfo = append(podsInfo, podInfo)
	}

	return podsInfo, nil
}

func setMetrics(podInfo *v1alpha1.PodInfo, podIp string, port int) error {
	// NOTE: we're not using a proxy as a design choice in order
	// to have a faster turnaround.
	req, err := http.NewRequest("GET", fmt.Sprintf("http://%s:%d/%s", podIp, port, platform.DefaultObservabilityMetrics), nil)
	if err != nil {
		return err
	}
	// Quarkus runtime specific, see https://github.com/apache/camel-quarkus/issues/7405
	req.Header.Add("Accept", "text/plain, */*")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		podInfo.ObservabilityService.MetricsEndpoint = platform.DefaultObservabilityMetrics
		podInfo.ObservabilityService.MetricsPort = port

		if podInfo.Runtime == nil {
			podInfo.Runtime = &v1alpha1.RuntimeInfo{}
		}
		if podInfo.Runtime.Exchange == nil {
			podInfo.Runtime.Exchange = &v1alpha1.ExchangeInfo{}
		}

		metrics, err := parseMetrics(resp.Body)
		if err != nil {
			return err
		}
		if metric, ok := metrics["app_info"]; ok {
			populateRuntimeInfo(metric, "app_info", podInfo)
		}
		if metric, ok := metrics["camel_exchanges_last_timestamp"]; ok {
			populateExchangesLastTimestamp(metric, "camel_exchanges_last_timestamp", podInfo)
		}
		if metric, ok := metrics["camel_exchanges_total"]; ok {
			populateExchangesTotal(metric, "camel_exchanges_total", podInfo)
		}
		if metric, ok := metrics["camel_exchanges_failed_total"]; ok {
			populateExchangesFailedTotal(metric, "camel_exchanges_failed_total", podInfo)
		}
		if metric, ok := metrics["camel_exchanges_succeeded_total"]; ok {
			populateExchangesSucceededTotal(metric, "camel_exchanges_succeeded_total", podInfo)
		}
		if metric, ok := metrics["camel_exchanges_inflight"]; ok {
			populateExchangesInflight(metric, "camel_exchanges_inflight", podInfo)
		}

		return nil
	}

	return fmt.Errorf("HTTP status not OK, it was %d", resp.StatusCode)
}

func parseMetrics(reader io.Reader) (map[string]*dto.MetricFamily, error) {
	var parser expfmt.TextParser
	mf, err := parser.TextToMetricFamilies(reader)
	if err != nil {
		return nil, err
	}

	return mf, nil
}

func populateRuntimeInfo(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) != 1 {
		log.Infof("WARN: expected exactly one %s metric, got %d", metricName, len(metric.GetMetric()))
		return
	}

	for _, label := range metric.GetMetric()[0].GetLabel() {
		switch ptr.Deref(label.Name, "") {
		case "camel_runtime_provider":
			podInfo.Runtime.RuntimeProvider = ptr.Deref(label.Value, "")
		case "camel_runtime_version":
			podInfo.Runtime.RuntimeVersion = ptr.Deref(label.Value, "")
		case "camel_version":
			podInfo.Runtime.CamelVersion = ptr.Deref(label.Value, "")
		}
	}
}

func populateExchangesTotal(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) == 0 {
		log.Infof("WARN: expected at least 1 %s metric, got %d", metricName, len(metric.GetMetric()))
		return
	}
	if metric.GetMetric()[0].GetCounter() == nil {
		log.Infof("WARN: expected %s metric to be a counter", metricName)
		return
	}

	podInfo.Runtime.Exchange.Total = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateExchangesFailedTotal(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) == 0 {
		log.Infof("WARN: expected at least 1 %s metric, got %d", metricName, len(metric.GetMetric()))
		return
	}
	if metric.GetMetric()[0].GetCounter() == nil {
		log.Infof("WARN: expected %s metric to be a counter", metricName)
		return
	}

	podInfo.Runtime.Exchange.Failed = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateExchangesSucceededTotal(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) == 0 {
		log.Infof("WARN: expected at least 1 %s metric, got %d", metricName, len(metric.GetMetric()))
		return
	}
	if metric.GetMetric()[0].GetCounter() == nil {
		log.Infof("WARN: expected %s metric to be a counter", metricName)
		return
	}

	podInfo.Runtime.Exchange.Succeeded = int(ptr.Deref(metric.GetMetric()[0].GetCounter().Value, 0))
}

func populateExchangesInflight(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) == 0 {
		log.Infof("WARN: expected at least 1 %s metric, got %d", metricName, len(metric.GetMetric()))
		return
	}
	if metric.GetMetric()[0].GetGauge() == nil {
		log.Infof("WARN: expected %s metric to be a gauge", metricName)
		return
	}

	podInfo.Runtime.Exchange.Pending = int(ptr.Deref(metric.GetMetric()[0].GetGauge().Value, 0))
}

func populateExchangesLastTimestamp(metric *dto.MetricFamily, metricName string, podInfo *v1alpha1.PodInfo) {
	if len(metric.GetMetric()) == 0 {
		log.Debugf("expected at least 1 exchanges_last_timestamp metric, got %d", len(metric.GetMetric()))
		return
	}
	if metric.GetMetric()[0].GetGauge() == nil {
		log.Debugf("expected %s metric to be a gauge", metricName)
		return
	}

	lastExchangeTimestamp := int64(ptr.Deref(metric.GetMetric()[0].GetGauge().Value, 0))
	if lastExchangeTimestamp == 0 {
		return
	}
	timeUnixMilli := time.UnixMilli(lastExchangeTimestamp)
	podInfo.Runtime.Exchange.LastTimestamp = &metav1.Time{Time: timeUnixMilli}
}

func setHealth(podInfo *v1alpha1.PodInfo, podIp string, port int) error {
	// NOTE: we're not using a proxy as a design choice in order
	// to have a faster turnaround.
	resp, err := http.Get(fmt.Sprintf("http://%s:%d/%s", podIp, port, platform.DefaultObservabilityHealth))
	if err != nil {
		return err
	}
	status := "Unknown"
	defer resp.Body.Close()
	// The endpoint reports 503 when the service is down, but still provide the
	// health information
	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable {
		podInfo.ObservabilityService.HealthPort = port
		podInfo.ObservabilityService.HealthEndpoint = platform.DefaultObservabilityHealth

		status, err = parseHealthStatus(resp.Body)
		if err != nil {
			return err
		}
	}
	if podInfo.Runtime == nil {
		podInfo.Runtime = &v1alpha1.RuntimeInfo{}
	}
	podInfo.Runtime.Status = status

	return nil
}

func parseHealthStatus(reader io.Reader) (string, error) {
	var healthContent map[string]any
	err := json.NewDecoder(reader).Decode(&healthContent)
	if err != nil {
		return "", err
	}
	status, ok := healthContent["status"].(string)
	if !ok {
		return "", errors.New("health endpoint syntax error: missing .status property")
	}

	return string(status), nil
}

func (app *nonManagedCamelDeployment) getObservabilityPort() int {
	defaultPort := platform.GetObservabilityPort()
	if app.GetAnnotations() == nil || app.GetAnnotations()[v1alpha1.AppObservabilityServicesPort] == "" {
		return defaultPort
	}

	port, err := strconv.Atoi(app.GetAnnotations()[v1alpha1.AppObservabilityServicesPort])
	if err == nil {
		return port
	} else {
		log.Error(err, "could not properly parse application observability services port configuration, "+
			"fallback to default operator value %d", defaultPort)
	}

	return defaultPort
}
