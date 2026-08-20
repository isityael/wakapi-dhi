#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
definition="${repo_root}/dhi/wakapi.yaml"
dockerfile="${repo_root}/Dockerfile"
prepare_definition="${repo_root}/.ci/prepare-dhi-release-definition.sh"
gomod="${repo_root}/go.mod"
pipeline="${repo_root}/.woodpecker/build.yaml"
ci_pipeline="${repo_root}/.woodpecker/ci.yaml"
validate_pipeline="${repo_root}/.woodpecker/validate.yaml"
tag_workflow="${repo_root}/.forgejo/workflows/release-tag.yaml"

grep -Eq 'dhi\.io/golang:1\.26\.[6-9]-alpine3\.24-dev@sha256:' "${definition}" || {
  echo "DHI build must use Go 1.26.6 or newer for the current Go vulnerability fixes" >&2
  exit 1
}

dockerfile_go_reference="$(sed -n 's/^ARG GO_BASE=//p' "${dockerfile}" | head -n 1)"
definition_go_reference="$(sed -n 's/^[[:space:]]*GOLANG_REFERENCE:[[:space:]]*//p' "${definition}" | head -n 1)"
test -n "${dockerfile_go_reference}" \
  && test "${dockerfile_go_reference}" = "${definition_go_reference}" || {
    echo "Dockerfile and native DHI definition must pin the same Go image" >&2
    exit 1
  }

release_fixture="$(mktemp -d)"
trap 'rm -r "${release_fixture}"' EXIT
release_commit="0123456789abcdef0123456789abcdef01234567"
release_definition="${release_fixture}/wakapi.yaml"
"${prepare_definition}" "${definition}" "${release_definition}" "${release_commit}"
grep -Fq "  COMMIT_SHA: ${release_commit}" "${release_definition}" || {
  echo "release definition must build the triggering commit" >&2
  exit 1
}

if "${prepare_definition}" "${definition}" "${release_definition}" "not-a-commit" 2>/dev/null; then
  echo "release definition preparer must reject malformed commit SHAs" >&2
  exit 1
fi

ambiguous_definition="${release_fixture}/ambiguous.yaml"
cp "${definition}" "${ambiguous_definition}"
printf '%s\n' '  COMMIT_SHA: 89abcdef0123456789abcdef0123456789abcdef' >> "${ambiguous_definition}"
if "${prepare_definition}" "${ambiguous_definition}" "${release_definition}" "${release_commit}" 2>/dev/null; then
  echo "release definition preparer must reject ambiguous COMMIT_SHA variables" >&2
  exit 1
fi

if grep -RFn '1.26.4' "${repo_root}/.woodpecker" "${gomod}"; then
  echo "CI and module metadata must not retain the vulnerable Go 1.26.4 toolchain" >&2
  exit 1
fi

grep -Eq 'golang\.org/x/text v0\.(39|[4-9][0-9])\.' "${gomod}" || {
  echo "Wakapi must vendor a golang.org/x/text release containing the CVE-2026-56852 fix" >&2
  exit 1
}

grep -Eq 'golang\.org/x/mod v0\.([4-9][0-9])\.' "${gomod}" || {
  echo "Wakapi must vendor golang.org/x/mod 0.40.0 or newer for the current module verification fixes" >&2
  exit 1
}

grep -Fq 'candidate-${CI_COMMIT_SHA}' "${pipeline}" || {
  echo "release pipeline must publish an isolated candidate tag" >&2
  exit 1
}

if ! grep -Fq '.ci/prepare-dhi-release-definition.sh dhi/wakapi.yaml .ci/wakapi-release.yaml "$${CI_COMMIT_SHA}"' "${pipeline}" \
  || ! grep -Fq 'dockerfile: .ci/wakapi-release.yaml' "${pipeline}"; then
  echo "release pipeline must build the commit-pinned generated definition" >&2
  exit 1
fi

grep -Fq 'name: scan-candidate' "${pipeline}" || {
  echo "release pipeline must scan the candidate before promotion" >&2
  exit 1
}

grep -Fq 'name: promote-release' "${pipeline}" || {
  echo "release pipeline must promote aliases only after the scan passes" >&2
  exit 1
}

grep -Fq 'name: sign-candidate' "${pipeline}" || {
  echo "release pipeline must sign the scanned candidate" >&2
  exit 1
}

grep -Fq 'cosign login ghcr.io' "${pipeline}" || {
  echo "release pipeline must authenticate Cosign before uploading signatures" >&2
  exit 1
}

grep -Eq 'WAKAPI_VERSION=.*dhi/wakapi\.yaml' "${pipeline}" || {
  echo "release aliases must come from the DHI definition" >&2
  exit 1
}

if grep -Fq '"2.17.4-yaelmoshi.2"' "${pipeline}"; then
  echo "release pipeline must not hard-code a stale Wakapi alias" >&2
  exit 1
fi

grep -Fq 'event: tag' "${pipeline}" \
  && grep -Fq 'ci/woodpecker/push/validate' "${tag_workflow}" \
  && grep -Fq 'event: push' "${validate_pipeline}" || {
  echo "release publication must be triggered by a success-gated Forgejo tag" >&2
  exit 1
}

grep -Fq '.ci/test-migration-workflow.sh' "${ci_pipeline}" \
  && grep -Fq '.ci/test-dhi-release-policy.sh' "${ci_pipeline}" || {
    echo "CI must enforce the checked-in workflow policy tests" >&2
    exit 1
  }

echo "DHI release security policy passed"
