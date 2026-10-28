# 0002. Deduplicate on `event_id` alone

Status: accepted, 2026-10-28

## Context

The spec partitions `events` by month on `ts` and makes `(ts, event_id)` the primary key, because
Postgres requires the partition key in a unique index. Re-reading a source must not add rows. But an
adapter can produce the same event with a different `ts` on a second read: Grok's `chat_history.jsonl`
has no timestamps and is joined with `updates.jsonl`, so a first read before the matching update
arrives and a later read give different times. With `(ts, event_id)` that is a duplicate row, and the
idempotency test ("run the same backfill twice, the row count does not change") would pass on clean
data and fail in production.

## Decision

The store keeps a small unpartitioned table `event_ids (event_id text primary key, ts timestamptz)`.
Ingest inserts ids there first (`ON CONFLICT DO NOTHING RETURNING`) and writes to `events` only the
events whose id was new. The first copy wins. `events` keeps its `(ts, event_id)` key for the partitioning.

Producers keep `ts` stable anyway: it comes from the source, never from the moment of reading.
Event ids are globally unique (see `envelope.ID`, `ScopedID`); an id that is only unique inside a
session must be scoped, because a collision silently drops a real event.

## Consequences

One more index of every id (about 70 bytes per event). A corrected timestamp never replaces a stored one.
