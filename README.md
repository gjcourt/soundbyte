<!-- readme-type: service -->
# SoundByte

Streams raw PCM audio over UDP from a pipe or Spotify Connect source to LAN clients

Streaming audio from a source machine to a speaker elsewhere on the LAN
usually means running a full media server or dealing with codec negotiation
and buffering. SoundByte skips that: a pure Go server reads raw PCM from
stdin or a named pipe and sends fixed 5ms frames over UDP, and a client
buffers and plays them back — no codec, no ACKs, no retransmission. It pairs
with `librespot` for Spotify Connect, or any program that writes PCM to a
pipe, and can optionally authenticate packets with HMAC-SHA256.

**Status:** actively developed since 2026-02 (last commit 2026-09-24);
publishes GHCR images on every push to `main` but is not deployed in the
homelab yet.

```text
$ go build -o server ./cmd/server && ./server -h
Usage of ./server:
  -addr string
    	Target UDP address (default "255.255.255.255:5004")
  -input string
    	Path to input pipe/file (or 'stdin') (default "stdin")
  -token string
    	Shared secret for HMAC-SHA256 packet authentication (optional)
```

## Quick start

Needs: Go 1.25.7+, plus ALSA headers (`libasound2-dev`) on Linux for the client.

```bash
git clone https://github.com/gjcourt/soundbyte && cd soundbyte
go run ./cmd/client -port 5004
```

In a second terminal, feed it 48kHz stereo 16-bit PCM:

```bash
ffmpeg -i track.mp3 -f s16le -ac 2 -ar 48000 - | go run ./cmd/server -addr 127.0.0.1:5004
```

## Usage

Point the server at a named pipe instead of stdin:

```bash
mkfifo /tmp/audio_pipe
go run ./cmd/server -addr 127.0.0.1:5004 -input /tmp/audio_pipe
```

Feed it Spotify Connect audio via `librespot`, resampled from 44.1kHz to 48kHz:

```bash
librespot --name "SoundByte" --bitrate 320 --backend pipe --device /tmp/spotifypipe --initial-volume 100 &

tail -f /tmp/spotifypipe | \
  sox -t raw -r 44100 -e signed -b 16 -c 2 - -t raw -r 48000 - | \
  go run ./cmd/server -addr <CLIENT_IP>:5004
```

Run both containers with Docker Compose instead of `go run` — the server
sends to the `client` service, and the client needs `/dev/snd`, so audio
output only works on a Linux host:

```bash
docker-compose build
docker-compose up
```

For the full flag list: `go run ./cmd/server -h` or `go run ./cmd/client -h`.

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
only starts playback once enough packets have accumulated; `gopxl/beep`
renders the reassembled PCM to the host's default audio device.

```
librespot/pipe → stdin → server (framing, optional HMAC) → UDP → client (verify, jitter buffer) → beep → speakers
```

The code follows a hexagonal (ports & adapters) layout, enforced in CI by
`go-arch-lint`. The full component diagram, the ports/adapters map, and the
doc index (design proposals, operations runbooks, migration plans, protocol
reference) are in [`docs/architecture.md`](docs/architecture.md) and
[`docs/README.md`](docs/README.md).

## Development

Requires Go 1.25.7+ and, for the client only, ALSA headers on Linux:

```bash
sudo apt install libasound2-dev
```

```bash
make build   # compile ./server and ./client
make lint    # golangci-lint run ./... — CI lint job
make test    # go test -race -v ./... — CI test job
```

The CI `arch-lint` job checks the hexagonal boundaries; to run the same check
locally:

```bash
go install github.com/fe3dback/go-arch-lint@v1.18.0
go-arch-lint check
```

Conventions for contributors and agents: [AGENTS.md](AGENTS.md).

## Deployment

`.github/workflows/image.yml` publishes two images to GHCR on every push to
`main`: `ghcr.io/gjcourt/soundbyte` (server, `linux/amd64` and
`linux/arm64`) and `ghcr.io/gjcourt/soundbyte-client` (client, `linux/amd64`
only — needs CGO and ALSA). See [AGENTS.md](AGENTS.md) for the tag/mutability
contract. Not deployed in the homelab yet, so there's no runbook.

## License

[Apache-2.0](LICENSE).
