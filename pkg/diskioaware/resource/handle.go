package resource

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
)

type Handle interface {
	Name() string
	Run(ExtendedCache, informers.SharedInformerFactory, kubernetes.Interface) error
}

type CacheHandle interface {
	// AddCacheNodeInfo(string, ioiv1.ResourceConfigSpec) error
	// InitNodeStatus(string, ioiv1.ResourceConfigSpec)
	DeleteCacheNodeInfo(string) error
	// UpdateCacheNodeStatus(string, *ioiv1.NodeIOStatusStatus) error
	CriticalPod(annotations map[string]string) bool
	// GetRequestByAnnotation(*v1.Pod, string, bool) (map[string]*aggregator.IOResourceRequest, interface{}, error)
	NodeRegistered(string) bool
	// AddPod(*v1.Pod, string, interface{}, kubernetes.Interface) ([]*aggregator.IOResourceRequest, bool, error) // input pod request interface
	RemovePod(*v1.Pod, string, kubernetes.Interface) error
	AdmitPod(*v1.Pod, string) (interface{}, error)               // output pod request interface
	CanAdmitPod(string, interface{}) (bool, interface{}, error)  // input pod request interface, output node's resource status interface
	NodePressureRatio(interface{}, interface{}) (float64, error) // input pod request interface, and node's resource status interface
	PrintCacheInfo()
	NodeHasIoiLabel(*v1.Node) bool
	// InitBENum(string, ioiv1.ResourceConfigSpec)
	UpdatePVC(*v1.PersistentVolumeClaim, string, string, int64, bool) error
}

type HandleBase struct {
	EC ExtendedCache
}

func (h *HandleBase) RemovePod(pod *v1.Pod, nodeName string, client kubernetes.Interface) error {
	r := h.EC.GetExtendedResource(nodeName)
	if r != nil {
		err := r.RemovePod(pod, client)
		if err != nil {
			return err
		}
	}
	h.EC.PrintCacheInfo()
	return nil
}

// todo
func (h *HandleBase) AddPod(pod *v1.Pod, nodeName string, request interface{}, client kubernetes.Interface) ([]struct{}, bool, error) {
	r := h.EC.GetExtendedResource(nodeName)
	if r != nil {
		return r.AddPod(pod, request, client)
	}
	return nil, true, fmt.Errorf("cannot get extended resource: %v", nodeName)
}

func (h *HandleBase) AdmitPod(pod *v1.Pod, nodeName string) (interface{}, error) {
	r := h.EC.GetExtendedResource(nodeName)
	h.EC.PrintCacheInfo()
	if r != nil {
		return r.AdmitPod(pod)
	}

	return nil, fmt.Errorf("failed to get the extended resource on node: %s", nodeName)
}

func (h *HandleBase) PrintCacheInfo() {
	h.EC.PrintCacheInfo()
}

func (h *HandleBase) NodeRegistered(node string) bool {
	if obj := h.EC.GetExtendedResource(node); obj != nil {
		return true
	}
	return false
}
