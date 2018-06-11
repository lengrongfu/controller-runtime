/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"fmt"

	"github.com/kubernetes-sigs/controller-runtime/pkg/eventhandler"
	"github.com/kubernetes-sigs/controller-runtime/pkg/internal/controller"
	"github.com/kubernetes-sigs/controller-runtime/pkg/manager"
	"github.com/kubernetes-sigs/controller-runtime/pkg/predicate"
	"github.com/kubernetes-sigs/controller-runtime/pkg/reconcile"
	logf "github.com/kubernetes-sigs/controller-runtime/pkg/runtime/log"
	"github.com/kubernetes-sigs/controller-runtime/pkg/source"
	"k8s.io/client-go/util/workqueue"
)

var log = logf.KBLog.WithName("controller")

// Options are the arguments for creating a new Controller
type Options struct {
	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.
	MaxConcurrentReconciles int

	// Reconcile reconciles an object
	Reconcile reconcile.Reconcile
}

// Controller is a work queue that watches for changes to objects (i.e. Create / Update / Delete events) and
// then reconciles an object (i.e. make changes to ensure the system state matches what is specified in the object).
type Controller interface {
	// Watch takes events provided by a Source and uses the EventHandler to enqueue reconcile.Requests in
	// response to the events.
	//
	// Watch may be provided one or more Predicates to filter events before they are given to the EventHandler.
	// Events will be passed to the EventHandler iff all provided Predicates evaluate to true.
	Watch(src source.Source, evthdler eventhandler.EventHandler, prct ...predicate.Predicate) error

	// Start starts the controller.  Start blocks until stop is closed or a controller has an error starting.
	Start(stop <-chan struct{}) error
}

func New(name string, mrg manager.Manager, options Options) (Controller, error) {
	if options.Reconcile == nil {
		return nil, fmt.Errorf("must specify Reconcile")
	}

	if len(name) == 0 {
		return nil, fmt.Errorf("must specify Name for Controller")
	}

	if options.MaxConcurrentReconciles <= 0 {
		options.MaxConcurrentReconciles = 1
	}

	// Inject dependencies into Reconcile
	if err := mrg.SetFields(options.Reconcile); err != nil {
		return nil, err
	}

	// Create controller with dependencies set
	c := &controller.Controller{
		Reconcile: options.Reconcile,
		Cache:     mrg.GetCache(),
		Config:    mrg.GetConfig(),
		Scheme:    mrg.GetScheme(),
		Client:    mrg.GetClient(),
		Queue:     workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), name),
		MaxConcurrentReconciles: options.MaxConcurrentReconciles,
		Name: name,
	}

	// Add the controller as a Manager componentsw
	mrg.Add(c)
	return c, nil
}
