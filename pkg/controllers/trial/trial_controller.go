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
	"fmt"
	"strings"
	"time"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	"github.com/alibaba/morphling/pkg/controllers/trial/dbclient"

	"github.com/alibaba/morphling/pkg/controllers/util"
)

const (
	ControllerName     = "trial-controller"
	defaultMetricValue = "0.0"
)

var (
	log = logf.Log.WithName(ControllerName)
)

// NewReconciler returns a new reconcile.ReconcileTrial
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

// Add Watch of resources
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

	//// Watch for changes to service pod
	//err = c.Watch(
	//	&source.Kind{Type: &corev1.Pod{}},
	//	&handler.EnqueueRequestForOwner{
	//		IsController: true,
	//		OwnerType:    &morphlingv1alpha1.Trial{},
	//	})
	//if err != nil {
	//	log.Error(err, "Service Pod watch error")
	//	return err
	//}

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

// TrialReconciler reconciles a Trial object
type ReconcileTrial struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	recorder record.EventRecorder
	dbclient.DBClient
	updateStatusHandler updateStatusFunc
	//collector           *TrialsCollector // collector is a wrapper for experiment metrics.
}

// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tuning.kubedl.io,resources=trials/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch

func (r *ReconcileTrial) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("trial", req.NamespacedName)

	// Fetch the trial instance
	original := &morphlingv1alpha1.Trial{}
	err := r.Get(context.TODO(), req.NamespacedName, original)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
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
		// Todo (Controller) Should delete this
		//if instance.Status.CompletionTime == nil {
		//	now := metav1.Now()
		//	instance.Status.CompletionTime = &now
		//}
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
		if r.updateStatusHandler != nil {
			err = r.updateStatusHandler(instance)
			if err != nil {
				logger.Error(err, "Update trial instance status error")
				return reconcile.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

// reconcileServiceDeployment reconciles a service deployment
func (r *ReconcileTrial) reconcileServiceDeployment(instance *morphlingv1alpha1.Trial, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Get service deploy
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.GetName(), Namespace: deploy.GetNamespace()}, deploy)
	if err != nil && !util.IsCompletedTrial(instance) {
		// If not created, create the service pod
		if errors.IsNotFound(err) {
			if util.IsCompletedTrial(instance) {
				return nil, nil
			}
			logger.Info("Creating service pod", "name", deploy.GetName())
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				logger.Error(err, "Create service pod error")
				return nil, err
			}

		} else {
			logger.Error(err, "Trial Get error")
			return nil, err
		}
	} else {
		// If service pod has already been created
		if util.IsCompletedTrial(instance) {
			// Delete the service pod upon the completion of the trial
			if deploy.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
				logger.Info("Deleting service pod")
				return nil, nil
			}
			// Delete the service pod
			if err = r.Delete(context.TODO(), deploy, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				logger.Error(err, "Delete service pod error")
				return nil, err
			} else {
				logger.Info("Delete service pod succeeded")
				return nil, nil
			}
		}
	}
	return deploy, nil
}

