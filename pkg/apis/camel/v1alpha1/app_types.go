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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AppKind --.
	AppKind string = "App"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
// Important: Run "make generate" to regenerate code after modifying this file

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=apps,scope=Namespaced,shortName=capp,categories=camel
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`,description="The Camel App phase"
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.status.image`,description="The Camel App image"
// +kubebuilder:printcolumn:name="Replicas",type=string,JSONPath=`.status.replicas`,description="The Camel App Pods"
// +kubebuilder:printcolumn:name="Info",type=string,JSONPath=`.status.info`,description="The Camel App info"
// +kubebuilder:subresource:status
// +kubebuilder:storageversion

// App is the Schema for the Camel Applications API.
type App struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// the desired App specification
	Spec AppSpec `json:"spec,omitempty"`
	// the status of the App
	Status AppStatus `json:"status,omitempty"`
}

// AppSpec specifies the configuration of an App.
type AppSpec struct {
}

// AppStatus defines the observed state of an App.
type AppStatus struct {
	// the actual phase
	Phase AppPhase `json:"phase,omitempty"`
	// the image used to run the application
	Image string `json:"image,omitempty"`
	// Some information about the pods backing the application
	Pods []PodInfo `json:"pods,omitempty"`
	// The number of replicas (pods running)
	Replicas *int32 `json:"replicas,omitempty"`
	// The number of replicas (pods running)
	Info string `json:"info,omitempty"`
}

// +kubebuilder:object:root=true

// AppList contains a list of Apps.
type AppList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []App `json:"items"`
}

// AppPhase --.
type AppPhase string

// AppPhaseRunning --.
const AppPhaseRunning AppPhase = "Running"

// AppPhaseError --.
const AppPhaseError AppPhase = "Error"

// PodInfo contains a set of information related to the Pod running the Camel application.
type PodInfo struct {
	// the Pod name
	Name string `json:"name,omitempty"`
	// the Pod ip
	InternalIP string `json:"internalIp,omitempty"`
	// the Pod status
	Status string `json:"status,omitempty"`
	// Observability services information
	ObservabilityService ObservabilityServiceInfo `json:"observe,omitempty"`
	// Some information about the Camel runtime
	Runtime *RuntimeInfo `json:"runtime,omitempty"`
}

// RuntimeInfo contains a set of information related to the Camel application runtime.
type RuntimeInfo struct {
	// the name of Camel context
	ContextName string `json:"contextName,omitempty"`
	// the runtime provider
	RuntimeProvider string `json:"runtimeProvider,omitempty"`
	// the runtime version
	RuntimeVersion string `json:"runtimeVersion,omitempty"`
	// the Camel core version
	CamelVersion string `json:"camelVersion,omitempty"`
	// Information about the exchange
	Exchange *ExchangeInfo `json:"exchange,omitempty"`
}

// ObservabilityServiceInfo contains the endpoints that can be possibly used to scrape more information.
type ObservabilityServiceInfo struct {
	// the health endpoint
	HealthEndpoint string `json:"healthEndpoint,omitempty"`
	// the metrics endpoint
	MetricsEndpoint string `json:"metricsEndpoint,omitempty"`
}

// ExchangeInfo contains the endpoints that can be possibly used to scrape more information.
type ExchangeInfo struct {
	// The total number of exchanges
	Total int `json:"total,omitempty"`
	// The total number of exchanges succeeded
	Succeeded int `json:"succeed,omitempty"`
	// The total number of exchanges failed
	Failed int `json:"failed,omitempty"`
	// The total number of exchanges pending (in Camel jargon, inflight exchanges)
	Pending int `json:"pending,omitempty"`
}
