package paste

import "time"

// Paste is a single paste stored in memory.
//
// ExpiresAt is nil when the paste never expires. time.Time values are
// marshalled as RFC3339 by encoding/json.
type Paste struct {
	ID        string     `json:"id"`
	Content   string     `json:"content"`
	Language  string     `json:"language"`
	CreatedAt time.Time  `json:"created_at"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
