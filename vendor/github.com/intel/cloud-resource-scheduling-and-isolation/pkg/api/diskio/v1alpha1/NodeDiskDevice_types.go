package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

type NodeDiskDevice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeDiskDeviceSpec   `json:"spec,omitempty"`
	Status NodeDiskDeviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NodeDiskDeviceList contains a list of NodeDiskDevice
type NodeDiskDeviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeDiskDevice `json:"items"`
}

type NodeDiskDeviceSpec struct {
	NodeName string                `json:"nodeName,omitempty"`
	Devices  map[string]DiskDevice `json:"devices,omitempty"`
}

type DiskDevice struct {
	// Device name
	Name string `json:"name"`
	// Device vendor
	Vendor string `json:"vendor,omitempty"`
	// Device model
	Model string `json:"model,omitempty"`
	// Default or not
	Type string `json:"type,omitempty"`
	// Read/Write io bandwidth limit
	Default IOBandwidth `json:"default,omitempty"`
	// Profile result of io bandwidth capacity
	Capacity IOBandwidth `json:"capacity,omitempty"`
}

type NodeDiskDeviceStatus struct {
}
