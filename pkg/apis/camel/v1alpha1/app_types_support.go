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

package v1alpha1

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

const (
	// camelPrefix is used to identify Camel prefix labels/annotations.
	camelPrefix = "camel.apache.org"
	// AppLabel is used to tag k8s object created by a given Camel Application.
	AppLabel = "camel.apache.org/app"
	// AppSyntheticLabel is used to tag k8s synthetic Camel Applications.
	AppSyntheticLabel = "camel.apache.org/is-synthetic"
	// AppImportedKindLabel specifies from what kind of resource an App was imported.
	AppImportedKindLabel = "camel.apache.org/imported-from-kind"
	// AppImportedNameLabel specifies from what resource an App was imported.
	AppImportedNameLabel = "camel.apache.org/imported-from-name"
	// AppPollingIntervalSecondsAnnotation is used to instruct a given application to poll interval.
	AppPollingIntervalSecondsAnnotation = "camel.apache.org/polling-interval-seconds"
	// AppObservabilityServicesPort is used to instruct an application to use a specific port for metrics scraping.
	AppObservabilityServicesPort = "camel.apache.org/observability-services-port"
	// AppSLIExchangeErrorPercentageAnnotation is used to instruct a given application error percentage SLI Exchange.
	AppSLIExchangeErrorPercentageAnnotation = "camel.apache.org/sli-exchange-error-percentage"
	// AppSLIExchangeWarningPercentageAnnotation is used to instruct a given application warning percentage SLI Exchange.
	AppSLIExchangeWarningPercentageAnnotation = "camel.apache.org/sli-exchange-warning-percentage"
)

func NewApp(namespace string, name string) CamelApp {
	return CamelApp{
		TypeMeta: metav1.TypeMeta{
			APIVersion: SchemeGroupVersion.String(),
			Kind:       AppKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
	}
}

func (appStatus *CamelAppStatus) AddCondition(condition metav1.Condition) {
	if appStatus.Conditions == nil {
		appStatus.Conditions = []metav1.Condition{}
	}
	appStatus.Conditions = append(appStatus.Conditions, condition)
}

// ImportCamelAnnotations copies all camel annotations from the deployment to the App.
func (app *CamelApp) ImportCamelAnnotations(annotations map[string]string) {
	for k, v := range annotations {
		if strings.HasPrefix(k, camelPrefix) {
			app.Annotations[k] = v
		}
	}
}

func (appStatus *CamelAppStatus) GetCondition(condType string) *metav1.Condition {
	for _, cond := range appStatus.Conditions {
		if cond.Type == condType {
			return &cond
		}
	}

	return nil
}

// DoesExposeMetrics returns true if the app was reconciled and has metrics availability.
func (appStatus *CamelAppStatus) DoesExposeMetrics() bool {
	return len(appStatus.Pods) > 0 &&
		appStatus.Pods[0].ObservabilityService != nil &&
		appStatus.Pods[0].ObservabilityService.MetricsEndpoint != "" &&
		appStatus.Pods[0].ObservabilityService.MetricsPort != 0
}

// GetOwnerReferences returns the owner references to this app.
func (app *CamelApp) GetOwnerReferences() []metav1.OwnerReference {
	return []metav1.OwnerReference{
		{
			APIVersion:         app.APIVersion,
			Kind:               app.Kind,
			Name:               app.Name,
			UID:                app.UID,
			Controller:         ptr.To(true),
			BlockOwnerDeletion: ptr.To(true),
		},
	}
}
