package admin

import (
	"context"
	"log"

	"grounded_llm_server/internal/store"
)

type autoIngestResult struct {
	Queued         bool
	AlreadyRunning bool
	JobID          int64
	OutboxFlushed  int
}

func maybeFlushAutoIngest(
	ctx context.Context,
	st *store.ChatStore,
	tenantID, domainID, actor, source string,
	sync bool,
) autoIngestResult {
	var out autoIngestResult
	if st == nil || !store.KBAutoIngestEnabled() {
		return out
	}
	job, alreadyRunning, flushed, err := st.FlushKBIngestOutbox(ctx, tenantID, domainID, actor, source, nil)
	if err != nil {
		log.Printf("KB auto-ingest flush tenant=%s domain=%s: %v", tenantID, domainID, err)
		return out
	}
	if flushed == 0 {
		return out
	}
	out.OutboxFlushed = flushed
	out.JobID = job.ID
	out.AlreadyRunning = alreadyRunning
	out.Queued = true
	if !alreadyRunning {
		startIngestWorker(job, sync)
	}
	return out
}
