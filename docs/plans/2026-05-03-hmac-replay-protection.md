---
title: "HMAC replay protection"
status: "Draft"
created: "2026-05-03"
updated: "2026-05-03"
updated_by: "george"
tags: ["security", "protocol", "udp", "auth"]
---

# HMAC replay protection

## Problem statement

The current packet authentication scheme (`pkg/auth/auth.go`,
`internal/adapters/udp/transport.go`) appends an HMAC-SHA256 tag over the
12-byte header (`Sequence` + `Timestamp`) and the PCM payload. The receiver
verifies the tag and strips it, but **never inspects the timestamp or sequence
fields for freshness**:

- `internal/adapters/udp/transport.go:46-59` — `Receive` returns the verified
  bytes directly; no clock check.
- `internal/domain/packet.go:30-37` — `Sequence` and `Timestamp` are decoded
  but only `Sequence` is used downstream, and only as a jitter-buffer sort key
  (`internal/domain/jitter.go:30-32`). There is no last-seen sequence cursor.
- `internal/app/streaming.go:48-50` — sender sets `Sequence` starting at 0 on
  every server start and uses `time.Now().UnixNano()` for `Timestamp`. There
  is no session/epoch identifier.

A passive on-path attacker on the LAN can therefore capture any signed packet
and replay it (or a burst of them) verbatim, indefinitely, without any cached
HMAC key. The signature still validates because every byte in the captured
packet is unchanged.

## Threat model

**In scope** (LAN attacker, the realistic threat for a homelab broadcast):

1. **Replay of a captured packet.** Single old PCM frame is re-injected; the
   client's jitter buffer accepts it. Because `Sequence` is used only as a
   sort key, a packet with `Sequence=42` replayed an hour later is sorted into
   the buffer wherever 42 currently falls relative to the live sequence
   cursor — typically far in the past, so it is dropped on the way out of the
   buffer once `len(packets) >= minCount` and 42 sorts to head while live
   seqs are e.g. ~720,000. In practice this *currently* mitigates replay by
   accident, but it is not a security guarantee — see below.

2. **Replay storm to induce duplicate audio / DoS the buffer.** An attacker
   replays a recent window (say the last 1 s = ~200 packets) at high rate.
   The packets all verify; they all push into `Buffer` under `b.mu.Lock()`
   and trigger `sort.Slice` over a growing slice (jitter perf finding,
   moderate in critique). At sustained replay rates the lock and the O(n log n)
   sort become a denial-of-service vector against the receive goroutine.

3. **Server restart attack.** When the server restarts, `Sequence` resets to
   0 (`internal/app/streaming.go:31`). An attacker who captured packets from
   *before* the restart with seq 100..200 can replay them after the restart
   while the live server is emitting seq 0..99, and the jitter buffer will
   happily accept and play them as future audio. This is the strongest
   concrete attack and it is not mitigated by the accidental sort-order
   defense.

4. **Pre-recorded session injection.** Attacker captures a full session,
   later replays the entire stream against the same key. With no clock
   binding, the receiver cannot tell yesterday's audio from today's.

**Out of scope** (would not be addressed by this plan, called out for
clarity):

- Off-path attacker — they don't have the key, HMAC already covers it.
- Active MITM that can drop and rewrite packets — UDP broadcast on a trusted
  LAN; MITM on switched LAN is a different threat tier.
- Key compromise / forward secrecy — separate problem; key rotation is
  discussed only as it interacts with this design.

## Design options

### Option A — Timestamp window only

Receiver rejects any verified packet whose `Timestamp` deviates from
`time.Now().UnixNano()` by more than `replayWindow` (e.g. 2 s).

```go
// pseudocode in udp.Receiver.Receive
if absDelta(now, pkt.Timestamp) > replayWindowNs {
    return nil, raddr, ErrStale
}
```

| Property | Verdict |
|---|---|
| Server-side state | None |
| Packet overhead | 0 bytes |
| Multi-client | Trivial — pure stateless check on each receiver |
| Key rotation | Orthogonal |
| False positives | Clock skew between server and client; LAN NTP usually <50 ms but edge cases exist |
| Stops single-packet replay | Only if attacker replays >2 s after capture |
| Stops fast replay | **No** — attacker can replay within the window forever |
| Stops post-restart replay | Partially — only mitigates packets older than the window |

Cheap, weak. Real-time audio means the legitimate window can be very small
(<1 s) which helps, but it does not stop a determined fast replayer at all.

### Option B — Timestamp window + nonce + bounded seen-cache

Add a 16-byte random nonce per packet. Receiver enforces:

1. Timestamp inside `[now - W, now + skew]` (W ~= 2 s).
2. Nonce not present in a server-side seen-set scoped by the timestamp window.

