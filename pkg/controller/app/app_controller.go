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
	"strconv"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/event"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/monitoring"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

func Add(ctx context.Context, mgr manager.Manager, c client.Client) error {
	return add(mgr, newReconciler(mgr, c))
}

func newReconciler(mgr manager.Manager, c client.Client) reconcile.Reconciler {
	return monitoring.NewInstrumentedReconciler(
		&reconcileApp{
			client:   c,
			reader:   mgr.GetAPIReader(),
			scheme:   mgr.GetScheme(),
			recorder: mgr.GetEventRecorderFor("camel-dashboard-app-controller"),
		},
		schema.GroupVersionKind{
			Group:   v1alpha1.SchemeGroupVersion.Group,
			Version: v1alpha1.SchemeGroupVersion.Version,
			Kind:    v1alpha1.AppKind,
		},
	)
}

func add(mgr manager.Manager, r reconcile.Reconciler) error {
	return builder.ControllerManagedBy(mgr).
		Named("app-controller").
		For(&v1alpha1.CamelApp{}, builder.WithPredicates(UpdateFalsePredicate{})).
		Complete(r)
}

// reconcileApp reconciles an App object.
type reconcileApp struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the API server
	client   client.Client
	reader   ctrl.Reader
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

func (r *reconcileApp) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	rlog := Log.WithValues("request-namespace", request.Namespace, "request-name", request.Name)
	var instance v1alpha1.CamelApp
	if err := r.client.Get(ctx, request.NamespacedName, &instance); err != nil {
		if k8serrors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	actions := []Action{
		NewMonitorAction(),
	}
	var err error

	target := instance.DeepCopy()
	targetLog := rlog.ForApp(target)

	for _, a := range actions {
		a.InjectClient(r.client)
		a.InjectLogger(targetLog)

		if !a.CanHandle(target) {
			continue
		}
		targetLog.Debugf("Invoking action %s", a.Name())

		target, err = a.Handle(ctx, target)
		if err != nil {
			event.NotifyAppError(ctx, r.client, r.recorder, &instance, target, err)
			if target != nil {
				_ = r.update(ctx, &instance, target, &targetLog)
			}
			return reconcile.Result{}, err
		}

		if target != nil {
			if err := r.update(ctx, &instance, target, &targetLog); err != nil {
				event.NotifyAppError(ctx, r.client, r.recorder, &instance, target, err)
				return reconcile.Result{}, err
			}
		}
		event.NotifyAppUpdated(ctx, r.client, r.recorder, &instance, target)
	}

	return reconcile.Result{RequeueAfter: getPollingInterval(target)}, nil
}

func (r *reconcileApp) update(ctx context.Context, base *v1alpha1.CamelApp, target *v1alpha1.CamelApp, log *log.Logger) error {
	if err := r.client.Status().Patch(ctx, target, ctrl.MergeFrom(base)); err != nil {
		event.NotifyAppError(ctx, r.client, r.recorder, base, target, err)
		return err
	}

	if target.Status.Phase != base.Status.Phase {
		log.Info(
			"State transition",
			"phase-from", base.Status.Phase,
			"phase-to", target.Status.Phase,
		)
	}

	return nil
}

func getPollingInterval(target *v1alpha1.CamelApp) time.Duration {
	defaultPolling := platform.GetPollingInterval()
	if target.Annotations == nil || target.Annotations[v1alpha1.AppPollingIntervalSecondsAnnotation] == "" {
		return defaultPolling
	}

	interval, err := strconv.Atoi(target.Annotations[v1alpha1.AppPollingIntervalSecondsAnnotation])
	if err == nil {
		return time.Duration(interval) * time.Second
	} else {
		log.Error(err, "could not properly parse polling interval annotation, fallback to default operator value")
	}

	return defaultPolling
}

func getSLIExchangeErrorThreshold(target *v1alpha1.CamelApp) int {
	defaultValue := platform.GetSLIExchangeErrorThreshold()
	if target.Annotations == nil || target.Annotations[v1alpha1.AppSLIExchangeErrorPercentageAnnotation] == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(target.Annotations[v1alpha1.AppSLIExchangeErrorPercentageAnnotation])
	if err == nil {
		return val
	} else {
		log.Error(err, "could not properly parse SLI error percentage, fallback to default operator value")
	}

	return defaultValue
}

func getSLIExchangeWarningThreshold(target *v1alpha1.CamelApp) int {
	defaultValue := platform.GetSLIExchangeWarningThreshold()
	if target.Annotations == nil || target.Annotations[v1alpha1.AppSLIExchangeWarningPercentageAnnotation] == "" {
		return defaultValue
	}

	val, err := strconv.Atoi(target.Annotations[v1alpha1.AppSLIExchangeWarningPercentageAnnotation])
	if err == nil {
		return val
	} else {
		log.Error(err, "could not properly parse SLI warning percentage, fallback to default operator value")
	}

	return defaultValue
}
