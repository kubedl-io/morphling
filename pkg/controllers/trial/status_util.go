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
	"fmt"
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/util"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type updateStatusFunc func(instance *morphlingv1alpha1.Trial) error

func (r *ReconcileTrial) UpdateTrialStatusByClientJob(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job) error {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	// Update trial result
	if err := r.updateTrialResult(instance, deployedJob); err != nil {
		logger.Error(err, "Update trial result error")
		return err
	}
	// Update trial condition
	jobCondition := deployedJob.Status.Conditions
	r.updateTrialStatusCondition(instance, deployedJob, jobCondition)
	return nil
}

func (r *ReconcileTrial) UpdateTrialStatusByServiceDeployment(instance *morphlingv1alpha1.Trial, deployedDeployment *appsv1.Deployment) {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	ServiceDeploymentCondition := deployedDeployment.Status.Conditions
	if util.IsServiceDeplomentFail(ServiceDeploymentCondition) {
		message := "Trial service pod failed"
		objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
		metric := morphlingv1alpha1.Metric{Name: objectiveMetricName, Value: "0.0"}
		instance.Status.TrialResult = &morphlingv1alpha1.TrialResult{}
		instance.Status.TrialResult.ObjectiveMetricsObserved = []morphlingv1alpha1.Metric{metric}
		util.MarkTrialStatusFailed(instance, message)
		logger.Info("Service deployment is failed", "name", deployedDeployment.GetName())
	} else {
		message := "Trial service pod pending"
		util.MarkTrialStatusPendingTrial(instance, message)
		logger.Info("Service deployment is pending", "name", deployedDeployment.GetName())
	}
}

func (r *ReconcileTrial) updateTrialStatusCondition(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job, jobCondition []batchv1.JobCondition) {

	if jobCondition == nil || instance == nil || deployedJob == nil {
		msg := "Trial is running"
		util.MarkTrialStatusRunning(instance, msg)
		return
	}

	now := metav1.Now()
	if util.IsJobSucceeded(jobCondition) {
		// Client-side stress test job is completed
		if isTrialResultAvailable(instance) {
			msg := "Client-side stress test job has completed"
			util.MarkTrialStatusSucceeded(instance, corev1.ConditionTrue, msg)
			instance.Status.CompletionTime = &now
			eventMsg := fmt.Sprintf("Client-side stress test job %s has succeeded", deployedJob.GetName())
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "JobSucceeded", eventMsg)
		} else {
			// Client job has NOT recorded the trial result
			msg := "Trial results are not available"
			util.MarkTrialStatusSucceeded(instance, corev1.ConditionFalse, msg)
		}
	} else if util.IsJobFailed(jobCondition) {
		// Client-side stress test job is failed
		msg := "Client-side stress test job has failed"
		util.MarkTrialStatusFailed(instance, msg)
		instance.Status.CompletionTime = &now
	} else {
		// Client-side stress test job is still running
		msg := "Client-side stress test job is running"
		util.MarkTrialStatusRunning(instance, msg)
	}
}

func (r *ReconcileTrial) updateTrialResult(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job) error {
	logger := log.WithValues("Trial", types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})

	jobCondition := deployedJob.Status.Conditions
	if util.IsJobSucceeded(jobCondition) {
		logger.Info("Client Job is Completed", "name", deployedJob.GetName())
		// Update trial observation
		if err := r.updateTrialResultForSucceededTrial(instance); err != nil {
			logger.Error(err, "Update trial result error")
			return err
		}
	} else if util.IsJobFailed(jobCondition) {
		logger.Info("Client Job is Failed", "name", deployedJob.GetName())
		r.updateTrialResultForFailedTrial(instance)
	}
	return nil
}

func (r *ReconcileTrial) updateTrialResultForSucceededTrial(instance *morphlingv1alpha1.Trial) error {
	if &instance.Spec.Objective == nil || &instance.Spec.Objective.ObjectiveMetricName == nil || r.DBClient == nil {
		return nil
	}
	reply, err := r.GetTrialResult(instance)
	if err != nil {
		return err
	}
	if reply != nil {
		instance.Status.TrialResult = reply
	}
	return nil
}

func (r *ReconcileTrial) updateTrialResultForFailedTrial(instance *morphlingv1alpha1.Trial) {
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

func isTrialResultAvailable(instance *morphlingv1alpha1.Trial) bool {
	if instance == nil || &instance.Spec.Objective == nil || &instance.Spec.Objective.ObjectiveMetricName == nil {
		return false
	}
	// Get the name of the objective metric
	objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName
	if instance.Status.TrialResult != nil {
		if instance.Status.TrialResult.ObjectiveMetricsObserved != nil {
			for _, metric := range instance.Status.TrialResult.ObjectiveMetricsObserved {
				// Find the objective metric record from trail status
				if metric.Name == objectiveMetricName {
					return true
				}
			}
		}
	}
	// Objective metric record Not found
	return false
}

func (r *ReconcileTrial) ControllerName() string {
	return ControllerName
}
