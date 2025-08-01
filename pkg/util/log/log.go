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

package log

import (
	"fmt"

	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/go-logr/logr"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

// Log --.
var Log Logger

func init() {
	Log = Logger{
		delegate: logf.Log.WithName("camel-dashboard"),
	}
}

// InitForCmd is required to avoid nil pointer exceptions from command line.
func InitForCmd() {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))
}

// Injectable identifies objects that can receive a Logger.
type Injectable interface {
	InjectLogger(logger Logger)
}

// Logger --.
type Logger struct {
	delegate logr.Logger
}

// Debugf --.
func (l Logger) Debugf(format string, args ...interface{}) {
	l.delegate.V(1).Info(fmt.Sprintf(format, args...))
}

// Infof --.
func (l Logger) Infof(format string, args ...interface{}) {
	l.delegate.Info(fmt.Sprintf(format, args...))
}

// Errorf --.
func (l Logger) Errorf(err error, format string, args ...interface{}) {
	l.delegate.Error(err, fmt.Sprintf(format, args...))
}

// Debug --.
func (l Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.delegate.V(1).Info(msg, keysAndValues...)
}

// Info --.
func (l Logger) Info(msg string, keysAndValues ...interface{}) {
	l.delegate.Info(msg, keysAndValues...)
}

// Error --.
func (l Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.delegate.Error(err, msg, keysAndValues...)
}

// WithName --.
func (l Logger) WithName(name string) Logger {
	return Logger{
		delegate: l.delegate.WithName(name),
	}
}

// WithValues --.
func (l Logger) WithValues(keysAndValues ...interface{}) Logger {
	return Logger{
		delegate: l.delegate.WithValues(keysAndValues...),
	}
}

// ForIntegration --.
func (l Logger) ForApp(target *v1alpha1.CamelApp) Logger {
	return l.WithValues(
		"api-version", target.APIVersion,
		"kind", target.Kind,
		"ns", target.Namespace,
		"name", target.Name,
	)
}

// AsLogger --.
func (l Logger) AsLogger() logr.Logger {
	return l.delegate
}

// ***********************************
//
// Helpers
//
// ***********************************

// WithName --.
func WithName(name string) Logger {
	return Log.WithName(name)
}

// WithValues --.
func WithValues(keysAndValues ...interface{}) Logger {
	return Log.WithValues(keysAndValues...)
}

// ForIntegration --.
func ForApp(target *v1alpha1.CamelApp) Logger {
	return Log.ForApp(target)
}

// ***********************************
//
//
//
// ***********************************

// Debugf --.
func Debugf(format string, args ...interface{}) {
	Log.Debugf(format, args...)
}

// Infof --.
func Infof(format string, args ...interface{}) {
	Log.Infof(format, args...)
}

// Errorf --.
func Errorf(err error, format string, args ...interface{}) {
	Log.Errorf(err, format, args...)
}

// Debug --.
func Debug(msg string, keysAndValues ...interface{}) {
	Log.Debug(msg, keysAndValues...)
}

// Info --.
func Info(msg string, keysAndValues ...interface{}) {
	Log.Info(msg, keysAndValues...)
}

// Error --.
func Error(err error, msg string, keysAndValues ...interface{}) {
	Log.Error(err, msg, keysAndValues...)
}
