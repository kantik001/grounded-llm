package store

import "strings"

// TelegramUser is the identity persisted in users / used for session ownership.
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// RAGFragment is a document chunk (full content for verify; excerpt for UI/DB citations).
type RAGFragment struct {
	Filename string `json:"filename"`
	Content  string `json:"content"`
	Page     int    `json:"page,omitempty"`
	Excerpt  string `json:"excerpt,omitempty"`
}

// Message is one LLM chat/completions message (also used for HistoryForLLM).
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatMessage is one persisted chat turn for UI and LLM context rebuild.
type ChatMessage struct {
	ID              int64         `json:"id,omitempty"`
	Role            string        `json:"role"`
	Content         string        `json:"content"`
	ImageDataURL    string        `json:"image_data_url,omitempty"`
	ImageURL        string        `json:"image_url,omitempty"`
	ImageToken      string        `json:"-"`
	ClassPrediction string        `json:"class_prediction,omitempty"`
	ClassConfidence float64       `json:"class_confidence,omitempty"`
	Kind            string        `json:"kind,omitempty"`
	Citations       []RAGFragment `json:"citations,omitempty"`
	FeedbackRating  *int          `json:"feedback_rating,omitempty"`
}

// TenantPurgeStats counts removed records and files for admin tenant purge.
type TenantPurgeStats struct {
	Sessions      int64 `json:"sessions"`
	Messages      int64 `json:"messages"`
	FeedbackRows  int64 `json:"feedback_rows"`
	AuditRows     int64 `json:"audit_rows"`
	AnalyticsRows int64 `json:"analytics_rows"`
	ReindexJobs   int64 `json:"reindex_jobs"`
	DataFiles     int   `json:"data_files"`
	UploadTokens  int64 `json:"upload_tokens"`
}

// NormalizeTenantID lowercases and trims a tenant id.
func NormalizeTenantID(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizeTenantID(raw string) string { return NormalizeTenantID(raw) }

// ToLLMMessage maps a stored chat message into an LLM message when possible.
func (m ChatMessage) ToLLMMessage() (Message, bool) {
	switch m.Role {
	case "assistant":
		if m.Content == "" {
			return Message{}, false
		}
		return Message{Role: "assistant", Content: m.Content}, true
	case "user":
		if m.ImageURL != "" || m.ImageDataURL != "" || m.ImageToken != "" {
			s := "The user sent an image."
			if t := trimUserCaption(m.Content); t != "" {
				s += " Caption: " + t
			}
			return Message{Role: "user", Content: s}, true
		}
		if t := trimUserCaption(m.Content); t != "" {
			return Message{Role: "user", Content: t}, true
		}
		return Message{}, false
	default:
		return Message{}, false
	}
}

func trimUserCaption(s string) string {
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\r", " "), "\n", " "))
}

// TrimHistoryMessages keeps the last max messages for LLM context.
func TrimHistoryMessages(msgs []Message, max int) []Message {
	if len(msgs) <= max {
		return msgs
	}
	return msgs[len(msgs)-max:]
}
