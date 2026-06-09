# Native DHI Distribution Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Convert `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi` to a native DHI YAML image build published by Woodpecker to public GHCR, then refresh the public OCI Helm chart and infra wrapper pins.

**Architecture:** The Wakapi image build source of truth moves from `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/Dockerfile` to `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/dhi/wakapi.yaml`. Woodpecker builds that DHI definition with `docker buildx build -f dhi/wakapi.yaml`, publishes public tags to `ghcr.io/yaelmoshi/wakapi-dhi`, then the chart in `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi` pins the tested image digest while `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi` consumes the refreshed chart without runtime renames.

**Tech Stack:** Docker Hardened Images YAML frontend (`# syntax=dhi.io/build:2-alpine3.23`), Docker Buildx, Woodpecker CI, GHCR, Go 1.26.4, Helm OCI charts, Kubernetes security contexts.

---

## File Structure

- Create `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/dhi/wakapi.yaml`: native DHI image definition.
- Create `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/scripts/validate_dhi_image.sh`: local smoke test for the produced image.
- Modify `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml`: switch publish input from `Dockerfile` to the DHI YAML definition.
- Modify `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/compose.yml`: bump DHI Postgres image from 17 to 18.
- Modify `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/testing/compose.yml`: bump DHI database test images.
- Modify `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/README.md`: document GHCR-hosted native DHI image build.
- Modify `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/Chart.yaml`: bump chart version and `appVersion` after the image publish.
- Modify `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/values.yaml`: pin the new GHCR image digest.
- Modify `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/README.md`: document DHI-built image and GHCR OCI chart.
- Modify `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.yaml`: update chart dependency after publishing.
- Regenerate `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.lock`.
- Regenerate `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/charts/wakapi-dhi-1.2.10.tgz`.
- Modify `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values.yaml`, `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-root.yaml`, and `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-homelab.yaml` only if the tested image tag or digest needs a deployment pin update.

## Task 1: Add the Native DHI Image Definition

**Files:**
- Create: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/dhi/wakapi.yaml`
- Test: local DHI build command in this task

- [ ] **Step 1: Create the DHI definition**

Use `apply_patch` to add this file:

```yaml
# syntax=dhi.io/build:2-alpine3.23

name: Wakapi DHI
image: ghcr.io/yaelmoshi/wakapi-dhi
variant: runtime
tags:
  - 2.17.3-yaelmoshi.2
  - latest
platforms:
  - linux/amd64
dates:
  release: "2026-06-09"
vars:
  WAKAPI_VERSION: 2.17.3-yaelmoshi.2
  GO_VERSION: 1.26.4
contents:
  repositories:
    - https://dhi.io/apk/alpine/v3.23/main
    - https://dl-cdn.alpinelinux.org/alpine/v3.23/main
    - https://dl-cdn.alpinelinux.org/alpine/v3.23/community
  keyring:
    - https://dhi.io/keyring/dhi-apk@docker-0F81AD7700D99184.rsa.pub
  packages:
    - alpine-baselayout-data
    - busybox
    - ca-certificates-bundle
    - tzdata
  builds:
    - name: wakapi
      contents:
        repositories:
          - https://dhi.io/apk/alpine/v3.23/main
          - https://dl-cdn.alpinelinux.org/alpine/v3.23/main
          - https://dl-cdn.alpinelinux.org/alpine/v3.23/community
        keyring:
          - https://dhi.io/keyring/dhi-apk@docker-0F81AD7700D99184.rsa.pub
        packages:
          - alpine-baselayout-data
          - busybox
          - ca-certificates-bundle
          - git
          - golang-1.26=1.26.4-r0
          - tzdata
      pipeline:
        - name: build
          runs: |
            set -eux -o pipefail

            mkdir -p /src /staging/app /staging/data
            cp -a "$BUILDKIT_CONTEXT_KEEP_GIT_DIR"/. /src/ 2>/dev/null || cp -a /workspace/. /src/
            cd /src

            go mod download
            GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} CGO_ENABLED=0 GOEXPERIMENT=jsonv2 go build -ldflags "-s -w" -v -o /staging/app/wakapi main.go
            GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} CGO_ENABLED=0 go build -ldflags "-s -w" -v -o /staging/app/healthcheck scripts/healthcheck.go

            cp config.default.yml /staging/app/config.yml
            sed -i 's/listen_ipv6: ::1/listen_ipv6: "-"/g' /staging/app/config.yml
            chmod 0555 /staging/app/wakapi /staging/app/healthcheck
            chmod 0444 /staging/app/config.yml
            chown -R 65532:65532 /staging/app /staging/data
      outputs:
        - source: /staging/app
          target: /app
          uid: 65532
          gid: 65532
        - source: /staging/data
          target: /data
          uid: 65532
          gid: 65532
