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

package monitoring

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type fakeReconciler struct {
	result reconcile.Result
	err    error
	calls  int
}

func (f *fakeReconciler) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	f.calls++
	return f.result, f.err
}

func TestResultLabelFor(t *testing.T) {
	tests := []struct {
		name     string
		result   reconcile.Result
		err      error
		expected string
	}{
		{
			name:     "reconciled",
			result:   reconcile.Result{},
			expected: string(reconciled),
		},
		{
			name:     "requeued",
			result:   reconcile.Result{RequeueAfter: time.Second},
			expected: string(requeued),
		},
		{
			name:     "errored",
			err:      errors.New("boom"),
			expected: string(errored),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			label := resultLabelFor(tt.result, tt.err)
			if label != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, label)
			}
		})
	}
}

// histogramSampleCount returns the sample count for a HistogramVec with the given labels
func histogramSampleCount(t *testing.T, labels prometheus.Labels) uint64 {
	t.Helper()

	metric := &dto.Metric{}
	h := loopDuration.With(labels)
	if err := h.(prometheus.Metric).Write(metric); err != nil {
		t.Fatalf("failed to write metric: %v", err)
	}

	return metric.GetHistogram().GetSampleCount()
}

func TestInstrumentedReconciler_Reconciled(t *testing.T) {
	rec := &fakeReconciler{
		result: reconcile.Result{},
	}

	gvk := schema.GroupVersionKind{
		Group:   "camel.apache.org",
		Version: "v1",
		Kind:    "CamelApp",
	}

	r := NewInstrumentedReconciler(rec, gvk)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "default",
		},
	}

	labels := map[string]string{
		namespaceLabel: "default",
		groupLabel:     gvk.Group,
		versionLabel:   gvk.Version,
		kindLabel:      gvk.Kind,
		resultLabel:    string(reconciled),
		tagLabel:       "",
	}

	before := histogramSampleCount(t, labels)

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after := histogramSampleCount(t, labels)

	if rec.calls != 1 {
		t.Fatalf("expected reconciler to be called once, got %d", rec.calls)
	}

	if after != before+1 {
		t.Fatalf("expected histogram to increment by 1, before=%v after=%v", before, after)
	}
}

func TestInstrumentedReconciler_Requeued(t *testing.T) {
	rec := &fakeReconciler{
		result: reconcile.Result{RequeueAfter: time.Second},
	}

	gvk := schema.GroupVersionKind{
		Group:   "camel.apache.org",
		Version: "v1",
		Kind:    "CamelApp",
	}

	r := NewInstrumentedReconciler(rec, gvk)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "ns1",
		},
	}

	labels := map[string]string{
		namespaceLabel: "ns1",
		groupLabel:     gvk.Group,
		versionLabel:   gvk.Version,
		kindLabel:      gvk.Kind,
		resultLabel:    string(requeued),
		tagLabel:       "",
	}

	before := histogramSampleCount(t, labels)

	_, err := r.Reconcile(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	after := histogramSampleCount(t, labels)

	if after != before+1 {
		t.Fatalf("expected histogram to increment by 1")
	}
}

func TestInstrumentedReconciler_Errored(t *testing.T) {
	rec := &fakeReconciler{
		err: errors.New("boom"),
	}

	gvk := schema.GroupVersionKind{
		Group:   "camel.apache.org",
		Version: "v1",
		Kind:    "CamelApp",
	}

	r := NewInstrumentedReconciler(rec, gvk)

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Namespace: "errors",
		},
	}

	labels := map[string]string{
		namespaceLabel: "errors",
		groupLabel:     gvk.Group,
		versionLabel:   gvk.Version,
		kindLabel:      gvk.Kind,
		resultLabel:    string(errored),
		tagLabel:       string(platformError),
	}

	before := histogramSampleCount(t, labels)

	_, err := r.Reconcile(context.Background(), req)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	after := histogramSampleCount(t, labels)

	if after != before+1 {
		t.Fatalf("expected histogram to increment by 1")
	}
}
