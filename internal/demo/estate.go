package demo

type Estate struct {
	Name       string     `json:"name"`
	Watermark  string     `json:"watermark"`
	Identities []Identity `json:"identities"`
}

type Identity struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Provider string `json:"provider"`
}

func Seed() Estate {
	return NewEstate(
		Identity{ID: "demo-service-account", Name: "payments-deployer", Kind: "service-account", Provider: "synthetic"},
		Identity{ID: "demo-workload", Name: "checkout-worker", Kind: "workload", Provider: "synthetic"},
	)
}

func NewEstate(identities ...Identity) Estate {
	syntheticIdentities := append([]Identity(nil), identities...)
	for index := range syntheticIdentities {
		syntheticIdentities[index].Provider = "synthetic"
	}
	return Estate{
		Name:       "Nomyr synthetic estate",
		Watermark:  "demo data",
		Identities: syntheticIdentities,
	}
}
