package diskio

import (
	"sync"

	// utils "sigs.k8s.io/IOIsolation/pkg"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type NodeInfo struct {
	DisksStatus    map[string]*DiskInfo // disk device uuid : diskStatus
	EmptyDirDevice string               // disk device uuid
}

type DiskInfo struct {
	DiskName       string
	BEDefaultRead  float64
	BEDefaultWrite float64
	// GA                 *utils.IOPoolStatus
	// BE                 *utils.IOPoolStatus
	Capacity int64
}

func (d *DiskInfo) DeepCopy() *DiskInfo {
	newDiskInfo := &DiskInfo{
		DiskName: d.DiskName,
		// ReadRatio:          make([]utils.MappingRatio, len(d.ReadRatio)),
		// WriteRatio:         make([]utils.MappingRatio, len(d.WriteRatio)),
		BEDefaultRead:  d.BEDefaultRead,
		BEDefaultWrite: d.BEDefaultWrite,
		// GA:                 &utils.IOPoolStatus{},
		// BE:                 &utils.IOPoolStatus{},
		Capacity: d.Capacity,
	}
	// newDiskInfo.GA = d.GA.DeepCopy()
	// newDiskInfo.BE = d.BE.DeepCopy()

	return newDiskInfo
}

func NewDiskInfo() *DiskInfo {
	diskInfo := &DiskInfo{
		// ReadRatio:  []utils.MappingRatio{},
		// WriteRatio: []utils.MappingRatio{},
		// GA:         &utils.IOPoolStatus{},
		// BE:         &utils.IOPoolStatus{},
	}
	return diskInfo
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
