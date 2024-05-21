package iodriver

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type IODriver struct {
	mu                 sync.RWMutex
	podLister          corelisters.PodLister
	clientset          versioned.Interface
	nodeName           string
	diskInfos          *DiskInfos            // key:device id
	podCrInfos         map[string]*PodCrInfo // key:podUid
	observedGeneration *int64
}

func NewIODriver(ctx context.Context) (*IODriver, error) {
	n, ok := os.LookupEnv("Node_Name")
	if !ok {
		return nil, fmt.Errorf("init client failed: unable to get node name")
	}

	l, versioned, err := initClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("init client failed: %v", err)
	} else {
		log.Println("init client succeeded")
	}
	return &IODriver{
		clientset:          versioned,
		podLister:          l,
		nodeName:           n,
		diskInfos:          &DiskInfos{},
		observedGeneration: nil,
	}, nil
}

func initClient(ctx context.Context) (corelisters.PodLister, versioned.Interface, error) {
	var err error
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, nil, err
	}

	coreclient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("core client init err: %v", err)
	}

	//todo: change resync period
	coreFactory := informers.NewSharedInformerFactory(coreclient, 0)
	coreFactory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), coreFactory.Core().V1().Pods().Informer().HasSynced) {
		return nil, nil, fmt.Errorf("timed out waiting for caches to sync pods")
	}

	clientset, err := versioned.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("clientset init err: %v", err)
	}

	return coreFactory.Core().V1().Pods().Lister(), clientset, nil
}

func (c *IODriver) Run(ctx context.Context) {
	// get profile result and create DiskDevice CR
	info := GetDiskProfile()
	if err := c.ProcessProfile(info); err != nil {
		log.Fatal(err)
	}

	// watch NodeDiskIOStats and its handlers
	if err := c.WatchNodeDiskIOStats(ctx); err != nil {
		log.Fatal(err)
	}

	// update IO stats periodically to CR
	go wait.UntilWithContext(ctx, c.periodicUpdate, PeriodicUpdateInterval)

	<-ctx.Done()
	log.Println("IODriver stopped")
}

func (c *IODriver) periodicUpdate(ctx context.Context) {
	c.mu.RLock()
	bw := c.calDiskAllocatable(c.diskInfos.EmptyDir)
	c.mu.Unlock()

	toUpdate := make(map[string]v1alpha1.IOBandwidth)
	toUpdate[c.diskInfos.EmptyDir] = *bw
	err := c.UpdateNodeDiskIOInfoStatus(toUpdate)
	if err != nil {
		log.Printf("failed to update node io stats: %v", err)
	}
}
