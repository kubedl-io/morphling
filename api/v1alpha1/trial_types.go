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
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrialSpec defines the pressure test task under a specific configuration
type TrialSpec struct {
	// Sampling results, including all the parameter values to be tuned
	SamplingResult []ParameterAssignment `json:"samplingResult,omitempty"`

	// Describes the objective of the experiment.
	Objective ObjectiveSpec `json:"objective,omitempty"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// The client template to trigger the test against target service.
	ClientTemplate v1beta1.JobTemplateSpec `json:"clientTemplate,omitempty"`

	// The target service pod/deployment whose parameters to be tuned
	ServicePodTemplate corev1.PodTemplate `json:"servicePodTemplate,omitempty"`

	// The maximum time in seconds for a deployment to make progress before it is considered to be failed.
	ServiceProgressDeadline *int32 `json:"serviceProgressDeadline,omitempty"`
}

// TrialStatus defines the status of this pressure test
type TrialStatus struct {
	// Output of the trial, including the TrialAssignment and the Objective value (e.g., QPS)
	TrialResult *TrialResult `json:"trialResult,omitempty"`

	// Observed runtime condition for this Trial.
	Conditions []TrialCondition `json:"conditions,omitempty"`

	// The time this trial was started.
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// The time this trial was completed.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

type TrialCondition struct {
	// Type of trial condition.
	Type TrialConditionType `json:"type"`

	// Standard Kubernetes object's LastTransitionTime
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

type TrialConditionType string

const (
	TrialRunning   TrialConditionType = "Running"
	TrialSucceeded TrialConditionType = "Succeeded"
	TrialFailed    TrialConditionType = "Failed"
	TrialCreated   TrialConditionType = "Created"
	TrialPending   TrialConditionType = "Pending"
	TrialKilled    TrialConditionType = "Killed"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Object Name",type=string,JSONPath=`.status.trialResult.objectiveMetricsObserved[-1:].name`
// +kubebuilder:printcolumn:name="Object Value",type=string,JSONPath=`.status.trialResult.objectiveMetricsObserved[-1:].value`
// +kubebuilder:printcolumn:name="Parameters",type=string,JSONPath=`.spec.samplingResult`

// Trial is the Schema for the trials API
type Trial struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrialSpec   `json:"spec,omitempty"`
	Status TrialStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TrialList contains a list of Trial
type TrialList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trial `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trial{}, &TrialList{})
}
