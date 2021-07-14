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

package util

import (
	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
	"github.com/alibaba/morphling/pkg/controllers/consts"
)

// SamplingLabels returns the expected sampling labels.
func SamplingLabels(instance *morphlingv1alpha1.Sampling) map[string]string {
	res := make(map[string]string)
	for k, v := range instance.Labels {
		res[k] = v
	}
	res[consts.LabelDeploymentName] = GetAlgorithmDeploymentName(instance)
	res[consts.LabelExperimentName] = instance.Name
	res[consts.LabelSamplingName] = instance.Name

	return res
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
