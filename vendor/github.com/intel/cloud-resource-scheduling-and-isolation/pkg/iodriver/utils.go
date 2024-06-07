package iodriver

import (
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	NodeDiskDeviceCRSuffix string        = "nodediskdevice"
	NodeDiskIOInfoCRSuffix string        = "nodediskiostats"
	CRNameSpace            string        = "ioi-system"
	DiskIOAnnotation       string        = "blockio.kubernetes.io/resources"
	NodeIOStatusCR         string        = "NodeDiskIOStats"
	APIVersion             string        = "ioi.intel.com/v1"
	PeriodicUpdateInterval time.Duration = 5 * time.Second
	Mi                     float64       = 1024 * 1024
	EmptyDir               DeviceType    = "emptyDir"
	Others                 DeviceType    = "others"
)

var (
	MinDefaultIOBW      resource.Quantity = resource.MustParse("5Mi")
	MinDefaultTotalIOBW resource.Quantity = resource.MustParse("10Mi")

	UpdateBackoff = wait.Backoff{
		Steps:    3,
		Duration: 100 * time.Millisecond, // 0.1s
		Jitter:   1.0,
	}
)

func GetCRName(n string, suff string) string {
	return fmt.Sprintf("%v-%v", n, suff)
}

func GetSliceIdx(target string, s []string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}
