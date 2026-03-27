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

package v1alpha1

type Dashboard struct {
	Title         string  `json:"title"`
	Panels        []Panel `json:"panels"`
	SchemaVersion int     `json:"schemaVersion"`
	Version       int     `json:"version"`
}

type Panel struct {
	Datasource  string      `json:"datasource"`
	Type        string      `json:"type"`
	Title       string      `json:"title"`
	Targets     []Target    `json:"targets"`
	FieldConfig FieldConfig `json:"fieldConfig,omitempty"`
}

type Target struct {
	Expr string `json:"expr"`
}

type FieldConfig struct {
	Defaults FieldDefaults `json:"defaults"`
}

type FieldDefaults struct {
	Unit       string         `json:"unit,omitempty"`
	Min        float64        `json:"min,omitempty"`
	Max        float64        `json:"max,omitempty"`
	Thresholds *Thresholds    `json:"thresholds,omitempty"`
	Custom     *CustomOptions `json:"custom,omitempty"`
}

type CustomOptions struct {
	ThresholdsStyle *ThresholdsStyle `json:"thresholdsStyle,omitempty"`
}

type ThresholdsStyle struct {
	Mode string `json:"mode"`
}

type Thresholds struct {
	Mode  string          `json:"mode"`
	Steps []ThresholdStep `json:"steps"`
}

type ThresholdStep struct {
	Color     string   `json:"color"`
	Value     *float64 `json:"value,omitempty"`
	LineStyle string   `json:"lineStyle,omitempty"`
	Fill      float64  `json:"fill,omitempty"`
}
