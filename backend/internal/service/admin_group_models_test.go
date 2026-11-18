//go:build unit

package service

import "testing"

func TestDefaultModelsListCandidateIDs_KiroIncludesGPT56Variants(t *testing.T) {
	models := defaultModelsListCandidateIDs(PlatformKiro)
	available := make(map[string]struct{}, len(models))
	for _, model := range models {
		available[model] = struct{}{}
	}

	for _, model := range []string{"gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna"} {
		if _, found := available[model]; !found {
			t.Fatalf("expected Kiro group model candidates to include %q", model)
		}
	}
}
