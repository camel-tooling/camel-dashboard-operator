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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	authorizationv1 "k8s.io/api/authorization/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	clientgotesting "k8s.io/client-go/testing"
)

func TestCheckPermission_Allowed(t *testing.T) {
	ctx := context.Background()

	client := fake.NewClientset()

	// Reactor simulates the SAR response
	client.PrependReactor("create", "selfsubjectaccessreviews", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{
				Allowed: true,
			},
		}, nil
	})

	allowed, err := CheckPermission(ctx, client, "apps", "deployments", "default", "my-deploy", "get")
	require.NoError(t, err)
	require.True(t, allowed)
}

func TestCheckPermission_Denied(t *testing.T) {
	ctx := context.Background()

	// fake client does not actually enforce SAR logic; we simulate response
	client := fake.NewClientset()

	// Override the reactor to simulate denied access
	client.PrependReactor("create", "selfsubjectaccessreviews", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		sar := &authorizationv1.SelfSubjectAccessReview{
			Status: authorizationv1.SubjectAccessReviewStatus{
				Allowed: false,
			},
		}
		return true, sar, nil
	})

	allowed, err := CheckPermission(ctx, client, "apps", "deployments", "default", "my-deploy", "delete")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestCheckPermission_ForbiddenError(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientset()

	// Simulate forbidden error
	client.PrependReactor("create", "selfsubjectaccessreviews", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, k8serrors.NewForbidden(authorizationv1.Resource("deployments"), "my-deploy", nil)
	})

	allowed, err := CheckPermission(ctx, client, "apps", "deployments", "default", "my-deploy", "update")
	require.NoError(t, err)
	require.False(t, allowed)
}

func TestCheckPermission_OtherError(t *testing.T) {
	ctx := context.Background()
	client := fake.NewClientset()

	// Simulate network or other error
	client.PrependReactor("create", "selfsubjectaccessreviews", func(action clientgotesting.Action) (bool, runtime.Object, error) {
		return true, nil, k8serrors.NewBadRequest("bad request")
	})

	allowed, err := CheckPermission(ctx, client, "apps", "deployments", "default", "my-deploy", "update")
	require.Error(t, err)
	require.False(t, allowed)
}
