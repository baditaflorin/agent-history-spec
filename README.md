# agent-history-spec

The contract of the agent-history family: one searchable store for every prompt, reply, tool call and bus
message of every coding agent (Claude Code, Codex, Gemini CLI, Grok Build, and more) on every machine.

This repository holds no service. It holds what the services agree on:

| Path | What it is |
|---|---|
| `envelope/` | Go package: the `Event` type, `Validate`, event-id rules, NDJSON reader and writer |
| `schema/event.v1.schema.json` | the same envelope as JSON Schema, for implementers outside Go |
| `openapi/agent-history.v1.yaml` | the one HTTP API of `agent-store`, which everything else speaks |
| `vectors/` | conformance vectors: event ids, valid events, invalid events with the field each must blame |
| `docs/adr/` | why it is shaped this way |

## The services

Each one is its own repository and passes the vectors here. Only `agent-redact` is released so far; the
others are being built against this contract.

| Service | Job | Status |
|---|---|---|
| [`agent-redact`](https://github.com/baditaflorin/agent-redact) | remove secrets from text and JSON | v0.1.0 |
| `agent-transcripts` | read Claude, Codex, Gemini, Grok and audit-log files into envelopes | in progress |
| `agent-store` | ingest and query API over Postgres with pgvector | in progress |
| `agent-tail` | collector: watch, spool, ship | planned |
| `agent-embed` | embedding worker | planned |
| `agent-find` | query CLI | planned |
| `agent-mcp` | MCP server over stdio | planned |
| `agent-nudge` | prompt-time hook: has someone already worked on this? | planned |
| `agent-eval` | golden-set runner and load generator | planned |

## Using the envelope

```go
import "github.com/baditaflorin/agent-history-spec/envelope"

e := envelope.Event{
    EventID:    envelope.ID(nativeID, sessionKey, lineNo, text),
    TS:         tsFromTheSource,
    Host:       "laptop1",
    Agent:      "claude",
    SessionKey: envelope.SessionKey("claude", "laptop1", sessionID),
    Role:       envelope.RoleToolCall,
    Command:    "docker compose down",
}
if err := envelope.Validate(&e); err != nil { /* *envelope.FieldError names the field */ }
_ = envelope.WriteNDJSON(w, []envelope.Event{e})
```

Rules that matter:

- `ts` comes from the source, never from the moment of reading.
- `event_id` is globally unique and stable. Use a native id only if it is unique across sessions;
  otherwise `ScopedID`; with none, `HashID`.
- Content over 32 KiB or raw over 1 MiB is cut by the producer (`truncated: true`); the store refuses
  a larger event instead of cutting it.
- Unknown fields are ignored, so a newer producer does not break an older store.

## Tests

```
go test ./...
```

The vectors are the specification. A change to `envelope/` that changes a vector is a breaking change.

## Status

Pre-1.0. The envelope and API are `v1` but may change until the first services are released.
Licensed under Apache-2.0.
