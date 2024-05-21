package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type NodeDiskIOStats struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeDiskIOStatsSpec   `json:"spec,omitempty"`
	Status NodeDiskIOStatsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeDiskIOStatsList contains a list of NodeDiskIOStats
type NodeDiskIOStatsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeDiskIOStats `json:"items"`
}

type NodeDiskIOStatsSpec struct {
	NodeName string `json:"nodeName"`
	// a slice of reserved pod uids
	ReservedPods []string `json:"reservedPods,omitempty"`
}

// NodeDiskIOStatsStatus defines the observed state of node disks
type NodeDiskIOStatsStatus struct {
	// ObservedGeneration is the most recent generation observed by this IO Driver.
	ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
	// the key of the map is the device id
	AllocatableBandwidth map[string]IOBandwidth `json:"allocableBandwidth,omitempty"`
}

type IOBandwidth struct {
	// Normalized total IO throughput capacity
	Total resource.Quantity `json:"total,omitempty"`
	// Normalized read IO throughput capacity
	Read resource.Quantity `json:"read,omitempty"`
	// Normalized write IO throughput capacity
	Write resource.Quantity `json:"write,omitempty"`
}
