package trial

import (
	"context"
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"github.com/alibaba/morphling/pkg/controllers/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"strings"
	"time"
)

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
				if errors.IsNotFound(err) {
					logger.Info("Delete client operation is redundant")
					return nil, nil
				}
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
func (r *ReconcileTrial) getDesiredJobSpec(instance *morphlingv1alpha1.Trial) (*batchv1.Job, error) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	job := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:        util.GetStressTestJobName(instance),
			Namespace:   instance.GetNamespace(),
			Labels:      util.ServiceDeploymentLabels(instance),
			Annotations: instance.Annotations,
		},
	}
	if &instance.Spec.ClientTemplate != nil {
		instance.Spec.ClientTemplate.Spec.DeepCopyInto(&job.Spec)
	}
	// The default restart policy for a pod is not acceptable in the context of a job
	if job.Spec.Template.Spec.RestartPolicy == "" {
		job.Spec.Template.Spec.RestartPolicy = corev1.RestartPolicyNever
	}
	// The default backoff limit will restart the trial job which is unlikely to produce desirable results
	if job.Spec.BackoffLimit == nil {
		job.Spec.BackoffLimit = new(int32)
	}
	// Expose the current assignments as environment variables to every container
	for i := range job.Spec.Template.Spec.Containers {
		c := &job.Spec.Template.Spec.Containers[i]
		c.Env = appendJobEnv(instance, c.Env)
	}
	if err := controllerutil.SetControllerReference(instance, job, r.Scheme); err != nil {
		logger.Error(err, "Set client job controller reference error", "name", job.GetName())
		return nil, err
	}
	return job, nil
}

// appendJobEnv appends an environment variable for jobs
func appendJobEnv(t *morphlingv1alpha1.Trial, env []corev1.EnvVar) []corev1.EnvVar {
	env = append(env, corev1.EnvVar{Name: "RequestTemplate", Value: fmt.Sprintf(t.Spec.RequestTemplate)})
	env = append(env, corev1.EnvVar{Name: "ServiceName", Value: util.GetServiceEndpoint(t)})
	env = append(env, corev1.EnvVar{Name: "TrialName", Value: fmt.Sprintf(t.Name)})
	env = append(env, corev1.EnvVar{Name: "Namespace", Value: fmt.Sprintf(t.Namespace)})
	env = append(env, corev1.EnvVar{Name: "DBNamespace", Value: fmt.Sprintf(consts.DefaultControllerNamespace)})
	env = append(env, corev1.EnvVar{Name: "DBPort", Value: fmt.Sprintf(consts.DefaultMorphlingDBManagerServicePort)})
	for _, cat := range t.Spec.SamplingResult {
		name := strings.ReplaceAll(strings.ToUpper(cat.Name), ".", "_")
		env = append(env, corev1.EnvVar{Name: name, Value: fmt.Sprintf(cat.Value)})
	}
	return env
}
