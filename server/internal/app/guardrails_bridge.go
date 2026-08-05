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
