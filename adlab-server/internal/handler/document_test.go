package handler

import (
	"strings"
	"testing"
)

func TestDefaultDocsReflectCurrentSDKAPIFlow(t *testing.T) {
	t.Parallel()

	cases := map[string][]string{
		"ios": {
			"/api/v1/sdk/init",
			"/api/v1/sdk/init_complete",
			"/api/v1/ad/request",
			"/api/v1/track",
		},
		"android": {
			"/api/v1/sdk/init",
			"/api/v1/sdk/init_complete",
			"/api/v1/ad/request",
			"/api/v1/track",
		},
		"web": {
			"/api/v1/sdk/init",
			"/api/v1/ad/request",
			"/api/v1/c2s/result",
			"/api/v1/track",
		},
		"api": {
			"/api/v1/sdk/init",
			"/api/v1/sdk/heartbeat",
			"/api/v1/ad/request",
			"/api/v1/c2s/result",
			"/api/v1/track",
		},
	}

	for key, expectedSnippets := range cases {
		doc, ok := defaultDocs[key]
		if !ok {
			t.Fatalf("default doc %q not found", key)
		}
		for _, snippet := range expectedSnippets {
			if !strings.Contains(doc.Content, snippet) {
				t.Fatalf("default doc %q should contain %q", key, snippet)
			}
		}
	}
}
