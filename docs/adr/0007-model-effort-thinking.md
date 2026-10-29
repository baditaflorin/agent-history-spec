# 0007. The answering model, the effort level and thinking are fields of the envelope

Status: accepted, 2026-10-29 (contract v0.2.0)

## Context

People want "only the sessions that Sonnet answered" or "only what ran at effort xhigh". Claude Code writes the
answering model, the requested effort and a thinking block on every assistant record; Codex writes `model` and
`effort` once per turn. All of it was already inside `raw`, but a filter over `raw` cannot use an index and means
nothing to a tool that does not know each agent's file format.

## Decision

Three optional envelope fields, set by the producer on the events of an assistant turn (assistant text and tool
calls) and empty elsewhere:

- `model`: the model that produced the event, as the tool reports it, lower-case (`claude-sonnet-5-5`).
  Placeholders such as `<synthetic>` are not a model and are left empty.
- `effort`: the thinking or reasoning level the tool reports, lower-case (`low`, `medium`, `high`, `xhigh`, `max`, ...).
  The set is open; the store does not enumerate it.
- `thinking`: true when the turn contained a thinking or reasoning block. It is only ever sent as true, so absent
  means "not reported", never "did not think".

A producer never guesses: a tool that does not say (today the Gemini and Grok adapters, see 0006) leaves the fields
empty. `envelope.CleanModel` and `CleanEffort` turn whatever a tool wrote into a valid value or an empty string,
so an odd value costs the field and never the event. `Validate` refuses a non-empty value that is not lower-case or
is too long.

The API gains `model` and `effort` as filters on `/v1/search` and `/v1/sessions`: exact, case-insensitive, and a
trailing `*` on `model` matches a prefix. A session matches when at least one of its answers does, so a session that
changed model matches each of them. Hits carry `model` and `effort`; a session summary carries `models` and
`efforts`, the sorted distinct values seen. `thinking` is stored but not a filter or an output field in v0.2.0.

## Consequences

Backward compatible: unknown fields are ignored, so an older store accepts a newer collector and an older collector
sends none. A store older than v0.2.0 drops the values silently and does not filter. Old events stay without the
fields until they are re-sent or backfilled from `raw`, so a filter only sees what the store could derive. Effort
can change inside a session; it belongs to the event, and the session keeps the set.
