package resource

import (
	"fmt"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type Handle interface {
	Name() string
	Run(ExtendedCache, kubernetes.Interface) error
}

type CacheHandle interface {
	AddCacheNodeInfo(string, map[string]v1alpha1.DiskDevice)
	DeleteCacheNodeInfo(string) error
	UpdateCacheNodeStatus(string, v1alpha1.NodeDiskIOStatsStatus) error
	IsIORequired(annotations map[string]string) bool
	NodeRegistered(string) bool
	AddPod(pod *v1.Pod, nodeName string, request v1alpha1.IOBandwidth) error
	RemovePod(*v1.Pod, string) error
	CanAdmitPod(string, v1alpha1.IOBandwidth) (bool, error)
	NodePressureRatio(string, v1alpha1.IOBandwidth) (float64, error)
	GetDiskNormalizeModel(string) (string, error)
	PrintCacheInfo()
}

type HandleBase struct {
	EC ExtendedCache
}

func (h *HandleBase) RemovePod(pod *v1.Pod, nodeName string) error {
	r := h.EC.GetExtendedResource(nodeName)
	if r != nil {
		err := r.RemovePod(pod)
		if err != nil {
			return err
		}
	}
	h.EC.PrintCacheInfo()
	return nil
}

func (h *HandleBase) AddPod(pod *v1.Pod, nodeName string, request v1alpha1.IOBandwidth) error {
	r := h.EC.GetExtendedResource(nodeName)
	if r != nil {
		return r.AddPod(pod, request)
	}
	return fmt.Errorf("cannot get extended resource: %v", nodeName)
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
