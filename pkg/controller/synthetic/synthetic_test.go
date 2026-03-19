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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	v1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonManagedUnsupported(t *testing.T) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-pod",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "my-cnt",
					Image: "my-img",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	nilAdapter, err := NonManagedCamelApplicationFactory(pod)
	require.Error(t, err)
	assert.Equal(t, "unsupported Pod object kind", err.Error())
	assert.Nil(t, nilAdapter)
}

func TestNonManagedDeployment(t *testing.T) {
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-deploy",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						v1.AppLabel: "my-imported-it",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-cnt",
							Image: "my-img",
						},
					},
				},
			},
		},
	}

	expectedIt := v1.NewApp("ns", "my-imported-it")
	expectedIt.SetAnnotations(map[string]string{
		v1.AppImportedNameLabel: "my-deploy",
		v1.AppImportedKindLabel: "Deployment",
		v1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       deploy.Name,
			UID:        deploy.UID,
			Controller: &controller,
		},
	}
	expectedIt.SetOwnerReferences(references)

	deploymentAdapter, err := NonManagedCamelApplicationFactory(deploy)
	require.NoError(t, err)
	assert.NotNil(t, deploymentAdapter)
	assert.Equal(t, expectedIt.ObjectMeta, deploymentAdapter.CamelApp(context.Background(), nil).ObjectMeta)
}

func TestNonManagedCronJob(t *testing.T) {
	cron := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: batchv1.SchemeGroupVersion.String(),
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-cron",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								v1.AppLabel: "my-imported-it",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "my-cnt",
									Image: "my-img",
								},
							},
						},
					},
				},
			},
		},
	}

	expectedIt := v1.NewApp("ns", "my-imported-it")
	expectedIt.SetAnnotations(map[string]string{
		v1.AppImportedNameLabel: "my-cron",
		v1.AppImportedKindLabel: "CronJob",
		v1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       cron.Name,
			UID:        cron.UID,
			Controller: &controller,
		},
	}
	expectedIt.SetOwnerReferences(references)
	cronJobAdapter, err := NonManagedCamelApplicationFactory(cron)
	require.NoError(t, err)
	assert.NotNil(t, cronJobAdapter)
	assert.Equal(t, expectedIt, *cronJobAdapter.CamelApp(context.Background(), nil))
}

func TestSyntheticOnAddCreateAppUnsupported(t *testing.T) {
	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "pod1",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			PodIP: "10.0.0.1",
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}
	fakeClient, err := internal.NewFakeClient(pod1)
	require.NoError(t, err)
	onAdd(context.TODO(), fakeClient, pod1)
	// No need to check anything as the func is void return.
}

func TestSyntheticOnAddCreateAppOnDelete(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-app",
			Namespace:   "default",
			UID:         "1234",
			Annotations: map[string]string{"foo": "bar"},
			Labels:      map[string]string{v1alpha1.AppLabel: "my-app"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(3)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main", Image: "my-image:v1"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          3,
			AvailableReplicas: 2,
		},
	}
	fakeClient, err := internal.NewFakeClient(deploy)
	require.NoError(t, err)
	onAdd(context.TODO(), fakeClient, deploy)
	createdApp, err := getSyntheticCamelApp(context.TODO(), fakeClient, "default", "my-app")
	require.NoError(t, err)
	assert.Equal(t, "default", createdApp.GetNamespace())
	assert.Equal(t, "my-app", createdApp.GetName())
	assert.Equal(t, map[string]string{
		"camel.apache.org/imported-from-kind": "Deployment",
		"camel.apache.org/imported-from-name": "my-app",
		"camel.apache.org/is-synthetic":       "true",
	}, createdApp.GetAnnotations())

	// Let's try to delete
	onDelete(context.TODO(), fakeClient, deploy)
	_, err = getSyntheticCamelApp(context.TODO(), fakeClient, "default", "my-app")
	require.Error(t, err)
	assert.Equal(t, "camelapps.camel.apache.org \"my-app\" not found", err.Error())
}

func TestSyntheticOnAddDuplicatedCamelApp(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-app",
			Namespace:   "default",
			UID:         "1234",
			Annotations: map[string]string{"foo": "bar"},
			Labels:      map[string]string{v1alpha1.AppLabel: "my-app"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(3)),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main", Image: "my-image:v1"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          3,
			AvailableReplicas: 2,
		},
	}
	existingCamelApp := v1alpha1.NewApp("default", "my-app")
	fakeClient, err := internal.NewFakeClient(&existingCamelApp, deploy)
	require.NoError(t, err)
	onAdd(context.TODO(), fakeClient, deploy)

	// Existing app had no annotations as it was not imported by the operator!
	existingApp, err := getSyntheticCamelApp(context.TODO(), fakeClient, "default", "my-app")
	require.NoError(t, err)
	assert.Equal(t, "default", existingApp.GetNamespace())
	assert.Equal(t, "my-app", existingApp.GetName())
	assert.Nil(t, existingApp.GetAnnotations())

	// Created app, however has those annotations and a different name as the
	// operator detect naming collision and create a new app suffixing the resource name
	createdApp, err := getSyntheticCamelApp(context.TODO(), fakeClient, "default", "my-app-my-app")
	require.NoError(t, err)
	assert.Equal(t, "default", createdApp.GetNamespace())
	assert.Equal(t, "my-app-my-app", createdApp.GetName())
	assert.Equal(t, map[string]string{
		"camel.apache.org/imported-from-kind": "Deployment",
		"camel.apache.org/imported-from-name": "my-app",
		"camel.apache.org/is-synthetic":       "true",
	}, createdApp.GetAnnotations())
}
