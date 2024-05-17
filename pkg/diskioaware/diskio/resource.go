package diskio

import (
	"fmt"
	"sync"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type NodeInfo struct {
	DisksStatus   map[string]*DiskInfo // disk device uuid : diskStatus
	DefaultDevice string               // disk device uuid
}

type DiskInfo struct {
	NormalizerName string
	DiskName       string
	Capacity       v1alpha1.IOBandwidth
	Allocatable    v1alpha1.IOBandwidth
}

func (d *DiskInfo) DeepCopy() *DiskInfo {
	newDiskInfo := &DiskInfo{
		DiskName:       d.DiskName,
		Capacity:       d.Capacity,
		Allocatable:    d.Allocatable,
		NormalizerName: d.NormalizerName,
	}

	return newDiskInfo
}

type Resource struct {
	nodeName string
	info     *NodeInfo
	ch       resource.CacheHandle
	sync.RWMutex
}

func (ps *Resource) Name() string {
	return "BlockIO"
}

func (ps *Resource) AddPod(pod *v1.Pod, request v1alpha1.IOBandwidth) error {
	dev := ps.info.DefaultDevice
	if _, ok := ps.info.DisksStatus[dev]; !ok {
		return fmt.Errorf("cannot find default device %s in cache", dev)
	}
	ps.info.DisksStatus[dev].Allocatable.Read.Sub(request.Read)
	ps.info.DisksStatus[dev].Allocatable.Write.Sub(request.Write)
	ps.info.DisksStatus[dev].Allocatable.Total.Sub(request.Total)
	return nil
}

func (ps *Resource) RemovePod(pod *v1.Pod) error {
	request, err := resource.IoiContext.GetPodRequest(string(pod.UID))
	if err != nil {
		return fmt.Errorf("cannot get pod request: %v", err)
	}
	dev := ps.info.DefaultDevice
	if _, ok := ps.info.DisksStatus[dev]; !ok {
		return fmt.Errorf("cannot find default device %s in cache", dev)
	}
	ps.info.DisksStatus[dev].Allocatable.Read.Add(request.Read)
	ps.info.DisksStatus[dev].Allocatable.Write.Add(request.Write)
	ps.info.DisksStatus[dev].Allocatable.Total.Add(request.Total)
	return nil
}

//	func (ps *Resource) AdmitPod(pod *v1.Pod) (v1alpha1.IOBandwidth, error) {
//		return v1alpha1.IOBandwidth{}, nil
//	}
func (ps *Resource) PrintInfo() {
	for disk, diskInfo := range ps.info.DisksStatus {
		klog.Info("***device: ", disk, " ***")
		klog.Info("device name: ", diskInfo.DiskName)
		klog.Info("normalizer name: ", diskInfo.NormalizerName)
		klog.Info("capacity read: ", diskInfo.Capacity.Read.String())
		klog.Info("capacity write: ", diskInfo.Capacity.Write.String())
		klog.Info("capacity total: ", diskInfo.Capacity.Total.String())
		klog.Info("allocatable read: ", diskInfo.Capacity.Read.String())
		klog.Info("allocatable write: ", diskInfo.Capacity.Write.String())
		klog.Info("allocatable total: ", diskInfo.Capacity.Total.String())
	}
}
