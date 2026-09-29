# SoundByte

SoundByte streams audio over UDP as raw 16-bit PCM chunks. A pure Go server
reads PCM from stdin or a named pipe and blasts it to one or more clients,
which buffer and play it back with `gopxl/beep`. It's built for a homelab LAN:
feed it `librespot` for Spotify Connect, or any program that writes PCM to a
pipe.

There's no codec, no ACKs, and no retransmission — just fixed-size frames,
a 12-byte header, and an optional HMAC signature.

## Quickstart

Start a client to listen for audio:

```bash
go run ./cmd/client -port 5004
```

Feed the server PCM. Audio must be 48kHz, stereo, 16-bit signed little-endian:

```bash
ffmpeg -i track.mp3 -f s16le -ac 2 -ar 48000 - | go run ./cmd/server -addr 127.0.0.1:5004
```

Or point the server at a named pipe instead of stdin:

```bash
mkfifo /tmp/audio_pipe
go run ./cmd/server -addr 127.0.0.1:5004 -input /tmp/audio_pipe
```

### Spotify Connect via librespot

`librespot` outputs 44.1kHz PCM, so resample to 48kHz before it reaches the
server:

```bash
librespot --name "SoundByte" --bitrate 320 --backend pipe --device /tmp/spotifypipe --initial-volume 100 &

tail -f /tmp/spotifypipe | \
  sox -t raw -r 44100 -e signed -b 16 -c 2 - -t raw -r 48000 - | \
  go run ./cmd/server -addr <CLIENT_IP>:5004
```

## Configuration

Both binaries are configured entirely by flags; there are no config files or
environment variables.

| Binary | Flag | Default | Meaning |
|---|---|---|---|
| server | `-addr` | `255.255.255.255:5004` | Target UDP address (broadcast by default) |
| server | `-input` | `stdin` | `stdin` or a path to a file/named pipe |
| server | `-token` | `""` | Shared secret for HMAC-SHA256 packet auth; empty disables it |
| client | `-port` | `5004` | UDP port to listen on |
| client | `-buf` | `20` | Jitter buffer size in packets (20 × 5ms ≈ 100ms) |
| client | `-token` | `""` | Must match the server's `-token`, or be empty on both |

Authentication is off by default. When `-token` is set on both ends, every
packet is signed with HMAC-SHA256 and unsigned or mismatched packets are
dropped silently. There's no encryption and no replay protection — see
[`docs/reference/2026-05-02-authentication.md`](docs/reference/2026-05-02-authentication.md)
for the threat model.

## How it works

The server splits incoming PCM into fixed 5ms frames (960 bytes at 48kHz
stereo 16-bit), wraps each in a 12-byte header (sequence + timestamp), signs
it if a token is set, and sends it over UDP. The client verifies, decodes,
and pushes frames into a jitter buffer that reorders by sequence number and
only starts playback once enough packets have accumulated. `gopxl/beep`
renders the reassembled PCM to the host's default audio device.

```
librespot/pipe → stdin → server (framing, optional HMAC) → UDP → client (verify, jitter buffer) → beep → speakers
```

The code follows a hexagonal (ports & adapters) layout — `internal/domain`
for the wire format and buffer, `internal/ports` for the interfaces,
`internal/adapters` for stdin/UDP I/O, `internal/app` for the server-side
streaming use case — enforced in CI by `go-arch-lint`. The full breakdown,
including the component diagram and why the client skips the app layer, is in
[`docs/architecture.md`](docs/architecture.md).

## Development

Requires Go 1.25+ and, for the client only, ALSA headers on Linux:

```bash
sudo apt install libasound2-dev
```

```bash
make build   # compile ./server and ./client
make test    # go test -race -v ./...
make all     # test + build
```

Before pushing:

```bash
golangci-lint run ./...
go test -race ./...
```

Boundary checks (`go-arch-lint`) run in CI on every PR; to run the same check
locally:

```bash
go install github.com/fe3dback/go-arch-lint@v1.18.0
go-arch-lint check
```

## Docker

```bash
docker-compose build
docker-compose up server   # run just the server, e.g. to pipe audio into it
```

The compose file wires a server and client together for a local end-to-end
test, with the client on the host's ALSA device (`/dev/snd`). Running the
client in Docker only produces audio on Linux — on macOS or Windows run it
natively with `go run ./cmd/client` instead.

`.github/workflows/image.yml` publishes two images to GHCR on every push to
`main`: `ghcr.io/gjcourt/soundbyte` (server, `linux/amd64` and `linux/arm64`)
and `ghcr.io/gjcourt/soundbyte-client` (client, `linux/amd64` only, since it
needs CGO and ALSA). Each build is tagged with the date, `latest`, and an
immutable `<date>-<sha7>`.

## Documentation

Further docs live under [`docs/`](docs/), organized by topic — architecture,
design proposals, operations, migration plans, protocol/auth reference, and
research spikes. Start at [`docs/README.md`](docs/README.md).