accounts:
  run-as: nonroot
  users:
    - name: nonroot
      uid: 65532
      gid: 65532
  groups:
    - name: nonroot
      gid: 65532
      members:
        - nonroot
os-release:
  name: Docker Hardened Images (Alpine)
  id: alpine
  version-id: "3.23"
  pretty-name: Docker Hardened Images/Alpine Linux v3.23
  home-url: https://docker.com/products/hardened-images/
  bug-report-url: https://docker.com/support/
work-dir: /app
environment:
  ENVIRONMENT: prod
  SSL_CERT_FILE: /etc/ssl/certs/ca-certificates.crt
  WAKAPI_ALLOW_SIGNUP: "true"
  WAKAPI_DB_HOST: ""
  WAKAPI_DB_NAME: /data/wakapi.db
  WAKAPI_DB_PASSWORD: ""
  WAKAPI_DB_TYPE: sqlite3
  WAKAPI_DB_USER: ""
  WAKAPI_INSECURE_COOKIES: "true"
  WAKAPI_LISTEN_IPV4: 0.0.0.0
  WAKAPI_PASSWORD_SALT: ""
annotations:
  org.opencontainers.image.url: https://github.com/yaelmoshi/wakapi-dhi
  org.opencontainers.image.documentation: https://github.com/muety/wakapi
  org.opencontainers.image.source: https://github.com/yaelmoshi/wakapi-dhi
  org.opencontainers.image.title: Wakapi (DHI-native)
  org.opencontainers.image.licenses: MIT
  org.opencontainers.image.description: Wakapi native DHI-built distribution published on GHCR
entrypoint:
  - /app/wakapi
ports:
  - 3000/tcp
volumes:
  - /data
```

- [ ] **Step 2: Run the first local DHI build**

Run:

```bash
docker buildx build . \
  -f dhi/wakapi.yaml \
  --sbom=generator=dhi.io/scout-sbom-indexer:1 \
  --provenance=1 \
  --tag ghcr.io/yaelmoshi/wakapi-dhi:dhi-local \
  --load
```

Expected: build succeeds and local image `ghcr.io/yaelmoshi/wakapi-dhi:dhi-local` exists.

- [ ] **Step 3: If the build context copy path fails, patch the build stage**

If Step 2 fails because neither `$BUILDKIT_CONTEXT_KEEP_GIT_DIR` nor `/workspace` exists in the DHI build stage, replace the source copy block in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/dhi/wakapi.yaml` with this tar-based fallback:

```yaml
            mkdir -p /src /staging/app /staging/data
            tar -C /src -xf /context.tar
            cd /src
```

Then rebuild with:

```bash
tar --exclude='.git' --exclude='node_modules' --exclude='coverage' -cf /tmp/wakapi-dhi-context.tar .
docker buildx build . \
  -f dhi/wakapi.yaml \
  --build-context context.tar=/tmp/wakapi-dhi-context.tar \
  --sbom=generator=dhi.io/scout-sbom-indexer:1 \
  --provenance=1 \
  --tag ghcr.io/yaelmoshi/wakapi-dhi:dhi-local \
  --load
```

Expected: build succeeds. Keep the working source staging method in the YAML and do not keep both code paths.

- [ ] **Step 4: Commit the DHI definition**

Run:

```bash
git add dhi/wakapi.yaml
git commit -m "build: add native DHI image definition"
```

Expected: commit succeeds.

## Task 2: Add Local Runtime Image Validation

**Files:**
- Create: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/scripts/validate_dhi_image.sh`
- Test: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/scripts/validate_dhi_image.sh ghcr.io/yaelmoshi/wakapi-dhi:dhi-local`

- [ ] **Step 1: Add the validation script**

Use `apply_patch` to add:

