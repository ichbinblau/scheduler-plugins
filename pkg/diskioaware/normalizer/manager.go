package normalizer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
	corelisters "k8s.io/client-go/listers/core/v1"
	"k8s.io/klog/v2"
)

const (
	diskModelConfig = "/etc/kubernetes/diskModels.properties"
	resyncDuration  = 90 * time.Second
)

type PlList []PlConfig

type PlConfig struct {
	Vendor  string `json:"vendor"`
	Model   string `json:"model"`
	URL     string `json:"url"`
	SkipTLS bool   `json:"skipTLS"`
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

func (pm *NormalizerManager) Run(ctx context.Context, num int, cmLister corelisters.ConfigMapLister) {
	var periodJob = func(ctx context.Context) {
		data, err := os.ReadFile(diskModelConfig)
		if err != nil {
			klog.Errorf("failed to load disk model config: %v", err)
		}
		pls := &PlList{}
		if err := json.Unmarshal(data, pls); err != nil {
			klog.Errorf("failed to deserialize disk model config: %v", err)
		}
		pm.LoadPlugins(ctx, *pls, num)
	}
	go wait.UntilWithContext(ctx, periodJob, resyncDuration)
}

// LoadPlugin implements the interface method
func (pm *NormalizerManager) LoadPlugins(ctx context.Context, l PlList, workers int) error {
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
				norm, err := pm.loader.LoadPlugin(ctx, pl)
				if err != nil {
					klog.Infof("Failed to load %v: %v\n", pl.URL, err)
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