The seen-set is a ring of two `map[[16]byte]struct{}` buckets sized for
`expectedPPS * W` (e.g. 200 pkt/s * 2 s = 400 entries) with rotation when
the active window slides forward.

| Property | Verdict |
|---|---|
| Server-side state | ~64 KB per receiver (bounded by window × pps) |
| Packet overhead | +16 bytes per packet |
| Multi-client | Each receiver maintains its own cache — fine |
| Key rotation | Orthogonal; cache resets are safe |
| False positives | Only on duplicate UDP delivery (rare, network-stack double-send) — same packet bytes including same nonce; an arguable correct rejection |
| Stops single-packet replay | Yes |
| Stops fast replay | Yes — every replayed nonce is rejected for as long as the timestamp keeps it inside the window, then it falls outside the window and is rejected by the timestamp check |
| Stops post-restart replay | Yes — old nonces' timestamps are now outside the window |

Provides full coverage of the threat model. The 16-byte nonce overhead at
200 pps is 3.2 KB/s — negligible against the 192 KB/s PCM rate.

### Option C — Per-session ID + monotonic sequence + max-seen counter

Replace the existing 4-byte `Sequence` with: 8-byte session ID (random per
server start) + 8-byte sequence (monotonic per session). Keep timestamp
window. Receiver tracks `(sessionID, maxSeq)` per session; rejects any
packet with `seq <= maxSeq` for that session.

To tolerate UDP reordering, allow a small reorder window: accept
`seq > maxSeq - reorderWindow` and track a bitmap of recent seen seqs in
that window (RFC 4303 / IPsec ESP anti-replay window, 64-bit bitmap).

| Property | Verdict |
|---|---|
| Server-side state | Per session: 8B sessionID + 8B maxSeq + 8B bitmap = 24 B per known session |
| Packet overhead | +12 bytes (sessionID is the new field; sequence widens 4→8) |
| Multi-client | Each receiver keeps a map of sessions seen; bounded per server lifetime |
| Key rotation | Orthogonal |
| False positives | Reorder beyond the 64-pkt window (= 320 ms at 5 ms framing) gets dropped; today's jitter buffer is 100 ms (20 packets) so this is comfortable |
| Stops single-packet replay | Yes — seq already seen |
| Stops fast replay | Yes |
| Stops post-restart replay | Yes — new sessionID means the old `(sessionID, seq)` pairs are not in any active replay window; the receiver sees the new session start at seq 0 cleanly |
| Bonus | Eliminates the per-restart accidental-sort fragility; also gives `Buffer` a real ordering signal (sessionID + monotonic seq) |

Slightly more bytes than B, but with the upside that the protocol gains a
proper notion of "session" that the rest of the code currently fakes by
chance. State is per-session not per-window, which scales better if pps
rises (e.g. larger frames, more channels).

### Comparison summary

| | A (window) | B (nonce + cache) | C (session + seq) |
|---|---|---|---|
| Stops the threat model fully | No | Yes | Yes |
| Bytes/packet | 0 | +16 | +12 |
| Server state | 0 | O(W × pps) | O(sessions) |
| Reorder-tolerant | n/a | n/a | yes (64-bit window) |
| Adds protocol-level concept | None | None | Session — useful elsewhere |
| Implementation complexity | Trivial | Moderate | Moderate |

## Recommendation

**Option C** — per-session ID + monotonic sequence + 64-bit anti-replay
window, plus the cheap timestamp sanity check from Option A as a
defense-in-depth gate ahead of the per-session bookkeeping.

Justification:

1. It is the only option that gives the receiver a real, durable answer to
   "have I seen this before?" without growing state with traffic rate.
2. The codebase already has a 4-byte `Sequence` field that is **load-bearing
   for ordering** in `Buffer.Push` (`internal/domain/jitter.go:30-32`). C
   strengthens that field rather than ignoring it. After C, the jitter
   buffer can cheaply discard `seq <= lastPlayed` for *all* sources, not
   just the security path.
3. Server-restart replay is the strongest concrete attack and it is the
   one Option A cannot stop. C kills it cleanly via the random session ID.
4. The +12-byte overhead is 1.3% of a 960-byte payload — well within the
   "no per-chunk framing overhead" spirit of AGENTS.md (the constraint is
   about not framing inside the PCM, not about a small constant header).
5. State scales with sessions, not with packet rate. For a homelab with one
   server today and at most a handful tomorrow, this is essentially free.
6. RFC 4303 anti-replay-window is well-trodden; the 64-bit bitmap is ~10
   lines of Go and trivially testable.

## Wire-format change

### Before (v1)

```
offset  size  field
0       4     Sequence            uint32 BE
4       8     Timestamp           uint64 BE (ns since epoch)
12      N     Payload (PCM)       []byte
12+N    32    HMAC-SHA256(0..12+N) (only when auth enabled)
```

