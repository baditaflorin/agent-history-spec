# 0006. Adapters without a real sample are labelled unverified

Status: accepted, 2026-10-28

## Context

Claude Code and Codex transcripts exist on the owner's machines and are used as golden files. No
Gemini CLI, Antigravity CLI or Grok CLI data was found on any of the four machines when this was
decided. An adapter written from documentation alone can drop fields silently when a real file differs.

## Decision

`agent-transcripts` marks an adapter `verified` only when it is tested on scrubbed real files of the
CLI version it names. The others report `unverified` in `agent-transcripts adapters` and in their
first log line, and carry golden files built from the documented format. The first real session
that appears is turned into a golden file and the label is lifted. `parse_errors_total{agent,agent_version}`
is the warning that a CLI changed its format.

## Consequences

Gemini and Grok support can ship early without claiming more than it has been shown to do.
