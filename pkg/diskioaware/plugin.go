package diskioaware

import (
	"context"
	"encoding/json"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/apis/config"
	"sigs.k8s.io/scheduler-plugins/apis/config/validation"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/normalizer"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type DiskIO struct {
	rhs    resource.Handle
	scorer Scorer
	nm     *normalizer.NormalizerManager
	client kubernetes.Interface
}

const (
	Name           = "DiskIO"
	stateKeyPrefix = "DiskIO-"
	diskVendorKey  = "diskVendors"
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

func loadModels(config *config.DiskIOArgs, cmLister corelisters.ConfigMapLister) (*normalizer.NormalizerManager, error) {
	name, ns := config.ConfigMapName, config.ConfigMapNamespace
	cm, err := cmLister.ConfigMaps(name).Get(ns)
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %v", ns, name, err)
	}
	data, ok := cm.Data[diskVendorKey]
	if !ok {
		return nil, fmt.Errorf("failed to load disk vendor data %v: %v", cm.Data, err)
	}
	pls := &normalizer.PlList{}
	if err := json.Unmarshal([]byte(data), pls); err != nil {
		return nil, fmt.Errorf("failed to deserialize configmap %s/%s: %v", ns, name, err)
	}
	nm := normalizer.NewNormalizerManager(baseModelDir, maxRetries)
	go nm.LoadPlugins(*pls, workers)
	return nm, nil
}

// New initializes a new plugin and returns it.
func New(configuration runtime.Object, handle framework.Handle) (framework.Plugin, error) {
	args, ok := configuration.(*config.DiskIOArgs)
	if !ok {
		return nil, fmt.Errorf("want args to be of type DiskIOArgs, got %T", args)
	}

	// validate args
	if err := validation.ValidateDiskIOArgs(nil, args); err != nil {
		return nil, err
	}

	// load disk vendor normalize functions
	nm, err := loadModels(args, handle.SharedInformerFactory().Core().V1().ConfigMaps().Lister())
	if err != nil {
		return nil, err
	}

	// initialize scorer
	scorer, err := getScorer(args.ScoreStrategy)
	if err != nil {
		return nil, err
	}

	// todo: initialize event handling

	// todo: initialize workqueue to send reserve request

	return &DiskIO{
		rhs:    nil,
		scorer: scorer,
		nm:     nm,
		client: handle.ClientSet(),
	}, nil
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
		{Resource: framework.Node, ActionType: framework.Delete | framework.UpdateNodeLabel},
		{Resource: framework.GVK(fmt.Sprintf("nodestaticioinfoes.v1alpha1.%v", "xyz")), ActionType: framework.All},
	}
	return ce
}
