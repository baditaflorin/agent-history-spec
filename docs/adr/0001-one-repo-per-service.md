# 0001. One repository per service, one shared contract

Status: accepted, 2026-10-28

## Context

The first design (Build Spec v0.1) put seven binaries in one Go module. The services are being open
sourced, and people will want one of them (a transcript parser, a redactor, an MCP server) without the
rest.

## Decision

- Each service is its own repository with its own `go.mod`, README, LICENSE (Apache-2.0), releases and
  tests. Repository name equals binary name: `agent-tail`, `agent-store`, `agent-find`, `agent-mcp`,
  `agent-nudge`, `agent-eval`, `agent-embed`, `agent-redact`, `agent-transcripts`.
- The contract is this repository. It holds the event envelope (Go package `envelope`, JSON Schema),
  the OpenAPI file and the conformance vectors. A service claims conformance by passing the vectors.
- `envelope` is the only shared code, because event-id derivation must be identical everywhere. No other
  library is shared; two small copies beat a dependency between services.
- Services talk HTTP against `openapi/agent-history.v1.yaml`. Only `agent-store` reads or writes the database.

## Consequences

Ten repositories to release. A change to the contract is a new tag here first; breaking changes bump the
`v1` in the file names. The reference deployment composes the services; nothing requires it.
