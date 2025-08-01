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
	"strconv"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/controller/synthetic"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewMonitorAction returns an action that monitors the App.
func NewMonitorAction() Action {
	return &monitorAction{}
}

type monitorAction struct {
	baseAction
}

func (action *monitorAction) Name() string {
	return "monitor"
}

func (action *monitorAction) CanHandle(app *v1alpha1.CamelApp) bool {
	return true
}

func (action *monitorAction) Handle(ctx context.Context, app *v1alpha1.CamelApp) (*v1alpha1.CamelApp, error) {
	action.L.Infof("Monitoring App %s/%s with status %s", app.Namespace, app.Name, app.Status.Phase)
	objOwner, err := lookupObject(ctx, action.client,
		app.Annotations[v1alpha1.AppImportedKindLabel], app.Namespace, app.Annotations[v1alpha1.AppImportedNameLabel])
	if err != nil {
		return nil, err
	}
	if objOwner == nil {
		return nil, fmt.Errorf("deployment %s/%s does not exist", app.Namespace, app.Name)
	}
	nonManagedApp, err := synthetic.NonManagedCamelApplicationFactory(*objOwner)
	if err != nil {
		return nil, err
	}
	targetApp := app.DeepCopy()
	targetApp.Status = v1alpha1.CamelAppStatus{}
	targetApp.ImportCamelAnnotations(nonManagedApp.GetAnnotations())

	deployImage := nonManagedApp.GetAppImage()
	appPhase := nonManagedApp.GetAppPhase()
	targetApp.Status.Phase = appPhase
	targetApp.Status.Image = deployImage
	pods, err := nonManagedApp.GetPods(ctx, action.client)
	if err != nil {
		return targetApp, err
	}
	targetApp.Status.Pods = pods
	targetApp.Status.Replicas = nonManagedApp.GetReplicas()
	targetRuntimeInfo := getInfo(pods)
	if targetRuntimeInfo != nil {
		targetApp.Status.Info = formatRuntimeInfo(targetRuntimeInfo)
	}
	appRuntimeInfo := getInfo(app.Status.Pods)
	if appRuntimeInfo != nil && targetRuntimeInfo != nil {
		pollingInterval := getPollingInterval(targetApp)
		sliErrPerc := getSLIExchangeErrorThreshold(targetApp)
		sliWarnPerc := getSLIExchangeWarningThreshold(targetApp)
		targetApp.Status.SuccessRate = getSLIExchangeSuccessRate(*appRuntimeInfo, *targetRuntimeInfo, &pollingInterval, sliErrPerc, sliWarnPerc)
	}

	message := "Success"
	if app.Status.Replicas != nil && len(pods) != int(*app.Status.Replicas) {
		message = fmt.Sprintf("%d out of %d pods available", len(pods), int(*app.Status.Replicas))
	}

	if len(pods) > 0 && allPodsReady(pods) {
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Monitored",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MonitoringComplete",
			Message:            message,
		})
	} else {
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Monitored",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MonitoringComplete",
			Message:            "Some pod is not ready. See specific pods statuses messages.",
		})
	}

	if len(pods) > 0 && allPodsUp(pods) {
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Healthy",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "HealthCheckCompleted",
			Message:            "All pods are reported as healthy.",
		})
	} else {
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Healthy",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "HealthCheckCompleted",
			Message:            "Some pod is not healthy. See specific pods statuses messages.",
		})
	}

	return targetApp, nil
}

