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

package kubernetes

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"k8s.io/client-go/kubernetes/scheme"

	"github.com/stretchr/testify/require"
)

func TestLoadResourceFromYaml(t *testing.T) {
	// Prepare a simple Pod YAML
	podYAML := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: test-container
    image: nginx
`

	// Load the Pod using your function
	obj, err := LoadResourceFromYaml(scheme.Scheme, podYAML)
	require.NoError(t, err)
	require.NotNil(t, obj)

	// Type assertion to a Pod
	pod, ok := obj.(*corev1.Pod)
	require.True(t, ok, "expected a *corev1.Pod")

	// Verify fields
	require.Equal(t, "test-pod", pod.Name)
	require.Equal(t, "default", pod.Namespace)
	require.Len(t, pod.Spec.Containers, 1)
	require.Equal(t, "nginx", pod.Spec.Containers[0].Image)
}

func TestLoadResourceFromYaml_InvalidYAML(t *testing.T) {
	invalidYAML := `
apiVersion: v1
kind: Pod
metadata:
  name: test-pod
  namespace: default
spec:
  containers:
  - name: test-container
    image: nginx
  invalid_field-something
`

	obj, err := LoadResourceFromYaml(scheme.Scheme, invalidYAML)
	require.Error(t, err)
	require.Nil(t, obj)
}

func TestLoadResourceFromYaml_UnknownKind(t *testing.T) {
	unknownYAML := `
apiVersion: v1
kind: UnknownKind
metadata:
  name: unknown
`

	obj, err := LoadResourceFromYaml(scheme.Scheme, unknownYAML)
	require.Error(t, err)
	require.Nil(t, obj)
}
