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

package resource

import (
	"context"
	"fmt"
	"sync"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/utils"
)

var IoiContext *ResourceIOContext

type StorageInfo struct {
	DevID            string
	RequestedStorage int64
	NodeName         string
}

type SyncContext struct {
	Node string
}

type ResourceIOContext struct {
	VClient     versioned.Interface
	Reservedpod map[string][]string             // nodename -> PodList
	PodRequests map[string]v1alpha1.IOBandwidth // poduid -> bw
	NsWhiteList []string
	queue       workqueue.RateLimitingInterface
	// lastUpdatedGen map[string]int64
	sync.Mutex
}

func NewContext(rl workqueue.RateLimiter, wl []string, h framework.Handle) (*ResourceIOContext, error) {
	queue := workqueue.NewNamedRateLimitingQueue(rl, "ResourceIO plugin")
	cfg := h.KubeConfig()
	cfg.ContentType = "application/json"
	c, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	return &ResourceIOContext{
		Reservedpod: make(map[string][]string),
		PodRequests: make(map[string]v1alpha1.IOBandwidth),
		NsWhiteList: wl,
		VClient:     c,
		queue:       queue,
		// lastUpdatedGen: make(map[string]int64),
	}, nil
}

func (c *ResourceIOContext) RunWorkerQueue(ctx context.Context) {
	for {
		obj, shutdown := c.queue.Get()
		if shutdown {
			break
		}
		err := func() error {
			defer c.queue.Done(obj)

			switch obj := obj.(type) {
			case *SyncContext: // update Reserved Pods
				pl, ok := c.Reservedpod[obj.Node]
				if !ok {
					return fmt.Errorf("node %v doesn't exist in reserved pod", obj.Node)
				}
				return utils.UpdateNodeIOStatus(ctx, c.VClient, obj.Node, pl)
			default:
				klog.Warningf("unexpected work item %#v", obj)
			}
			return nil
		}()
		if err != nil {
			klog.Errorf("work queue handle data error: %v", err)
			klog.Warningf("Retrying %#v after %d failures", obj, c.queue.NumRequeues(obj))
			c.queue.AddRateLimited(obj)
		} else {
			c.queue.Forget(obj)
		}
	}
}

func (c *ResourceIOContext) GetNsWhiteList() []string {
	return c.NsWhiteList
}

func (c *ResourceIOContext) GetReservedPods(node string) ([]string, error) {
	if v, ok := c.Reservedpod[node]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("node %v doesn't exist in reserved pod", node)
}

func (c *ResourceIOContext) GetPodRequest(pod string) (v1alpha1.IOBandwidth, error) {
	if v, ok := c.PodRequests[pod]; ok {
		return v, nil
	}
	return v1alpha1.IOBandwidth{}, fmt.Errorf("node %v doesn't exist in podRequests", pod)
}

func (c *ResourceIOContext) SetReservedPods(node string, pl []string) {
	c.Reservedpod[node] = pl
}

func (c *ResourceIOContext) SetPodRequests(podid string, req v1alpha1.IOBandwidth) {
	c.PodRequests[podid] = v1alpha1.IOBandwidth{
		Read:  req.Read.DeepCopy(),
		Write: req.Write.DeepCopy(),
		Total: req.Total.DeepCopy(),
	}
}

func (c *ResourceIOContext) RemoveNode(node string) {
	delete(c.Reservedpod, node)
}

func (c *ResourceIOContext) removePodRequest(podid string) {
	delete(c.PodRequests, podid)
}

func (c *ResourceIOContext) InNamespaceWhiteList(ns string) bool {
	for _, n := range c.NsWhiteList {
		if ns == n {
			return true
		}
	}
	return false
}

func (c *ResourceIOContext) AddPod(pod *corev1.Pod, nodeName string, bw v1alpha1.IOBandwidth) error {
	c.Lock()
	defer c.Unlock()

	pl, err := c.GetReservedPods(nodeName)
	if err != nil {
		return fmt.Errorf("get reserved pods error: %v", err)
	}
	pl = append(pl, fmt.Sprintf("%s-%s", pod.Name, pod.UID))
	c.SetPodRequests(string(pod.UID), bw)
	c.SetReservedPods(nodeName, pl)
	c.queue.Add(&SyncContext{Node: nodeName})
	return nil
}

func (c *ResourceIOContext) RemovePod(pod *corev1.Pod, nodeName string) error {
	c.Lock()
	defer c.Unlock()
	v, err := c.GetReservedPods(nodeName)
	if err != nil {
		return fmt.Errorf("get reserved pods error: %v", err)
	}
	for i, p := range v {
		if p == fmt.Sprintf("%s-%s", pod.Name, pod.UID) {
			v = append(v[:i], v[i+1:]...)
			break
		}
	}
	c.removePodRequest(string(pod.UID))
	c.SetReservedPods(nodeName, v)
	c.queue.Add(&SyncContext{Node: nodeName})
	return nil
}