//reconcileTrial reconcile the trial with core functions
func (r *ReconcileTrial) reconcileTrial(instance *morphlingv1alpha1.Trial) error {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Get desired service, and reconcile it
	service, err := r.DesiredService(instance)
	if err != nil {
		return err
	}
	_, err = r.reconcileService(instance, service)
	if err != nil {
		return err
	}

	// Prepare the service deployment
	desiredDeploy, err := r.getDesiredDeploymentSpec(instance)
	if err != nil {
		logger.Error(err, "Service Pod Spec Get error")
		return err
	}

	// Prepare the client job spec
	desiredJob, err := r.getDesiredJobSpec(instance)
	if err != nil {
		logger.Error(err, "Job Spec Get error")
		return err
	}

	// Reconcile the service deployment
	deployedDeployment, err := r.reconcileServiceDeployment(instance, desiredDeploy)
	if err != nil {
		logger.Error(err, "Reconcile Service Pod error")
		return err
	}

	if deployedDeployment == nil {
		// If the service pod is deleted, reconcile the client job, check if the job need to be deleted
		_, err := r.reconcileJob(instance, desiredJob)
		if err != nil {
			logger.Error(err, "Reconcile Client job error")
			return err
		}
		return nil
	}
	//else {
	//	if instance.Status.ServiceDeployment == nil {
	//		instance.Status.ServiceDeployment = &appsv1.Deployment{}
	//	}
	//	deployedDeployment.DeepCopyInto(instance.Status.ServiceDeployment)
	//}

	ServiceDeploymentCondition := deployedDeployment.Status.Conditions

	// Create client job once the service pod is ready
	if util.IsServiceDeplomentReady(ServiceDeploymentCondition) {
		logger.Info("Service Pod is ready", "name", deployedDeployment.GetName())

		deployedJob, err := r.reconcileJob(instance, desiredJob)
		if err != nil || deployedJob == nil {
			logger.Error(err, "Reconcile Client job error")
			return err
		}
		//else {
		//	if instance.Status.StressTestJob == nil {
		//		instance.Status.StressTestJob = &batchv1.Job{}
		//	}
		//	deployedJob.DeepCopyInto(instance.Status.StressTestJob)
		//}

		// Update trial observation when the job is succeeded.
		jobCondition := deployedJob.Status.Conditions
		if util.IsJobSucceeded(jobCondition) {
			logger.Info("Client Job is Completed", "name", desiredJob.GetName())
			// Update trial observation
			if err = r.UpdateTrialStatusObservation(instance); err != nil {
				logger.Error(err, "Update trial status observation error")
				return err
			}
		}

		if util.IsJobFailed(jobCondition) {
			logger.Info("Client Job is Failed", "name", desiredJob.GetName())
			instance.Status.TrialResult = &morphlingv1alpha1.TrialResult{
				TunableParameters:        nil,
				ObjectiveMetricsObserved: nil,
			}

			instance.Status.TrialResult.TunableParameters = make([]morphlingv1alpha1.ParameterAssignment, 0)
			for _, assignment := range instance.Spec.SamplingResult {
				instance.Status.TrialResult.TunableParameters = append(instance.Status.TrialResult.TunableParameters, morphlingv1alpha1.ParameterAssignment{
					Name:     assignment.Name,
					Value:    assignment.Value,
					Category: assignment.Category,
				})
			}
			instance.Status.TrialResult.ObjectiveMetricsObserved = append(instance.Status.TrialResult.ObjectiveMetricsObserved, morphlingv1alpha1.Metric{
				Name:  instance.Spec.Objective.ObjectiveMetricName,
				Value: defaultMetricValue,
			})
		}

		// Update Trial condition. If Service Pod is ready, the trail condition depends on the Client Job
		r.UpdateTrialStatusCondition(instance, deployedJob, jobCondition)

	} else {
		if util.IsServiceDeplomentFail(ServiceDeploymentCondition) {
			message := "Trial service pod failed"

			objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
			metric := morphlingv1alpha1.Metric{Name: objectiveMetricName, Value: "0.0"}
			instance.Status.TrialResult = &morphlingv1alpha1.TrialResult{}
			instance.Status.TrialResult.ObjectiveMetricsObserved = []morphlingv1alpha1.Metric{metric}

			util.SetConditionTrial(instance, morphlingv1alpha1.TrialFailed, corev1.ConditionTrue, message)
		} else {
			message := "Trial service pod pending"
			util.SetConditionTrial(instance, morphlingv1alpha1.TrialPending, corev1.ConditionTrue, message)
		}
	}
	return nil
}

//reconcileJob reconcile the client job
func (r *ReconcileTrial) reconcileJob(instance *morphlingv1alpha1.Trial, job *batchv1.Job) (*batchv1.Job, error) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		return nil, err
	}
	err := r.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, job)
	if err != nil {
		// If the client job is not created, create it
		if errors.IsNotFound(err) {
			if util.IsCompletedTrial(instance) {
				return nil, nil
			}
			logger.Info("Creating Client", "name", job.GetName())
			time.Sleep(5 * time.Second)
			err = r.Create(context.TODO(), job)
			if err != nil {
				logger.Error(err, "Create Client Job error")
				return nil, err
			}
		} else {
			logger.Error(err, "Trial Get error")
			return nil, err
		}
	} else {
		// If the client job has already been created
		if util.IsCompletedTrial(instance) {
			// Delete the client job upon the completion of the trial
			if err = r.Delete(context.TODO(), job, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				logger.Error(err, "Delete Client error")
				return nil, err
			} else {
				return nil, nil
			}
		}
	}
	return job, nil
}

