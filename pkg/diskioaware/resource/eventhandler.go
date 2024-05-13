package resource

import (
	"context"
	"fmt"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/clientset/versioned"
	externalinformer "github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
)

type IOEventHandler struct {
	cache     Handle              // resourceType -> handle
	vClient   versioned.Interface // versioned client for CRDs
	podLister corelisters.PodLister
	nsLister  corelisters.NamespaceLister
}

func NewIOEventHandler(cache Handle, h framework.Handle, pl corelisters.PodLister, nl corelisters.NamespaceLister) *IOEventHandler {
	return &IOEventHandler{
		cache:     cache,
		vClient:   IoiContext.VClient,
		podLister: pl,
		nsLister:  nl,
	}
}

func (ps *IOEventHandler) BuildEvtHandler(ctx context.Context, podInformer cache.SharedIndexInformer, iof externalinformer.SharedInformerFactory) error {
	ddInformer := iof.Diskio().V1alpha1().NodeDiskDevices().Informer()
	diInformer := iof.Diskio().V1alpha1().NodeDiskIOStatses().Informer()
	dh := cache.ResourceEventHandlerFuncs{
		AddFunc:    ps.AddDiskDevice,
		DeleteFunc: ps.DeleteDiskDevice,
	}
	if _, err := ddInformer.AddEventHandler(dh); err != nil {
		return err
	}
	// NodeIOStatus event handler
	ih := cache.ResourceEventHandlerFuncs{
		UpdateFunc: ps.UpdateNodeDiskIOStats,
	}
	if _, err := diInformer.AddEventHandler(ih); err != nil {
		return err
	}
	iof.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), ddInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync resource NodeStaticIOInfo")
	}
	if !cache.WaitForCacheSync(ctx.Done(), diInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync resource NodeIOStatus")
	}
	fhandler := cache.FilteringResourceEventHandler{
		FilterFunc: func(obj interface{}) bool {
			pod, ok := obj.(*v1.Pod)
			if !ok {
				klog.Errorf("cannot convert to *v1.Pod: %v", obj)
				return false
			}
			if IoiContext.InNamespaceWhiteList(pod.Namespace) {
				return false
			}
			if pod.Spec.NodeName == "" {
				return false
			}
			return true
		},
		Handler: cache.ResourceEventHandlerFuncs{
			DeleteFunc: ps.DeletePod,
		},
	}
	if _, err := podInformer.AddEventHandler(fhandler); err != nil {
		return err
	}
	if !cache.WaitForCacheSync(ctx.Done(), podInformer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync resource NodeIOStatus")
	}
	return nil
}

func (ps *IOEventHandler) AddDiskDevice(obj interface{}) {
	_, ok := obj.(*v1alpha1.NodeDiskDevice)
	if !ok {
		klog.Errorf("[AddDiskDevice]cannot convert obj to NodeDiskDevice: %v", obj)
		return
	}
}

func (ps *IOEventHandler) DeleteDiskDevice(obj interface{}) {
	_, ok := obj.(*v1alpha1.NodeDiskDevice)
	if !ok {
		klog.Errorf("[DeleteDiskDevice]cannot convert obj to NodeDiskDevice: %v", obj)
		return
	}
}

func (ps *IOEventHandler) UpdateNodeDiskIOStats(oldObj, newObj interface{}) {
}

func (ps *IOEventHandler) DeletePod(obj interface{}) {
	_, ok := obj.(*v1.Pod)
	if !ok {
		klog.Errorf("[DeletePod]cannot convert to *v1.Pod: %v", obj)
		return
	}
}
