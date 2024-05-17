package resource

import (
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

// ExtendedResource specifies extended resources's aggregation methods
// It is inteneded to be triggered by Pod/Node events
type ExtendedResource interface {
	Name() string
	AddPod(pod *v1.Pod, request v1alpha1.IOBandwidth) error
	RemovePod(pod *v1.Pod) error
	// AdmitPod(pod *v1.Pod) (v1alpha1.IOBandwidth, error)
	PrintInfo()
}
