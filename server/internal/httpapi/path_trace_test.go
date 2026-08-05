package httpapi_test

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"grounded_llm_server/internal/httpapi"
)

func TestPathTraceStepsIncludeRequestID(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	tr := httpapi.BeginPathTrace("abc123ef", "tenant-a")
	if tr == nil {
		t.Fatal("expected pathTrace")
	}
	tr.Step("message.accept", map[string]any{"domain_id": "default", "stream": false})
	tr.Step("retrieve", map[string]any{"ms": int64(12), "fragments": 3, "ok": true})
	time.Sleep(time.Millisecond)
	tr.Step("done", map[string]any{"ms": tr.ElapsedMS(), "ok": true, "verify_pass": true})

	out := buf.String()
	for _, want := range []string{
		"req=abc123ef",
		"step=message.accept",
		"step=retrieve",
		"step=done",
		"tenant=tenant-a",
		"fragments=3",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("log missing %q; got:\n%s", want, out)
		}
	}
}

func TestPathTraceFromContext(t *testing.T) {
	tr := httpapi.BeginPathTrace("deadbeef", "default")
	ctx := httpapi.ContextWithPathTrace(context.Background(), tr)
	got := httpapi.PathTraceFrom(ctx)
	if got == nil || got.RequestID() != "deadbeef" {
		t.Fatalf("expected trace in context, got %#v", got)
	}
	if httpapi.PathTraceFrom(context.Background()) != nil {
		t.Fatal("empty context should have no trace")
	}
}

func TestAttachRequestID(t *testing.T) {
	body := httpapi.AttachRequestID(nil, nil)
	if body == nil {
		t.Fatal("expected non-nil map")
	}
}
