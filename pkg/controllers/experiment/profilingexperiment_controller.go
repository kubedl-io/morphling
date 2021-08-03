/*
Copyright 2021 The Alibaba Authors.

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

package experiment

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/experiment/sampling_client"
	"github.com/alibaba/morphling/pkg/controllers/util"
)

const (
	ControllerName = "experiment-controller"
)

var (
	log = logf.Log.WithName(ControllerName)
)

// NewReconciler returns a new reconcile.Reconciler
func NewReconciler(mgr manager.Manager) *ProfilingExperimentReconciler {
	r := &ProfilingExperimentReconciler{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
	}
	r.Sampling = sampling_client.New(mgr.GetScheme(), mgr.GetClient())
	r.updateStatusHandler = r.updateStatus
	return r
}

func (r *ProfilingExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "Failed to create experiment controller")
		return err
	}
	// Add watch
	if err = addWatch(c); err != nil {
		log.Error(err, "Trial watch failed")
		return err
	}
	log.Info("Experiment controller created")
	return nil
}

// Add Watch of resources
func addWatch(c controller.Controller) error {
	// Watch for changes to Experiment
	err := c.Watch(&source.Kind{Type: &morphlingv1alpha1.ProfilingExperiment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Experiment watch failed")
		return err
	}

	// Watch for trials for the experiments
	err = c.Watch(
		&source.Kind{Type: &morphlingv1alpha1.Trial{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &morphlingv1alpha1.ProfilingExperiment{},
		})
	if err != nil {
		log.Error(err, "Trial watch failed")
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ProfilingExperimentReconciler{}

type updateStatusFunc func(instance *morphlingv1alpha1.ProfilingExperiment) error

// ProfilingExperimentReconciler reconciles a ProfilingExperiment object
type ProfilingExperimentReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	sampling_client.Sampling
	updateStatusHandler updateStatusFunc
}

// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=profilingexperiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=profilingexperiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a trial object and makes changes based on the state read
func (r *ProfilingExperimentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := log.WithValues("Experiment", req.NamespacedName)

	// Fetch the profiling experiment instance
	original := &morphlingv1alpha1.ProfilingExperiment{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return. Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		logger.Error(err, "Profiling experiment get error")
		return reconcile.Result{}, err
	}
	instance := original.DeepCopy()

	// Cleanup upon completion
	if util.IsCompletedExperiment(instance) {
		if !util.HasRunningTrials(instance) {
			return reconcile.Result{}, nil
		}
	}
	if !util.IsCreatedExperiment(instance) {
		// Create the experiment
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		message := "Experiment is created"
		util.MarkExperimentStatusCreated(instance, message)
	} else {
		// Reconcile experiment
		err := r.ReconcileExperiment(instance)
		if err != nil {
			logger.Error(err, "Reconcile experiment error")
			r.recorder.Eventf(instance, corev1.EventTypeWarning, "ReconcileFailed", "Failed to reconcile: %v", err)
			return reconcile.Result{}, err
		}
	}

	// Update experiment status
	if !equality.Semantic.DeepEqual(original.Status, instance.Status) {
		err = r.updateStatusHandler(instance)
		if err != nil {
			logger.Error(err, "Update experiment status error")
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// ReconcileExperiment is the main reconcile loop.
func (r *ProfilingExperimentReconciler) ReconcileExperiment(instance *morphlingv1alpha1.ProfilingExperiment) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Fetch trials
	trials, err := r.fetchTrials(instance)
	if err != nil {
		logger.Error(err, "Fetch trials error")
	}

	// Update trials results
	if len(trials.Items) > 0 {
		updateTrialsSummary(instance, trials)
	}

	// Update experiment status
	if !util.IsCompletedExperiment(instance) {
		updateExperimentStatusCondition(instance)
	}

	// Reconcile trials
	if !util.IsCompletedExperiment(instance) {
		err := r.ReconcileTrials(instance, trials.Items)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ProfilingExperimentReconciler) updateStatus(instance *morphlingv1alpha1.ProfilingExperiment) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err) {
			return err
		}
	}
	return nil
}
