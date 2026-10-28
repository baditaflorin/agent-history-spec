# 0003. Redact at the source and again at the store

Status: accepted, 2026-10-28

## Context

The store holds every secret an agent has ever seen. Redaction at the collector is right, because
a secret should never leave the laptop. But one buggy or outdated collector, or a pusher that was never
given a redactor (feed guard, the bus bridge), would put a key into the shared store.

## Decision

`agent-tail` redacts before writing its spool. `agent-store` runs the same `agent-redact` rules over
`content`, `command` and `raw` of every accepted event before inserting. Matches become
`[REDACTED:<rule>]`. A daily canary session containing fake keys of every rule is ingested; a query
counts stored rows that match any key pattern and must return 0.

## Consequences

Ingest pays one more pass over each event. A redaction rule added to `agent-redact` protects old
collectors immediately. The rules are public once the repository is; the canary check is what catches
a rule that misses.
