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

package samplingclient

import (
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	samplingapi "github.com/alibaba/morphling/api/v1alpha1/manager"
)

// appendAlgorithmSettingsFromSampling appends the algorithm settings
// in sampling to Experiment.
// Algorithm settings in sampling will overwrite the settings in experiment.
func appendAlgorithmSettingsFromSampling(experiment *morphlingv1alpha1.ProfilingExperiment, algoSettingsInSampling []morphlingv1alpha1.AlgorithmSetting) {
	algoSettingsInExperiment := experiment.Spec.Algorithm
	for _, setting := range algoSettingsInSampling {
		if index, found := contains(
			algoSettingsInExperiment.AlgorithmSettings, setting.Name); found {
			// If the setting is found in Experiment, update it.
			algoSettingsInExperiment.AlgorithmSettings[index].Value = setting.Value
		} else {
			// If not found, append it.
			algoSettingsInExperiment.AlgorithmSettings = append(
				algoSettingsInExperiment.AlgorithmSettings, setting)
		}
	}
}

func updateAlgorithmSettings(sampling *morphlingv1alpha1.Sampling, algorithm *samplingapi.AlgorithmSpec) {
	for _, setting := range algorithm.AlgorithmSetting {
		if setting != nil {
			if index, found := contains(sampling.Spec.Algorithm.AlgorithmSettings, setting.Name); found {
				// If the setting is found in Sampling, update it.
				sampling.Spec.Algorithm.AlgorithmSettings[index].Value = setting.Value
			} else {
				// If not found, append it.
				sampling.Spec.Algorithm.AlgorithmSettings = append(sampling.Spec.Algorithm.AlgorithmSettings, morphlingv1alpha1.AlgorithmSetting{
					Name:  setting.Name,
					Value: setting.Value,
				})
			}
		}
	}
}

func contains(algorithmSettings []morphlingv1alpha1.AlgorithmSetting,
	name string) (int, bool) {
	for i, s := range algorithmSettings {
		if s.Name == name {
			return i, true
		}
	}
	return -1, false
}
