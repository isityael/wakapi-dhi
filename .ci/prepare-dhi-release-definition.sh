#!/bin/sh
set -eu

source_definition="$1"
output_definition="$2"
release_commit="$3"

case "${release_commit}" in
  *[!0-9a-f]*|'')
    echo "release commit must be a lowercase hexadecimal SHA" >&2
    exit 1
    ;;
esac

if [ "${#release_commit}" -ne 40 ]; then
  echo "release commit must contain exactly 40 hexadecimal characters" >&2
  exit 1
fi

if [ "$(grep -c '^  COMMIT_SHA:' "${source_definition}")" -ne 1 ]; then
  echo "source definition must contain exactly one COMMIT_SHA variable" >&2
  exit 1
fi

sed "s/^  COMMIT_SHA:.*/  COMMIT_SHA: ${release_commit}/" "${source_definition}" > "${output_definition}"
