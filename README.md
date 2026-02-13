# SoundByte

A pure Go client/server application to stream audio over UDP using Raw PCM chunks (for simplicity and zero-dependency).
Designed to support Spotify Connect (via `librespot`) and Linux Pipes.

## Architecture

*   **Server**: Reads PCM audio from `stdin` or a named pipe, chunks it into 5ms frames, and streams it over UDP.
*   **Client**: Receives UDP stream, buffers packets (jitter buffer), and plays via default output device using `gopxl/beep`.

## Prerequisites

*   **Go 1.23+**
*   **Audio Dependency (Linux only)**: `libasound2-dev`
    ```bash
    sudo apt install libasound2-dev
    ```
*   **Spotify Connect**: Requires [librespot](https://github.com/librespot-org/librespot).
    *   Download a binary release or build with Cargo (`cargo install librespot`).
*   **SoX (Optional but Recommended)**: To resample audio sources to 48kHz.
    ```bash
    sudo apt install sox
    ```

## Installation

```bash
git clone <repo>
cd soundbyte
go mod download
```

## Usage

### 1. Start the Client

The client listens for UDP packets and plays them.

```bash
# Listen on port 5004
go run ./cmd/client -port 5004
```

### 2. Start the Server

The server reads standard input. You need to feed it 48kHz, Stereo, 16-bit Signed Little-Endian PCM.

#### Simple Test (File Source)
If you have a wav file:
```bash
ffmpeg -i track.mp3 -f s16le -ac 2 -ar 48000 - | go run ./cmd/server -addr 127.0.0.1:5004
```

#### Linux Pipe / Metadata
```bash
# Create a pipe
mkfifo /tmp/audio_pipe

# Run server reading from pipe
go run ./cmd/server -addr 127.0.0.1:5004 -input /tmp/audio_pipe
```

### 3. Spotify Connect Integration

`librespot` output is typically 44.1kHz. Ideally, resample to 48kHz for Opus.

**Command Chain:**
```bash
# 1. Start Client (on speaker machine)
go run ./cmd/client

# 2. Start Librespot -> SoX -> Server (on source machine)
librespot --name "GoStreamer" --bitrate 320 --backend pipe --device /tmp/spotifypipe --initial-volume 100 &

# 3. Read pipe, resample, send
tail -f /tmp/spotifypipe | \
sox -t raw -r 44100 -e signed -b 16 -c 2 - -t raw -r 48000 - | \
go run ./cmd/server -addr <CLIENT_IP>:5004
```

## Docker

### Build
```bash
docker-compose build
```

### Run
**Note:** Running the *Client* in Docker on macOS/Windows often results in no audio due to OS limitations. Running the Server in Docker is fine.

```bash
# Server only
docker-compose up server
```
