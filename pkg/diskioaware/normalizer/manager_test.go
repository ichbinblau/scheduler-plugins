package normalizer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func getBytes(fileName string) ([]byte, error) {
	if len(fileName) == 0 {
		return nil, fmt.Errorf("file name is empty")
	}
	p, err := os.ReadFile(fmt.Sprintf("../sample-plugin/%v.so", fileName))
	if err != nil {
		return nil, fmt.Errorf("failed to example.so: %v", err)
	}
	return p, nil
}

func TestNormalizerManager_LoadPlugins(t *testing.T) {
	p, err := getBytes("foo")
	if err != nil {
		t.Fatalf("failed to foo.so: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(p)
	}))
	defer ts.Close()

	tests := []struct {
		name     string
		plugins  *PlList
		expected bool
	}{
		{
			name:     "Empty plugin list",
			plugins:  &PlList{},
			expected: false,
		},
		{
			name: "Successful single plugin loading",
			plugins: &PlList{
				{Vendor: "Intel", Model: "P1111", URL: ts.URL + "/m1"},
			},
			expected: true,
		},
		{
			name: "Failed plugin loading",
			plugins: &PlList{
				{Vendor: "vendor3", Model: "model3", URL: "http://localhost:8080"},
				{Vendor: "vendor4", Model: "model4", URL: "http://localhost:8080"},
			},
			expected: true,
		},
	}

	nm := NewNormalizerManager("../sample-plugin", 3)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := nm.LoadPlugins(context.Background(), *tt.plugins, 2)
			if (err == nil) != tt.expected {
				t.Errorf("case: %v failed expected error=%v", tt.name, tt.expected)
			}
			for _, p := range *tt.plugins {
				os.Remove(nm.loader.getFilePath(p))
			}
		})
	}
}

func TestNormalizerManager_UnloadPlugin(t *testing.T) {
	s := NewnStore()
	s.Set("plugin1", func(in string) (string, error) {
		return in, nil
	})
	manager := &NormalizerManager{
		store:  s,
		loader: &PluginLoader{},
	}

	tests := []struct {
		name     string
		pName    string
		expected bool
	}{
		{
			name:     "Successful plugin unloading",
			pName:    "plugin1",
			expected: true,
		},
		{
			name:     "Failed plugin unloading",
			pName:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.UnloadPlugin(tt.pName)
			if (err == nil) != tt.expected {
				t.Errorf("case: %v failed expected error=%v", tt.name, tt.expected)
			}
		})
	}
}

func TestNormalizerManager_GetNormalizer(t *testing.T) {
	s := NewnStore()
	s.Set("plugin1", func(in string) (string, error) {
		return in, nil
	})
	manager := &NormalizerManager{
		store:  s,
		loader: &PluginLoader{},
	}

	tests := []struct {
		name     string
		pName    string
		expected bool
	}{
		{
			name:     "Existing plugin",
			pName:    "plugin1",
			expected: true,
		},
		{
			name:     "Non-existing plugin",
			pName:    "plugin2",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.GetNormalizer(tt.pName)
			if (err == nil) != tt.expected {
				t.Errorf("case: %v failed got err=%v expected error=%v", tt.name, err, tt.expected)
			}
		})
	}
}
