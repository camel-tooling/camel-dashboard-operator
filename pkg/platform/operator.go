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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
)

const (
	OperatorWatchNamespaceEnvVariable = "WATCH_NAMESPACE"
	operatorNamespaceEnvVariable      = "NAMESPACE"
	CamelAppLabelSelector             = "LABEL_SELECTOR"

	CamelAppPollIntervalSeconds             = "POLL_INTERVAL_SECONDS"
	DefaultPollingIntervalSeconds           = 60
	SLIExchangeErrorPercentage              = "SLI_ERR_PERCENTAGE"
	defaultSLIExchangeErrorPercentage       = 5
	SLIExchangeWarningPercentage            = "SLI_WARN_PERCENTAGE"
	defaultSLIExchangeWarningPercentage     = 10
	CamelAppObservabilityPort               = "OBSERVABILITY_PORT"
	defaultObservabilityPort            int = 9876
	DefaultObservabilityMetrics             = "observe/metrics"
	DefaultObservabilityHealth              = "observe/health"

	OperatorLockName = "camel-dashboard-lock"
)

// IsCurrentOperatorGlobal returns true if the operator is configured to watch all namespaces.
func IsCurrentOperatorGlobal() bool {
	if watchNamespace, envSet := os.LookupEnv(OperatorWatchNamespaceEnvVariable); !envSet || strings.TrimSpace(watchNamespace) == "" {
		log.Debug("Operator is global to all namespaces")
		return true
	}

	log.Debug("Operator is local to namespace")
	return false
}

// GetOperatorWatchNamespace returns the namespace the operator watches.
func GetOperatorWatchNamespace() string {
	if namespace, envSet := os.LookupEnv(OperatorWatchNamespaceEnvVariable); envSet {
		return namespace
	}
	return ""
}

// GetOperatorNamespace returns the namespace where the current operator is located (if set).
func GetOperatorNamespace() string {
	if podNamespace, envSet := os.LookupEnv(operatorNamespaceEnvVariable); envSet {
		return podNamespace
	}
	return ""
}

// GetOperatorLockName returns the name of the lock lease that is electing a leader on the particular namespace.
func GetOperatorLockName(operatorID string) string {
	return fmt.Sprintf("%s-lock", operatorID)
}

// GetAppLabelSelector returns the label selector used to determine a Camel application.
func GetAppLabelSelector() string {
	if labelSelector, envSet := os.LookupEnv(CamelAppLabelSelector); envSet && labelSelector != "" {
		return labelSelector
	}
	return v1alpha1.AppLabel
}

// getOperatorEnvAsInt returns a generic operator environment variable as an it. It fallbacks to default value if the env var is missing.
func getOperatorEnvAsInt(envVar, envVarDescription string, defaultValue int) int {
	if envVarVal, envSet := os.LookupEnv(envVar); envSet && envVarVal != "" {
		v, err := strconv.Atoi(envVarVal)
		if err == nil {
			return v
		} else {
			log.Errorf(err, "could not properly parse Operator %s, "+
				"fallback to default value %d", envVarDescription, defaultValue)
		}
	}

	return defaultValue
}

// getPollingIntervalSeconds returns the polling interval (in seconds) for the operator. It fallbacks to default value.
func getPollingIntervalSeconds() int {
	return getOperatorEnvAsInt(CamelAppPollIntervalSeconds, "polling interval configuration", DefaultPollingIntervalSeconds)
}

// GetPollingInterval returns the polling interval for the operator. It fallbacks to default value.
func GetPollingInterval() time.Duration {
	return time.Duration(getPollingIntervalSeconds()) * time.Second
}

// GetObservabilityPort returns the observability port set for the operator. It fallbacks to default value.
func GetObservabilityPort() int {
	return getOperatorEnvAsInt(CamelAppObservabilityPort, "observability port configuration", defaultObservabilityPort)
}

// GetSLIExchangeErrorThreshold returns the SLI Exchange error threshold configuration. It fallbacks to default value.
func GetSLIExchangeErrorThreshold() int {
	return getOperatorEnvAsInt(SLIExchangeErrorPercentage, "SLI exchange error threshold", defaultSLIExchangeErrorPercentage)
}

// GetSLIExchangeWarnThreshold returns the SLI Exchange warning threshold configuration. It fallbacks to default value.
func GetSLIExchangeWarningThreshold() int {
	return getOperatorEnvAsInt(SLIExchangeWarningPercentage, "SLI exchange warning threshold", defaultSLIExchangeWarningPercentage)
}
