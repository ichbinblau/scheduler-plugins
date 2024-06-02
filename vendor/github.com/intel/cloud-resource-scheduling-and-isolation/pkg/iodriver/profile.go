package iodriver

import (
	"log"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	FakeDeviceID = "fakeDevice"
)

var BlockSize = []string{"512", "1k", "4k", "8k", "16k", "32k"}

func GetFakeDevice() *DiskInfo {
	f := &DiskInfo{
		Name:       FakeDeviceID,
		Type:       EmptyDir,
		MajorMinor: "7:1",
		TotalRBPS:  resource.MustParse("1000Mi"),
		TotalWBPS:  resource.MustParse("1000Mi"),
		TotalBPS:   resource.MustParse("2000Mi"),
		Model:      "P4510",
		Vendor:     "Intel",
	}
	readRatios := []float64{1, 2, 3, 4, 5}
	writeRatios := []float64{1, 2, 3, 4, 6.2}
	for index, ratio := range readRatios {
		f.ReadRatio[BlockSize[index]] = ratio
	}
	for index, ratio := range writeRatios {
		f.WriteRatio[BlockSize[index]] = ratio
	}
	return f
}

// GetDiskProfile returns the disk profile result
// with fake device id and fake device info
// Customize your own profile tool to profile disks
func GetDiskProfile() *DiskInfos {
	pf := &DiskInfos{}
	pf.Info = make(map[string]*DiskInfo)
	pf.Info[FakeDeviceID] = GetFakeDevice()
	pf.EmptyDir = FakeDeviceID
	return pf
}

// get disk profile result and create CR
func (c *IODriver) ProcessProfile(di *DiskInfos) error {
	log.Printf("now in disk ProcessProfile: %v", di)

	devList := make(map[string]v1alpha1.DiskDevice)
	for n, dev := range di.Info {
		devList[n] = v1alpha1.DiskDevice{
			Name:   dev.Name,
			Vendor: dev.Vendor,
			Model:  dev.Model,
			Type:   string(dev.Type),
			Capacity: v1alpha1.IOBandwidth{
				Total: dev.TotalBPS,
				Read:  dev.TotalRBPS,
				Write: dev.TotalWBPS,
			},
		}
	}
	s := &v1alpha1.NodeDiskDevice{}
	n := c.nodeName
	s.Name = GetCRName(n, NodeDiskDeviceCRSuffix)
	s.Namespace = CRNameSpace
	s.Spec.NodeName = n
	s.Spec.Devices = devList

	err := c.CreateNodeDiskDeviceCR(s)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.diskInfos = di

	return nil
}
