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

package sampling

import (
	"context"

	"k8s.io/apimachinery/pkg/api/equality"

	morphlingv1alpha1 "github.com/alibaba/morphling/api/v1alpha1"
)

const (
	SamplingDeploymentReady    = "DeploymentReady"
	SamplingDeploymentNotReady = "DeploymentNotReady"
	SamplingRunningReason      = "SamplingRunning"
	SamplingFailedReason       = "SamplingFailed"
)

func (r *SamplingReconciler) updateStatus(s *morphlingv1alpha1.Sampling, oldS *morphlingv1alpha1.Sampling) error {
	if !equality.Semantic.DeepEqual(s.Status, oldS.Status) {
		if err := r.Update(context.TODO(), s); err != nil {
			return err
		}
	}
	return nil
}

func (r *SamplingReconciler) updateStatusCondition(s *morphlingv1alpha1.Sampling, oldS *morphlingv1alpha1.Sampling) error {
	if !equality.Semantic.DeepEqual(s.Status.Conditions, oldS.Status.Conditions) {
		newConditions := s.Status.Conditions
		s.Status = oldS.Status
		s.Status.Conditions = newConditions
		if err := r.Update(context.TODO(), s); err != nil {
			return err
		}
	}
	return nil
}
