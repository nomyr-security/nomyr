package demo

import "testing"

func TestSeedIsWatermarkedSyntheticFixture(t *testing.T) {
	estate := Seed()
	if estate.Name != "Nomyr synthetic estate" || estate.Watermark != "demo data" {
		t.Fatalf("fixture is not visibly watermarked: %#v", estate)
	}
	if len(estate.Identities) == 0 {
		t.Fatal("fixture has no identities")
	}
	for _, identity := range estate.Identities {
		if identity.Provider != "synthetic" {
			t.Errorf("identity %q is not marked synthetic", identity.ID)
		}
	}
}

func TestNewEstateCopiesReusableFixtureInput(t *testing.T) {
	identities := []Identity{{ID: "fixture", Provider: "real-provider"}}
	estate := NewEstate(identities...)
	identities[0].ID = "changed"
	if estate.Identities[0].ID != "fixture" {
		t.Fatal("fixture retained caller-owned identity storage")
	}
	if estate.Identities[0].Provider != "synthetic" {
		t.Fatal("fixture retained a non-synthetic provider label")
	}
}
