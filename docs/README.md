# SoundByte Documentation

SoundByte streams audio over UDP as raw PCM chunks. This folder is organized into a fixed six-folder taxonomy. Each folder's `README.md` describes what belongs there.

## Folders

- [`architecture/`](architecture/README.md) — how the system is built today.
- [`design/`](design/README.md) — proposals, RFCs, in-flight or recently shipped designs.
- [`operations/`](operations/README.md) — runbooks, smoke tests, dev setup.
- [`plans/`](plans/README.md) — phased migrations, rollout sequencing.
- [`reference/`](reference/README.md) — wire protocol, auth, configuration.
- [`research/`](research/README.md) — spikes, investigations.

## Quick links

- **Architecture overview** → [`architecture/2026-05-02-overview.md`](architecture/2026-05-02-overview.md).
- **Wire protocol / API** → [`reference/2026-05-02-api.md`](reference/2026-05-02-api.md).
- **Authentication** → [`reference/2026-05-02-authentication.md`](reference/2026-05-02-authentication.md).

## Conventions

- All non-index docs use frontmatter (`title`, `status`, `created`, `updated`, `updated_by`, `tags`).
- Filenames carry topic and creation date (`<yyyy-mm-dd>-<topic>.md`); state lives in `status:` frontmatter, never in the filename.
- See `AGENTS.md` for the full convention.
