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

	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// nonManagedCamelCronjob represents a cron Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelCronjob struct {
	cron *batchv1.CronJob
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelCronjob) CamelApp(ctx context.Context, c client.Client) *v1alpha1.CamelApp {
	newApp := v1alpha1.NewApp(app.cron.Namespace, app.cron.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.cron.Name,
		v1alpha1.AppImportedKindLabel: "CronJob",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       app.cron.Name,
			UID:        app.cron.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)
	return &newApp
}

// GetAppPhase returns the phase of the backing Camel application.
func (app *nonManagedCamelCronjob) GetAppPhase() v1alpha1.CamelAppPhase {
	return v1alpha1.CamelAppPhase("TBD")
}

// GetReplicas returns the number of desired replicas for the backing Camel application.
func (app *nonManagedCamelCronjob) GetReplicas() *int32 {
	return ptr.To(int32(-1))
}

// GetAppImage returns the container image of the backing Camel application.
func (app *nonManagedCamelCronjob) GetAppImage() string {
	return ""
}

// GetPods returns the container image of the backing Camel application.
func (app *nonManagedCamelCronjob) GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error) {
	return nil, nil
}

// GetAnnotations returns the backing deployment object annotations.
func (app *nonManagedCamelCronjob) GetAnnotations() map[string]string {
	return app.cron.Annotations
}
