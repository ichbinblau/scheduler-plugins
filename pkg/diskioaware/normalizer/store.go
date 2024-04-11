package normalizer

import "fmt"

type nStore struct {
	executer map[string]Normalize
}

func NewnStore() *nStore {
	return &nStore{
		executer: make(map[string]Normalize),
	}
}

func (ns *nStore) Set(name string, n Normalize) error {
	if len(name) == 0 {
		return fmt.Errorf("normalizer name cannot be empty")
	}
	if n == nil {
		return fmt.Errorf("normalizer func cannot be empty")
	}
	ns.executer[name] = n
	return nil
}

func (ns *nStore) Contains(name string) bool {
	_, ok := ns.executer[name]
	return ok
}

func (ns *nStore) Delete(name string) {
	delete(ns.executer, name)
}

func (ns *nStore) Get(name string) (Normalize, error) {
	if len(name) == 0 {
		return nil, fmt.Errorf("normalizer name cannot be empty")
	}
	n, ok := ns.executer[name]
	if !ok {
		return nil, fmt.Errorf("normalizer %s not found", name)
	}
	return n, nil
}
