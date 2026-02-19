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

package platform

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsCurrentOperatorGlobal(t *testing.T) {
	t.Run("empty env => global", func(t *testing.T) {
		t.Setenv(OperatorWatchNamespaceEnvVariable, "")
		assert.True(t, IsCurrentOperatorGlobal())
	})

	t.Run("whitespace env => global", func(t *testing.T) {
		t.Setenv(OperatorWatchNamespaceEnvVariable, "   ")
		assert.True(t, IsCurrentOperatorGlobal())
	})

	t.Run("set env => local", func(t *testing.T) {
		t.Setenv(OperatorWatchNamespaceEnvVariable, "my-namespace")
		assert.False(t, IsCurrentOperatorGlobal())
	})
}

func TestGetOperatorWatchNamespace(t *testing.T) {
	t.Setenv(OperatorWatchNamespaceEnvVariable, "ns1")

	assert.Equal(t, "ns1", GetOperatorWatchNamespace())
}

func TestGetOperatorNamespace(t *testing.T) {
	t.Setenv(operatorNamespaceEnvVariable, "operator-ns")

	assert.Equal(t, "operator-ns", GetOperatorNamespace())
}

func TestGetOperatorEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		envValue     string
		expected     int
		defaultValue int
	}{
		{"valid int", "42", 42, 10},
		{"invalid int", "abc", 10, 10},
		{"empty value", "", 10, 10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_INT_ENV", tt.envValue)

			val := getOperatorEnvAsInt("TEST_INT_ENV", "test int env", tt.defaultValue)
			assert.Equal(t, tt.expected, val)
		})
	}
}

func TestGetPollingInterval(t *testing.T) {
	t.Setenv(CamelAppPollIntervalSeconds, "30")

	assert.Equal(t, 30*time.Second, GetPollingInterval())
}

func TestGetObservabilityPort_Default(t *testing.T) {
	t.Setenv(CamelAppObservabilityPort, "")

	assert.Equal(t, defaultObservabilityPort, GetObservabilityPort())
}

func TestSLIThresholds(t *testing.T) {
	t.Setenv(SLIExchangeErrorPercentage, "7")
	assert.Equal(t, 7, GetSLIExchangeErrorThreshold())

	t.Setenv(SLIExchangeWarningPercentage, "15")
	assert.Equal(t, 15, GetSLIExchangeWarningThreshold())
}
