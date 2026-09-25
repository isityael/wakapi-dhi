#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ci="${repo_root}/.woodpecker/ci.yaml"
api="${repo_root}/.woodpecker/api.yaml"

for workflow in "${ci}" "${api}"; do
  [[ "$(yq '.when[] | select(.event == "push") | .branch' "${workflow}")" == "master" ]] || {
    echo "push workflows must run only on master to avoid duplicating pull-request pipelines" >&2
    exit 1
  }
  [[ "$(yq '.depends_on // [] | length' "${workflow}")" == "0" ]] || {
    echo "test workflows must run in parallel; the release gate waits for all of them" >&2
    exit 1
  }
done

for retired in migration.yaml migration-sqlite.yaml validate.yaml; do
  [[ ! -e "${repo_root}/.woodpecker/${retired}" ]] || {
    echo ".woodpecker/${retired} is retired; its checks live in ci.yaml and api.yaml" >&2
    exit 1
  }
done

[[ "$(yq '.services | length' "${api}")" == "1" ]] \
  && [[ "$(yq '.services[0].image' "${api}")" == postgres:* ]] || {
  echo "the API workflow must start exactly one Postgres service (the production dialect)" >&2
  exit 1
}

grep -Fq 'ghcr.io/isityael/wakapi-dhi:latest' "${api}" \
  && grep -Fq 'run_api_tests.sh postgres --migration' "${api}" || {
  echo "upgrade tests must start from the last published fork image on Postgres" >&2
  exit 1
}

[[ "$(yq '.steps[] | select(.name == "api-tests") | .backend_options.kubernetes.resources.limits.memory' "${api}")" == "2Gi" ]] || {
  echo "API workflow memory limits must be nested under Kubernetes resources" >&2
  exit 1
}

[[ "$(yq '.concurrency.limit' "${api}")" == "2" ]] \
  && [[ "$(yq '.concurrency.group' "${api}")" == "wakapi-migration" ]] || {
  echo "API workflow concurrency must be capped" >&2
  exit 1
}

echo "migration workflow resource contract passed"
