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

// type Plugin struct {
// 	Name    string
// 	URL     string
// 	Retries int
// }

type PluginLoader struct {
	baseDir string
}

func (pl *PluginLoader) getFilePath(p PlConfig) string {
	return filepath.Join(pl.baseDir, fmt.Sprintf("%s-%s.so", p.Vendor, p.Model))
}

func (pl *PluginLoader) loadPlugin(p PlConfig) error {
	klog.Infof("Loading plugin %s-%s from %s\n", p.Vendor, p.Model, p.Source)

	return downloadFile(p.Source, pl.getFilePath(p))
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
func downloadFile(url string, filepath string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
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
		return err
	}
	return nil
}

func (pl *PluginLoader) LoadPlugin(p PlConfig) (Normalize, error) {
	klog.Infof("Loading plugin %s/%s\n", p.Vendor, p.Model)
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
	sym, err := plugin.Lookup("Normalize")
	if err != nil {
		return nil, fmt.Errorf("failed to lookup Normalize symbol: %v", err)
	}

	// get the normalizer
	norm, ok := sym.(Normalize)
	if !ok {
		return nil, errors.New("unexpected type from module symbol")
	}
	return norm, nil
}