Total header: 12 B. Total tag: 32 B when auth on.

### After (v2)

```
offset  size  field
0       1     Version             uint8     (= 0x02)
1       1     Flags               uint8     (reserved, must be 0)
2       2     Reserved            uint16 BE (must be 0; future use)
4       8     SessionID           [8]byte   (random per server start)
12      8     Sequence            uint64 BE (monotonic per session, 0-indexed)
20      8     Timestamp           uint64 BE (ns since epoch, sender clock)
28      M     Payload (PCM)       []byte
28+M    32    HMAC-SHA256(0..28+M) (only when auth enabled)
```

Total header: 28 B (was 12 B; +16 B). Tag: unchanged 32 B.

Go types in `internal/domain/packet.go`:

```go
const (
    HeaderSizeV1 = 12
    HeaderSizeV2 = 28
    ProtocolV1   = 0x01 // implicit; v1 has no version byte
    ProtocolV2   = 0x02
)

type Packet struct {
    Version    uint8     // 1 or 2
    SessionID  [8]byte   // zero in v1
    Sequence   uint64    // widened
    Timestamp  uint64
    Data       []byte
}
```

`Decode` peeks the first byte: if it is `0x02`, parse v2; otherwise treat
the datagram as v1 (legacy path: `Sequence` is read as uint32 from
offset 0, the high 32 bits of the in-memory `Sequence` are 0,
`SessionID` is zero).

### Why these specific bytes

- **Version at byte 0.** Lets the receiver branch before any other parse.
  v1 has no version byte but starts with a uint32 sequence; treating any
  first byte != 0x02 (and != 0x03..0x7F reserved) as v1 is unambiguous as
  long as the sender never produces v2 with version != 0x02.
- **Flags + reserved.** 4-byte alignment for the SessionID; future bits
  for nonce-mode, AES-GCM upgrade, etc., without another wire bump.
- **SessionID 8 bytes / 64 bits.** Random per server start (`crypto/rand`).
  At one restart per minute the birthday collision probability over a year
  is ~1.4e-9; ample.
- **Sequence widens to 64 bits.** Today at 200 pps a 32-bit counter wraps
  in 248 days. With a session-scoped seq we want zero wrap during a
  realistic run. 64-bit costs 4 B and removes the wrap entirely.
- **HMAC still over the entire datagram including version/flags.** No
  unauthenticated header bytes exist.

### Migration story

1. Bump `internal/domain/packet.go` constants and add v2 encode/decode
   alongside v1.
2. Server gains a `--protocol-version` flag, default still `1` for one
   release.
3. Receiver accepts both formats unconditionally — v1 path still has zero
   replay protection (documented), v2 path enforces the full check.
4. Once a release lands with v2 receivers in the field, flip server default
   to `v2` in the next minor release.
5. One release later, remove v1 send path. Keep v1 receive for one more
   release behind a `--allow-legacy` flag, then remove.

This matches AGENTS.md's "wire format is fixed; changes break both sides"
invariant by being explicit about it: v2 is a deliberate break, gated by
an explicit version byte and a phased rollout.

## Implementation phases

Each phase is a single PR.

### Phase 1 — domain types + v2 codec (no behavior change)

- Add `Version`, `SessionID` fields to `domain.Packet`.
- Add `EncodeV2` / `DecodeV2`; keep `Encode` / `Decode` as v1.
- New `domain.ParseAny(data) (*Packet, error)` that branches on first byte.
- Pure additive; no caller wired up yet.
- Tests: round-trip both versions, mixed-version `ParseAny`, malformed
  inputs (truncated header at every offset).

### Phase 2 — anti-replay tracker (server-agnostic)

- New `internal/domain/replay.go`: `type Window struct { ... }` with
  `Check(sessionID [8]byte, seq uint64, nowNs int64) error`.
- 64-bit bitmap per session, evict idle sessions after `2 * replayWindow`
  of silence.
- Returns `ErrReplay`, `ErrStale`, `ErrSessionEvicted`.
- Tests: in-order, reorder within window, reorder beyond window, exact
  duplicate, replay after eviction, stale timestamp, future timestamp,
  multi-session interleave, bitmap edge (seq jumps by 63, 64, 65, 1000).

### Phase 3 — receiver integration

- `udp.Receiver` accepts an optional `*domain.Window`. When set and the
  packet is v2, run the check after HMAC verify and before returning.
- v1 packets pass through unchanged with a one-time slog warning per
  source address.
- New `--replay-window` flag in `cmd/client` (default 2 s).
- Integration test using the existing `internal/testdoubles/`: simulate
  legitimate stream, then re-feed a captured datagram, assert
  `ErrReplay`.

### Phase 4 — sender adopts v2

- `streamingService` generates a session ID at construction
  (`crypto/rand.Read`) and emits v2 packets when configured.
