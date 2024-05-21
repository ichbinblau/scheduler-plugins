package iodriver

import (
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type IORequest struct {
	Rbps      string `json:"rbps"`
	Wbps      string `json:"wbps"`
	BlockSize string `json:"blockSize"`
}

type DeviceType string

type DiskInfo struct {
	Name       string             `json:"name,omitempty"`
	Model      string             `json:"model,omitempty"`
	Vendor     string             `json:"vendor,omitempty"`
	MajorMinor string             `json:"majMin,omitempty"`
	Type       DeviceType         `json:"type,omitempty"`
	MountPoint string             `json:"mountPoint,omitempty"`
	Capacity   string             `json:"capacity,omitempty"`
	TotalBPS   resource.Quantity  `json:"totalBps,omitempty"`
	TotalRBPS  resource.Quantity  `json:"totalRbps,omitempty"`
	TotalWBPS  resource.Quantity  `json:"totalWbps,omitempty"`
	ReadRatio  map[string]float64 `json:"read_ratio,omitempty"`
	WriteRatio map[string]float64 `json:"write_ratio,omitempty"`
}

type DiskInfos struct {
	EmptyDir string
	Info     map[string]*DiskInfo // the map's key is disk id
}

type PodCrInfo struct {
	bw    *v1alpha1.IOBandwidth
	devId string // diskId
}
