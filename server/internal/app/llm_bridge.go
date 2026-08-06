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
	return llm.Redis()
}

func mockLLMCompletion(messages []Message) (string, error) {
	return llm.MockComplete(messages)
}

func callLLMCompletion(messages []Message) (string, error) {
	return llm.Complete(messages)
}
