package main

import (
	"encoding/json"
	"fmt"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/iodriver"
	"k8s.io/apimachinery/pkg/api/resource"
)

type barNormalizer struct{}

func (n barNormalizer) Name() string {
	return "Intel Sample Disk"
}

// ioRequest example: {"rbps": "30Mi", "wbps": "20Mi", "blocksize": "4k"}
func (n barNormalizer) EstimateRequest(ioReq string) (string, error) {
	var req = &iodriver.IORequest{}
	var resp = &v1alpha1.IOBandwidth{}

	if len(ioReq) == 0 {
		resp = &v1alpha1.IOBandwidth{
			Read:  iodriver.MinDefaultIOBW,
			Write: iodriver.MinDefaultIOBW,
			Total: iodriver.MinDefaultTotalIOBW,
		}
	} else {
		err := json.Unmarshal([]byte(ioReq), req)
		if err != nil {
			return "", err
		}
		resp, err = normalize(req)
		if err != nil {
			return "", err
		}
	}
	normalized, err := json.Marshal(resp)
	if err != nil {
		return "", err
	}
	return string(normalized), nil
}

// customized normalization method
func normalize(ioRequest *iodriver.IORequest) (*v1alpha1.IOBandwidth, error) {
	r := resource.MustParse(ioRequest.Rbps)
	w := resource.MustParse(ioRequest.Wbps)

	rbpsValue, _ := r.AsInt64()
	wbpsValue, _ := w.AsInt64()

	rout := float64(rbpsValue) / iodriver.Mi
	wout := float64(wbpsValue) / iodriver.Mi

	return &v1alpha1.IOBandwidth{
		Read:  resource.MustParse(fmt.Sprintf("%fMi", rout)),
		Write: resource.MustParse(fmt.Sprintf("%fMi", wout)),
		Total: resource.MustParse(fmt.Sprintf("%fMi", rout+wout)),
	}, nil
}

// Exported as a symbol named "Normalizer"
var Normalizer barNormalizer
