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

package trial

import (
	"context"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/trial/dbclient"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
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

	"github.com/alibaba/morphling/pkg/controllers/util"
)

const (
	ControllerName     = "trial-controller"
	defaultMetricValue = "0.0"
)

var (
	log = logf.Log.WithName(ControllerName)
)

func NewReconciler(mgr manager.Manager) *ReconcileTrial {
	r := &ReconcileTrial{
		Client:   mgr.GetClient(),
		Scheme:   mgr.GetScheme(),
		DBClient: dbclient.NewTrialDBClient(),
		recorder: mgr.GetEventRecorderFor(ControllerName),
		Log:      logf.Log.WithName(ControllerName),
	}
	r.updateStatusHandler = r.updateStatus
	return r
}

func (r *ReconcileTrial) updateStatus(instance *morphlingv1alpha1.Trial) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err) {
			return err
		}
	}
	return nil
}

func (r *ReconcileTrial) SetupWithManager(mgr ctrl.Manager) error {
	c, err := controller.New(ControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		log.Error(err, "Failed to create experiment controller")
		return err
	}
	if err = addWatch(c); err != nil {
		log.Error(err, "Trial watch failed")
		return err
	}
	log.Info("Experiment controller created")
	return nil
}

func addWatch(c controller.Controller) error {
	// Watch for changes to trial
	err := c.Watch(&source.Kind{Type: &morphlingv1alpha1.Trial{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "Trial watch error")
		return err
	}
	// Watch for changes to client job
	err = c.Watch(
		&source.Kind{Type: &batchv1.Job{}},
		&handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &morphlingv1alpha1.Trial{},
		})
	if err != nil {
		log.Error(err, "Client Job watch error")
		return err
	}
	// Watch for changes to service deployment
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &morphlingv1alpha1.Trial{},
	})
	if err != nil {
		log.Error(err, "Service Deployment watch error")
		return err
	}
	// Watch for changes to service
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &morphlingv1alpha1.Trial{},
	})
	if err != nil {
		log.Error(err, "Service Service watch error")
		return err
	}
	return nil
}

type ReconcileTrial struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	dbclient.DBClient
	updateStatusHandler updateStatusFunc
}

// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

// Reconcile reads that state of the cluster for a trial object and makes changes based on the state read
func (r *ReconcileTrial) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("trial", req.NamespacedName)

	// Fetch the trial instance
	original := &morphlingv1alpha1.Trial{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("try to get trial, but it has been deleted", "key", req.String())
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Trial Get error")
		return reconcile.Result{}, err
	}

	instance := original.DeepCopy()
	// If not created, create the trial
	if !util.IsCreatedTrial(instance) {
		if instance.Status.StartTime == nil {
			now := metav1.Now()
			instance.Status.StartTime = &now
		}
		msg := "Trial is created"
		util.MarkTrialStatusCreatedTrial(instance, msg)
	} else {
		// Reconcile trial
		err := r.reconcileTrial(instance)
		if err != nil {
			logger.Error(err, "Reconcile trial error")
			return reconcile.Result{}, err
		}
	}
	// Update trial status
	if !equality.Semantic.DeepEqual(original.Status, instance.Status) {
		err = r.updateStatusHandler(instance)
		if err != nil {
			logger.Error(err, "Update trial instance status error")
			return reconcile.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

//reconcileTrial reconcile the trial with core functions
func (r *ReconcileTrial) reconcileTrial(instance *morphlingv1alpha1.Trial) error {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Get desired service, and reconcile it
	service, err := r.getDesiredService(instance)
	if err != nil {
		logger.Error(err, "ML service get error")
		return err
	}
	// Get desired deployment
	desiredDeploy, err := r.getDesiredDeploymentSpec(instance)
	if err != nil {
		logger.Error(err, "Service deployment construction error")
		return err
	}
	// Get desired client job
	desiredJob, err := r.getDesiredJobSpec(instance)
	if err != nil {
		logger.Error(err, "Client-side job construction error")
		return err
	}

	// Reconcile the service
	err = r.reconcileService(instance, service)
	if err != nil {
		logger.Error(err, "Reconcile ML service error")
		return err
	}
	// Reconcile the deployment
	deployedDeployment, err := r.reconcileServiceDeployment(instance, desiredDeploy)
	if err != nil {
		logger.Error(err, "Reconcile ML deployment error")
		return err
	}
	// Check if the job need to be deleted
	if deployedDeployment == nil {
		_, err := r.reconcileJob(instance, desiredJob)
		if err != nil {
			logger.Error(err, "Reconcile client-side job error")
			return err
		}
		return nil
	}

	deployedJob := &batchv1.Job{}
	// Create client job
	if util.IsServiceDeplomentReady(deployedDeployment.Status.Conditions) {
		logger.Info("Service Pod is ready", "name", deployedDeployment.GetName())
		deployedJob, err = r.reconcileJob(instance, desiredJob)
		if err != nil || deployedJob == nil {
			logger.Error(err, "Reconcile client-side job error")
			return err
		}
	}
	// Update trial status (conditions and results)
	if util.IsServiceDeplomentReady(deployedDeployment.Status.Conditions) {
		if err = r.UpdateTrialStatusByClientJob(instance, deployedJob); err != nil {
			logger.Error(err, "Update trial status by client-side job condition error")
			return err
		}
	} else {
		r.UpdateTrialStatusByServiceDeployment(instance, deployedDeployment)
	}
	return nil
}
