#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
definition="${repo_root}/dhi/wakapi.yaml"
gomod="${repo_root}/go.mod"
pipeline="${repo_root}/.woodpecker/build.yaml"

grep -Eq 'dhi\.io/golang:1\.26\.[5-9]-alpine3\.24-dev@sha256:' "${definition}" || {
  echo "DHI build must use a Go release containing the CVE-2026-39822 fix" >&2
  exit 1
}

if grep -RFn '1.26.4' "${repo_root}/.woodpecker" "${gomod}"; then
  echo "CI and module metadata must not retain the vulnerable Go 1.26.4 toolchain" >&2
  exit 1
fi

grep -Eq 'golang\.org/x/text v0\.(39|[4-9][0-9])\.' "${gomod}" || {
  echo "Wakapi must vendor a golang.org/x/text release containing the CVE-2026-56852 fix" >&2
  exit 1
}

grep -Fq 'candidate-${CI_COMMIT_SHA}' "${pipeline}" || {
  echo "release pipeline must publish an isolated candidate tag" >&2
  exit 1
}

grep -Fq 'COMMIT_SHA: "${CI_COMMIT_SHA}"' "${pipeline}" || {
  echo "release pipeline must build the commit that triggered the pipeline" >&2
  exit 1
}

grep -Fq 'name: scan-candidate' "${pipeline}" || {
  echo "release pipeline must scan the candidate before promotion" >&2
  exit 1
}

grep -Fq 'name: promote-release' "${pipeline}" || {
  echo "release pipeline must promote aliases only after the scan passes" >&2
  exit 1
}

echo "DHI release security policy passed"
