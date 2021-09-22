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

// ProfilingExperimentSpec defines the desired state of ProfilingExperiment
type ProfilingExperimentSpec struct {
	// List of hyperparameter to be tuned.
	TunableParameters []ParameterCategory `json:"tunableParameters,omitempty"`

	// Describes the objective of the experiment.
	Objective ObjectiveSpec `json:"objective,omitempty"`

	// Sampling algorithm, e.g., Bayesian opt.
	Algorithm AlgorithmSpec `json:"algorithm,omitempty"`

	// Maximum number of trials
	MaxNumTrials *int32 `json:"maxNumTrials,omitempty"`

	// Parallelism is the number of concurrent trials.
	Parallelism *int32 `json:"parallelism,omitempty"`

	// The request template in json format, used for testing against the REST API of target service.
	RequestTemplate string `json:"requestTemplate,omitempty"`

	// Client Template to trigger the test against target service
	ClientTemplate v1beta1.JobTemplateSpec `json:"clientTemplate,omitempty"`

	// The target service pod/deployment whose parameters to be tuned
	ServicePodTemplate corev1.PodTemplate `json:"servicePodTemplate,omitempty"`

	// The maximum time in seconds for a deployment to make progress before it is considered to be failed.
	ServiceProgressDeadline *int32 `json:"serviceProgressDeadline,omitempty"`
}

type ProfilingExperimentStatus struct {
	// List of observed runtime conditions for this ProfilingExperiment.
	Conditions []ProfilingCondition `json:"conditions,omitempty"`

	// Current optimal parameters
	CurrentOptimalTrial TrialResult `json:"currentOptimalTrial,omitempty"`

	// Sampled configurations and the corresponding object values
	TrialResultList []TrialResult `json:"trialResultList,omitempty"`

	// Completion time of the experiment
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Start time of the experiment
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// List of trial names which are running.
	RunningTrialList []string `json:"runningTrialList,omitempty"`

	// List of trial names which are pending.
	PendingTrialList []string `json:"pendingTrialList,omitempty"`

	// List of trial names which have already failed.
	FailedTrialList []string `json:"failedTrialList,omitempty"`

	// List of trial names which have already succeeded.
	SucceededTrialList []string `json:"succeededTrialList,omitempty"`

	// List of trial names which have been killed.
	KilledTrialList []string `json:"killedTrialList,omitempty"`

	// TrialsTotal is the total number of trials owned by the experiment.
	TrialsTotal int32 `json:"trialsTotal,omitempty"`

	// How many trials have succeeded.
	TrialsSucceeded int32 `json:"trialsSucceeded,omitempty"`

	// How many trials have been killed.
	TrialsKilled int32 `json:"trialsKilled,omitempty"`

	// How many trials are pending.
	TrialsPending int32 `json:"trialsPending,omitempty"`

	// How many trials are running.
	TrialsRunning int32 `json:"trialsRunning,omitempty"`

	// How many trials have failed.
	TrialsFailed int32 `json:"trialsFailed,omitempty"`
}

type CollectorKind string

// +k8s:deepcopy-gen=true

// ProfilingCondition describes the state of the experiment at a certain point.
type ProfilingCondition struct {
	// Type of experiment condition.
	Type ProfilingConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// A human readable message indicating details about the transition.
	Message string `json:"message,omitempty"`

	// The last time this condition was updated.
	LastUpdateTime metav1.Time `json:"lastUpdateTime,omitempty"`
}

type TrialResult struct {
	// Current parameter assignment of the trial.
	TunableParameters []ParameterAssignment `json:"tunableParameters,omitempty"`

	// The observed value for the objective metrics under these parameters, e.g., QPS=100.
	ObjectiveMetricsObserved []Metric `json:"objectiveMetricsObserved,omitempty"`
}

// TrialAssignment is the assignment for one trial.
type TrialAssignment struct {
	// Sampling results
	ParameterAssignments []ParameterAssignment `json:"parameterAssignments,omitempty"`

	//Name of the sampling_client sampling_client result, used to start a trial
	Name string `json:"name,omitempty"`
}

type Metric struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

// ParameterType defines the type of hyper-parameter to be tuned, we support the following several kinds of parameters.
type ParameterType string

const (
	ParameterTypeDouble      ParameterType = "double"
	ParameterTypeInt         ParameterType = "int"
	ParameterTypeDiscrete    ParameterType = "discrete"
	ParameterTypeCategorical ParameterType = "categorical"
)

