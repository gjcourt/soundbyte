---
title: "Hexagonal architecture migration"
status: "Completed"
created: "2026-05-02"
updated: "2026-05-03"
updated_by: "george"
tags: ["architecture", "hex", "refactor"]
---

# Hexagonal architecture migration

## Pre-migration layout (factual)

```
cmd/
  server/main.go   — wires UDP sender + stdin/file source
  client/main.go   — wires UDP receiver + jitter buffer + beep player
pkg/
  auth/            — HMAC-SHA256 packet sign/verify
  jitter/          — packet reordering buffer (sequence-keyed)
  protocol/        — wire packet (header + PCM payload) encode/decode
  middleware/      — rate-limited byte/packet count logger
```

This repo never had a WebSocket server, JWT auth, or session/mixing logic.
The transport has always been one-way UDP carrying raw PCM. The migration
goal was to relocate the existing types and I/O into a canonical hex
layout and wire fakes for tests, without changing the wire protocol.

## Migration steps (executed)

1. **Define `internal/domain/`** — move `pkg/protocol.Packet` and
   `pkg/jitter.Buffer` into pure-domain types (no external deps). Already
   bootstrapped with `doc.go`. One or more PRs.

2. **Define outbound ports in `internal/ports/outbound/`** —
   `PCMSource` (replaces ad-hoc stdin reader in `cmd/server`),
   `PacketSender`, `PacketReceiver`. One PR.

3. **Define inbound ports in `internal/ports/inbound/`** —
   `StreamingService` for the server's read-encode-send loop. One PR.

4. **Create `internal/app/`** — move the streaming loop out of
   `cmd/server/main.go` into `app.streamingService`, depending only on
   port interfaces. One PR.

5. **Create `internal/adapters/`** — wrap stdin/file (`adapters/stdin`)
   and UDP I/O (`adapters/udp`, including `pkg/auth` integration) as
   named adapters implementing outbound ports. One PR per adapter.

6. **Thin `cmd/server/` and `cmd/client/` into wiring** — both `main`
   functions become composition roots that wire ports to adapters.
   One PR.

7. **Add function-field fakes** — add `FakePCMSource`,
   `FakePacketSender`, `FakePacketReceiver` to `testdoubles/`,
   aggregated in `ServerDeps`.

8. **Delete legacy packages** — remove `pkg/protocol/` and
   `pkg/jitter/` once `internal/domain/` is the source of truth.
   `pkg/auth/` and `pkg/middleware/` remain as shared utilities with no
   domain dependencies.

## Depguard rules (active)

| Rule | Status | Notes |
|---|---|---|
| `domain-no-other-internal` | Active | domain cannot import ports/app/adapters |
| `ports-no-impl` | Active | ports cannot import app/adapters |
| `app-no-adapters` | Active | app cannot import adapters |
| `adapters-no-app` | Active | adapters cannot import app |

Bootstrapped before any code was moved into `internal/`; no legacy allow
entries were needed because the migration moved code in one shot rather
than incrementally.

## Status

Tracked in `docs/plans/2026-05-02-hex-migration-status.md`. All 8 steps
landed in commit `787154c` (refactor/hex-layout). Follow-up work (tests
for jitter/streaming/UDP adapters, observability via slog) is tracked
under separate critique fixes.