func lookupObject(ctx context.Context, c client.Client, kind, ns string, name string) (*ctrl.Object, error) {
	var obj ctrl.Object
	switch kind {
	case "Deployment":
		obj = &appsv1.Deployment{
			TypeMeta: metav1.TypeMeta{
				Kind:       kind,
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Namespace: ns,
				Name:      name,
			},
		}
		// TODO more cases
	default:
		return nil, fmt.Errorf("cannot manage Camel application of type %s", kind)
	}
	key := ctrl.ObjectKey{
		Namespace: ns,
		Name:      name,
	}
	if err := c.Get(ctx, key, obj); err != nil && k8serrors.IsNotFound(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return &obj, nil
}

func getInfo(pods []v1alpha1.PodInfo) *v1alpha1.RuntimeInfo {
	runtimeInfo := v1alpha1.RuntimeInfo{
		Exchange: &v1alpha1.ExchangeInfo{},
	}

	for _, pod := range pods {
		// Collect runtime information only once
		if runtimeInfo.RuntimeProvider == "" && pod.Runtime != nil {
			runtimeInfo.RuntimeProvider = pod.Runtime.RuntimeProvider
			runtimeInfo.RuntimeVersion = pod.Runtime.RuntimeVersion
			runtimeInfo.CamelVersion = pod.Runtime.CamelVersion
		}
		// Sum all the exchanges processed
		if pod.Runtime != nil && pod.Runtime.Exchange != nil {
			runtimeInfo.Exchange.Total += pod.Runtime.Exchange.Total
			runtimeInfo.Exchange.Failed += pod.Runtime.Exchange.Failed
			runtimeInfo.Exchange.Pending += pod.Runtime.Exchange.Pending
			runtimeInfo.Exchange.Succeeded += pod.Runtime.Exchange.Succeeded

			// Set the major timestamp
			if pod.Runtime.Exchange.LastTimestamp != nil {
				if runtimeInfo.Exchange.LastTimestamp == nil || pod.Runtime.Exchange.LastTimestamp.After(runtimeInfo.Exchange.LastTimestamp.Time) {
					runtimeInfo.Exchange.LastTimestamp = pod.Runtime.Exchange.LastTimestamp
				}
			}
		}
	}

	if runtimeInfo.RuntimeProvider == "" && runtimeInfo.Exchange.Total == 0 {
		// Likely there was no available metric at all
		return nil
	}

	return &runtimeInfo
}

func allPodsReady(pods []v1alpha1.PodInfo) bool {
	for _, pod := range pods {
		if !pod.Ready {
			return false
		}
	}

	return true
}

func allPodsUp(pods []v1alpha1.PodInfo) bool {
	for _, pod := range pods {
		if pod.Runtime == nil || pod.Runtime.Status != "UP" {
			return false
		}
	}

	return true
}

func formatRuntimeInfo(runtimeInfo *v1alpha1.RuntimeInfo) string {
	if runtimeInfo.RuntimeProvider != "" {
		return fmt.Sprintf(
			"%s - %s (%s)",
			runtimeInfo.RuntimeProvider, runtimeInfo.RuntimeVersion, runtimeInfo.CamelVersion,
		)
	}
	return ""
}

func getSLIExchangeSuccessRate(app, target v1alpha1.RuntimeInfo, pollingInteval *time.Duration, sliErrPerc, sliWarnPerc int) *v1alpha1.SLIExchangeSuccessRate {
	var failureRate float64
	sliExchangeSuccessRate := v1alpha1.SLIExchangeSuccessRate{
		SamplingIntervalDuration: pollingInteval,
	}

	totalLastInterval := target.Exchange.Total - app.Exchange.Total
	failedLastInterval := target.Exchange.Failed - app.Exchange.Failed
	if totalLastInterval > 0 {
		failureRate = float64(failedLastInterval) / float64(totalLastInterval) * 100
		successRate := 100 - failureRate
		sliExchangeSuccessRate.SuccessPercentage = strconv.FormatFloat(successRate, 'f', 2, 64)
	}

	if totalLastInterval >= 0 {
		sliExchangeSuccessRate.SamplingIntervalTotal = totalLastInterval
	}
	if failedLastInterval >= 0 {
		sliExchangeSuccessRate.SamplingIntervalFailed = failedLastInterval
	}

	if failureRate > float64(sliWarnPerc) {
		sliExchangeSuccessRate.Status = v1alpha1.SLIExchangeStatusError
	} else if failureRate > float64(sliErrPerc) {
		sliExchangeSuccessRate.Status = v1alpha1.SLIExchangeStatusWarning
	} else if totalLastInterval > 0 {
		// We prevent to mark as success when there is no yet exchange
		sliExchangeSuccessRate.Status = v1alpha1.SLIExchangeStatusSuccess
	}

	if target.Exchange.LastTimestamp != nil {
		sliExchangeSuccessRate.LastTimestamp = target.Exchange.LastTimestamp
	}

	return &sliExchangeSuccessRate
}
