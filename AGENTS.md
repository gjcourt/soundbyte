# AGENTS.md

> SoundByte streams audio over UDP as raw PCM chunks — a pure-Go server reads from stdin or a named pipe and a client buffers and plays back via `gopxl/beep`. Designed for Spotify Connect via `librespot` and Linux pipe sources. — https://github.com/gjcourt/soundbyte

## Commands

| Command | Use |
|---------|-----|
| `make build` | Compile both server and client binaries |
| `make test` | Run all tests |
| `make all` | test + build |
| `make run-server` | Build and run the server |
| `make clean` | Remove built binaries |
| `docker-compose up` | Integrated server+client local run |

Single test: `go test ./pkg/<package> -run TestX -v`
Pre-push: `golangci-lint run ./... && go test -race ./...`

## Architecture

Pre-hexagonal layout — to be restructured into `internal/{domain,ports,app,adapters}` as part of the cross-repo hex migration. Today:

- `cmd/server/` — server entry point (reads PCM, chunks into 5ms frames, streams over UDP).
- `cmd/client/` — client entry point (receives UDP, jitter-buffers, plays through `gopxl/beep`).
- `server/` — server-side logic (PCM ingestion, framing, UDP transport).
- `client/` — client-side logic (UDP receive, buffering, playback).
- `pkg/` — shared library code reused by both binaries.

Protocol is intentionally simple raw PCM over UDP — zero framing overhead.

## Conventions

- **Server reads PCM from stdin or a named pipe** — never opens audio devices itself.
- **Client owns playback** — only the client touches `gopxl/beep` / system audio.
- **Framing is fixed at 5ms** — changing the frame size requires changes on both sides.
- **Conventional Commits** for every commit (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `ci:`).
- **Branch names** follow `<type>/<description>`.

## Invariants

- The wire format is raw PCM — no framing, no headers. Adding fields breaks compatibility on both sides.
- Audio system libraries (PortAudio or platform equivalent) are required only on the client.
- The compiled binaries live at `./server` and `./client`; never committed.

## What NOT to Do

- Do not call `gopxl/beep` from the server — server is headless / device-free.
- Do not introduce per-chunk framing overhead in the protocol — keep raw PCM.
- Do not commit binaries (`server`, `client`, `*_unix`).
- Do not skip race testing for code paths that touch the UDP socket or jitter buffer.

## Domain

Self-hosted audio streaming for the homelab. Sources (Spotify Connect via `librespot`, Linux pipes from media players) feed PCM into the server, which broadcasts to one or more UDP-connected clients on the LAN. Latency budget is small (<50ms) since playback is synchronous with the source.

## Cross-service dependencies

| Service | Purpose |
|---|---|
| `librespot` (Spotify Connect) | Optional PCM source feeding the server's stdin |
| Linux named pipe | Generic PCM source |
| Snapcast (homelab) | Related multi-room audio system; SoundByte covers the simpler pipe-based use case |

## Quality gate before push

1. `golangci-lint run ./...`
2. `go test -race ./...`
3. `make build`

## Documentation

`docs/` taxonomy: `architecture/` · `design/` · `operations/` · `plans/` · `reference/` · `research/`. See each folder's `README.md` for scope. Index: `docs/README.md`.

## Observability

Logs to stderr in slog text format. No metrics endpoint today; cluster-level pod status is the source of health signal when running in the homelab.

When you learn a new convention or invariant in this repo, update this file.