```bash
#!/usr/bin/env bash
set -o errexit -o nounset -o pipefail

image="${1:-ghcr.io/yaelmoshi/wakapi-dhi:dhi-local}"
container="wakapi-dhi-validate"

cleanup() {
  docker rm -f "$container" >/dev/null 2>&1 || true
}
trap cleanup EXIT

cleanup

docker run -d \
  --name "$container" \
  -p 3000:3000 \
  -e WAKAPI_PASSWORD_SALT=validation-salt \
  -e WAKAPI_DB_TYPE=sqlite3 \
  -e WAKAPI_DB_NAME=/data/wakapi.db \
  "$image" >/dev/null

for _ in $(seq 1 60); do
  if curl --fail --silent --output /dev/null http://127.0.0.1:3000/api/health; then
    docker exec "$container" /app/healthcheck
    exit 0
  fi
  sleep 1
done

docker logs "$container"
exit 1
```

- [ ] **Step 2: Make it executable**

Run:

```bash
chmod +x scripts/validate_dhi_image.sh
```

Expected: script is executable.

- [ ] **Step 3: Run image validation**

Run:

```bash
scripts/validate_dhi_image.sh ghcr.io/yaelmoshi/wakapi-dhi:dhi-local
```

Expected: command exits `0`; `/api/health` and `/app/healthcheck` both succeed.

- [ ] **Step 4: Commit the validation script**

Run:

```bash
git add scripts/validate_dhi_image.sh
git commit -m "test: add DHI image smoke validation"
```

Expected: commit succeeds.

## Task 3: Refresh DHI Database Images Used by Compose Tests

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/compose.yml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/testing/compose.yml`
- Test: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/testing/run_api_tests.sh`

- [ ] **Step 1: Patch the Compose images**

Use `apply_patch` to make these replacements:

```diff
--- a/compose.yml
+++ b/compose.yml
@@
-    image: dhi.io/postgres:17-debian13
+    image: dhi.io/postgres:18-debian13
--- a/testing/compose.yml
+++ b/testing/compose.yml
@@
-    image: dhi.io/postgres:17-debian13
+    image: dhi.io/postgres:18-debian13
@@
-    image: dhi.io/mysql:8
+    image: dhi.io/mysql:9-debian13
@@
-    image: dhi.io/mariadb:11
+    image: dhi.io/mariadb:12-debian13
```

- [ ] **Step 2: Validate Compose syntax**

Run:

```bash
docker compose -f compose.yml config >/tmp/wakapi-compose.rendered.yml
docker compose -f testing/compose.yml config >/tmp/wakapi-testing-compose.rendered.yml
```

Expected: both commands exit `0`.

- [ ] **Step 3: Run database-backed startup tests**

Run:

```bash
testing/run_api_tests.sh postgres
testing/run_api_tests.sh mysql
testing/run_api_tests.sh mariadb
```

Expected: each script starts the selected database image, starts Wakapi, waits for `/api/health`, and exits `0`.

- [ ] **Step 4: Run the SQLite API collection if Bruno is installed**

Run:

```bash
testing/run_api_tests.sh sqlite
```

Expected: if `bru` is installed, the Bruno API collection passes. If `bru` is not installed, record that the SQLite API collection was skipped because `testing/run_api_tests.sh` requires `bru`.

- [ ] **Step 5: Commit the Compose image refresh**

Run:

```bash
git add compose.yml testing/compose.yml
git commit -m "test: refresh DHI database images"
```

Expected: commit succeeds.

