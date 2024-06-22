/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
	ps.Lock()
	defer ps.Unlock()
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
	resource.IoiContext.Lock()
	request, err := resource.IoiContext.GetPodRequest(string(pod.UID))
	if err != nil {
		return fmt.Errorf("cannot get pod request: %v", err)
	}
	resource.IoiContext.Unlock()
	ps.Lock()
	defer ps.Unlock()
	dev := ps.info.DefaultDevice
	if _, ok := ps.info.DisksStatus[dev]; !ok {
		return fmt.Errorf("cannot find default device %s in cache", dev)
	}
	ps.info.DisksStatus[dev].Allocatable.Read.Add(request.Read)
	ps.info.DisksStatus[dev].Allocatable.Write.Add(request.Write)
	ps.info.DisksStatus[dev].Allocatable.Total.Add(request.Total)
	return nil
}

func (ps *Resource) PrintInfo() {
	for disk, diskInfo := range ps.info.DisksStatus {
		klog.V(6).Info("***device: ", disk, " ***")
		klog.V(6).Info("device name: ", diskInfo.DiskName)
		klog.V(6).Info("normalizer name: ", diskInfo.NormalizerName)
		klog.V(6).Info("capacity read: ", diskInfo.Capacity.Read.String())
		klog.V(6).Info("capacity write: ", diskInfo.Capacity.Write.String())
		klog.V(6).Info("capacity total: ", diskInfo.Capacity.Total.String())
		klog.V(6).Info("allocatable read: ", diskInfo.Allocatable.Read.String())
		klog.V(6).Info("allocatable read: ", diskInfo.Allocatable.Read.String())
		klog.V(6).Info("allocatable write: ", diskInfo.Allocatable.Write.String())
		klog.V(6).Info("allocatable total: ", diskInfo.Allocatable.Total.String())
	}
}
