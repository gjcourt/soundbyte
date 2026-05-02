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
| `domain-no-other-internal` | Active ✓ | Only `internal/domain/doc.go` exists |
| `app-no-adapters` | Active ✓ | No `internal/app/` yet |
| `adapters-isolation` | Active ✓ | No `internal/adapters/` yet |

Rules are defined but only activate when `internal/` code exists. All green.

## Migration checklist

- [x] `internal/domain/doc.go` bootstrapped
- [x] `internal/testdoubles/` bootstrapped with `NewServerDeps()`
- [ ] Step 1 — extract domain types from `server/` + `pkg/` → `internal/domain/`
- [ ] Step 2 — define outbound ports in `internal/ports/outbound/`
- [ ] Step 3 — define inbound ports in `internal/ports/inbound/`
- [ ] Step 4 — create `internal/app/` (business logic)
- [ ] Step 5 — create `internal/adapters/` (infrastructure wrappers)
- [ ] Step 6 — thin `server/` into driving adapter
- [ ] Step 7 — add fakes to `testdoubles/`, wire `ServerDeps`
- [ ] Step 8 — delete legacy `server/`, `client/`, `pkg/`