## Task 4: Migrate Woodpecker to the Native DHI Build

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml`
- Test: `docker buildx build` local command and Woodpecker pipeline after push

- [ ] **Step 1: Patch Woodpecker path filters and build step**

Use `apply_patch` to update `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml` so the push include list watches `dhi/**` and the build settings use `dhi/wakapi.yaml`:

```yaml
# Wakapi - DHI-native fork
# Builds and pushes to ghcr.io/yaelmoshi/wakapi-dhi
when:
  - event: manual
  - event: push
    branch: master
    path:
      include: ["**/*.go", "go.*", "dhi/**", ".woodpecker/**"]

steps:
  - name: build-and-push-ghcr
    image: woodpeckerci/plugin-docker-buildx:6.0.4@sha256:765ebfaa6f71383197babb2352c7902bafe141e6232a62cafb740510a788c550
    privileged: true
    settings:
      registry: ghcr.io
      repo: ghcr.io/yaelmoshi/wakapi-dhi
      dockerfile: dhi/wakapi.yaml
      platforms: linux/amd64
      tags:
        - "${CI_COMMIT_SHA:0:8}"
        - "2.17.3-yaelmoshi.2"
        - latest
      cache_images:
        - ghcr.io/yaelmoshi/wakapi-dhi:buildcache
      logins:
        - registry: ghcr.io
          username: yaelmoshi
          password:
            from_secret: github_token
        - registry: dhi.io
          username:
            from_secret: dhi_username
          password:
            from_secret: dhi_password
      provenance: mode=max
      buildkit_driveropt: network=host
```

- [ ] **Step 2: Validate the local equivalent build command**

Run:

```bash
docker buildx build . \
  -f dhi/wakapi.yaml \
  --sbom=generator=dhi.io/scout-sbom-indexer:1 \
  --provenance=1 \
  --tag ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2-local \
  --load
```

Expected: build succeeds.

- [ ] **Step 3: Smoke-test the locally built CI-equivalent image**

Run:

```bash
scripts/validate_dhi_image.sh ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2-local
```

Expected: command exits `0`.

- [ ] **Step 4: Commit the Woodpecker migration**

Run:

```bash
git add .woodpecker/build.yaml
git commit -m "ci: build native DHI image in Woodpecker"
```

Expected: commit succeeds.

## Task 5: Run App-Level Verification Before Publish

**Files:**
- Test only

- [ ] **Step 1: Run Go tests**

Run:

```bash
go test ./...
```

Expected: all packages pass.

- [ ] **Step 2: Run scripts module tests or compile check**

Run:

```bash
(cd scripts && go test ./...)
```

Expected: all packages pass or no test packages fail.

- [ ] **Step 3: Validate Docker Compose files**

Run:

```bash
docker compose -f compose.yml config >/tmp/wakapi-compose.rendered.yml
docker compose -f testing/compose.yml config >/tmp/wakapi-testing-compose.rendered.yml
```

Expected: both commands exit `0`.

- [ ] **Step 4: Check whitespace**

Run:

```bash
git diff --check
```

Expected: no whitespace errors.

- [ ] **Step 5: Push the Wakapi DHI image migration**

Run:

```bash
git push origin master
git push github master
```

Expected: both remotes accept the commits.

- [ ] **Step 6: Watch Woodpecker publish the image**

Run:

```bash
woodpecker-cli repo ps m0sh1/wakapi-dhi --output table
```

Expected: a new pipeline for the pushed commit appears and reaches `success`.

- [ ] **Step 7: Capture the published image digest**

Run:

```bash
docker buildx imagetools inspect ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2
docker buildx imagetools inspect ghcr.io/yaelmoshi/wakapi-dhi:$(git rev-parse --short=8 HEAD)
```

Expected: both references exist. Record the `Digest: sha256:...` for the semantic tag.

## Task 6: Refresh the Public OCI Helm Chart

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/Chart.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/values.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/README.md`
- Test: Helm lint and template commands

- [ ] **Step 1: Patch chart metadata**

In `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/Chart.yaml`, make these changes:

```diff
-version: 1.2.9
-appVersion: "2.17.3-yaelmoshi.1"
+version: 1.2.10
+appVersion: "2.17.3-yaelmoshi.2"
```

- [ ] **Step 2: Capture the published image digest**

Run from `/Users/yaelmeya/git/m0sh1.cc/helm-charts`:

```bash
native_dhi_digest="$(docker buildx imagetools inspect ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2 | awk '/^Digest:/ {print $2; exit}')"
test -n "$native_dhi_digest"
printf '%s\n' "$native_dhi_digest"
```

Expected: prints a value beginning with `sha256:`.

- [ ] **Step 3: Patch the default image digest**

In `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/values.yaml`, replace the old digest pin:

```yaml
image:
  repository: ghcr.io/yaelmoshi/wakapi-dhi
  pullPolicy: IfNotPresent
  tag: "2.17.3-yaelmoshi.2@${native_dhi_digest}"
```

Use the digest captured in Step 2. The final committed YAML must contain the literal digest, not the shell variable.

- [ ] **Step 4: Patch README wording**

Update `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi/README.md` so the opening says:

```markdown
Public OCI Helm chart for the [yaelmoshi/wakapi-dhi](https://github.com/yaelmoshi/wakapi-dhi) native DHI-built Wakapi distribution.

The default image is built from a native Docker Hardened Images YAML definition in the Wakapi fork and published publicly to GHCR. The chart itself is also published as a public OCI artifact on GHCR.
```

Also update install examples from `1.2.9` to `1.2.10` and image examples from `2.17.3-yaelmoshi.1` to `2.17.3-yaelmoshi.2` with the digest captured in Step 2.

- [ ] **Step 5: Run chart validation**

Run from `/Users/yaelmeya/git/m0sh1.cc/helm-charts`:

```bash
helm lint charts/wakapi-dhi
helm template wakapi charts/wakapi-dhi -f charts/wakapi-dhi/ci/test-values.yaml >/tmp/wakapi-dhi-chart.yaml
rg -n "name: wakapi$|ghcr.io/yaelmoshi/wakapi-dhi|runAsNonRoot|allowPrivilegeEscalation|seccompProfile" /tmp/wakapi-dhi-chart.yaml
git diff --check
```

Expected: lint succeeds, template succeeds, rendered resource names remain `wakapi`, rendered image references the new DHI-built GHCR image, and security context remains strict.

- [ ] **Step 6: Commit the chart refresh**

Run:

```bash
git add charts/wakapi-dhi/Chart.yaml charts/wakapi-dhi/values.yaml charts/wakapi-dhi/README.md
git commit -m "chore(wakapi): publish native DHI chart defaults"
```

Expected: commit succeeds.

- [ ] **Step 7: Publish the OCI chart**

Run:

```bash
helm package charts/wakapi-dhi --destination /tmp
helm registry login ghcr.io -u yaelmoshi
helm push /tmp/wakapi-dhi-1.2.10.tgz oci://ghcr.io/yaelmoshi/charts
```

Expected: `ghcr.io/yaelmoshi/charts/wakapi-dhi:1.2.10` is pushed successfully.

- [ ] **Step 8: Verify public chart pull**

Run:

```bash
helm pull oci://ghcr.io/yaelmoshi/charts/wakapi-dhi --version 1.2.10 --destination /tmp
```

Expected: pull succeeds.

- [ ] **Step 9: Push helm-charts changes**

Run:

```bash
git push origin main
git push github main
```

Expected: both remotes accept the chart commit.

## Task 7: Refresh the Infra Wrapper

**Files:**
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.lock`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/charts/wakapi-dhi-1.2.10.tgz`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-root.yaml`
- Modify: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-homelab.yaml`
- Test: wrapper lint and template commands

- [ ] **Step 1: Patch wrapper chart metadata**

In `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.yaml`, make these changes:

```diff
-version: 0.1.17
+version: 0.1.18
 appVersion: "2.17.3-yaelmoshi.1"
+appVersion: "2.17.3-yaelmoshi.2"
@@
   - name: wakapi-dhi
-    version: "1.2.9"
+    version: "1.2.10"
```

If the patch leaves two `appVersion` lines, keep only:

```yaml
appVersion: "2.17.3-yaelmoshi.2"
```

- [ ] **Step 2: Capture the published image digest**

Run from `/Users/yaelmeya/git/m0sh1.cc/infra`:

```bash
native_dhi_digest="$(docker buildx imagetools inspect ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2 | awk '/^Digest:/ {print $2; exit}')"
test -n "$native_dhi_digest"
printf '%s\n' "$native_dhi_digest"
```

Expected: prints a value beginning with `sha256:`.

- [ ] **Step 3: Patch wrapper image pins**

In each of these files:

- `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values.yaml`
- `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-root.yaml`
- `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/values-homelab.yaml`

set:

```yaml
wakapi:
  image:
    repository: ghcr.io/yaelmoshi/wakapi-dhi
    tag: "2.17.3-yaelmoshi.2@${native_dhi_digest}"
    pullPolicy: IfNotPresent
```

Use the digest captured in Step 2. The final committed YAML files must contain the literal digest, not the shell variable.

- [ ] **Step 4: Refresh Helm dependencies**

Run from `/Users/yaelmeya/git/m0sh1.cc/infra`:

```bash
helm dependency update apps/user/wakapi
```

Expected: `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/Chart.lock` references `wakapi-dhi` version `1.2.10`, and `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi/charts/wakapi-dhi-1.2.10.tgz` exists.

- [ ] **Step 5: Run wrapper validation**

Run from `/Users/yaelmeya/git/m0sh1.cc/infra`:

```bash
helm lint apps/user/wakapi
helm template wakapi apps/user/wakapi -f apps/user/wakapi/values.yaml >/tmp/wakapi-wrapper-default.yaml
helm template wakapi apps/user/wakapi -f apps/user/wakapi/values-root.yaml >/tmp/wakapi-wrapper-root.yaml
helm template wakapi apps/user/wakapi -f apps/user/wakapi/values-homelab.yaml >/tmp/wakapi-wrapper-homelab.yaml
rg -n "name: wakapi$|ghcr.io/yaelmoshi/wakapi-dhi|2.17.3-yaelmoshi.2|runAsNonRoot|seccompProfile" /tmp/wakapi-wrapper-default.yaml /tmp/wakapi-wrapper-root.yaml /tmp/wakapi-wrapper-homelab.yaml
git diff --check
```

Expected: lint and templates succeed, image pin is rendered, runtime object names remain `wakapi`, and security context remains strict.

- [ ] **Step 6: Commit wrapper refresh**

Run:

```bash
git add apps/user/wakapi/Chart.yaml apps/user/wakapi/Chart.lock apps/user/wakapi/charts apps/user/wakapi/values.yaml apps/user/wakapi/values-root.yaml apps/user/wakapi/values-homelab.yaml
git commit -m "chore: deploy native DHI wakapi image"
```

Expected: commit succeeds without staging unrelated dirty files such as `/Users/yaelmeya/git/m0sh1.cc/infra/apps/cluster/valkey/Chart.yaml` or `/Users/yaelmeya/git/m0sh1.cc/infra/tools/cli`.

- [ ] **Step 7: Push infra changes**

Run:

```bash
git push origin main
```

Expected: Forgejo accepts the commit.

## Task 8: Final Verification and Memory Update

**Files:**
- No code files modified unless Basic Memory MCP note is updated
- Test: repo statuses, public artifact pulls, optional read-only ArgoCD checks

- [ ] **Step 1: Verify all three repos are synced**

Run:

```bash
git -C /Users/yaelmeya/git/m0sh1.cc/wakapi-dhi status --short
git -C /Users/yaelmeya/git/m0sh1.cc/helm-charts status --short
git -C /Users/yaelmeya/git/m0sh1.cc/infra status --short
```

Expected: Wakapi and helm-charts are clean. Infra may still show unrelated pre-existing dirt, but no Wakapi-related files should be uncommitted.

- [ ] **Step 2: Verify public image and chart pulls**

Run:

```bash
docker buildx imagetools inspect ghcr.io/yaelmoshi/wakapi-dhi:2.17.3-yaelmoshi.2
helm pull oci://ghcr.io/yaelmoshi/charts/wakapi-dhi --version 1.2.10 --destination /tmp
```

Expected: both public pulls succeed.

- [ ] **Step 3: Optional read-only ArgoCD verification**

Run only read-only or explicitly allowed reconciliation commands:

```bash
argocd app get wakapi --grpc-web
argocd app diff wakapi --grpc-web
```

Expected: command output is collected for reporting. Do not use `--prune` or `--force`.

- [ ] **Step 4: Update the existing Basic Memory MCP note**

Search Basic Memory MCP project `main` for the existing Wakapi DHI note and update it with:

- native DHI YAML image definition path
- GHCR public image tag and digest
- OCI chart version and digest
- infra wrapper version
- validation commands and results
- any DHI build pitfalls found during implementation

Expected: existing note is updated rather than duplicating a second note.

## Self-Review

- Spec coverage: the plan covers native DHI image definition, Woodpecker build authority, public GHCR image, public GHCR OCI chart, digest pinning, infra wrapper refresh, and no runtime rename.
- Red-flag scan: the plan contains no unresolved committed-file markers; digest values are captured with shell variables and must be written as literal digests before committing chart or infra files.
- Type consistency: version string is consistently `2.17.3-yaelmoshi.2`; chart version is consistently `1.2.10`; infra wrapper version is consistently `0.1.18`.
