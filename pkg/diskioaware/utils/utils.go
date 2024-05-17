package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"reflect"
	"sort"
	"time"

	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/api/diskio/v1alpha1"
	"github.com/intel/cloud-resource-scheduling-and-isolation/pkg/generated/clientset/versioned"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	clientretry "k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
	hashutil "k8s.io/kubernetes/pkg/util/hash"
)

const (
	EmptyDir         string = "emptyDir"
	Others           string = "others"
	DiskIOAnnotation string = "diskio.intel.com/io-bandwidth"

	NodeDiskDeviceCRSuffix string = "-nodediskdevice"
	NodeDiskIOInfoCRSuffix string = "-nodediskiostats"
	NodeIOStatusCR         string = "NodeDiskIOStats"
	APIVersion             string = "ioi.intel.com/v1"
	CRNameSpace            string = "ioi-system"
)

var UpdateBackoff = wait.Backoff{
	Steps:    3,
	Duration: 100 * time.Millisecond, // 0.1s
	Jitter:   1.0,
}

type IORequest struct {
	Rbps      string `json:"rbps"`
	Wbps      string `json:"wbps"`
	BlockSize string `json:"blockSize"`
}

func RequestStrToQuantity(reqStr string) (v1alpha1.IOBandwidth, error) {
	q := v1alpha1.IOBandwidth{}
	err := json.Unmarshal([]byte(reqStr), &q)
	if err != nil {
		return v1alpha1.IOBandwidth{}, fmt.Errorf("unmarshal request error: %v", err)
	}
	q.Total = q.Read.DeepCopy()
	q.Total.Add(q.Write)
	return q, nil
}

func GetNodeIOStatus(client versioned.Interface, n string) (*v1alpha1.NodeDiskIOStats, error) {
	if client == nil {
		return nil, fmt.Errorf("kubernetes configmap client cannot be nil")
	}
	if len(n) == 0 {
		return nil, fmt.Errorf("node name cannot be empty")
	}
	obj, err := client.DiskioV1alpha1().NodeDiskIOStatses(CRNameSpace).Get(context.Background(), n+NodeDiskIOInfoCRSuffix, v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func CreateNodeIOStatus(client versioned.Interface, node string, pl []string) error {
	nodeStatusInfo := &v1alpha1.NodeDiskIOStats{
		TypeMeta: v1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       NodeIOStatusCR,
		},
		ObjectMeta: v1.ObjectMeta{
			Namespace: CRNameSpace,
			Name:      node + NodeDiskIOInfoCRSuffix,
		},
		Spec: v1alpha1.NodeDiskIOStatsSpec{
			NodeName:     node,
			ReservedPods: pl,
		},
		Status: v1alpha1.NodeDiskIOStatsStatus{},
	}

	_, err := client.DiskioV1alpha1().NodeDiskIOStatses(CRNameSpace).Create(context.TODO(), nodeStatusInfo, v1.CreateOptions{})
	if err != nil {
		klog.Error("CreateNodeIOStatus fails: ", err)
		return err
	}
	return nil
}

func UpdateNodeIOStatus(client versioned.Interface, node string, pl []string) error {

	return clientretry.RetryOnConflict(UpdateBackoff, func() error {
		sts, err := GetNodeIOStatus(client, node)
		if err != nil {
			return err
		}
		sts.Spec.ReservedPods = pl
		_, err = client.DiskioV1alpha1().NodeDiskIOStatses(CRNameSpace).Update(context.TODO(), sts, v1.UpdateOptions{})
		if err != nil {
			klog.Error("UpdateNodeIOStatus fails: ", err)
			return err
		}
		return nil
	})
}

func ComparePodList(pl1, pl2 []string) bool {
	sort.Strings(pl1)
	sort.Strings(pl2)
	return reflect.DeepEqual(pl1, pl2)
}

func HashObject(obj interface{}) uint64 {
	if obj == nil {
		return 0
	}
	hash := fnv.New32a()

	hashutil.DeepHashObject(hash, obj)
	return uint64(hash.Sum32())
}
