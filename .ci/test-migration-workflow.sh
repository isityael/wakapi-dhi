#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
external="${repo_root}/.woodpecker/migration.yaml"
sqlite="${repo_root}/.woodpecker/migration-sqlite.yaml"
ci="${repo_root}/.woodpecker/ci.yaml"
validate="${repo_root}/.woodpecker/validate.yaml"

for workflow in "${ci}" "${external}" "${sqlite}"; do
  [[ "$(yq '.when[] | select(.event == "push") | .branch' "${workflow}")" == "master" ]] || {
    echo "push workflows must run only on master to avoid duplicating pull-request pipelines" >&2
    exit 1
  }
done

[[ "$(yq '.matrix' "${external}")" == "null" ]] || {
  echo "external migrations run against Postgres only; MySQL/MariaDB are unsupported in this fork" >&2
  exit 1
}

[[ "$(yq '.services | length' "${external}")" == "1" ]] || {
  echo "each external migration workflow must start exactly one database service" >&2
  exit 1
}

[[ "$(yq '.services[0].image' "${external}")" == postgres:* ]] || {
  echo "external migration service must be a Postgres image" >&2
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

[[ "$(yq '.depends_on[]' "${sqlite}")" == "ci" ]] \
  && [[ "$(yq '.depends_on[]' "${external}")" == "migration-sqlite" ]] \
  && [[ "$(yq '.depends_on[]' "${validate}")" == "migration" ]] || {
  echo "migration workflows must be staged to stay within the shared runner quota" >&2
  exit 1
}

[[ "$(yq '.steps[0].backend_options.kubernetes.resources.limits.memory' "${sqlite}")" == "2Gi" ]] || {
  echo "SQLite workflow memory limits must be nested under Kubernetes resources" >&2
  exit 1
}

[[ "$(yq '.concurrency.limit' "${external}")" == "2" ]] \
  && [[ "$(yq '.concurrency.group' "${external}")" == "wakapi-migration" ]] || {
  echo "external migration concurrency must be capped" >&2
  exit 1
}

echo "migration workflow resource contract passed"
