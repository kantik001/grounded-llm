package main

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"
)

func TestPathTraceStepsIncludeRequestID(t *testing.T) {
	var buf bytes.Buffer
	prev := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(prev)

	tr := beginPathTrace("abc123ef", "tenant-a")
	if tr == nil {
		t.Fatal("expected pathTrace")
	}
	tr.step("message.accept", map[string]any{"domain_id": "default", "stream": false})
	tr.step("retrieve", map[string]any{"ms": int64(12), "fragments": 3, "ok": true})
	time.Sleep(time.Millisecond)
	tr.step("done", map[string]any{"ms": tr.elapsedMS(), "ok": true, "verify_pass": true})

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
	tr := beginPathTrace("deadbeef", "default")
	ctx := contextWithPathTrace(context.Background(), tr)
	got := pathTraceFrom(ctx)
	if got == nil || got.reqID != "deadbeef" {
		t.Fatalf("expected trace in context, got %#v", got)
	}
	if pathTraceFrom(context.Background()) != nil {
		t.Fatal("empty context should have no trace")
	}
}

func TestAttachRequestID(t *testing.T) {
	// gin.Context without middleware: attachRequestID should leave body alone when empty id
	body := attachRequestID(nil, nil)
	if body == nil {
		t.Fatal("expected non-nil map")
	}
}