// getDesiredJobSpec returns a new trial run job from the template on the trial
func (r *ReconcileTrial) getDesiredJobSpec(t *morphlingv1alpha1.Trial) (*batchv1.Job, error) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: t.GetName(), Namespace: t.GetNamespace()})

	job := &batchv1.Job{}
	//job.Name = t.Name + "-client-job"
	//job.Namespace = t.GetNamespace()
	//job.Labels = util.ClientLabels(t)
	job.Labels = make(map[string]string)
	// Start with the job template
	if &t.Spec.ClientTemplate != nil {
		t.Spec.ClientTemplate.Spec.DeepCopyInto(&job.Spec)
		if &t.Spec.ClientTemplate.ObjectMeta != nil {
			t.Spec.ClientTemplate.ObjectMeta.DeepCopyInto(&job.ObjectMeta)
		}
	}
	job.Name = util.GetStressTestJobName(t) //t.Name + "-client-job"
	job.Namespace = t.GetNamespace()
	//job.Labels = util.ClientLabels(t)
	if job.Labels == nil {
		job.Labels = make(map[string]string)
	}
	if &t.Labels != nil {
		for k, v := range t.Labels {
			job.Labels[k] = v
		}
	}
	job.Labels[consts.LabelTrialName] = t.Name

	// The default restart policy for a pod is not acceptable in the context of a job
	if job.Spec.Template.Spec.RestartPolicy == "" {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	}
	//t.Spec.ServiceName = util.GetServiceEndpoint(t)

	// The default backoff limit will restart the trial job which is unlikely to produce desirable results
	if job.Spec.BackoffLimit == nil {
		job.Spec.BackoffLimit = new(int32)
	}

	// Expose the current assignments as environment variables to every container (except the default sleep container added below)
	for i := range job.Spec.Template.Spec.Containers {
		c := &job.Spec.Template.Spec.Containers[i]
		c.Env = AppendJobEnv(t, c.Env)
	}

	if err := controllerutil.SetControllerReference(t, job, r.Scheme); err != nil {
		logger.Error(err, "Set Client Job controller reference error")
		return nil, err
	}
	return job, nil
}

func (r *ReconcileTrial) DesiredService(t *morphlingv1alpha1.Trial) (*corev1.Service, error) {
	ports := []corev1.ServicePort{
		{
			Name: consts.DefaultServicePortName,
			Port: consts.DefaultServicePort,
		},
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetServiceName(t),
			Namespace: t.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: util.ServicePodLabels(t),
			Ports:    ports,
			Type:     corev1.ServiceTypeLoadBalancer,
		},
	}

	// Add owner reference to the service so that it could be GC after the sampling is deleted
	if err := controllerutil.SetControllerReference(t, service, r.Scheme); err != nil {
		return nil, err
	}

	return service, nil
}

func (r *ReconcileTrial) reconcileService(instance *morphlingv1alpha1.Trial, service *corev1.Service) (*corev1.Service, error) {
	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) && !util.IsCompletedTrial(instance) {
		log.Info("Creating Service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return nil, err
	} else {
		if util.IsCompletedTrial(instance) {
			// Delete the service pod upon the completion of the trial
			if foundService.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {

				return nil, nil
			}
			// Delete the service pod
			if err = r.Delete(context.TODO(), foundService, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {

				return nil, err
			} else {
				return nil, nil
			}
		}
	}
	return foundService, nil
}

// getDesiredPodSpec returns a new trial r22un job from the template on the trial
func (r *ReconcileTrial) getDesiredDeploymentSpec(t *morphlingv1alpha1.Trial) (*appsv1.Deployment, error) {

	podTemplate := &corev1.PodTemplateSpec{}
	//podTemplate.Name = t.Name + "-service-pod"
	//podTemplate.Namespace = t.GetNamespace()
	//podTemplate.Labels = util.ServicePodLabels(t)
	podTemplate.Labels = make(map[string]string)
	if &t.Spec.ServicePodTemplate != nil {
		//t.Spec.ServicePodTemplate.ObjectMeta.DeepCopyInto(&pod.ObjectMeta)
		t.Spec.ServicePodTemplate.Template.Spec.DeepCopyInto(&podTemplate.Spec)
		if &t.Spec.ServicePodTemplate.Template.ObjectMeta != nil {
			t.Spec.ServicePodTemplate.Template.ObjectMeta.DeepCopyInto(&podTemplate.ObjectMeta)
		}
	}
	podTemplate.Name = t.Name + "-service-pod"
	podTemplate.Namespace = t.GetNamespace()

	if podTemplate.Labels == nil {
		podTemplate.Labels = make(map[string]string)
		log.Info("podTemplate.Labels =make(map[string]string)")
	}

	// podTemplate.Labels = util.ServicePodLabels(t)
	if &t.Labels != nil {
		for k, v := range t.Labels {
			podTemplate.Labels[k] = v
		}
	}

	podTemplate.Labels["trial"] = t.Name
	podTemplate.Labels[consts.LabelDeploymentName] = util.GetServiceDeploymentName(t)

	for i := range podTemplate.Spec.Containers {
		c := &podTemplate.Spec.Containers[i]
		c.Env, c.Args, c.Resources = AppendAssignmentEnv(t, c.Env, c.Args, c.Resources)
		c.Ports = []corev1.ContainerPort{
			{
				Name:          consts.DefaultServicePortName,
				ContainerPort: consts.DefaultServicePort,
			},
		}
	}
	//dealine := int32(900)
	d := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetServiceDeploymentName(t),
			Namespace:   t.GetNamespace(),
			Labels:      util.ServiceDeploymentLabels(t),
			Annotations: t.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: util.ServicePodLabels(t),
			},
			Template: *podTemplate,
			//ProgressDeadlineSeconds: &dealine,
		},
	}
	if t.Spec.ServiceProgressDeadline != nil {
		d.Spec.ProgressDeadlineSeconds = t.Spec.ServiceProgressDeadline
	}

	if err := controllerutil.SetControllerReference(t, d, r.Scheme); err != nil {
		return nil, err
	}

	return d, nil
}

