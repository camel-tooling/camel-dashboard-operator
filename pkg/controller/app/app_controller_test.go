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
	"github.com/stretchr/testify/require"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
)

func TestReconcileApp_Reconcile(t *testing.T) {
	app := &v1alpha1.CamelApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-app",
			Namespace: "default",
		},
	}
	fakeClient, err := internal.NewFakeClient([]runtime.Object{app})
	require.NoError(t, err)
	r := &reconcileApp{
		client: fakeClient,
		scheme: fakeClient.Scheme(),
	}
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "test-app",
			Namespace: "default",
		},
	}

	res, err := r.Reconcile(context.TODO(), req)

	require.NoError(t, err)
	require.True(t, res.RequeueAfter >= 0)
}

func TestReconcileApp_NotFound(t *testing.T) {
	fakeClient, err := internal.NewFakeClient(nil)
	require.NoError(t, err)
	r := &reconcileApp{
		client: fakeClient,
		scheme: fakeClient.Scheme(),
	}
	req := ctrl.Request{
		NamespacedName: types.NamespacedName{
			Name:      "missing",
			Namespace: "default",
		},
	}

	res, err := r.Reconcile(context.TODO(), req)

	require.NoError(t, err)
	require.Equal(t, ctrl.Result{}, res)
}
