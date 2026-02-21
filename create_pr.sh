#!/bin/bash
set -e

cd /Users/george/src/soundbyte

git checkout -b feat/multi-arch-build || git checkout feat/multi-arch-build
git add Makefile scripts/build_and_push_image.sh
git commit -m "feat: add multi-arch build support and list-images command" || true
git push -u origin feat/multi-arch-build

cat << 'EOF' > pr_body.md
## What changed
- Updated `Makefile` to set `PLATFORM ?= linux/amd64,linux/arm64`
- Added `list-images` command to `Makefile` using `gh api`
- Updated `scripts/build_and_push_image.sh` to use `docker buildx build --platform`

## Why
Fixes an issue where pulling the image on an `amd64` platform failed with `no match for platform in manifest: not found`. The image is now built for both `amd64` and `arm64`. The `list-images` command provides an easy way to view available tags in `ghcr.io`.

## Notes
Requires `gh` CLI authenticated with `read:packages` scope to use `make list-images`.

## Checklist
- [x] `golangci-lint run ./...` passes with 0 issues
- [x] `go test -race ./...` passes
- [x] Documentation updated (if applicable)
EOF

gh pr create --title "feat: add multi-arch build support and list-images command" --body-file pr_body.md || true
rm pr_body.md
