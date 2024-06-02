package normalizer

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"plugin"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

const (
	defaultMaxRetry = 1
	defaultTimeout  = 360 * time.Second
)

type PluginLoader struct {
	baseDir string
	client  *http.Client
}

func NewPluginLoader(base string) *PluginLoader {
	return &PluginLoader{
		baseDir: base,
		client:  http.DefaultClient,
	}
}

func (pl *PluginLoader) getFilePath(p PlConfig) string {
	return filepath.Join(pl.baseDir, fmt.Sprintf("%s-%s.so", p.Vendor, p.Model))
}

func (pl *PluginLoader) loadPlugin(ctx context.Context, p PlConfig) (string, error) {
	klog.Infof("Loading plugin %s-%s.so from %s\n", p.Vendor, p.Model, p.URL)

	firstColon := strings.IndexByte(p.URL, ':')
	if firstColon == -1 {
		return "", fmt.Errorf("invalid URL: %s", p.URL)
	}

	scheme := p.URL[:firstColon]
	switch scheme {
	case "http", "https":
		dst := pl.getFilePath(p)
		if err := downloadFile(ctx, pl.client, p.URL, dst); err != nil {
			return "", err
		}
		return dst, nil
	case "file":
		localPath := p.URL[7:] // strip file://
		if _, err := os.Stat(localPath); err != nil {
			return "", fmt.Errorf("local file not found: %s", localPath)
		}
		return filepath.Clean(localPath), nil
	default:
		return "", fmt.Errorf("unsupported URL scheme: %s", scheme)
	}
}

// DownloadFile will download a url to a local file.
func downloadFile(ctx context.Context, client *http.Client, url string, filepath string) error {
	// c, cancel := context.WithTimeout(ctx, defaultTimeout)
	// defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	// klog.Infof("download begin %s", url)
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download plugin: %v", err)
	}
	defer resp.Body.Close()
	// klog.Infof("download done %s", url)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received %v status code from %q", resp.StatusCode, url)
	}
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

func (pl *PluginLoader) LoadPlugin(ctx context.Context, p PlConfig) (Normalize, error) {
	if p.Vendor == "" || p.Model == "" || p.URL == "" {
		return nil, fmt.Errorf("invalid plugin configuration")
	}
	dst, err := pl.loadPlugin(ctx, p)
	if err != nil {
		return nil, err
	}

	// load the plugin
	plugin, err := plugin.Open(dst)
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
