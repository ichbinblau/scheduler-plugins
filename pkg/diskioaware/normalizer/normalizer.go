package normalizer

type Normalize func(string) string

type Normalizer interface {
	Name() string
	EstimateRequest(string) string
}
