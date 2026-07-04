package main

import (
	"os"
	"path/filepath"
	"testing"
)

func setupMockConfig(t *testing.T) {
	t.Helper()
	t.Setenv("LLM_MOCK", "true")
	t.Setenv("RAG_MOCK", "true")
	t.Setenv("DOMAINS_CONFIG_PATH", domainsConfigForTest(t))
	config = loadConfig()
	domainCatalog = domainsFile{}
	if err := loadDomainCatalog(); err != nil {
		t.Fatalf("loadDomainCatalog: %v", err)
	}
	initLocaleConfig(config)
}

func TestMockRAGContextResponse_Vacation(t *testing.T) {
	out := mockRAGContextResponse("How many vacation days?", "default")
	if !out.Success || len(out.Fragments) == 0 {
		t.Fatalf("expected mock RAG success, got %+v", out)
	}
	if out.Fragments[0].Content == "" {
		t.Fatal("expected fragment content")
	}
}

func TestMockLLMCompletion_Vacation(t *testing.T) {
	setupMockConfig(t)
	answer, err := mockLLMCompletion([]Message{{Role: "user", Content: "vacation days?"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer == "" {
		t.Fatal("expected non-empty mock LLM answer")
	}
}

func TestAnswerWithRAG_MockMode(t *testing.T) {
	setupMockConfig(t)
	wd, _ := os.Getwd()
	_ = wd
	result := answerWithRAG(
		"How many paid vacation days do employees get?",
		"default",
		"default",
		"en",
		nil,
		"test-session",
	)
	if !result.OK {
		t.Fatalf("expected OK, got err=%q soft=%v", result.ErrMsg, result.SoftFail)
	}
	if !result.VerifyPass {
		t.Fatalf("expected verify pass for mock answer, answer=%q", result.Answer)
	}
	if len(result.Citations) == 0 {
		t.Fatal("expected citations from mock fragments")
	}
}

func TestFetchRAGContext_MockSkipsHTTP(t *testing.T) {
	setupMockConfig(t)
	if !ragMockEnabled() {
		t.Fatal("RAG mock should be enabled")
	}
	out, err := fetchRAGContext("vacation", "default", "default", "en")
	if err != nil {
		t.Fatal(err)
	}
	if !out.Success {
		t.Fatalf("expected success, got %+v", out)
	}
}

func TestCallLLMCompletion_MockSkipsHTTP(t *testing.T) {
	setupMockConfig(t)
	answer, err := callLLMCompletion([]Message{{Role: "user", Content: "How many vacation days?"}})
	if err != nil {
		t.Fatal(err)
	}
	if answer == "" {
		t.Fatal("expected mock answer")
	}
}

func TestDomainsConfigPathFromEnv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(wd, "..", "config", "domains.json")
	if _, err := os.Stat(p); err != nil {
		t.Skip("domains.json not found")
	}
	t.Setenv("DOMAINS_CONFIG_PATH", p)
	got := domainsConfigForTest(t)
	if got != p {
		t.Fatalf("expected %q, got %q", p, got)
	}
}
