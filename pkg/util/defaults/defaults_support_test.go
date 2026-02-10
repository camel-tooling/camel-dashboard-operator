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

package defaults

import (
	"testing"
)

func TestEnvOrDefault(t *testing.T) {
	const envName1 = "TEST_ENV1"
	const envName2 = "TEST_ENV2"

	t.Run("first env set", func(t *testing.T) {
		t.Setenv(envName1, "value1")
		t.Setenv(envName2, "value2") // should be ignored

		got := envOrDefault("default", envName1, envName2)
		if got != "value1" {
			t.Errorf("EnvOrDefault(...) = %q; want %q", got, "value1")
		}
	})

	t.Run("first env empty, second env set", func(t *testing.T) {
		t.Setenv(envName1, "")
		t.Setenv(envName2, "value2")

		got := envOrDefault("default", envName1, envName2)
		if got != "value2" {
			t.Errorf("EnvOrDefault(...) = %q; want %q", got, "value2")
		}
	})

	t.Run("no env set", func(t *testing.T) {
		t.Setenv(envName1, "")
		t.Setenv(envName2, "")

		got := envOrDefault("default", envName1, envName2)
		if got != "default" {
			t.Errorf("EnvOrDefault(...) = %q; want %q", got, "default")
		}
	})
}

func TestOperatorID(t *testing.T) {
	t.Run("env set", func(t *testing.T) {
		t.Setenv("OPERATOR_ID", "op-123")
		if got := OperatorID(); got != "op-123" {
			t.Errorf("OperatorID() = %q; want %q", got, "op-123")
		}
	})

	t.Run("env unset", func(t *testing.T) {
		t.Setenv("OPERATOR_ID", "")
		if got := OperatorID(); got != "" {
			t.Errorf("OperatorID() = %q; want empty string", got)
		}
	})
}
