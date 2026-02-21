# Soundbyte API Documentation

Soundbyte is a real-time audio streaming application. The API provides endpoints for managing audio streams and user authentication.

## Endpoints

### Authentication
- `POST /api/auth/login`: Authenticate a user and receive a token.

### Streams
- `GET /api/streams`: List available audio streams.
- `POST /api/streams`: Create a new audio stream.
- `WS /api/streams/{id}`: Connect to an audio stream via WebSocket.
