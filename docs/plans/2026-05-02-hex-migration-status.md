---
title: "Hex migration status"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "tracking"]
---

# Hex migration status

## Depguard rules

| Rule | Status | Notes |
|---|---|---|
| `domain-no-other-internal` | Active ✓ | Prevents domain importing ports/app/adapters |
| `ports-no-impl` | Active ✓ | Prevents ports importing app/adapters |
| `app-no-adapters` | Active ✓ | Prevents app importing adapters |
| `adapters-no-app` | Active ✓ | Prevents adapters importing app layer |

All four canonical rules active and enforced by golangci-lint.

## Migration checklist

- [x] `internal/domain/doc.go` bootstrapped
- [x] `internal/testdoubles/` bootstrapped with `NewServerDeps()`
- [x] Step 1 — extract domain types from `pkg/` → `internal/domain/` (`Packet`, `Buffer`, protocol constants)
- [x] Step 2 — define outbound ports in `internal/ports/outbound/` (`PCMSource`, `PacketSender`, `PacketReceiver`)
- [x] Step 3 — define inbound ports in `internal/ports/inbound/` (`StreamingService`)
- [x] Step 4 — create `internal/app/streaming.go` (`streamingService` implementing `StreamingService`)
- [x] Step 5 — create `internal/adapters/stdin/` and `internal/adapters/udp/` with interface assertions
- [x] Step 6 — rewrite `cmd/server/main.go` to wire hex components; `cmd/client/main.go` uses domain types directly
- [x] Step 7 — add fakes to `testdoubles/`, wire `ServerDeps` (`FakePCMSource`, `FakePacketSender`, `FakePacketReceiver`)
- [x] Step 8 — delete legacy `pkg/protocol/` and `pkg/jitter/` (moved to `internal/domain/`)
