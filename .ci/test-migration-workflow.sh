#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
external="${repo_root}/.woodpecker/migration.yaml"
sqlite="${repo_root}/.woodpecker/migration-sqlite.yaml"
ci="${repo_root}/.woodpecker/ci.yaml"

for workflow in "${ci}" "${external}" "${sqlite}"; do
  [[ "$(yq '.when[] | select(.event == "push") | .branch' "${workflow}")" == "master" ]] || {
    echo "push workflows must run only on master to avoid duplicating pull-request pipelines" >&2
    exit 1
  }
done

[[ "$(yq '.matrix.include | length' "${external}")" == "3" ]] || {
  echo "external migration matrix must contain exactly three database variants" >&2
  exit 1
}

[[ "$(yq '.services | length' "${external}")" == "1" ]] || {
  echo "each external migration workflow must start exactly one database service" >&2
  exit 1
}

[[ "$(yq '.services[0].image' "${external}")" == '${DB_IMAGE}' ]] || {
  echo "external migration service image must come from the matrix" >&2
  exit 1
}

[[ -f "${sqlite}" ]] || {
  echo "SQLite migrations must run in a service-free workflow" >&2
  exit 1
}

[[ "$(yq '.services | length' "${sqlite}")" == "0" ]] || {
  echo "SQLite migration workflow must not allocate a database service" >&2
  exit 1
}

echo "migration workflow resource contract passed"
