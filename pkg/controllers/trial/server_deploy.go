package trial

import (
	"context"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
)

// getDesiredService returns a new k8s service for ML service test
func (r *ReconcileTrial) getDesiredService(t *morphlingv1alpha1.Trial) (*corev1.Service, error) {
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetServiceName(t),
			Namespace: t.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: util.ServicePodLabels(t),
			Ports:    []corev1.ServicePort{
				{
					Name: consts.DefaultServicePortName,
					Port: consts.DefaultServicePort,
				},
			},
			Type:     corev1.ServiceTypeClusterIP,
		},
	}
	// ToDo: SetControllerReference here is useless, as the controller delete svc upon trial completion
	// Add owner reference to the service so that it could be GC
	if err := controllerutil.SetControllerReference(t, service, r.Scheme); err != nil {
		return nil, err
	}
	return service, nil
}

// reconcileService reconciles a k8s service for ML service
func (r *ReconcileTrial) reconcileService(instance *morphlingv1alpha1.Trial, service *corev1.Service) error {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	foundService := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	// Create svc
	if err != nil && errors.IsNotFound(err) && !util.IsCompletedTrial(instance) {
		logger.Info("Creating ML service", "namespace", service.Namespace, "name", service.Name)
		err = r.Create(context.TODO(), service)
		return err
	}
	// Delete svc
	if util.IsCompletedTrial(instance) {
		// Delete svc upon trial completions
		if foundService.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
			logger.Info("Deleting ML service")
			return nil
		}
		if err = r.Delete(context.TODO(), foundService, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			if errors.IsNotFound(err) {
				logger.Info("Delete ML service operation is redundant")
				return nil
			}
			return err
		}
	}
	return nil
}

// getDesiredPodSpec returns a new deployment containing the ML service under test
func (r *ReconcileTrial) getDesiredDeploymentSpec(instance *morphlingv1alpha1.Trial) (*appsv1.Deployment, error) {
	// Prepare podTemplate and embed tunable parameters
	podTemplate := &corev1.PodTemplateSpec{}
	if &instance.Spec.ServicePodTemplate != nil {
		instance.Spec.ServicePodTemplate.Template.Spec.DeepCopyInto(&podTemplate.Spec)
	}
	podTemplate.Labels = util.ServicePodLabels(instance)
	for i := range podTemplate.Spec.Containers {
		c := &podTemplate.Spec.Containers[i]
		c.Env, c.Args, c.Resources = appendServiceEnv(instance, c.Env, c.Args, c.Resources)
	}
	// Prepare k8s deployment
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetServiceDeploymentName(instance),
			Namespace:   instance.GetNamespace(),
			Labels:      util.ServiceDeploymentLabels(instance),
			Annotations: instance.Annotations,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: util.ServicePodLabels(instance),},
			Template: *podTemplate,
		},
	}
	if instance.Spec.ServiceProgressDeadline != nil {
		deploy.Spec.ProgressDeadlineSeconds = instance.Spec.ServiceProgressDeadline
	}
	// ToDo: SetControllerReference here is useless, as the controller delete svc upon trial completion
	// Add owner reference to the service so that it could be GC
	if err := controllerutil.SetControllerReference(instance, deploy, r.Scheme); err != nil {
		return nil, err
	}
	return deploy, nil
}

// reconcileServiceDeployment reconciles the ML deployment containing the ML service under test
func (r *ReconcileTrial) reconcileServiceDeployment(instance *morphlingv1alpha1.Trial, deploy *appsv1.Deployment) (*appsv1.Deployment, error) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.GetName(), Namespace: deploy.GetNamespace()}, deploy)
	if err != nil && !util.IsCompletedTrial(instance) {
		// If not created, create the service deployment
		if errors.IsNotFound(err) {
			if util.IsCompletedTrial(instance) {
				return nil, nil
			}

			logger.Info("Creating ML service deployment", "name", deploy.GetName())
			err = r.Create(context.TODO(), deploy)
			if err != nil {
				logger.Error(err, "Create service deployment error", "name", deploy.GetName())
				return nil, err
			}
		} else {
			logger.Error(err, "Get service deployment error", "name", deploy.GetName())
			return nil, err
		}
	} else {
		if util.IsCompletedTrial(instance) {
			if deploy.ObjectMeta.DeletionTimestamp != nil || errors.IsNotFound(err) {
				logger.Info("Deleting ML deployment", "name", deploy.GetName())
				return nil, nil
			}
			// // Delete ML deployments upon trial completions
			if err = r.Delete(context.TODO(), deploy, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
				if errors.IsNotFound(err) {
					logger.Info("Delete ML deployment operation is redundant", "name", deploy.GetName())
					return nil, nil
				}
				logger.Error(err, "Delete ML deployment error","name", deploy.GetName())
				return nil, err
			} else {
				logger.Info("Delete ML deployment succeeded","name", deploy.GetName())
				return nil, nil
			}
		}
	}
	return deploy, nil
}

// AppendAssignmentEnv appends an environment variable for service pods
func appendServiceEnv(t *morphlingv1alpha1.Trial, env []corev1.EnvVar, args []string, resources corev1.ResourceRequirements) ([]corev1.EnvVar, []string, corev1.ResourceRequirements) {
	for _, a := range t.Spec.SamplingResult {
		switch a.Category {
			case morphlingv1alpha1.CategoryEnv: {
				name := strings.ReplaceAll(strings.ToUpper(a.Name), ".", "_")
				env = append(env, corev1.EnvVar{Name: name, Value: fmt.Sprintf(a.Value)})
			}
			case morphlingv1alpha1.CategoryArgs: {
				args = append(args, fmt.Sprintf(a.Value))
			}
			case morphlingv1alpha1.CategoryResource: {
				var resourceClass = corev1.ResourceCPU
				switch a.Name {
					case "cpu":
						resourceClass = corev1.ResourceCPU
					case "memory":
						resourceClass = corev1.ResourceMemory
					case "storage":
						resourceClass = corev1.ResourceStorage
					case "nvidia.com/gpu":
						resourceClass = "nvidia.com/gpu"
					case "nvidia.com/gpumem":
						resourceClass = "nvidia.com/gpumem"
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
	}
	return env, args, resources
}
