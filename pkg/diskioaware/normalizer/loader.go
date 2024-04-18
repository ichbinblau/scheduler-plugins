package normalizer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"plugin"

	"k8s.io/klog/v2"
)

const defaultMaxRetry = 1

type PluginLoader struct {
	baseDir    string
	maxRetries int
}

func NewPluginLoader(base string, maxRetries int) *PluginLoader {
	if maxRetries <= 1 {
		maxRetries = defaultMaxRetry
	}
	return &PluginLoader{
		baseDir:    base,
		maxRetries: maxRetries,
	}
}

func (pl *PluginLoader) getFilePath(p PlConfig) string {
	return filepath.Join(pl.baseDir, fmt.Sprintf("%s-%s.so", p.Vendor, p.Model))
}

func (pl *PluginLoader) loadPlugin(p PlConfig) error {
	klog.Infof("Loading plugin %s-%s.so from %s\n", p.Vendor, p.Model, p.Source)

	return pl.downloadFile(p.Source, pl.getFilePath(p))
}

func isValidURL(u string) bool {
	parsedURL, err := url.Parse(u)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false
	}
	return true
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func (pl *PluginLoader) downloadFile(url string, filepath string) error {
	for i := 0; i < pl.maxRetries; i++ {
		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			return err
		}
		defer out.Close()

		// Write the body to file
		if _, err = io.Copy(out, resp.Body); err != nil {
			continue
		}

		return nil
	}
	return fmt.Errorf("failed to download %s after %d attempts", url, 3)
}

func (pl *PluginLoader) LoadPlugin(p PlConfig) (Normalize, error) {
	klog.Infof("Loading plugin %s/%s.so\n", p.Vendor, p.Model)
	if err := pl.loadPlugin(p); err != nil {
		return nil, fmt.Errorf("failed to download plugin: %v", err)
	}

	// todo: verify signature
	// load the plugin
	plugin, err := plugin.Open(pl.getFilePath(p))
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %v", err)
	}

	// find symbol
	normSym, err := plugin.Lookup("Normalizer")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup Normalizer symbol: %v", err)
	}

	// get the normalizer class
	var claz Normalizer
	claz, ok := normSym.(Normalizer)
	if !ok {
		return nil, errors.New("unexpected type from module symbol")
	}

	return claz.EstimateRequest, nil
}