// ParameterCategory id the category of parameters, high-level parameter divisions, including env, args, resource
type ParameterCategory struct {
	Category   Category        `json:"category,omitempty"`
	Parameters []ParameterSpec `json:"parameters,omitempty"`
}

// FeasibleSpace defines the range of the hyper-parameters to be tuned
type FeasibleSpace struct {
	// The max value of the search space.
	Max string `json:"max,omitempty"`

	// The min value of the search space.
	Min string `json:"min,omitempty"`

	// The list of possible value.
	List []string `json:"list,omitempty"`

	// The step of sampling_client.
	Step string `json:"step,omitempty"`
}

// Category of the hyper-parameters to be tuned, for patching use.
type Category string

const (
	// Computing resources, including cpu, memory, gpumem.
	CategoryResource Category = "resource"

	// Environment variables, set for service pods/deployments.
	CategoryEnv Category = "env"

	// Args for codes running in service pods/deployments.
	CategoryArgs Category = "args"
)

// ParameterSpec is the meta data of a hyper-parameter to be tuned
type ParameterSpec struct {
	Name          string        `json:"name,omitempty"`
	ParameterType ParameterType `json:"parameterType,omitempty"`
	FeasibleSpace FeasibleSpace `json:"feasibleSpace,omitempty"`
}

// ObjectiveType is the optimization obj classes: minimize or maximize
type ObjectiveType string

const (
	ObjectiveTypeMinimize ObjectiveType = "minimize"
	ObjectiveTypeMaximize ObjectiveType = "maximize"
)

// ObjectiveSpec defines the optimization obj, e.g., minimize the resource cost per QPS
type ObjectiveSpec struct {
	// The type of the objective, including minimize or maximize
	Type ObjectiveType `json:"type,omitempty"`

	// Metric name, e.g., GPUMemConsumptionPerQPS
	ObjectiveMetricName string `json:"objectiveMetricName,omitempty"`
}

// AlgorithmSetting defines the parameters key-value pair of the Opt. algorithm
type AlgorithmSetting struct {
	// The name of the key-value pair.
	Name string `json:"name,omitempty"`

	// The value of the key.
	Value string `json:"value,omitempty"`
}

// AlgorithmName is the supported searching algorithms
type AlgorithmName string

const (
	BayesianOpt  AlgorithmName = "BayesianOpt"
	RandomSearch AlgorithmName = "random"
	GridSearch   AlgorithmName = "grid"
)

// AlgorithmSpec is the specification of Opt. algorithm
type AlgorithmSpec struct {
	// The name of algorithm for sampling_client: random, grid, bayesian optimization.
	AlgorithmName AlgorithmName `json:"algorithmName,omitempty"`

	// The key-value pairs representing settings for sampling_client algorithms.
	AlgorithmSettings []AlgorithmSetting `json:"algorithmSettings,omitempty"`
}

// ParameterAssignment defines the current hyper-parameter value (key-value pair)
type ParameterAssignment struct {
	Name     string   `json:"name,omitempty"`
	Value    string   `json:"value,omitempty"`
	Category Category `json:"category,omitempty"`
}

// ProfilingConditionType defines the status of the ProfilingExperiment
type ProfilingConditionType string

const (
	ProfilingCreated    ProfilingConditionType = "Created"
	ProfilingRunning    ProfilingConditionType = "Running"
	ProfilingRestarting ProfilingConditionType = "Restarting"
	ProfilingSucceeded  ProfilingConditionType = "Succeeded"
	ProfilingFailed     ProfilingConditionType = "Failed"
	ProfilingCompleted  ProfilingConditionType = "Completed"
)

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="State",type=string,JSONPath=`.status.conditions[-1:].type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Objective-Name",type=string,JSONPath=`.status.currentOptimalTrial.objectiveMetricsObserved[-1:].name`
// +kubebuilder:printcolumn:name="Optimal-Objective-Value",type=string,JSONPath=`.status.currentOptimalTrial.objectiveMetricsObserved[-1:].value`
// +kubebuilder:printcolumn:name="Optimal-Parameters",type=string,JSONPath=`.status.currentOptimalTrial.tunableParameters`
// +kubebuilder:resource:shortName="pe"
// +kubebuilder:subresource:status

// ProfilingExperiment is the Schema for the profilingexperiments API
type ProfilingExperiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProfilingExperimentSpec   `json:"spec,omitempty"`
	Status ProfilingExperimentStatus `json:"status,omitempty"`
}

// ProfilingExperimentList contains a list of ProfilingExperiment
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ProfilingExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProfilingExperiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProfilingExperiment{}, &ProfilingExperimentList{})
}
