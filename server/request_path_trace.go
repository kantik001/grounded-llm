package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

type pathTraceKey struct{}

// pathTrace correlates RAG pipeline steps under one request_id (MVP debug trail).
type pathTrace struct {
	reqID  string
	tenant string
	start  time.Time
}

func beginPathTrace(reqID, tenant string) *pathTrace {
	if reqID == "" {
		return nil
	}
	return &pathTrace{reqID: reqID, tenant: tenant, start: time.Now()}
}

func contextWithPathTrace(ctx context.Context, tr *pathTrace) context.Context {
	if tr == nil {
		return ctx
	}
	return context.WithValue(ctx, pathTraceKey{}, tr)
}

func pathTraceFrom(ctx context.Context) *pathTrace {
	if ctx == nil {
		return nil
	}
	tr, _ := ctx.Value(pathTraceKey{}).(*pathTrace)
	return tr
}

func (t *pathTrace) step(name string, fields map[string]any) {
	if t == nil {
		return
	}
	msg := "req=" + t.reqID + " step=" + name
	if t.tenant != "" {
		msg += " tenant=" + t.tenant
	}
	for k, v := range fields {
		msg += " " + k + "=" + formatLogValue(v)
	}
	log.Print(msg)
}

func (t *pathTrace) elapsedMS() int64 {
	if t == nil {
		return 0
	}
	return time.Since(t.start).Milliseconds()
}

func attachRequestID(c *gin.Context, body gin.H) gin.H {
	if body == nil {
		body = gin.H{}
	}
	if c == nil {
		return body
	}
	if id := ctxRequestID(c); id != "" {
		body["request_id"] = id
	}
	return body
}

func msSince(start time.Time) int64 {
	return time.Since(start).Milliseconds()
}
