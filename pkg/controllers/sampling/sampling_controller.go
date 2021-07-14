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

package sampling

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/sampling/composer"
	samplingclient "github.com/alibaba/morphling/pkg/controllers/sampling/sampling_client"
	"github.com/alibaba/morphling/pkg/controllers/util"
)

const (
	ControllerName = "sampling-controller"
)

var log = logf.Log.WithName(ControllerName)

// Add creates a new Sampling Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &SamplingReconciler{
		Client:         mgr.GetClient(),
		SamplingClient: samplingclient.New(mgr.GetClient()),
		Scheme:         mgr.GetScheme(),
		Composer:       composer.New(mgr),
	}
}

// add adds watch
func add(mgr manager.Manager, r reconcile.Reconciler) error {

	log.Info("Create sampling controller ing")
	// Create a new controller
	c, err := controller.New("sampling-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &morphlingv1alpha1.Sampling{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &morphlingv1alpha1.Sampling{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &morphlingv1alpha1.Sampling{},
	})
	if err != nil {
		return err
	}
	log.Info("Sampling controller created")
	return nil
}

var _ reconcile.Reconciler = &SamplingReconciler{}

// SamplingReconciler reconciles a Sampling object
type SamplingReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	samplingclient.SamplingClient
	composer.Composer
}

// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=samplings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=samplings/status,verbs=get;update;patch

func (r *SamplingReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := log.WithValues("Sampling", req.NamespacedName)

	// Fetch the Sampling instance
	oldS := &morphlingv1alpha1.Sampling{}
	err := r.Get(context.TODO(), req.NamespacedName, oldS)
	if err != nil {
		if errors.IsNotFound(err) {
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	instance := oldS.DeepCopy()

	// Sampling is marked as succeeded once the experiment is completed
	// TODO: We comment the following parts, as we are using long-running algorithm server
	if util.IsSucceededSampling(instance) {
		err = r.deleteDeployment(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = r.deleteService(instance)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// If not created, create the sampling instance
	if !util.IsCreatedSampling(instance) {
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		msg := "Sampling is created"
		util.MarkSamplingStatusCreated(instance, msg)
	} else {
		// Reconcile the sampling functions
		err := r.ReconcileSampling(instance)
		if err != nil {
			// Try updating just the status condition when possible
			// Status conditions might need to be  updated even in error
			_ = r.updateStatusCondition(instance, oldS)
			logger.Error(err, "Reconcile Sampling error")
			return reconcile.Result{}, err
		}
	}

	if err := r.updateStatus(instance, oldS); err != nil {
		return reconcile.Result{}, err
	}
	return ctrl.Result{}, nil
}

// ReconcileSampling is the main reconcile loop.
func (r *SamplingReconciler) ReconcileSampling(instance *morphlingv1alpha1.Sampling) error {
	logger := log.WithValues("Sampling", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Get desired sampling service, and reconcile it
	service, err := r.DesiredService(instance)
	if err != nil {
		return err
	}
	_, err = r.reconcileService(service)
	if err != nil {
		return err
	}

	// Get desired sampling deployment, and reconcile it
	deploy, err := r.DesiredDeployment(instance)
	if err != nil {
		return err
	}
	if foundDeploy, err := r.reconcileDeployment(deploy); err != nil {
		return err
	} else {
		if isReady := r.checkDeploymentReady(foundDeploy); isReady != true {
			msg := "Sampling deployment is not ready"
			util.MarkSamplingStatusDeploymentReady(instance, corev1.ConditionFalse, SamplingDeploymentNotReady, msg)
			return nil
		} else {
			msg := "Sampling deployment is ready"
			util.MarkSamplingStatusDeploymentReady(instance, corev1.ConditionTrue, SamplingDeploymentReady, msg)
		}

	}

	experiment := &morphlingv1alpha1.ProfilingExperiment{}
	trials := &morphlingv1alpha1.TrialList{}
	// Fetch experiments
	if err := r.Get(context.TODO(), types.NamespacedName{
		Name:      instance.Name,
		Namespace: instance.Namespace,
	}, experiment); err != nil {
		logger.Error(err, "Experiment get error when reconciling sampling")
		return err
	}
	// Fetch trials
	lo := &client.ListOptions{}
	sel := labels.SelectorFromSet(util.TrialLabels(experiment))
	lo.LabelSelector = sel
	lo.Namespace = instance.Namespace
	if err := r.List(context.TODO(), trials, lo); err != nil {
		logger.Error(err, "Trial list get error when reconciling sampling")
		return err
	}

	// Check status of the sampling instance
	if !util.IsRunningSampling(instance) && !util.IsSucceededSampling(instance) {
		if err = r.ValidateAlgorithmSettings(instance, experiment); err != nil {
			logger.Error(err, "Marking sampling failed as algorithm settings validation failed")
			msg := fmt.Sprintf("Validation failed: %v", err)
			util.MarkSamplingStatusFailed(instance, SamplingFailedReason, msg)
			// return nil since it is a terminal condition
			return nil
		}
		msg := "Sampling is running"
		util.MarkSamplingStatusRunning(instance, SamplingRunningReason, msg)
	}
	logger.Info("Sync assignments", "suggestions", instance.Spec.NumSamplingsRequested)

	// Load sampling result in instance.Status.SamplingResult
	if err = r.SyncAssignments(instance, experiment, trials.Items); err != nil {
		return err
	}

	return nil
}

// checkDeploymentReady checks if the algorithm server is ready
func (r *SamplingReconciler) checkDeploymentReady(deploy *appsv1.Deployment) bool {
	if deploy == nil {
		return false
	} else {
		for _, cond := range deploy.Status.Conditions {
			if cond.Type == appsv1.DeploymentAvailable && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func (r *SamplingReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&morphlingv1alpha1.ProfilingExperiment{}).
		Complete(r)
}
