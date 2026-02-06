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

	"github.com/stretchr/testify/require"
)

func TestJolokiaEnabled(t *testing.T) {
	tests := []struct {
		name string
		pod  corev1.Pod
		want bool
	}{
		{
			name: "pod with jolokia port",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Ports: []corev1.ContainerPort{
								{Name: "http", ContainerPort: 8080},
								{Name: "jolokia", ContainerPort: 8778},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod without jolokia port",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app",
							Ports: []corev1.ContainerPort{
								{Name: "http", ContainerPort: 8080},
								{Name: "metrics", ContainerPort: 9090},
							},
						},
					},
				},
			},
			want: false,
		},
		{
			name: "pod with multiple containers, one has jolokia",
			pod: corev1.Pod{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "app1",
							Ports: []corev1.ContainerPort{
								{Name: "http", ContainerPort: 8080},
							},
						},
						{
							Name: "app2",
							Ports: []corev1.ContainerPort{
								{Name: "jolokia", ContainerPort: 8778},
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "pod with no containers",
			pod:  corev1.Pod{}, // empty pod
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JolokiaEnabled(tt.pod)
			require.Equal(t, tt.want, got)
		})
	}
}
