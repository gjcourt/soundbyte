---
title: SoundByte authentication
status: Stable
created: 2026-05-02
updated: 2026-05-03
updated_by: gjcourt
tags: [reference, auth, hmac]
---

# SoundByte authentication

SoundByte authenticates UDP packets with a shared HMAC-SHA256 secret. There
are no user logins, tokens, JWTs, sessions, or WebSocket handshakes — the
auth layer is a per-packet signature only. Source of truth:
`pkg/auth/auth.go` and `internal/adapters/udp/transport.go`.

## Mode selection

The shared secret is supplied via the `--token` flag on **both** the server
and the client. If `--token` is empty (the default), packets are sent and
accepted as-is (auth disabled) — useful for local development on a trusted
LAN.

When the token is set, both sides must agree byte-for-byte. The token is
used directly as the HMAC key; there is no hashing or KDF.

## Wire format

When auth is enabled the sender appends a 32-byte HMAC-SHA256 tag computed
over the entire encoded packet (header + PCM payload):

```
[ 12-byte header ][ PCM payload ][ 32-byte HMAC-SHA256 tag ]
```

`MACSize = 32` (`pkg/auth.MACSize`).

## Sender (`auth.Sign`)

```go
mac := hmac.New(sha256.New, key)
mac.Write(data)
return append(data, mac.Sum(nil)...)
```

Called by `internal/adapters/udp.Sender.Send` before writing to the UDP
connection.

## Receiver (`auth.Verify`)

1. If `key == nil`, return `data` unchanged (auth disabled).
2. Reject `len(data) < MACSize` with `ErrInvalidMAC`.
3. Split: `payload := data[:len(data)-MACSize]`, `tag := data[len(data)-MACSize:]`.
4. Recompute HMAC-SHA256 over `payload`; compare with `hmac.Equal`
   (constant-time).
5. On mismatch, return `ErrInvalidMAC`. The client's receive loop drops the
   packet silently.

Called by `internal/adapters/udp.Receiver.Receive` immediately after
`ReadFromUDP`.

## Threat model and limitations

- **In scope.** Tampering with the packet body, payload truncation, and
  unauthenticated injection from an attacker who does not hold the shared
  secret. The MAC is checked in constant time.
- **Out of scope.** Confidentiality (the PCM payload is not encrypted),
  replay protection (an attacker on-path can capture and replay valid
  signed packets indefinitely — see `docs/plans/` for any open work), and
  forward secrecy. Operate on a trusted LAN.

## Operational notes

- Rotating the token requires restarting both server and client; there is
  no online rekey.
- The token must be transported out-of-band (e.g. config file, env var).
  Do not commit tokens to the repo.
