package normalizer

import (
	"errors"
	"fmt"
	"sync"

	"k8s.io/klog/v2"
)

type PlList []PlConfig

type PlConfig struct {
	Vendor string `json:"vendor"`
	Model  string `json:"model"`
	Source string `json:"source"`
}

type NormalizerManager struct {
	sync.RWMutex
	store  *nStore
	loader *PluginLoader
}

func NewNormalizerManager(base string) *NormalizerManager {
	return &NormalizerManager{
		store:  NewnStore(),
		loader: &PluginLoader{baseDir: base},
	}
}

// LoadPlugin implements the interface method
func (pm *NormalizerManager) LoadPlugins(l PlList, maxRetries int, workers int) error {
	// load plugins
	if len(l) == 0 {
		return errors.New("plugin list cannot be empty")
	}

	pls := make(chan PlConfig, 100) // queue of download jobs
	var wg sync.WaitGroup

	pm.Lock()
	defer pm.Unlock()
	// start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for pl := range pls {
				if !isValidURL(pl.Source) {
					klog.Infof("Invalid url: %s for %s\n", pl.Source, pl.Model)
				}
				norm, err := pm.loader.LoadPlugin(pl)
				if err != nil {
					retries := maxRetries
					if retries > 0 {
						klog.Infof("Retrying load for %s\n", pl.Source)
						retries--
						pls <- pl // requeue job
					} else {
						klog.Infof("Failed to load %s after %i attempts\n", maxRetries, pl.Source)
					}
				}
				// use Vendor+Model as key
				pm.store.Set(fmt.Sprintf("%s-%s", pl.Vendor, pl.Model), norm)
			}
		}()
	}

	// enqueue plugins to load
	for _, p := range l {
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
	// read lock or not?
	pm.Lock()
	defer pm.Unlock()
	exec, err := pm.store.Get(name)
	if err != nil {
		return nil, err
	}

	return exec, nil
}
