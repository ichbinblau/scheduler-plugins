package normalizer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/scheduler-plugins/apis/config"
)

const (
	diskVendorKey  = "diskModels.properties"
	resyncDuration = 30 * time.Second
)

type PlList []PlConfig

type PlConfig struct {
	Vendor   string `json:"vendor"`
	Model    string `json:"model"`
	Source   string `json:"source"`
	SkipTLS  bool   `json:"skipTLS"`
	CertPath string `json:"certPath"`
}

type NormalizerManager struct {
	sync.RWMutex
	store  *nStore
	loader *PluginLoader
}

func NewNormalizerManager(base string, m int) *NormalizerManager {
	return &NormalizerManager{
		store:  NewnStore(),
		loader: &PluginLoader{baseDir: base, maxRetries: m},
	}
}

func (pm *NormalizerManager) Run(ctx context.Context, config *config.DiskIOArgs, num int, cmLister corelisters.ConfigMapLister) {
	name, ns := config.DiskIOModelConfig, config.DiskIOModelConfigNS
	var periodJob = func(ctx context.Context) {
		cm, err := cmLister.ConfigMaps(name).Get(ns)
		if err != nil {
			klog.Errorf("failed to get configmap %s/%s: %v", ns, name, err)
		}

		data, ok := cm.Data[diskVendorKey]
		if !ok {
			klog.Errorf("failed to load disk vendor data %v: %v", cm.Data, err)
		}
		pls := &PlList{}
		if err := json.Unmarshal([]byte(data), pls); err != nil {
			klog.Errorf("failed to deserialize configmap %s/%s: %v", ns, name, err)
		}

		pm.LoadPlugins(*pls, num)
	}
	go wait.UntilWithContext(ctx, periodJob, resyncDuration)
}

// LoadPlugin implements the interface method
func (pm *NormalizerManager) LoadPlugins(l PlList, workers int) error {
	// todo: cancel with context
	// load plugins
	if len(l) == 0 {
		return errors.New("plugin list cannot be empty")
	}

	pls := make(chan PlConfig, 10) // queue of download jobs
	var wg sync.WaitGroup

	pm.Lock()
	defer pm.Unlock()
	// start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pl := range pls {
				// use Vendor+Model as key,
				key := fmt.Sprintf("%s-%s", pl.Vendor, pl.Model)
				if !isValidURL(pl.Source) {
					klog.Infof("Invalid url: %s for %s\n", pl.Source, pl.Model)
				}
				norm, err := pm.loader.LoadPlugin(pl)
				if err != nil {
					klog.Infof("Failed to load %v: %v\n", pl.Source, err)
				}
				// normalizer functions as value
				pm.store.Set(key, norm)
			}
		}()
	}

	// enqueue not existing plugins to load
	for _, p := range l {
		key := fmt.Sprintf("%s-%s", p.Vendor, p.Model)
		if pm.store.Contains(key) {
			continue
		}
		pls <- p
	}

	close(pls) // no more plugins will be added
	wg.Wait()  // wait for all tasks to complete
	return nil
}

// UnloadPlugin implements the interface method
func (pm *NormalizerManager) UnloadPlugin(name string) error {
	if len(name) == 0 {
		return errors.New("plugin name cannot be empty")
	}

	pm.Lock()
	defer pm.Unlock()
	pm.store.Delete(name)

	return nil
}

// GetPlugin implements the interface method
func (pm *NormalizerManager) GetNormalizer(name string) (Normalize, error) {

	pm.Lock()
	defer pm.Unlock()
	exec, err := pm.store.Get(name)
	if err != nil {
		return nil, err
	}

	return exec, nil
}
