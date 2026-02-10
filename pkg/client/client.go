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

package client

import (
	"sync"

	"k8s.io/client-go/scale"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"

	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	camel "github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/clientset/versioned"
	camelv1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/client/camel/clientset/versioned/typed/camel/v1alpha1"
)

var newClientMutex sync.Mutex

// Client is an abstraction for a k8s client.
type Client interface {
	ctrl.Client
	kubernetes.Interface
	CamelV1alpha1() camelv1alpha1.CamelV1alpha1Interface
	GetScheme() *runtime.Scheme
	GetConfig() *rest.Config
	ServerOrClientSideApplier() ServerOrClientSideApplier
	ScalesClient() (scale.ScalesGetter, error)
}

// Injectable identifies objects that can receive a Client.
type Injectable interface {
	InjectClient(client Client)
}

// Provider is used to provide a new instance of the Client each time it's required.
type Provider struct {
	Get func() (Client, error)
}

type defaultClient struct {
	ctrl.Client
	kubernetes.Interface
	camel  camel.Interface
	scheme *runtime.Scheme
	config *rest.Config
}

// Check interface compliance.
var _ Client = &defaultClient{}

func (c *defaultClient) CamelV1alpha1() camelv1alpha1.CamelV1alpha1Interface {
	return c.camel.CamelV1alpha1()
}

func (c *defaultClient) GetScheme() *runtime.Scheme {
	return c.scheme
}

func (c *defaultClient) GetConfig() *rest.Config {
	return c.config
}

// NewClientWithConfig creates a new k8s client that can be used from outside or in the cluster.
func NewClientWithConfig(fastDiscovery bool, cfg *rest.Config) (Client, error) {

	// The below call to apis.AddToScheme is not thread safe in the k8s API
	// We try to synchronize here across all k8s clients
	// https://github.com/apache/camel-dashboard/issues/5315
	newClientMutex.Lock()
	defer newClientMutex.Unlock()

	var err error
	clientScheme := scheme.Scheme
	if !clientScheme.IsVersionRegistered(v1alpha1.SchemeGroupVersion) {
		// Setup Scheme for all resources
		err = apis.AddToScheme(clientScheme)
		if err != nil {
			return nil, err
		}
	}

	var clientset kubernetes.Interface
	if clientset, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}

	var camelClientset camel.Interface
	if camelClientset, err = camel.NewForConfig(cfg); err != nil {
		return nil, err
	}

	var mapper meta.RESTMapper
	if fastDiscovery {
		mapper = newFastDiscoveryRESTMapper(cfg)
	}

	// Create a new client to avoid using cache (enabled by default with controller-runtime client)
	clientOptions := ctrl.Options{
		Scheme: clientScheme,
		Mapper: mapper,
	}
	dynClient, err := ctrl.New(cfg, clientOptions)
	if err != nil {
		return nil, err
	}

	return &defaultClient{
		Client:    dynClient,
		Interface: clientset,
		camel:     camelClientset,
		scheme:    clientOptions.Scheme,
		config:    cfg,
	}, nil
}

// FromManager creates a new k8s client from a manager object.
func FromManager(manager manager.Manager) (Client, error) {
	var err error
	var clientset kubernetes.Interface
	if clientset, err = kubernetes.NewForConfig(manager.GetConfig()); err != nil {
		return nil, err
	}
	var camelClientset camel.Interface
	if camelClientset, err = camel.NewForConfig(manager.GetConfig()); err != nil {
		return nil, err
	}

	return &defaultClient{
		Client:    manager.GetClient(),
		Interface: clientset,
		camel:     camelClientset,
		scheme:    manager.GetScheme(),
		config:    manager.GetConfig(),
	}, nil
}
