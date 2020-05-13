package heartbeat

import (
	"fmt"
)

// Heartbeat is a structure representing activity for a user on a some entity.
type Heartbeat struct {
	Branch         string     `json:"branch"`
	Category       Category   `json:"category"`
	CursorPosition int        `json:"cursorpos"`
	Dependencies   []string   `json:"dependencies"`
	Entity         string     `json:"entity"`
	EntityType     EntityType `json:"type"`
	IsWrite        bool       `json:"is_write"`
	Language       string     `json:"language"`
	LineNumber     int        `json:"lineno"`
	Lines          int        `json:"lines"`
	Project        string     `json:"project"`
	Time           int64      `json:"time"`
	UserAgent      string     `json:"user_agent"`
}

// ID returns an ID generated drom the heartbeat data.
func (h Heartbeat) ID() string {
	return fmt.Sprintf("%d-%s-%s-%s-%s-%s-%t",
		h.Time,
		h.EntityType,
		h.Category,
		h.Project,
		h.Branch,
		h.Entity,
		h.IsWrite,
	)
}
