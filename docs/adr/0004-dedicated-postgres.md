# 0004. The store ships its own Postgres with pgvector

Status: accepted, 2026-10-28

## Context

The fleet already has a Postgres 17 host. It is a shared 4-core machine without pgvector. The store needs
`pg_trgm` and `vector`, large GIN and HNSW indexes, and monthly partitions, and it will be open sourced.

## Decision

`agent-store` documents and ships a compose file with `pgvector/pgvector:pg17` and its own volume. The
reference deployment runs it on a fleet Docker host, not on the shared Postgres. Backups are a nightly
dump through the same role that provisions it. Extensions are never installed on shared database hosts for this.

## Consequences

One more database to run and back up. Nothing else in the fleet can be hurt by the indexes, and anyone
can reproduce the setup with one `docker compose up`.
