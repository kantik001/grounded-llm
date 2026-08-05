package app

import (
	"grounded_llm_server/internal/guardrails"
)

type GuardrailsMode = guardrails.Mode

const (
	GuardrailsModeLocal  = guardrails.ModeLocal
	GuardrailsModeRemote = guardrails.ModeRemote
	GuardrailsModeHybrid = guardrails.ModeHybrid
)

func initGuardrailsClient(cfg *Config) {
	guardrails.Init(cfg)
}

func closeGuardrailsClient() {
	guardrails.Close()
}

func normalizeGuardrailsMode(s string) GuardrailsMode {
	return guardrails.NormalizeMode(s)
}

func verifyViaGuardrails(answer, contextText, tenantID string, piiBlock bool) (bool, string, error) {
	return guardrails.VerifyText(answer, contextText, tenantID, piiBlock)
}
