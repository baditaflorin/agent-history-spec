# 0004. The store runs on its own Postgres with pgvector

Status: accepted, 2026-10-28

## Context

The store needs the `pg_trgm` and `vector` extensions, large GIN and HNSW indexes, and monthly partitions.
A Postgres you already run is often shared with other workloads and may not offer pgvector at all, and
heavy index builds on it would be felt by everything else that uses it.

## Decision

`agent-store` documents and ships a compose file with the `pgvector/pgvector:pg17` image and its own
volume. Backups are a nightly dump. Extensions are not installed into a shared database for this.

## Consequences

One more database to run and back up. Nothing else can be hurt by the indexes, and anyone can reproduce
the setup with one `docker compose up`.