// AppendAssignmentEnv appends an environment variable for service pods
func AppendAssignmentEnv(t *morphlingv1alpha1.Trial, env []corev1.EnvVar, args []string, resources corev1.ResourceRequirements) ([]corev1.EnvVar, []string, corev1.ResourceRequirements) {
	//resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
	//resources.Requests = make(map[corev1.ResourceName]resource.Quantity)

	for _, a := range t.Spec.SamplingResult {
		if a.Category == morphlingv1alpha1.CategoryEnv {
			name := strings.ReplaceAll(strings.ToUpper(a.Name), ".", "_")
			env = append(env, corev1.EnvVar{Name: name, Value: fmt.Sprintf(a.Value)})
		} else if a.Category == morphlingv1alpha1.CategoryArgs {
			args = append(args, fmt.Sprintf(a.Value))
		} else if a.Category == morphlingv1alpha1.CategoryResource {
			var resourceClass = corev1.ResourceCPU
			switch a.Name {
			case "cpu":
				resourceClass = corev1.ResourceCPU
				break
			case "memory":
				resourceClass = corev1.ResourceMemory
				break
			case "storage":
				resourceClass = corev1.ResourceStorage
				break
			case "nvidia.com/gpu":
				resourceClass = "nvidia.com/gpu"
				break
			case "nvidia.com/gpumem":
				resourceClass = "nvidia.com/gpumem"
				break
			default:
				resourceClass = corev1.ResourceEphemeralStorage
			}
			if resources.Limits == nil {
				resources.Limits = make(map[corev1.ResourceName]resource.Quantity)
			}
			if resources.Requests == nil {
				resources.Requests = make(map[corev1.ResourceName]resource.Quantity)
			}
			resources.Limits[resourceClass] = resource.MustParse(a.Value)
			resources.Requests[resourceClass] = resource.MustParse(a.Value)
		}
	}
	return env, args, resources
}

// AppendJobEnv appends an environment variable for jobs
func AppendJobEnv(t *morphlingv1alpha1.Trial, env []corev1.EnvVar) []corev1.EnvVar {

	env = append(env, corev1.EnvVar{Name: "RequestTemplate", Value: fmt.Sprintf(t.Spec.RequestTemplate)})
	env = append(env, corev1.EnvVar{Name: "ServiceName", Value: util.GetServiceEndpoint(t)}) //fmt.Sprintf(t.Spec.ServiceName)}
	env = append(env, corev1.EnvVar{Name: "TrialName", Value: fmt.Sprintf(t.Name)})
	env = append(env, corev1.EnvVar{Name: "Namespace", Value: fmt.Sprintf(t.Namespace)})
	for _, cat := range t.Spec.SamplingResult {
		if cat.Category == morphlingv1alpha1.CategoryEnv && (cat.Name == "BATCH_SIZE" || cat.Name == "MODEL_NAME") {
			env = append(env, corev1.EnvVar{Name: cat.Name, Value: fmt.Sprintf(cat.Value)})
		}
	}
	return env
}
