package app

import (
	"github.com/redis/go-redis/v9"

	"grounded_llm_server/internal/llm"
)

func init() {
	llm.BindConfig(func() *Config { return config })
}

type (
	LLMRequest  = llm.Request
	LLMResponse = llm.Response
	LLMUsage    = llm.Usage
	Choice      = llm.Choice
)

func initRedis() {
	llm.InitRedis()
}

func llmRedis() redis.Cmdable {
	// Avoid the classic Go nil-interface trap: a nil *redis.Client boxed as
	// redis.Cmdable is a non-nil interface and would make WithRedis enable a
	// backend that panics on first use.
	c := llm.Redis()
	if c == nil {
		return nil
	}
	return c
}

func mockLLMCompletion(messages []Message) (string, error) {
	return llm.MockComplete(messages)
}

func callLLMCompletion(messages []Message) (string, error) {
	return llm.Complete(messages)
}
