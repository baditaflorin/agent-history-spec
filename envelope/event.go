// Package envelope is the wire format every agent-history service speaks: one Event per
// prompt, reply, tool call, tool result or bus message, shipped as NDJSON.
//
// It is the only code the services share. Event ids must come out byte-identical in the
// collector, the store and every pusher, so the rules live here and are pinned by the
// vectors in ../vectors.
package envelope

import (
	"encoding/json"
	"time"
)

// Roles an event can have.
const (
	RoleUser       = "user"
	RoleAssistant  = "assistant"
	RoleToolCall   = "tool_call"
	RoleToolResult = "tool_result"
	RoleBusMsg     = "bus_msg"
)

// Relations an Edge can carry.
const (
	RelBusMsg  = "bus_msg"
	RelSpawned = "spawned"
	RelRanIn   = "ran_in"
	RelTouched = "touched"
	RelSimilar = "similar"
)

// Event is one row of history. Everything the source knows and the store needs is here;
// received_at is the server's clock and is never part of the envelope.
type Event struct {
	// EventID is globally unique and stable across re-reads (see ID, HashID, ScopedID).
	EventID string `json:"event_id"`
	// TS is when it happened on the laptop, taken from the source, never from the time of reading.
	TS time.Time `json:"ts"`
	// Host is the laptop name. It never contains ':' or whitespace.
	Host string `json:"host"`
	// Agent is an open set matching ^[a-z][a-z0-9_-]{0,31}$: claude, codex, gemini, grok, feedguard, bus, ...
	Agent string `json:"agent"`
	// SessionKey is agent:host:native_session_id (see SessionKey).
	SessionKey string `json:"session_key"`
	// Role is one of the Role constants.
	Role string `json:"role"`

	ToolName string `json:"tool_name,omitempty"`
	// Command is set when the tool is a shell.
	Command string `json:"command,omitempty"`
	Cwd     string `json:"cwd,omitempty"`
	// ExitCode is a pointer so that exit code 0 is distinguishable from "not a command result".
	ExitCode *int `json:"exit_code,omitempty"`
	// Content is the redacted text, capped at MaxContentBytes.
	Content string `json:"content,omitempty"`
	// Raw is the full redacted source record, capped at MaxRawBytes.
	Raw json.RawMessage `json:"raw,omitempty"`
	// Truncated says Content or Raw was cut to fit its cap.
	Truncated bool `json:"truncated,omitempty"`

	// Model is the model that produced this event, as the tool reports it, lower-case:
	// "claude-sonnet-5-5", "gpt-5.6-terra". Producers set it on events of an assistant turn
	// (assistant text, tool calls) and leave it empty everywhere else and whenever the tool does
	// not say. It is never guessed.
	Model string `json:"model,omitempty"`
	// Effort is the thinking or reasoning level the tool reports for that turn, lower-case:
	// low, medium, high, xhigh, max, ... Same rules as Model.
	Effort string `json:"effort,omitempty"`
	// Thinking is true when the assistant turn contained a thinking or reasoning block. It is
	// only ever sent as true: absent means "not reported", not "did not think".
	Thinking bool `json:"thinking,omitempty"`

	// Session facts. The store keeps the latest non-empty value of each per session.
	NativeSessionID string `json:"native_session_id,omitempty"`
	AgentVersion    string `json:"agent_version,omitempty"`
	GitRepo         string `json:"git_repo,omitempty"`
	GitBranch       string `json:"git_branch,omitempty"`
	Title           string `json:"title,omitempty"`
	// ResumeCmd is how to reopen the session, for example "claude --resume <id>". The adapter
	// sets it because only the adapter knows its CLI.
	ResumeCmd string `json:"resume_cmd,omitempty"`

	// Edges are graph hints with this event's session as the source. The store adds ran_in itself.
	Edges []Edge `json:"edges,omitempty"`
}

// Edge links the event's session to another node: "session:<key>", "repo:<path>" or "file:<path>".
type Edge struct {
	Rel    string  `json:"rel"`
	Dst    string  `json:"dst"`
	Weight float32 `json:"weight,omitempty"`
}
