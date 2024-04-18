package main

import (
	"encoding/json"
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
)

type fooNormalizer struct{}
type IORequest struct {
	Rbps      string `json:"rbps"`
	Wbps      string `json:"wbps"`
	BlockSize string `json:"blockSize"`
}

func (n fooNormalizer) Name() string {
	return "Intel P4510 NVMe Disk"
}

// ioRequest example: {"rbps": "30M", "wbps": "20M", "blocksize": "4k"}
func (n fooNormalizer) EstimateRequest(ioReq string) (string, error) {
	var req = &IORequest{}

	err := json.Unmarshal([]byte(ioReq), req)
	if err != nil {
		return "", err
	}
	resp, err := normalize(req)
	if err != nil {
		return "", err
	}
	normalized, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}

// customized normalization method
func normalize(ioRequest *IORequest) (*IORequest, error) {
	r := resource.MustParse(ioRequest.Rbps)
	w := resource.MustParse(ioRequest.Wbps)
	return &IORequest{
		Rbps:      fmt.Sprint(r.Value() * 2),
		Wbps:      fmt.Sprint(w.Value() * 2),
		BlockSize: ioRequest.BlockSize,
	}, nil
}

// Exported as a symbol named "Normalizer"
var Normalizer fooNormalizer
