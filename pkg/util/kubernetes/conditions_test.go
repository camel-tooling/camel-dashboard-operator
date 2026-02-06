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

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestGetPodCondition(t *testing.T) {
	tests := []struct {
		name          string
		pod           corev1.Pod
		conditionType corev1.PodConditionType
		wantPresent   bool
		wantStatus    corev1.ConditionStatus
	}{
		{
			name: "condition exists",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
						{Type: corev1.PodScheduled, Status: corev1.ConditionFalse},
					},
				},
			},
			conditionType: corev1.PodReady,
			wantPresent:   true,
			wantStatus:    corev1.ConditionTrue,
		},
		{
			name: "condition exists but different type",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
					},
				},
			},
			conditionType: corev1.PodScheduled,
			wantPresent:   false,
		},
		{
			name:          "pod has no conditions",
			pod:           corev1.Pod{},
			conditionType: corev1.PodReady,
			wantPresent:   false,
		},
		{
			name: "multiple conditions, matching last one",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodScheduled, Status: corev1.ConditionFalse},
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
					},
				},
			},
			conditionType: corev1.PodReady,
			wantPresent:   true,
			wantStatus:    corev1.ConditionTrue,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cond := GetPodCondition(tt.pod, tt.conditionType)
			if tt.wantPresent {
				require.NotNil(t, cond)
				require.Equal(t, tt.wantStatus, cond.Status)
			} else {
				require.Nil(t, cond)
			}
		})
	}
}
