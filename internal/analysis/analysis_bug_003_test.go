package analysis

import "testing"

func TestOptionalPipelineDisabled(t *testing.T) {
	if OptionalPipeline(false) != nil {
		t.Fatal("disabled analyzer must be a nil interface")
	}
}

func TestRegistryAnalyzeOrDefault(t *testing.T) {
	r := NewRegistry()
	r.Register("default", 1, OptionalPipeline(false))
	got := r.AnalyzeOrDefault("default", 1, "alpha beta")
	if len(got) != 2 || got[0].Term != "alpha" || got[1].Term != "beta" {
		t.Fatalf("unexpected default tokens: %#v", got)
	}
}
