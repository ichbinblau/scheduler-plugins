package resource

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// ExtendedResource specifies extended resources's aggregation methods
// It is inteneded to be triggered by Pod/Node events
type ExtendedResource interface {
	Name() string
	// todo
	AddPod(pod *v1.Pod, request interface{}, client kubernetes.Interface) ([]struct{}, bool, error)
	RemovePod(pod *v1.Pod, client kubernetes.Interface) error
	Clone() ExtendedResource
	AdmitPod(pod *v1.Pod) (interface{}, error) // return pod requested resource and error msg
	PrintInfo()
}
