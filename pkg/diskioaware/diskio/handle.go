package diskio

import (
	"sync"

	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type Handle struct {
	resource.HandleBase
	client kubernetes.Interface
	sync.RWMutex
}

// ioiresource.Handle interface
func (h *Handle) Name() string {
	return "DiskIOHandle"
}

func (h *Handle) Run(c resource.ExtendedCache, cli kubernetes.Interface) error {
	h.EC = c
	h.client = cli
	return nil
}

// New Resource Handle
func New() resource.Handle {
	return &Handle{}
}
