---
title: "Hexagonal architecture migration"
status: "In progress"
created: "2026-05-02"
updated: "2026-05-02"
updated_by: "george"
tags: ["architecture", "hex", "refactor"]
---

# Hexagonal architecture migration

## Current layout

```
server/    — WebSocket server, session management, audio mixing
client/    — WebSocket client (thin wrapper)
cmd/       — main entry points
pkg/
  auth/       — JWT authentication
  jitter/     — jitter buffer
  middleware/ — HTTP middleware
  protocol/   — audio packet format
```

No `internal/` packages exist yet. This is the largest restructure across the
six Go repos in the hex migration wave.

## Migration steps

1. **Define `internal/domain/`** — extract audio packet types, session
   concepts, and jitter buffer logic from `server/` and `pkg/` into pure
   domain types with no external deps. Already bootstrapped with `doc.go`.
   One or more PRs.

2. **Define outbound ports in `internal/ports/outbound/`** — identify external
   dependencies (audio sinks, token verifiers, session stores) and express
   them as interfaces. One PR.

3. **Define inbound ports in `internal/ports/inbound/`** — interfaces for the
   WebSocket entry point and any public control surface. One PR.

4. **Create `internal/app/`** — move business logic (mix decisions, session
   lifecycle) out of `server/` into app services that depend only on port
   interfaces. One PR per use case.

5. **Create `internal/adapters/`** — wrap `pkg/auth`, WebSocket I/O, and
   other infrastructure as named adapters implementing outbound ports. Move
   from `pkg/` to `internal/adapters/<name>/`. One PR per adapter.

6. **Thin `server/` into a driving adapter** — `server/` becomes a thin
   wrapper that wires inbound requests to `ports/inbound/` calls. One PR.

7. **Add function-field fakes** — add fakes for each outbound port to
   `testdoubles/`, wire into `ServerDeps`.

8. **Delete legacy packages** — once everything is migrated, remove `server/`,
   `client/` (if wrapped), and `pkg/`. Update `cmd/` accordingly.

## Depguard notes

Bootstrap rules active for `internal/` packages only. Existing `server/`,
`client/`, and `pkg/` code is unaffected until migration steps move it to
`internal/`.

No legacy allow entries needed until code is first moved into `internal/`.
