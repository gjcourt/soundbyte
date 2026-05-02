---
title: SoundByte authentication
status: Stable
created: 2026-05-02
updated: 2026-05-02
updated_by: gjcourt
tags: [reference, auth, hmac]
---

# Soundbyte Authentication

Soundbyte uses token-based authentication for securing audio streams.

## Flow
1. User logs in via `/api/auth/login` and receives a token.
2. The token is used to authenticate WebSocket connections for audio streaming.
