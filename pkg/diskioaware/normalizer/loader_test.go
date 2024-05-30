package normalizer

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

var base string = "../sample-plugin/"

func getPath(base string, name string) string {
	return filepath.Join(base, name)
}

// Create a temporary plugin file for testing
func createTmpFile(n string, content string) (string, error) {
	tmpFile, err := os.CreateTemp("", n)
	if err != nil {
		return "", err
	}

	// Write some content to the temporary plugin file
	_, err = tmpFile.WriteString(content)
	if err != nil {
		return "", err
	}

	// Close the temporary plugin file
	err = tmpFile.Close()
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func TestPluginLoader_GetFilePath(t *testing.T) {
	pl := NewPluginLoader(base, 2)

	p := PlConfig{
		Vendor: "Intel",
		Model:  "P4510",
	}

	expectedFilePath := getPath(base, "Intel-P4510.so")

	actualFilePath := pl.getFilePath(p)

	if actualFilePath != expectedFilePath {
		t.Errorf("Expected file path: %s, but got: %s", expectedFilePath, actualFilePath)
	}
}

func TestPluginLoader_LoadPlugin(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("plugin content"))
	}))
	defer ts.Close()

	p := PlConfig{
		Vendor: "Intel",
		Model:  "P4510",
		URL:    ts.URL,
	}

	fName, err := createTmpFile(fmt.Sprintf("%s-%s.so", p.Vendor, p.Model), "plugin content")
	if err != nil {
		t.Fatalf("Failed to create temporary plugin file: %v", err)
	}
	defer os.Remove(fName)
	// Set the baseDir to the temporary directory
	pl := NewPluginLoader(filepath.Dir(fName), 3)

	// Load the plugin
	_, err = pl.loadPlugin(context.Background(), p)
	if err != nil {
		t.Errorf("Failed to load plugin: %v", err)
	}
}

// func TestIsValidURL(t *testing.T) {
// 	tests := []struct {
// 		name     string
// 		url      string
// 		expected bool
// 	}{
// 		{
// 			name:     "Invalid url: w/o protocol",
// 			url:      "invalid-url",
// 			expected: false,
// 		},
// 		{
// 			name:     "Invalid url: wrong protocol",
// 			url:      "file:///tmp/test.txt",
// 			expected: false,
// 		},
// 		{
// 			name:     "Valid URL",
// 			url:      "http://example.com",
// 			expected: true,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got := isValidURL(tt.url)
// 			if got != tt.expected {
// 				t.Errorf("case: %v failed got=%v expected=%v", tt.name, got, tt.expected)
// 			}
// 		})
// 	}
// }

func TestDownloadFile(t *testing.T) {
	content := "downloaded-file"
	// Create a test server to serve a file
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(content))
	}))
	defer ts.Close()

	// Create a temporary file to download the content
	tmpFile, err := createTmpFile("plugin.so", content)
	if err != nil {
		t.Fatalf("Failed to create temporary file: %v", err)
	}
	defer os.Remove(tmpFile)

	// pl := NewPluginLoader(base, 3)
	tests := []struct {
		name            string
		url             string
		success         bool
		expectedContent string
	}{
		{
			name:    "Invalid URL",
			url:     "123456",
			success: false,
		},
		{
			name:            "Valid URL and consistent content",
			url:             ts.URL,
			success:         true,
			expectedContent: content,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Download the file

			err = downloadFile(context.Background(), tt.url, tmpFile, false, 2)
			t.Log(err)
			if (err == nil) != tt.success {
				t.Errorf("case: %v failed expected=%v", tt.name, tt.success)
			} else if tt.success {
				// Read the downloaded file content
				content, err := os.ReadFile(tmpFile)
				if err != nil {
					t.Fatalf("Failed to read downloaded file: %v", err)
				}

				actualContent := string(content)

				if actualContent != tt.expectedContent {
					t.Errorf("Expected file content: %s, but got: %s", tt.expectedContent, actualContent)
				}
			}
		})
	}
}

func TestPluginLoader_LoadPlugin_Error(t *testing.T) {
	p, err := os.ReadFile(getPath(base, "foo.so"))
	if err != nil {
		t.Fatalf("failed to foo.so: %v", err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(p)
	}))
	defer ts.Close()
	pl := NewPluginLoader(base, 3)

	conf := PlConfig{
		Vendor: "Intel",
		Model:  "P4510",
		URL:    ts.URL,
	}

	_, err = pl.LoadPlugin(context.Background(), conf)
	if err != nil {
		t.Errorf("Failed load plugin: %v", err)
	}
	os.Remove(pl.getFilePath(conf))
}
