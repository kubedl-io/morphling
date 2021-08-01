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
	"k8s.io/apimachinery/pkg/api/errors"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/util"
)

type updateStatusFunc func(instance *morphlingv1alpha1.Trial) error

func (r *ReconcileTrial) updateStatus(instance *morphlingv1alpha1.Trial) error {
	err := r.Status().Update(context.TODO(), instance)
	if err != nil {
		if !errors.IsConflict(err){return err}
	}
	return nil
}

func (r *ReconcileTrial) UpdateTrialStatusCondition(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job, jobCondition []batchv1.JobCondition) {
	// Todo should mark trial status as 'unknown'
	if jobCondition == nil || instance == nil || deployedJob == nil {
		msg := "Trial is running"
		util.MarkTrialStatusRunning(instance, msg)
		return
	}

	now := metav1.Now()
	jobConditionType := (jobCondition[len(jobCondition)-1]).Type
	switch jobConditionType {
	
	}

	if jobConditionType == batchv1.JobComplete {
		// Client-side stress test job is completed
		if isTrialObservationAvailable(instance) {
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
	} else if jobConditionType == batchv1.JobFailed {
		// Client-side stress test job is failed
		msg := "Client-side stress test job has failed"
		util.MarkTrialStatusFailed(instance, msg)
		instance.Status.CompletionTime = &now
	} else {
		// Client-side stress test job is still running
		msg := "Client-side stress test job is running"
		util.MarkTrialStatusRunning(instance, msg)
	}

	return
}

func (r *ReconcileTrial) FindPodAssociatedWithServiceDeployment(instance *morphlingv1alpha1.Trial, deploy *appsv1.Deployment) (*corev1.PodList, error) {

	jobPod := &corev1.PodList{}

	// Select the pods associated with the job-name
	deployLabels := map[string]string{"deployment": deploy.GetName()}
	log.Info("labels: IP:", "name", deployLabels)
	log.Info("deployment: IP:", "name", deploy.GetName())
	lo := &client.ListOptions{}
	lo.LabelSelector = labels.SelectorFromSet(deployLabels)
	lo.Namespace = instance.Namespace
	log.Info("lo: IP:", "name", lo)

	// List the pods associated with the job-name
	if err := r.List(context.TODO(), jobPod, lo); err != nil {
		log.Error(err, "JobPod List error")
		return nil, err
	}
	log.Info("jobPod: IP:", "name", jobPod)
	return jobPod, nil
}

func (r *ReconcileTrial) UpdateTrialStatusObservation(instance *morphlingv1alpha1.Trial) error {
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

func isTrialObservationAvailable(instance *morphlingv1alpha1.Trial) bool {
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
