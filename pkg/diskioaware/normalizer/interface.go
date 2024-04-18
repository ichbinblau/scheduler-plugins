package normalizer

type Normalize func(string) (string, error)

type Normalizer interface {
	Name() string
	EstimateRequest(string) (string, error)
}
