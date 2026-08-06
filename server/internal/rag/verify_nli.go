package rag

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// NLIMode: off | assist (lexical first, NLI confirms failures) | replace (NLI only).
func NLIMode() string {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("VERIFY_NLI"))) {
	case "assist", "http", "1", "true", "on":
		return "assist"
	case "replace":
		return "replace"
	default:
		return "off"
	}
}

func nliURL() string {
	return strings.TrimSpace(os.Getenv("VERIFY_NLI_URL"))
}

type nliRequest struct {
	Premise     string `json:"premise"`
	Hypothesis  string `json:"hypothesis"`
	Context     string `json:"context,omitempty"`
	Claim       string `json:"claim,omitempty"`
}

type nliResponse struct {
	Entailed bool    `json:"entailed"`
	Label    string  `json:"label"`
	Score    float64 `json:"score"`
}

var nliHTTPClient = &http.Client{Timeout: 5 * time.Second}

// nliEntailed asks the optional NLI sidecar whether hypothesis is entailed by premise.
// Returns (entailed, ok). ok=false means the sidecar was unavailable — caller falls back.
func nliEntailed(premise, hypothesis string) (entailed bool, ok bool) {
	url := nliURL()
	if url == "" {
		return false, false
	}
	body, _ := json.Marshal(nliRequest{
		Premise:    premise,
		Hypothesis: hypothesis,
		Context:    premise,
		Claim:      hypothesis,
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return false, false
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := nliHTTPClient.Do(req)
	if err != nil {
		log.Printf("verify NLI request failed: %v", err)
		return false, false
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Printf("verify NLI status %d: %s", resp.StatusCode, string(raw)[:min(120, len(raw))])
		return false, false
	}
	var out nliResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return false, false
	}
	label := strings.ToLower(strings.TrimSpace(out.Label))
	if label == "entailment" || label == "entailed" {
		return true, true
	}
	if label == "contradiction" || label == "neutral" {
		return false, true
	}
	return out.Entailed, true
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// UnsupportedSentencesNLI extends lexical check with an optional HTTP NLI sidecar.
func UnsupportedSentencesNLI(body, contextText string, minSupport float64) []string {
	mode := NLIMode()
	if mode == "off" || nliURL() == "" {
		return UnsupportedSentences(body, contextText, minSupport)
	}

	contextStems := make(map[string]struct{})
	for _, s := range contentStems(contextText) {
		contextStems[s] = struct{}{}
	}

	var unsupported []string
	for _, sentence := range splitSentences(body) {
		stems := contentStems(sentence)
		if len(stems) < minSentenceTokens {
			continue
		}
		lexicalOK := scoreSupport(stems, contextStems) >= minSupport
		switch mode {
		case "replace":
			entailed, ok := nliEntailed(contextText, sentence)
			if !ok {
				// Sidecar down — fall back to lexical for this sentence.
				if !lexicalOK {
					unsupported = append(unsupported, sentence)
				}
				continue
			}
			if !entailed {
				unsupported = append(unsupported, sentence)
			}
		default: // assist
			if lexicalOK {
				continue
			}
			entailed, ok := nliEntailed(contextText, sentence)
			if ok && entailed {
				continue
			}
			unsupported = append(unsupported, sentence)
		}
	}
	return unsupported
}

// VerifyFaithfulnessLocal applies the configured faithfulness mode.
// Returns ok plus a human-readable reason for logs/analytics.
func VerifyFaithfulnessLocal(body, contextText string) (bool, string) {
	mode := FaithfulnessMode()
	if mode == FaithfulnessOff {
		return true, ""
	}
	unsupported := UnsupportedSentencesNLI(body, contextText, FaithfulnessMinSupport())
	if len(unsupported) == 0 {
		return true, ""
	}
	preview := unsupported[0]
	if r := []rune(preview); len(r) > 120 {
		preview = string(r[:120]) + "…"
	}
	reason := fmt.Sprintf("%d sentence(s) lack source support, e.g.: %q", len(unsupported), preview)
	if mode == FaithfulnessEnforce {
		return false, reason
	}
	return true, reason
}
