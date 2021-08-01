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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/experiment/sampling"
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
	r.Sampling = sampling.New(mgr.GetScheme(), mgr.GetClient())
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
	sampling.Sampling
	updateStatusHandler updateStatusFunc
}

// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=profilingexperiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=profilingexperiments/status,verbs=get;update;patch

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
		//// Terminate sampling, but not delete it
		//err := r.terminateSuggestion(instance)
		//if err != nil {
		//	logger.Error(err, "Terminate Suggestion error")
		//	return reconcile.Result{}, err
		//}
		// If experiment is completed with no running trials, stop reconcile
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
		// Todo (Controller) Should delete this
		//if instance.Status.CompletionTime == nil {
		//	now := metav1.Now()
		//	instance.Status.CompletionTime = &now
		//}
		message := "Experiment is created"
		util.MarkExperimentStatusCreated(instance, message)
	} else {
		// Reconcile experiment
		err := r.ReconcileExperiment(instance)
		if err != nil {
			logger.Error(err, "Reconcile experiment error")
			r.recorder.Eventf(instance,
				corev1.EventTypeWarning, "ReconcileFailed",
				"Failed to reconcile: %v", err)
			return reconcile.Result{}, err
		}
	}

	// Update experiment status
	if !equality.Semantic.DeepEqual(original.Status, instance.Status) {
		//tmp := &morphlingv1alpha1.ProfilingExperiment{}
		//err := r.Get(context.TODO(), req.NamespacedName, tmp)
		//
		//if original.GetResourceVersion() != tmp.GetResourceVersion() {
		//	logger.Info("Update experiment skipped")
		//	return reconcile.Result{}, nil
		//}

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
	trials := &morphlingv1alpha1.TrialList{}
	experimentLabels := map[string]string{consts.LabelExperimentName: instance.Name}
	lo := &client.ListOptions{}
	sel := labels.SelectorFromSet(experimentLabels)
	lo.LabelSelector = sel
	lo.Namespace = instance.Namespace
	if err := r.List(context.TODO(), trials, lo); err != nil {
		logger.Error(err, "Trial list error")
		return err
	}

	// Update trials summary and experiment status
	if len(trials.Items) > 0 {
		if err := UpdateExperimentStatus(instance, trials); err != nil {
			logger.Error(err, "Update experiment status error")
			return err
		}
	}

	// Check if the experiment is completed
	reconcileRequired := !util.IsCompletedExperiment(instance)
	if reconcileRequired {
		err := r.ReconcileTrials(instance, trials.Items)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReconcileTrials syncs trials.
func (r *ProfilingExperimentReconciler) ReconcileTrials(instance *morphlingv1alpha1.ProfilingExperiment, trials []morphlingv1alpha1.Trial) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	parallelCount := *instance.Spec.Parallelism
	activeCount := instance.Status.TrialsPending + instance.Status.TrialsRunning
	completedCount := instance.Status.TrialsSucceeded + instance.Status.TrialsFailed + instance.Status.TrialsKilled

	// If new trials are requested
	if activeCount < parallelCount {
		var requiredActiveCount int32
		if instance.Spec.MaxNumTrials == nil {
			requiredActiveCount = parallelCount
		} else {
			requiredActiveCount = *instance.Spec.MaxNumTrials - completedCount
			if requiredActiveCount > parallelCount {
				requiredActiveCount = parallelCount
			}
		}

		// addCount is the number of new trials to be started
		addCount := requiredActiveCount - activeCount
		if addCount < 0 {
			logger.Info("Invalid setting", "requiredActiveCount", requiredActiveCount, "MaxTrialCount",
				*instance.Spec.MaxNumTrials, "CompletedCount", completedCount)
			addCount = 0
		}
		logger.Info("Statistics",
			"requiredActiveCount", requiredActiveCount,
			"activeCount", activeCount,
			"completedCount", completedCount,
		)

		// Create "addCount" number of trials
		if addCount > 0 {
			logger.Info("CreateTrials", "addCount", addCount)
			if err := r.createTrials(instance, trials, addCount); err != nil {
				logger.Error(err, "Create trials error")
				return err
			}
		}
	}

	return nil
}

//createTrials gets sampling results and creates new trials
func (r *ProfilingExperimentReconciler) createTrials(instance *morphlingv1alpha1.ProfilingExperiment, trialList []morphlingv1alpha1.Trial, addCount int32) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	//// Fetch sampling results
	currentCount := int32(len(trialList))
	//trials, err := r.ReconcileSamplings(instance, currentCount, addCount)
	assignments, err := r.GetSamplings(addCount, instance, currentCount, trialList)

	if err != nil {
		logger.Error(err, "Get samplings error")
		return err
	}

	// Create new trials w.r.t. sampling results
	for _, assignment := range assignments {
		if err = r.createTrialInstance(instance, &assignment); err != nil {
			logger.Error(err, "Create trial instance error", "trial", assignment)
			continue
		}
	}
	return nil
}

// createTrialInstance creates a new trial instance
func (r *ProfilingExperimentReconciler) createTrialInstance(expInstance *morphlingv1alpha1.ProfilingExperiment, trialAssignment *morphlingv1alpha1.TrialAssignment) error {
	logger := log.WithValues("Experiment", types.NamespacedName{Name: expInstance.GetName(), Namespace: expInstance.GetNamespace()})

	// Init a new trial instance
	trial := &morphlingv1alpha1.Trial{}
	trial.Name = trialAssignment.Name // grid id
	trial.Namespace = expInstance.GetNamespace()
	trial.Labels = util.TrialLabels(expInstance)
	if err := controllerutil.SetControllerReference(expInstance, trial, r.Scheme); err != nil {
		logger.Error(err, "Set controller reference error")
		return err
	}

	// Set parameters for the new trial
	trial.Spec.ServiceProgressDeadline = expInstance.Spec.ServiceProgressDeadline
	trial.Spec.Objective = expInstance.Spec.Objective
	trial.Spec.RequestTemplate = expInstance.Spec.RequestTemplate
	expInstance.Spec.ServicePodTemplate.DeepCopyInto(&trial.Spec.ServicePodTemplate)
	expInstance.Spec.ClientTemplate.DeepCopyInto(&trial.Spec.ClientTemplate)
	trial.Spec.SamplingResult = make([]morphlingv1alpha1.ParameterAssignment, 0)
	for _, pa := range trialAssignment.ParameterAssignments {
		trial.Spec.SamplingResult = append(trial.Spec.SamplingResult, morphlingv1alpha1.ParameterAssignment{
			Name:     pa.Name,
			Value:    pa.Value,
			Category: pa.Category,
			//todo: Category
		})
	}

	// Create the new trial
	if err := r.Create(context.TODO(), trial); err != nil {
		logger.Error(err, "Trial create error", "Trial name", trial.Name)
		return err
	}
	return nil
}

func (r *ProfilingExperimentReconciler) updateStatus(instance *morphlingv1alpha1.ProfilingExperiment) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err){return err}
	}
	return nil
}
