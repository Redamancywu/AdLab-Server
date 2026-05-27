package handler

import (
	"encoding/json"
	"os"
	"testing"
)

func TestRepoSDKAPISpecMatchesEmbeddedSpec(t *testing.T) {
	t.Parallel()

	repoSpecBytes, err := os.ReadFile("../../docs/sdk-api.json")
	if err != nil {
		t.Fatalf("read repo sdk api spec failed: %v", err)
	}

	var embedded any
	if err := json.Unmarshal([]byte(openAPISpec), &embedded); err != nil {
		t.Fatalf("embedded sdk api spec should be valid json: %v", err)
	}

	var repo any
	if err := json.Unmarshal(repoSpecBytes, &repo); err != nil {
		t.Fatalf("repo sdk api spec should be valid json: %v", err)
	}

	embeddedNormalized, err := json.Marshal(embedded)
	if err != nil {
		t.Fatalf("normalize embedded spec failed: %v", err)
	}
	repoNormalized, err := json.Marshal(repo)
	if err != nil {
		t.Fatalf("normalize repo spec failed: %v", err)
	}

	if string(embeddedNormalized) != string(repoNormalized) {
		t.Fatalf("repo sdk api spec is out of sync with embedded runtime spec")
	}
}
