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

package util

import (
	"math"
	"testing"
)

func TestIToInt8(t *testing.T) {
	tests := []struct {
		name      string
		input     int
		want      int8
		expectErr bool
	}{
		{"min int8", math.MinInt8, math.MinInt8, false},
		{"max int8", math.MaxInt8, math.MaxInt8, false},
		{"zero", 0, 0, false},
		{"positive in range", 42, 42, false},
		{"negative in range", -42, -42, false},
		{"overflow positive", 128, 0, true},
		{"overflow negative", -129, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IToInt8(tt.input)
			if (err != nil) != tt.expectErr {
				t.Fatalf("IToInt8(%d) error = %v, wantErr %v", tt.input, err, tt.expectErr)
			}
			if !tt.expectErr {
				if got == nil {
					t.Fatalf("IToInt8(%d) returned nil, want %d", tt.input, tt.want)
				}
				if *got != tt.want {
					t.Errorf("IToInt8(%d) = %d, want %d", tt.input, *got, tt.want)
				}
			} else {
				if got != nil {
					t.Errorf("IToInt8(%d) = %v, want nil on overflow", tt.input, *got)
				}
			}
		})
	}
}
