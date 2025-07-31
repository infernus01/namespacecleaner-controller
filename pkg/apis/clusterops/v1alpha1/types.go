package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type NamespaceCleaner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceCleanerSpec   `json:"spec,omitempty"`
	Status NamespaceCleanerStatus `json:"status,omitempty"`
}

// what the cleaner should do
type NamespaceCleanerSpec struct {
	// Selector which namespaces to scan for old pods
	Selector metav1.LabelSelector `json:"selector,omitempty"`
}

// the current state
type NamespaceCleanerStatus struct {
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// a list of NamespaceCleaner
type NamespaceCleanerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NamespaceCleaner `json:"items"`
}
