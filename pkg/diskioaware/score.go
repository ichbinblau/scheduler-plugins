package diskioaware

import (
	"fmt"

	"k8s.io/kubernetes/pkg/scheduler/framework"
	"sigs.k8s.io/scheduler-plugins/pkg/diskioaware/resource"
)

type Scorer interface {
	Score(string, *stateData, resource.Handle) (int64, error)
}

func getScorer(scoreStrategy string) (Scorer, error) {
	switch scoreStrategy {
	case "LeastAllocated":
		return &LeastAllocatedScorer{}, nil
	case "MostAllocated":
		return &MostAllocatedScorer{}, nil
	default:
		return nil, fmt.Errorf("unknown score strategy %v", scoreStrategy)
	}
}

type MostAllocatedScorer struct{}

// todo: change algorithm
func (scorer *MostAllocatedScorer) Score(node string, state *stateData, rh resource.Handle) (int64, error) {
	if !state.nodeSupportIOI {
		return framework.MaxNodeScore, nil
	}
	ratio, err := rh.(resource.CacheHandle).NodePressureRatio(node, state.request)
	if err != nil {
		return 0, err
	}
	return int64(ratio * float64(framework.MaxNodeScore)), nil
}

type LeastAllocatedScorer struct{}

func (scorer *LeastAllocatedScorer) Score(node string, state *stateData, rh resource.Handle) (int64, error) {
	if !state.nodeSupportIOI {
		return framework.MaxNodeScore, nil
	}
	ratio, err := rh.(resource.CacheHandle).NodePressureRatio(node, state.request)
	if err != nil {
		return 0, err
	}
	return int64((1.0 - ratio) * float64(framework.MaxNodeScore)), nil
}
