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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

// RuntimeInfoApplyConfiguration represents a declarative configuration of the RuntimeInfo type for use
// with apply.
type RuntimeInfoApplyConfiguration struct {
	Status          *string                         `json:"status,omitempty"`
	RuntimeProvider *string                         `json:"runtimeProvider,omitempty"`
	RuntimeVersion  *string                         `json:"runtimeVersion,omitempty"`
	CamelVersion    *string                         `json:"camelVersion,omitempty"`
	Exchange        *ExchangeInfoApplyConfiguration `json:"exchange,omitempty"`
}

// RuntimeInfoApplyConfiguration constructs a declarative configuration of the RuntimeInfo type for use with
// apply.
func RuntimeInfo() *RuntimeInfoApplyConfiguration {
	return &RuntimeInfoApplyConfiguration{}
}

// WithStatus sets the Status field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Status field is set to the value of the last call.
func (b *RuntimeInfoApplyConfiguration) WithStatus(value string) *RuntimeInfoApplyConfiguration {
	b.Status = &value
	return b
}

// WithRuntimeProvider sets the RuntimeProvider field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RuntimeProvider field is set to the value of the last call.
func (b *RuntimeInfoApplyConfiguration) WithRuntimeProvider(value string) *RuntimeInfoApplyConfiguration {
	b.RuntimeProvider = &value
	return b
}

// WithRuntimeVersion sets the RuntimeVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RuntimeVersion field is set to the value of the last call.
func (b *RuntimeInfoApplyConfiguration) WithRuntimeVersion(value string) *RuntimeInfoApplyConfiguration {
	b.RuntimeVersion = &value
	return b
}

// WithCamelVersion sets the CamelVersion field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CamelVersion field is set to the value of the last call.
func (b *RuntimeInfoApplyConfiguration) WithCamelVersion(value string) *RuntimeInfoApplyConfiguration {
	b.CamelVersion = &value
	return b
}

// WithExchange sets the Exchange field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Exchange field is set to the value of the last call.
func (b *RuntimeInfoApplyConfiguration) WithExchange(value *ExchangeInfoApplyConfiguration) *RuntimeInfoApplyConfiguration {
	b.Exchange = value
	return b
}
