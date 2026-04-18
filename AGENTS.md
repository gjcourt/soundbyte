# SoundByte Agent Guidelines

## Repository Overview

SoundByte is a pure Go client/server application that streams audio over UDP using raw PCM chunks. Designed to support Spotify Connect (via `librespot`) and Linux pipes. The server reads from stdin or a named pipe; the client buffers and plays back via `gopxl/beep`.

## Project Structure

```
cmd/
  server/          ← server entry point (reads PCM, streams UDP)
  client/          ← client entry point (receives UDP, plays audio)
pkg/               ← shared library code
client/            ← client-specific logic
docs/              ← API, authentication, architecture docs
docker-compose.yml ← local dev / integration setup
Dockerfile         ← container image build
```

## Common Commands

```bash
make build         # compile server and client binaries
make test          # run tests
make all           # test + build

# Run via docker-compose for integrated testing
docker-compose up
```

## Architecture

- **Server**: Reads PCM audio from `stdin` or a named pipe → chunks into 5ms frames → streams over UDP.
- **Client**: Receives UDP stream → jitter buffer → plays via default audio output using `gopxl/beep`.
- Protocol is intentionally simple raw PCM over UDP — zero framing overhead.
- See `docs/architecture.md` for detailed design decisions.

## Development Notes

- Requires **Go 1.23+**.
- Audio output on the client uses `gopxl/beep` — ensure system audio libs are present (PortAudio or equivalent depending on platform).
- API reference: `docs/api.md`. Auth details: `docs/authentication.md`.
- The SoundByte server is related to the homelab Snapcast setup — see `../homelab/apps/base/snapcast/`.
