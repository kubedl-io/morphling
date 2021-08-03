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
	"github.com/alibaba/morphling/pkg/controllers/consts"
	"strconv"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	samplingClient "github.com/alibaba/morphling/pkg/controllers/experiment/sampling_client"
	"github.com/alibaba/morphling/pkg/controllers/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// updateTrialsSummary updates trials summary
func updateTrialsSummary(instance *morphlingv1alpha1.ProfilingExperiment, trials *morphlingv1alpha1.TrialList) {
	bestTrialValue, _ := strconv.ParseFloat(consts.DefaultMetricValue, 64)
	sts := &instance.Status
	sts.TrialsTotal = 0
	sts.RunningTrialList, sts.PendingTrialList, sts.FailedTrialList, sts.SucceededTrialList, sts.KilledTrialList = nil, nil, nil, nil, nil
	bestTrialIndex := -1
	objectiveType := instance.Spec.Objective.Type
	objectiveMetricName := instance.Spec.Objective.ObjectiveMetricName

	// Check the trial list
	for index, trial := range trials.Items {
		sts.TrialsTotal++
		if util.IsKilledTrial(&trial) {
			sts.KilledTrialList = append(sts.KilledTrialList, trial.Name)
		} else if util.IsFailedTrial(&trial) {
			sts.FailedTrialList = append(sts.FailedTrialList, trial.Name)
		} else if util.IsSucceededTrial(&trial) {
			sts.SucceededTrialList = append(sts.SucceededTrialList, trial.Name)
		} else if util.IsRunningTrial(&trial) {
			sts.RunningTrialList = append(sts.RunningTrialList, trial.Name)
		} else {
			sts.PendingTrialList = append(sts.PendingTrialList, trial.Name)
		}

		// Get trial results
		objectiveMetricValue := getObjectiveMetricValue(trial, objectiveMetricName)
		if objectiveMetricValue == nil {
			continue
		}

		// Initialize vars to objective metric value of the first trial
		if bestTrialIndex == -1 {
			bestTrialValue = *objectiveMetricValue
			bestTrialIndex = index
		}

		// Differentiate max / min objectives
		if objectiveType == morphlingv1alpha1.ObjectiveTypeMinimize {
			if *objectiveMetricValue < bestTrialValue {
				bestTrialValue = *objectiveMetricValue
				bestTrialIndex = index
			}
		} else if objectiveType == morphlingv1alpha1.ObjectiveTypeMaximize {
			if *objectiveMetricValue > bestTrialValue {
				bestTrialValue = *objectiveMetricValue
				bestTrialIndex = index
			}
		}
	}

	// Statistic summary
	sts.TrialsRunning = int32(len(sts.RunningTrialList))
	sts.TrialsPending = int32(len(sts.PendingTrialList))
	sts.TrialsSucceeded = int32(len(sts.SucceededTrialList))
	sts.TrialsFailed = int32(len(sts.FailedTrialList))
	sts.TrialsKilled = int32(len(sts.KilledTrialList))

	// if best trial is set
	if bestTrialIndex != -1 {
		bestTrial := trials.Items[bestTrialIndex]
		sts.CurrentOptimalTrial.TunableParameters = []morphlingv1alpha1.ParameterAssignment{}
		for _, parameterAssigment := range bestTrial.Spec.SamplingResult {
			sts.CurrentOptimalTrial.TunableParameters = append(sts.CurrentOptimalTrial.TunableParameters, parameterAssigment)
		}
		sts.CurrentOptimalTrial.ObjectiveMetricsObserved = []morphlingv1alpha1.Metric{}
		for _, metric := range bestTrial.Status.TrialResult.ObjectiveMetricsObserved {
			sts.CurrentOptimalTrial.ObjectiveMetricsObserved = append(sts.CurrentOptimalTrial.ObjectiveMetricsObserved, metric)
		}
	}
}

func getObjectiveMetricValue(trial morphlingv1alpha1.Trial, objectiveMetricName string) *float64 {
	if trial.Status.TrialResult == nil {
		return nil
	}
	for _, metric := range trial.Status.TrialResult.ObjectiveMetricsObserved {
		if objectiveMetricName == metric.Name {
			value, _ := strconv.ParseFloat(metric.Value, 0)
			return &value
		}
	}
	return nil
}

// updateExperimentStatusCondition updates the experiment status.
func updateExperimentStatusCondition(instance *morphlingv1alpha1.ProfilingExperiment) {
	completedTrialsCount := instance.Status.TrialsSucceeded + instance.Status.TrialsFailed + instance.Status.TrialsKilled
	now := metav1.Now()

	// Check if MaxTrialCount is reached.
	if (instance.Spec.MaxNumTrials != nil) && (completedTrialsCount >= *instance.Spec.MaxNumTrials) {
		msg := "Experiment has succeeded because max trial count has reached"
		util.MarkExperimentStatusSucceeded(instance, msg)
		instance.Status.CompletionTime = &now
		return
	}

	// Check if Sampling space is exhausted.
	if (CalculateMaximumSearchSpace(instance) > 0) && (int(completedTrialsCount) >= CalculateMaximumSearchSpace(instance)) {
		msg := "Experiment has succeeded because maximum search space has reached"
		util.MarkExperimentStatusSucceeded(instance, msg)
		instance.Status.CompletionTime = &now
		return
	}

	msg := "Experiment is running"
	util.MarkExperimentStatusRunning(instance, msg)
}

func CalculateMaximumSearchSpace(instance *morphlingv1alpha1.ProfilingExperiment) int {
	space := int(1)
	for _, cat := range instance.Spec.TunableParameters {
		for _, p := range cat.Parameters {
			feasibleSpace, err := samplingClient.ConvertFeasibleSpace(p.FeasibleSpace, p.ParameterType)
			if err != nil {
				log.Error(err, "failed to calculate maximum search space")
				return 0
			}
			if int(len(feasibleSpace)) <= 0 {
				log.Error(err, "failed to calculate maximum search space")
				return 0
			}
			space *= int(len(feasibleSpace))
		}
	}
	return space
}

// TrialLabels returns the expected trial labels.
func TrialLabels(instance *morphlingv1alpha1.ProfilingExperiment) map[string]string {
	res := make(map[string]string)
	for k, v := range instance.Labels {
		res[k] = v
	}
	res[consts.LabelExperimentName] = instance.Name

	return res
}
