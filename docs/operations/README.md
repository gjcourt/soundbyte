# operations/

Runbooks, smoke tests, and on-call / day-to-day operating procedures.

**Put here:**
- How to run the server and client, hook up `librespot`, verify roundtrip.
- Step-by-step procedures for known failure modes (jitter buffer underrun, client unable to reach server).

**Do not put here:**
- Wire-protocol or API specs — `reference/`.
- Architecture explanation — `architecture/`.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-running-locally.md`, `2026-09-01-spotify-connect-setup.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

Stale runbooks are dangerous. When a procedure changes, update the doc in the same PR.
