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

package patch

import (
	"encoding/json"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/utils/ptr"
)

type TestStruct struct {
	Name  string  `json:"name,omitempty"`
	Value *string `json:"value,omitempty"`
	Meta  *Meta   `json:"meta,omitempty"`
}

// DeepCopyObject satisfies runtime.Object
func (t TestStruct) DeepCopyObject() runtime.Object {
	copy := t
	if t.Value != nil {
		v := *t.Value
		copy.Value = &v
	}
	if t.Meta != nil {
		m := *t.Meta
		copy.Meta = &m
		if t.Meta.Enabled != nil {
			b := *t.Meta.Enabled
			copy.Meta.Enabled = &b
		}
	}
	return &copy
}

// GetObjectKind satisfies runtime.Object
func (t TestStruct) GetObjectKind() schema.ObjectKind {
	// For unit tests, a simple empty kind is enough
	return schema.EmptyObjectKind
}

type Meta struct {
	Enabled *bool `json:"enabled,omitempty"`
}

func TestMergePatch_SimpleMap(t *testing.T) {
	source := map[string]interface{}{
		"name": "foo",
		"val":  nil,
	}
	target := map[string]interface{}{
		"name": "bar",
		"val":  42,
	}

	patch, err := MergePatch(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(patch, &m); err != nil {
		t.Fatalf("invalid JSON patch: %v", err)
	}

	if m["val"] != 42.0 {
		t.Errorf("expected val=42, got %v", m["val"])
	}
	if m["name"] != "bar" {
		t.Errorf("expected name=bar, got %v", m["name"])
	}
}

func TestMergePatch_Unstructured(t *testing.T) {
	source := &unstructured.Unstructured{}
	target := &unstructured.Unstructured{
		Object: map[string]interface{}{"foo": "bar"},
	}

	patch, err := MergePatch(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(patch, &m); err != nil {
		t.Fatalf("invalid JSON patch: %v", err)
	}

	if m["foo"] != "bar" {
		t.Errorf("expected foo=bar, got %v", m["foo"])
	}
}

func TestMergePatch_TypedStruct(t *testing.T) {
	value := "x"
	source := TestStruct{
		Name:  "a",
		Value: nil,
		Meta:  &Meta{Enabled: nil},
	}
	target := TestStruct{
		Name:  "b",
		Value: &value,
		Meta:  &Meta{Enabled: ptr.To(true)},
	}

	patch, err := MergePatch(source, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(patch, &m); err != nil {
		t.Fatalf("invalid JSON patch: %v", err)
	}

	if m["name"] != "b" {
		t.Errorf("expected name=b, got %v", m["name"])
	}
	if m["value"] != "x" && m["value"] != value {
		t.Errorf("expected value to be non-nil, got %v", m["value"])
	}
	if meta, ok := m["meta"].(map[string]interface{}); ok {
		if _, exists := meta["enabled"]; !exists {
			t.Errorf("expected meta.enabled to exist")
		}
	} else {
		t.Errorf("expected meta to be map, got %T", m["meta"])
	}
}

func TestApplyPatch_Unstructured(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{"foo": "bar"},
	}

	patched, err := ApplyPatch(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patched != obj {
		t.Errorf("expected same unstructured object returned")
	}
}

func TestApplyPatch_TypedStruct(t *testing.T) {
	value := "x"
	obj := TestStruct{
		Name:  "a",
		Value: &value,
		Meta:  &Meta{Enabled: ptr.To(true)},
	}

	patched, err := ApplyPatch(obj)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if patched.Object["name"] != "a" {
		t.Errorf("expected name=a, got %v", patched.Object["name"])
	}
	if patched.Object["value"] != "x" {
		t.Errorf("expected value=x, got %v", patched.Object["value"])
	}
	if meta, ok := patched.Object["meta"].(map[string]interface{}); ok {
		if meta["enabled"] != true {
			t.Errorf("expected meta.enabled=true, got %v", meta["enabled"])
		}
	} else {
		t.Errorf("expected meta map, got %T", patched.Object["meta"])
	}
}

func TestRemoveNilValues_EmptyMapsDeleted(t *testing.T) {
	obj := map[string]interface{}{
		"a": nil,
		"b": map[string]interface{}{
			"x": nil,
			"y": map[string]interface{}{},
		},
		"c": "keep",
	}

	removeNilValues(reflect.ValueOf(obj), reflect.Value{})

	if _, exists := obj["a"]; exists {
		t.Errorf("expected a to be removed")
	}
	if _, exists := obj["b"]; exists {
		t.Errorf("expected b to be removed")
	}
	if obj["c"] != "keep" {
		t.Errorf("expected c to remain, got %v", obj["c"])
	}
}