- `cmd/server` flag `--protocol-version` defaults to `1` initially.
- End-to-end test under loopback UDP: server v2 + client v2 + replay
  attempt = rejected; server v1 + client v2 = accepted with warning.

### Phase 5 — flip default + telemetry

- `--protocol-version=2` becomes default.
- Slog metrics (counter): `replay_rejected_total`, `stale_rejected_total`,
  `legacy_v1_received_total`. Already aligned with the slog migration
  pending in the moderate critique.

### Phase 6 — deprecation

- One release after Phase 5: remove v1 send path.
- Two releases after Phase 5: remove v1 receive path or gate behind
  `--allow-legacy`.

## Test plan

### Unit (Phase 2 / `internal/domain/replay_test.go`)

| Case | Setup | Expectation |
|---|---|---|
| In-order baseline | seq 0..99, fresh ts | all accepted |
| Exact duplicate | feed seq 50 twice | second returns `ErrReplay` |
| Reorder within window | 0,1,2,4,3 | 3 accepted (in window) |
| Reorder at edge | window=64, feed 0..64 then 0 | replay |
| Beyond window | feed 0, then 70, then 0 | last is `ErrReplay` (or `ErrTooOld`) |
| Stale timestamp | ts = now - 5s, window = 2s | `ErrStale` |
| Future timestamp | ts = now + 5s | `ErrStale` |
| Multi-session | two sessionIDs, interleaved seq 0..9 | all accepted |
| Session eviction | session A silent for 2W, then re-emits seq 0 | accepted (new bitmap) |
| Bitmap shift correctness | feed 0, 1000 | both accepted; refeed 0 -> `ErrReplay` because still in eviction grace, then `ErrTooOld` after shift |

### Integration (Phase 3 / `internal/adapters/udp/transport_test.go`)

- Loopback UDP socket, real `Sender` + `Receiver` with an `auth.Key`.
- Send 10 packets, capture them via a tee.
- Replay packet[3] after a delay; receiver returns `ErrReplay`.
- Replay a v1-encoded packet against a v2 receiver; passes with one slog
  warning logged.

### Performance

- Benchmark `Window.Check` at 1k, 10k, 100k pps single-session — must be
  sub-microsecond on modern hw (it's a hash-map lookup plus bit ops).
- Bench receive-path under 200 pps continuous + 200 pps replay attempts;
  verify CPU stays flat (no allocation regression vs v1).
- Race-detector run (`go test -race ./...`) over all replay tests; the
  Window is touched from the receive goroutine and the slog logger only.

### Negative / fuzz

- `go test -fuzz=FuzzDecode` over `domain.ParseAny`. Inputs that flip
  version byte, truncate at every header offset, set reserved bytes,
  set session ID to all-zero (allowed) or all-FF (allowed) — none should
  panic.

## Open questions

These need maintainer input before Phase 1 lands:

1. **Replay window size.** 2 s is a defensible default for LAN audio (well
   above any realistic NTP skew, well below the legitimate jitter
   tolerance). Confirm or override.
2. **Reorder window size.** 64 packets = 320 ms at 5 ms framing. Today's
   jitter buffer is `--buf=20` (100 ms). Should the reorder window be
   parameterized, or is 64 a permanent constant?
3. **Eviction policy for idle sessions.** Bound by count (e.g. last 16
   sessions), bound by time (last 2 × window), or both? For a homelab
   with one server this rarely matters, but a parameter to bound state
   is cheap insurance against a malicious source ID-flooding the cache.
4. **v1 deprecation window.** How many releases between Phase 5 default
   flip and Phase 6 v1 removal? Default proposal: one minor release per
   step.
5. **Clock dependency.** The timestamp window assumes server and client
   clocks are within a few hundred ms. Homelab nodes use NTP; should
   the design call out a "clocks must agree to within W/2" precondition
   in `docs/reference/`?
6. **Multi-server (future).** If two servers ever feed the same client
   group, sessionIDs collide with probability ~2^-32 per pair. Worth
   widening to 16 bytes now (one-time wire change cost) or accept the
   risk and bump later behind another version byte? Recommendation:
   accept now (8 B is plenty for the homelab), but the question is
   worth answering on the record.

## References

- `pkg/auth/auth.go` — current HMAC-SHA256 sign/verify.
- `internal/adapters/udp/transport.go` — receive path, no replay check today.
- `internal/domain/packet.go` — wire format, the bytes this plan changes.
- `internal/domain/jitter.go` — consumer of `Sequence`; reorder-window
  numbers chosen to stay strictly larger than `Buffer.minCount`.
- `internal/app/streaming.go` — sender; generates `Sequence` and `Timestamp`.
- RFC 4303 §3.4.3 — IPsec ESP anti-replay window; the 64-bit bitmap pattern
  used in Option C.
