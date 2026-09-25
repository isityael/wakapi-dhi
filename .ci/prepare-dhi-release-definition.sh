#!/bin/sh
set -eu

source_definition="$1"
output_definition="$2"
release_commit="$3"
# Optional: full release version from the tag, e.g. 2.18.0-yael.3.
release_version="${4:-}"

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

if [ -z "${release_version}" ]; then
  sed "s/^  COMMIT_SHA:.*/  COMMIT_SHA: ${release_commit}/" "${source_definition}" > "${output_definition}"
  exit 0
fi

base_version="$(sed -n 's/^  WAKAPI_VERSION:[[:space:]]*//p' "${source_definition}")"
case "${release_version}" in
  "${base_version}-yael."*) build="${release_version#"${base_version}-yael."}" ;;
  *) build="" ;;
esac
case "${build}" in
  *[!0-9]*|'')
    echo "release version must be ${base_version}-yael.<n>, got ${release_version}" >&2
    exit 1
    ;;
esac

sed -e "s/^  COMMIT_SHA:.*/  COMMIT_SHA: ${release_commit}/" \
  -e "s/^  WAKAPI_VERSION:.*/  WAKAPI_VERSION: ${release_version}/" \
  "${source_definition}" > "${output_definition}"
