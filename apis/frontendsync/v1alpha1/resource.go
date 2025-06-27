package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// FrontendPageSpec defines the desired state of Frontend
type FrontendSyncSpec struct {
	Url          string `json:"url"`
	SyncInterval int    `json:"syncInterval"`
}

// +kubebuilder:object:root=true
type FrontendSync struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FrontendSyncSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// FrontendPageList contains a list of FrontendPage
type FrontendSyncList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []FrontendSync `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FrontendSync{}, &FrontendSyncList{})
}
