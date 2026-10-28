# 0005. Dashboards reach the store through a server-side proxy

Status: accepted, 2026-10-28

## Context

A web dashboard that wants to show history has its own login and, usually, a strict Content-Security-Policy
such as `connect-src 'self'`. The store holds every prompt and tool call of every agent, so it must not be
exposed to browsers, and it must not need to know about any particular dashboard.

## Decision

The dashboard keeps one read token for the store in its server configuration and exposes its own
`/api/history/*` routes that forward `search`, `sessions`, `sessions/{key}/events` and `graph` after the
dashboard's own authentication. The browser only ever calls the dashboard's origin. A dashboard that already
publishes tools to agents can publish history the same way, so agents that reach it need no `agent-mcp`.
This repository only fixes the API such a proxy calls; the dashboard code lives with the dashboard.

## Consequences

A dashboard is one more client of the OpenAPI file and needs no special endpoint. Revoking it means revoking
one read token.
