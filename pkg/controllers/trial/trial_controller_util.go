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

func (r *TrialReconciler) updateStatus(instance *morphlingv1alpha1.Trial) error {
	err := r.Update(context.TODO(), instance)
	if err != nil {
		return err
	}
	return nil
}

func (r *TrialReconciler) UpdateTrialStatusCondition(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job, jobCondition []batchv1.JobCondition) {
	if jobCondition == nil || instance == nil || deployedJob == nil {
		msg := "Trial is running"
		util.MarkTrialStatusRunning(instance, msg)
		return
	}

	now := metav1.Now()
	jobConditionType := (jobCondition[len(jobCondition)-1]).Type

	if jobConditionType == batchv1.JobComplete {
		// Client job is completed
		if isTrialObservationAvailable(instance) {
			// Client job has recorded the trial result
			msg := "Trial has succeeded"
			util.MarkTrialStatusSucceeded(instance, corev1.ConditionTrue, msg)
			instance.Status.CompletionTime = &now
			eventMsg := fmt.Sprintf("Client Job %s has succeeded", deployedJob.GetName())
			r.recorder.Eventf(instance, corev1.EventTypeNormal, "JobSucceeded", eventMsg)
			//r.collector.IncreaseTrialsSucceededCount(instance.Namespace)
		} else {
			// Client job has NOT recorded the trial result
			msg := "Metrics are not available"
			util.MarkTrialStatusSucceeded(instance, corev1.ConditionFalse, msg)
		}
	} else if jobConditionType == batchv1.JobFailed {
		// Client job is failed
		msg := "Trial has failed"
		util.MarkTrialStatusFailed(instance, msg)
		instance.Status.CompletionTime = &now
		//r.collector.IncreaseTrialsFailedCount(instance.Namespace)
	} else {
		// Client job is still running
		msg := "Trial is running"
		util.MarkTrialStatusRunning(instance, msg)
	}

	return
}

func (r *TrialReconciler) FindPodAssociatedWithServiceDeployment(instance *morphlingv1alpha1.Trial, deploy *appsv1.Deployment) (*corev1.PodList, error) {

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

func (r *TrialReconciler) UpdateTrialStatusObservation(instance *morphlingv1alpha1.Trial, deployedJob *batchv1.Job) error {
	if &instance.Spec.Objective == nil || &instance.Spec.Objective.ObjectiveMetricName == nil || r.ManagerClient == nil {
		return nil
	}

	// Get the name of the objective metric
	objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName

	// Get the pod of the client job
	//jobPod, err := r.FindPodAssociatedWithClientJob(instance, deployedJob)
	//if err != nil {
	//	return err
	//}

	// Get the trial result from the job pod (pod log)
	//if jobPod.Items != nil {
	reply, err := r.GetTrialObservationLog(instance)
	if err != nil {
		return err
	}
	if reply.ObservationLog != nil {
		// Get the value of the result
		value := reply.ObservationLog.MetricLogs[0].Metric.Value
		//value = "0.0"
		//bestObjectiveValue := string(value) //strconv.ParseFloat(value, 0)
		//bestObjectiveValue, _ := strconv.ParseFloat(value[:int(len(value)-1)], 0)
		//a, _ := strconv.ParseFloat(instance.Spec.SamplingResult[0].Value, 0)
		//bestObjectiveValue_ := int(bestObjectiveValue) //+ int32(a) + int32(rand.Intn(10))
		//int32(rand.Int())

		// Update the trial result to the status of the trial
		metric := morphlingv1alpha1.Metric{Name: objectiveMetricName, Value: value}
		instance.Status.TrialResult = &morphlingv1alpha1.TrialResult{}
		instance.Status.TrialResult.ObjectiveMetricsObserved = []morphlingv1alpha1.Metric{metric}
	}
	//}

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
