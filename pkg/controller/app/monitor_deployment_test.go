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
	"testing"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
)

func TestMonitorActionBakingDeploymentMissing(t *testing.T) {
	app := &v1alpha1.CamelApp{}
	app.Name = "test-app"
	app.Namespace = "default"
	app.Annotations = map[string]string{
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppImportedNameLabel: "test-deployment",
	}

	fakeClient, err := internal.NewFakeClient([]runtime.Object{app})
	require.NoError(t, err)

	action := &monitorAction{}
	action.InjectClient(fakeClient)

	_, err = action.Handle(context.TODO(), app)

	require.Error(t, err)
	require.Equal(t, "baking deployment does not exist for App default/test-app", err.Error())
}

func TestMonitorActionDeploymentScaledTo0(t *testing.T) {
	app := &v1alpha1.CamelApp{}
	app.Name = "test-app"
	app.Namespace = "default"
	app.Annotations = map[string]string{
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppImportedNameLabel: "my-test-deploy",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test-deploy", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Image: "my-camel-image"},
				},
			}},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "my-camel-app"}},
			Replicas: ptr.To(int32(0)),
		},
	}

	fakeClient, err := internal.NewFakeClient([]runtime.Object{app}, deployment)
	require.NoError(t, err)

	action := &monitorAction{}
	action.InjectClient(fakeClient)

	target, err := action.Handle(context.TODO(), app)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, "my-camel-image", target.Status.Image)
	assert.Equal(t, ptr.To(int32(0)), target.Status.Replicas)
	assert.Equal(t, v1alpha1.CamelAppPhasePaused, target.Status.Phase)
	monitored := target.Status.GetCondition("Monitored")
	assert.NotNil(t, monitored)
	assert.Equal(t, metav1.ConditionFalse, monitored.Status)
	assert.Equal(t, "No active Pod available", monitored.Message)
	healthy := target.Status.GetCondition("Healthy")
	assert.Nil(t, healthy)
}

func TestMonitorActionDeploymentNonActivePods(t *testing.T) {
	app := &v1alpha1.CamelApp{}
	app.Name = "test-app"
	app.Namespace = "default"
	app.Annotations = map[string]string{
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppImportedNameLabel: "my-test-deploy",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test-deploy", Namespace: "default"},
		Spec: appsv1.DeploymentSpec{
			Template: v1.PodTemplateSpec{Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Image: "my-camel-image"},
				},
			}},
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "my-camel-app"}},
			Replicas: ptr.To(int32(2)),
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          2,
			AvailableReplicas: 2,
		},
	}

	pod1 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod-1", Namespace: "default", Labels: map[string]string{"app": "my-camel-app"}},
		Status:     v1.PodStatus{Phase: corev1.PodPending},
	}
	pod2 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod-2", Namespace: "default", Labels: map[string]string{"app": "my-camel-app"}},
		Status:     v1.PodStatus{Phase: corev1.PodPending},
	}
	pod3 := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "my-pod-3", Namespace: "default", Labels: map[string]string{"app": "my-camel-app"}},
		Status:     v1.PodStatus{Phase: corev1.PodFailed},
	}

	fakeClient, err := internal.NewFakeClient([]runtime.Object{app}, deployment, pod1, pod2, pod3)
	require.NoError(t, err)

	action := &monitorAction{}
	action.InjectClient(fakeClient)

	target, err := action.Handle(context.TODO(), app)

	require.NoError(t, err)
	require.NotNil(t, target)
	assert.Equal(t, "my-camel-image", target.Status.Image)
	assert.Equal(t, ptr.To(int32(2)), target.Status.Replicas)
	assert.Equal(t, v1alpha1.CamelAppPhaseRunning, target.Status.Phase)
	assert.Len(t, target.Status.Pods, 3)
	// NOTE: we can only test not ready pods as otherwise the logic requires to access an HTTP endpoint
	// which is difficult to mock. The unit test for this logic is done separately.
	assert.Contains(t, target.Status.Pods, v1alpha1.PodInfo{Name: "my-pod-1", Status: "Pending"})
	assert.Contains(t, target.Status.Pods, v1alpha1.PodInfo{Name: "my-pod-2", Status: "Pending"})
	assert.Contains(t, target.Status.Pods, v1alpha1.PodInfo{Name: "my-pod-3", Status: "Failed"})

	monitored := target.Status.GetCondition("Monitored")
	assert.NotNil(t, monitored)
	assert.Equal(t, metav1.ConditionFalse, monitored.Status)
	healthy := target.Status.GetCondition("Healthy")
	assert.NotNil(t, healthy)
	assert.Equal(t, metav1.ConditionFalse, healthy.Status)
}
