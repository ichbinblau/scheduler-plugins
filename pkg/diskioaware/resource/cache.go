package resource

import (
	"k8s.io/klog/v2"
)

type ExtendedCache interface {
	SetExtendedResource(nodeName string, val ExtendedResource)
	GetExtendedResource(nodeName string) ExtendedResource
	DeleteExtendedResource(nodeName string)
	PrintCacheInfo()
}

type ResourceCache struct {
	Resources map[string]ExtendedResource
}

func NewExtendedCache() ExtendedCache {
	c := &ResourceCache{
		Resources: make(map[string]ExtendedResource),
	}

	return c
}

func (cache *ResourceCache) SetExtendedResource(nodeName string, val ExtendedResource) {
	cache.Resources[nodeName] = val
}

func (cache *ResourceCache) GetExtendedResource(nodeName string) ExtendedResource {
	val, ok := cache.Resources[nodeName]
	if !ok {
		return nil
	}
	return val
}

func (cache *ResourceCache) DeleteExtendedResource(nodeName string) {
	delete(cache.Resources, nodeName)
}

func (cache *ResourceCache) PrintCacheInfo() {
	klog.Info("Print Cache Info")
	for node, er := range cache.Resources {
		klog.Infof("==========Node: %s=============", node)
		er.PrintInfo()
	}
}
