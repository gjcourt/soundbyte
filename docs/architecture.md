# Soundbyte Architecture

Soundbyte is a real-time audio streaming application. It consists of a server (sender) and a client (receiver).

## Components

### Server
- Reads audio data from an input source (e.g., stdin).
- Encodes the audio data and sends it over UDP to connected clients.

### Client
- Receives audio data over UDP.
- Decodes the audio data and plays it using the host's sound device.

### Protocol
- Uses a custom protocol over UDP for low-latency audio streaming.
- Includes jitter buffering to handle network latency and packet loss.
