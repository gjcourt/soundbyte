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

Single test: `go test ./internal/<package>/... -run TestX -v`
Pre-push: `golangci-lint run ./... && go test -race ./...`

## Architecture

Hexagonal architecture (ports & adapters). Two entry points: `cmd/server/` and `cmd/client/`.

- `cmd/server/` — wires stdin/file source + UDP sender → app streaming service.
- `cmd/client/` — wires UDP receiver → domain jitter buffer → beep audio player.
- `internal/domain/` — `Packet`, `Buffer`, protocol constants, `Decode`/`Encode`.
- `internal/ports/inbound/` — `StreamingService` interface.
- `internal/ports/outbound/` — `PCMSource`, `PacketSender`, `PacketReceiver` interfaces.
- `internal/app/` — `streamingService` (implements `inbound.StreamingService`).
- `internal/adapters/stdin/` — `Source` (implements `outbound.PCMSource`).
- `internal/adapters/udp/` — `Sender` + `Receiver` with HMAC-SHA256 auth.
- `internal/testdoubles/` — function-field fakes for all outbound ports.
- `pkg/auth/` — HMAC-SHA256 sign/verify (shared utility, no domain dependencies).
- `pkg/middleware/` — rate-limited packet-count logging (shared utility).

Protocol is intentionally simple raw PCM over UDP — no framing overhead beyond the 12-byte packet header.

## Conventions

- **Server reads PCM from stdin or a named pipe** — never opens audio devices itself.
- **Client owns playback** — only the client touches `gopxl/beep` / system audio.
- **Framing is fixed at 5ms** — changing the frame size requires changes on both sides.
- **Conventional Commits** for every commit (`feat:`, `fix:`, `chore:`, `refactor:`, `docs:`, `test:`, `ci:`).
- **Branch names** follow `<type>/<description>`.

## Invariants

- The wire format is a 12-byte header (4B sequence + 8B timestamp) followed by raw PCM. Adding header fields breaks compatibility on both sides.
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
