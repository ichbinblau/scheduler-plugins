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
	"testing"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/clientset/versioned"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"
)

func TestResourceIOContext_AddPod(t *testing.T) {
	type fields struct {
		VClient     versioned.Interface
		Reservedpod map[string][]string
		PodRequests map[string]v1alpha1.IOBandwidth
		NsWhiteList []string
		queue       workqueue.RateLimitingInterface
	}
	type args struct {
		pod      *corev1.Pod
		nodeName string
		bw       v1alpha1.IOBandwidth
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Add pod to ResourceIOContext with success",
			fields: fields{
				Reservedpod: map[string][]string{
					"node1": {
						"pod2-456",
					},
				},
				PodRequests: map[string]v1alpha1.IOBandwidth{},
				queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test"),
			},
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
						UID:  "123",
					},
				},
				nodeName: "node1",
				bw: v1alpha1.IOBandwidth{
					Total: resource.MustParse("1000"),
					Read:  resource.MustParse("500"),
					Write: resource.MustParse("500"),
				},
			},
			wantErr: false,
		},
		{
			name: "Add pod to ResourceIOContext fails",
			fields: fields{
				Reservedpod: map[string][]string{},
				PodRequests: map[string]v1alpha1.IOBandwidth{},
				queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test"),
			},
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
						UID:  "123",
					},
				},
				nodeName: "node1",
				bw: v1alpha1.IOBandwidth{
					Total: resource.MustParse("1000"),
					Read:  resource.MustParse("500"),
					Write: resource.MustParse("500"),
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ResourceIOContext{
				VClient:     tt.fields.VClient,
				Reservedpod: tt.fields.Reservedpod,
				PodRequests: tt.fields.PodRequests,
				NsWhiteList: tt.fields.NsWhiteList,
				queue:       tt.fields.queue,
			}
			if err := c.AddPod(tt.args.pod, tt.args.nodeName, tt.args.bw); (err != nil) != tt.wantErr {
				t.Errorf("ResourceIOContext.AddPod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestResourceIOContext_RemovePod(t *testing.T) {
	type fields struct {
		VClient     versioned.Interface
		Reservedpod map[string][]string
		PodRequests map[string]v1alpha1.IOBandwidth
		NsWhiteList []string
		queue       workqueue.RateLimitingInterface
	}
	type args struct {
		pod      *corev1.Pod
		nodeName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Remove pod from ResourceIOContext with success",
			fields: fields{
				Reservedpod: map[string][]string{
					"node1": {
						"pod1-123",
						"pod2-456",
					},
				},
				PodRequests: map[string]v1alpha1.IOBandwidth{
					"123": {
						Total: resource.MustParse("1000"),
						Read:  resource.MustParse("500"),
						Write: resource.MustParse("500"),
					},
					"456": {
						Total: resource.MustParse("1000"),
						Read:  resource.MustParse("500"),
						Write: resource.MustParse("500"),
					},
				},
				queue: workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test"),
			},
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
						UID:  "123",
					},
				},
				nodeName: "node1",
			},
			wantErr: false,
		},
		{
			name: "Remove pod to ResourceIOContext fails",
			fields: fields{
				Reservedpod: map[string][]string{},
				PodRequests: map[string]v1alpha1.IOBandwidth{},
				queue:       workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "test"),
			},
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name: "pod1",
						UID:  "123",
					},
				},
				nodeName: "node1",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ResourceIOContext{
				VClient:     tt.fields.VClient,
				Reservedpod: tt.fields.Reservedpod,
				PodRequests: tt.fields.PodRequests,
				NsWhiteList: tt.fields.NsWhiteList,
				queue:       tt.fields.queue,
			}
			if err := c.RemovePod(tt.args.pod, tt.args.nodeName); (err != nil) != tt.wantErr {
				t.Errorf("ResourceIOContext.RemovePod() error = %v, wantErr %v", err, tt.wantErr)
			}
			klog.Info(c.Reservedpod)
			klog.Info(c.PodRequests)
		})
	}
}
