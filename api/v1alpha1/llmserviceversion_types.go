package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LLMServiceVersionSpec defines the desired state of LLMServiceVersion
type LLMServiceVersionSpec struct {
	// Version number of the LLM service
	Version string `json:"version"`

	// ModelName is the name of the LLM model
	ModelName string `json:"modelName"`

	// CreationTime is the time when this version was created
	CreationTime string `json:"creationTime"`

	// AssociatedExperimentSpec is the spec of the associated experiment
	AssociatedExperimentSpec ProfilingExperimentSpec `json:"associatedExperimentSpec"`
}

// LLMServiceVersionStatus defines the observed state of LLMServiceVersion
type LLMServiceVersionStatus struct {
	// TestCompletionTime is the time when testing for this version was completed
	TestCompletionTime metav1.Time `json:"testCompletionTime,omitempty"`

	// AssociatedExperimentStatus is the status of the associated experiment
	AssociatedExperimentStatus ProfilingExperimentStatus `json:"associatedExperimentStatus"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LLMServiceVersion is the Schema for the llmserviceversions API
type LLMServiceVersion struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LLMServiceVersionSpec   `json:"spec,omitempty"`
	Status LLMServiceVersionStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
// LLMServiceVersionList contains a list of LLMServiceVersion
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type LLMServiceVersionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LLMServiceVersion `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LLMServiceVersion{}, &LLMServiceVersionList{})
}
