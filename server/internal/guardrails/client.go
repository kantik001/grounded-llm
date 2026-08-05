package guardrails

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	pb "grounded_llm_server/gen/guardrails/v1"
	"grounded_llm_server/internal/config"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// Mode controls how RAG numeric/PII verification runs.
type Mode string

const (
	ModeLocal  Mode = "local"  // in-process (default, CI-safe)
	ModeRemote Mode = "remote" // grounded-guardrails gRPC only
	ModeHybrid Mode = "hybrid" // remote with local fallback
)

var (
	guardrailsMu     sync.RWMutex
	guardrailsClient pb.GuardrailsServiceClient
	guardrailsConn   *grpc.ClientConn
)

func Init(cfg *config.Config) {
	mode := NormalizeMode(cfg.GuardrailsMode)
	cfg.GuardrailsMode = string(mode)
	if mode == ModeLocal {
		log.Printf("Guardrails: local in-process verify")
		return
	}
	addr := strings.TrimSpace(cfg.GuardrailsGRPCAddr)
	if addr == "" {
		addr = "localhost:50052"
		cfg.GuardrailsGRPCAddr = addr
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		if mode == ModeRemote {
			log.Fatalf("Guardrails remote mode: dial %s: %v", addr, err)
		}
		log.Printf("Guardrails hybrid: dial %s failed (%v); using local verify", addr, err)
		return
	}
	if err := waitGuardrailsReady(ctx, conn); err != nil {
		_ = conn.Close()
		if mode == ModeRemote {
			log.Fatalf("Guardrails remote mode: connect %s: %v", addr, err)
		}
		log.Printf("Guardrails hybrid: connect %s failed (%v); using local verify", addr, err)
		return
	}
	guardrailsMu.Lock()
	guardrailsConn = conn
	guardrailsClient = pb.NewGuardrailsServiceClient(conn)
	guardrailsMu.Unlock()
	log.Printf("Guardrails: %s mode via gRPC %s", mode, addr)
}

func waitGuardrailsReady(ctx context.Context, conn *grpc.ClientConn) error {
	conn.Connect()
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return nil
		}
		if !conn.WaitForStateChange(ctx, state) {
			return fmt.Errorf("timeout waiting for Ready (last state=%s)", state.String())
		}
	}
}

func Close() {
	guardrailsMu.Lock()
	defer guardrailsMu.Unlock()
	if guardrailsConn != nil {
		_ = guardrailsConn.Close()
		guardrailsConn = nil
		guardrailsClient = nil
	}
}

func NormalizeMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "local":
		return ModeLocal
	case "remote":
		return ModeRemote
	case "hybrid":
		return ModeHybrid
	default:
		log.Printf("Guardrails: unknown GUARDRAILS_MODE=%q, using local", s)
		return ModeLocal
	}
}

func VerifyText(answer, contextText, tenantID string, piiBlock bool) (bool, string, error) {
	return verifyViaGuardrails(answer, contextText, tenantID, piiBlock)
}

func verifyViaGuardrails(answer, contextText, tenantID string, piiBlock bool) (bool, string, error) {
	guardrailsMu.RLock()
	client := guardrailsClient
	guardrailsMu.RUnlock()
	if client == nil {
		return false, "", fmt.Errorf("guardrails client not connected")
	}
	rules := []string{"numeric_verify"}
	if piiBlock {
		rules = append(rules, "pii_block")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	resp, err := client.VerifyText(ctx, &pb.TextRequest{
		Text:     answer,
		Context:  contextText,
		TenantId: tenantID,
		Rules:    rules,
	})
	if err != nil {
		return false, "", err
	}
	if resp.GetPassed() {
		return true, "Verification passed", nil
	}
	return false, formatGuardrailsViolations(resp.GetViolations()), nil
}

func formatGuardrailsViolations(violations []string) string {
	if len(violations) == 0 {
		return "Guardrails rejected the answer."
	}
	var nums []string
	var pii []string
	var other []string
	for _, v := range violations {
		switch {
		case strings.HasPrefix(v, "numeric_unmatched:"):
			nums = append(nums, strings.TrimPrefix(v, "numeric_unmatched:"))
		case strings.HasPrefix(v, "pii:"):
			pii = append(pii, v)
		default:
			other = append(other, v)
		}
	}
	var parts []string
	if len(nums) > 0 {
		parts = append(parts, fmt.Sprintf("Number(s) %v not found in sources.", nums))
	}
	if len(pii) > 0 {
		parts = append(parts, "PII detected: "+strings.Join(pii, ", "))
	}
	if len(other) > 0 {
		parts = append(parts, strings.Join(other, "; "))
	}
	return strings.Join(parts, " ")
}
