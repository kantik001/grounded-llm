package store

import "testing"

func TestKBAutoIngestEnabled(t *testing.T) {
	t.Setenv("KB_AUTO_INGEST", "")
	if KBAutoIngestEnabled() {
		t.Fatal("expected disabled by default")
	}
	t.Setenv("KB_AUTO_INGEST", "1")
	if !KBAutoIngestEnabled() {
		t.Fatal("expected enabled for KB_AUTO_INGEST=1")
	}
}
