package diskioaware

import (
	"context"
	"fmt"
	"time"

	externalinformer "github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/informers/externalversions"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"sigs.k8s.io/scheduler-plugins/apis/config/validation"
	"sigs.k8s.io/scheduler-plugins/apis/scheduling"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/diskio"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/normalizer"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type DiskIO struct {
	rh     resource.Handle
	scorer Scorer
	nm     *normalizer.NormalizerManager
	// client        versioned.Interface
	// sharedFactory externalversions.SharedInformerFactory
}

const (
	Name           = "DiskIO"
	stateKeyPrefix = "DiskIO-"
	maxRetries     = 3
	workers        = 2
	baseModelDir   = "/tmp"
)

var _ = framework.FilterPlugin(&DiskIO{})
var _ = framework.ScorePlugin(&DiskIO{})
var _ = framework.ReservePlugin(&DiskIO{})

type stateData struct {
	request           interface{}
	nodeResourceState interface{} // change name
	nodeSupportIOI    bool
}

func (d *stateData) Clone() framework.StateData {
	return d
}

// New initializes a new plugin and returns it.
func New(configuration runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	d := &DiskIO{
		rh: diskio.New(),
	}
	ctx := context.Background()
	args, ok := configuration.(*config.DiskIOArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type DiskIOArgs, got %T", args)
	}

	// validate args
	if err := validation.ValidateDiskIOArgs(nil, args); err != nil {
		return nil, err
	}

	// load disk vendor normalize functions
	// watch configmap with version
	d.nm = normalizer.NewNormalizerManager(baseModelDir, maxRetries)
	d.nm.Run(ctx, args, workers, handle.SharedInformerFactory().Core().V1().ConfigMaps().Lister())

	// initialize scorer
	scorer, err := getScorer(args.ScoreStrategy)
	if err != nil {
		return nil, err
	}
	d.scorer = scorer

	// initialize disk IO resource handler
	err = d.rh.Run(resource.NewExtendedCache(), handle.ClientSet())
	if err != nil {
		return nil, err
	}

	ratelimiter := workqueue.NewItemExponentialFailureRateLimiter(time.Second, 5*time.Second) // todo: load from config
	resource.IoiContext, err = resource.NewContext(ratelimiter, args.NSWhiteList, handle)
	if err != nil {
		return nil, err
	}
	go resource.IoiContext.RunWorkerQueue(ctx)

	// initialize event handling
	podLister := handle.SharedInformerFactory().Core().V1().Pods().Lister()
	nsLister := handle.SharedInformerFactory().Core().V1().Namespaces().Lister()
	eh := resource.NewIOEventHandler(d.rh, handle, podLister, nsLister)

	podInformer := handle.SharedInformerFactory().Core().V1().Pods().Informer()
	iof := externalinformer.NewSharedInformerFactory(resource.IoiContext.VClient, 0)
	err = eh.BuildEvtHandler(ctx, podInformer, iof)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (ps *DiskIO) Name() string {
	return Name
}

// Filter invoked at the filter extension point.
// Checks if a node has sufficient resources, such as cpu, memory, gpu, opaque int resources etc to run a pod.
// It returns a list of insufficient resources, if empty, then the node has all the resources requested by the pod.
func (ps *DiskIO) Filter(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeInfo *framework.NodeInfo) *framework.Status {

	return framework.NewStatus(framework.Success)
}

// Score invoked at the score extension point.
func (ps *DiskIO) Score(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) (int64, *framework.Status) {
	return 100, framework.NewStatus(framework.Success)
}

// ScoreExtensions of the Score plugin.
func (ps *DiskIO) ScoreExtensions() framework.ScoreExtensions {
	return nil
}

// Reserve is the functions invoked by the framework at "reserve" extension point.
func (ps *DiskIO) Reserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) *framework.Status {
	return framework.NewStatus(framework.Success, "")
}

func (ps *DiskIO) Unreserve(ctx context.Context, state *framework.CycleState, pod *v1.Pod, nodeName string) {
}

// EventsToRegister returns the possible events that may make a Pod
// failed by this plugin schedulable.
// NOTE: if in-place-update (KEP 1287) gets implemented, then PodUpdate event
// should be registered for this plugin since a Pod update may free up resources
// that make other Pods schedulable.
func (ps *DiskIO) EventsToRegister() []framework.ClusterEvent {
	// To register a custom event, follow the naming convention at:
	// https://git.k8s.io/kubernetes/pkg/scheduler/eventhandlers.go#L410-L422

	// todo: change action type
	ce := []framework.ClusterEvent{
		{Resource: framework.Pod, ActionType: framework.All},
		// {Resource: framework.Node, ActionType: framework.Delete | framework.UpdateNodeLabel},
		{Resource: framework.GVK(fmt.Sprintf("nodediskdevices.v1alpha1.%v", scheduling.GroupName)), ActionType: framework.Add | framework.Delete},
		{Resource: framework.GVK(fmt.Sprintf("nodediskiostatses.v1alpha1.%v", scheduling.GroupName)), ActionType: framework.Update},
	}
	return ce
}
