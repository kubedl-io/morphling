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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SamplingSpec defines the next configuration sampling point
type SamplingSpec struct {
	// Sampling Algorithm
	Algorithm AlgorithmSpec `json:"algorithm"`

	// Number of Samplings requested
	NumSamplingsRequested int32 `json:"numSamplingsRequested,omitempty"`

	// Node affinity of the Sampling pod
	Affinity *corev1.Affinity `json:"affinity,omitempty"`

	// Node toleration of the Sampling pod
	Toleration []corev1.Toleration `json:"toleration,omitempty"`
}

// SamplingStatus defines the next configuration sampling point
type SamplingStatus struct {
	// Sampling results
	SamplingResults []TrialAssignment `json:"samplingResults,omitempty"`

	// Observed runtime condition for this sampling.
	Conditions []SamplingCondition `json:"conditions,omitempty"`

	// Start time of the sampling instance
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Number of sampling results processed
	SamplingsProcessed int32 `json:"samplingsProcessed,omitempty"`
}

// TrialAssignment is the assignment for one trial.
type TrialAssignment struct {
	// Sampling results
	ParameterAssignments []ParameterAssignment `json:"parameterAssignments,omitempty"`

	//Name of the sampling sampling result, used to start a trial
	Name string `json:"name,omitempty"`
}

type SamplingCondition struct {
	// Type of the condition
	Type SamplingConditionType `json:"type"`

	// Standard Kubernetes object's LastUpdateTime
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`
}
type SamplingConditionType string

const (
	SamplingRunning         SamplingConditionType = "Running"
	SamplingSucceeded       SamplingConditionType = "Succeeded"
	SamplingFailed          SamplingConditionType = "Failed"
	SamplingDeploymentReady SamplingConditionType = "Ready"
	SamplingCreated         SamplingConditionType = "Created"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// Sampling is the Schema for the samplings API
type Sampling struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SamplingSpec   `json:"spec,omitempty"`
	Status SamplingStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SamplingList contains a list of Sampling
type SamplingList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Sampling `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Sampling{}, &SamplingList{})
}
