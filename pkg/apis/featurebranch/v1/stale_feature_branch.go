package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StaleFeatureBranchSpec defines the desired state of StaleFeatureBranch
type StaleFeatureBranchSpec struct {
	// +kubebuilder:validation:Required
	NamespaceSubstring string `json:"namespaceSubstring"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum=1
	AfterDaysWithoutDeploy int `json:"afterDaysWithoutDeploy"`

	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=30
	CheckEveryMinutes int `json:"checkEveryMinutes"`
}

// StaleFeatureBranchStatus defines the observed state of StaleFeatureBranch
type StaleFeatureBranchStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StaleFeatureBranch is the Schema for the stalefeaturebranches API
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=stalefeaturebranches,scope=Namespaced,shortName=sfb
type StaleFeatureBranch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StaleFeatureBranchSpec   `json:"spec,omitempty"`
	Status StaleFeatureBranchStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// StaleFeatureBranchList contains a list of StaleFeatureBranch
type StaleFeatureBranchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StaleFeatureBranch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StaleFeatureBranch{}, &StaleFeatureBranchList{})
}
