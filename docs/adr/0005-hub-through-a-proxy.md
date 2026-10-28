# 0005. The hub reaches the store through a server-side proxy

Status: accepted, 2026-10-28

## Context

`hub.scrapetheworld.org` is a private dashboard with its own login, a keystore for read-only agent
keys, and a Content-Security-Policy of `connect-src 'self'`. The store must not be exposed to browsers
and must not learn about the hub.

## Decision

The hub holds one read token for the store in its server configuration and exposes `/api/history/*`
that forwards `search`, `sessions`, `sessions/{key}/events` and `graph` after its own authentication.
The browser only calls the hub's origin. History is also published in the hub's agent contract, so
agents that already reach the hub through the MCP gateway can query history without `agent-mcp`.
The code lives in the private hub repository; this repository only fixes the API it calls.

## Consequences

The hub is one more client of the OpenAPI file and needs no special endpoint. Revoking the hub means
revoking one read token.
