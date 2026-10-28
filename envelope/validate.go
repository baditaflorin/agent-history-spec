package envelope

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// FieldError names the field of an Event that is invalid and why. Validate always returns
// one, so a caller can report Field in a Rejection.
type FieldError struct {
	Field  string
	Reason string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Reason }

var (
	agentPattern = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,31}$`)
	nodePattern  = regexp.MustCompile(`^(session|repo|file):.+$`)
)

func validRole(r string) bool {
	switch r {
	case RoleUser, RoleAssistant, RoleToolCall, RoleToolResult, RoleBusMsg:
		return true
	}
	return false
}

func validRel(r string) bool {
	switch r {
	case RelBusMsg, RelSpawned, RelRanIn, RelTouched, RelSimilar:
		return true
	}
	return false
}

func bad(field, format string, a ...any) error {
	return &FieldError{Field: field, Reason: fmt.Sprintf(format, a...)}
}

// short checks an optional text field: at most MaxFieldBytes and valid UTF-8.
func short(field, v string) error {
	if len(v) > MaxFieldBytes {
		return bad(field, "%d bytes, limit %d", len(v), MaxFieldBytes)
	}
	if !utf8.ValidString(v) {
		return bad(field, "not valid UTF-8")
	}
	return nil
}

// Validate checks an Event against the wire format and returns the first problem as a
// *FieldError. It checks shape and limits only; it does not look at the clock or at the store.
func Validate(e *Event) error {
	if e.EventID == "" {
		return bad("event_id", "required")
	}
	if len(e.EventID) > MaxIDBytes || strings.ContainsFunc(e.EventID, unicode.IsControl) {
		return bad("event_id", "must be at most %d bytes with no control characters", MaxIDBytes)
	}
	if e.TS.IsZero() {
		return bad("ts", "required")
	}
	if e.Host == "" || len(e.Host) > MaxFieldBytes || strings.ContainsAny(e.Host, ": \t\r\n") {
		return bad("host", "required, no ':' and no whitespace")
	}
	if !agentPattern.MatchString(e.Agent) {
		return bad("agent", "must match %s", agentPattern)
	}
	prefix := e.Agent + ":" + e.Host + ":"
	if len(e.SessionKey) > MaxFieldBytes || !strings.HasPrefix(e.SessionKey, prefix) || len(e.SessionKey) == len(prefix) {
		return bad("session_key", "must be agent:host:native_session_id")
	}
	if e.NativeSessionID != "" && e.SessionKey != prefix+e.NativeSessionID {
		return bad("session_key", "does not match agent, host and native_session_id")
	}
	if !validRole(e.Role) {
		return bad("role", "unknown role %q", e.Role)
	}
	for _, f := range []struct{ name, v string }{
		{"tool_name", e.ToolName}, {"cwd", e.Cwd}, {"native_session_id", e.NativeSessionID},
		{"agent_version", e.AgentVersion}, {"git_repo", e.GitRepo}, {"git_branch", e.GitBranch},
		{"title", e.Title}, {"resume_cmd", e.ResumeCmd},
	} {
		if err := short(f.name, f.v); err != nil {
			return err
		}
	}
	for _, f := range []struct{ name, v string }{{"command", e.Command}, {"content", e.Content}} {
		if len(f.v) > MaxContentBytes {
			return bad(f.name, "%d bytes, limit %d", len(f.v), MaxContentBytes)
		}
		if !utf8.ValidString(f.v) {
			return bad(f.name, "not valid UTF-8")
		}
	}
	if len(e.Raw) > MaxRawBytes {
		return bad("raw", "%d bytes, limit %d", len(e.Raw), MaxRawBytes)
	}
	if len(e.Raw) > 0 && !json.Valid(e.Raw) {
		return bad("raw", "not valid JSON")
	}
	if len(e.Edges) > MaxEdges {
		return bad("edges", "%d edges, limit %d", len(e.Edges), MaxEdges)
	}
	for i, ed := range e.Edges {
		field := fmt.Sprintf("edges[%d]", i)
		if !validRel(ed.Rel) {
			return bad(field+".rel", "unknown relation %q", ed.Rel)
		}
		if len(ed.Dst) > MaxFieldBytes || !nodePattern.MatchString(ed.Dst) {
			return bad(field+".dst", "must be session:<key>, repo:<path> or file:<path>")
		}
		if ed.Weight < 0 || math.IsNaN(float64(ed.Weight)) {
			return bad(field+".weight", "must be a number >= 0")
		}
	}
	return nil
}
