#!/usr/bin/env bash
set -euo pipefail

workflow=".github/workflows/release.yml"

require_marker() {
    local marker="$1"
    local message="$2"

    if ! grep -Fq "$marker" "$workflow"; then
        echo "$message" >&2
        exit 1
    fi
}

require_marker "contents: write" "release workflow must have contents: write so it can create GitHub releases"
require_marker "ref: \${{ github.event_name == 'workflow_dispatch' && inputs.tag || github.ref }}" "manual releases must check out the selected tag"
require_marker "Release tag must be valid v2 semantic versioning" "release workflow must validate v2 semantic-version tags"
require_marker "X-Fern-SDK-Version" "release workflow must match runtime metadata to the release tag"
require_marker "go mod verify" "release workflow must verify module dependencies"
require_marker "go build ./..." "release workflow must build the tagged SDK"
require_marker "go test ./..." "release workflow must test the tagged SDK"
require_marker "No changelog notes found" "release workflow must reject an empty changelog section"
require_marker "gh release create" "release workflow must create a GitHub release when one does not exist"
require_marker "gh release edit" "release workflow must update an existing GitHub release"
require_marker "release_notes.md" "release workflow must generate and use a release notes file"

echo "Release workflow checks passed."
