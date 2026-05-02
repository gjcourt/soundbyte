# architecture/

How the system is built **today** — the current shape of the code.

**Put here:**
- System-overview docs that describe layers and dataflow as they are right now.
- Diagrams and prose that explain the present architecture.

**Do not put here:**
- Proposals for future architecture (e.g., the hex migration target) — `design/`.
- Phased migration sequencing — `plans/`.
- Wire-protocol / API specs — `reference/`.
- Runbooks — `operations/`.

**Naming convention:** `<yyyy-mm-dd>-<topic>.md`
Examples: `2026-05-02-overview.md`.

**Allowed `status:` values:** `Stable`, `Superseded`.

When the architecture changes materially, supersede the existing doc with a new one and set `superseded_by:` on the old one.
