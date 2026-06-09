# Native DHI Wakapi Distribution Design

## Summary

This design converts `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi` from a DHI-based Dockerfile build into a native DHI-built distribution, while keeping public distribution on GitHub Container Registry.

The OCI image remains public at `ghcr.io/yaelmoshi/wakapi-dhi`.
The OCI Helm chart remains public at `ghcr.io/yaelmoshi/charts/wakapi-dhi`.
The source of truth for the image build becomes a native DHI definition rather than `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/Dockerfile`.

## Goals

- Make `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi` a native DHI-built project.
- Keep public artifact delivery on `ghcr.io`.
- Keep Forgejo and Woodpecker as the canonical build and publish path.
- Keep `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi` as the public OCI chart package.
- Keep runtime Kubernetes object names stable as `wakapi`.
- Align the image and chart with DHI best practices where they fit the current repository model.

## Non-Goals

- Publishing customized images or charts back to `dhi.io`.
- Renaming the deployed Kubernetes application from `wakapi` to `wakapi-dhi`.
- Restructuring `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi`.
- Introducing cluster writes outside the existing GitOps flow.
- Reworking the Wakapi application itself beyond what is required for the new image build path.

## Constraints

- There is no write access to `dhi.io`, so DHI-native source definitions must publish elsewhere.
- Public delivery must stay on `ghcr.io`.
- The existing Woodpecker pipeline in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml` remains the build authority.
- The public chart in `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi` must remain pullable without authentication.
- The infra wrapper in `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi` remains the ownership boundary for cluster-specific settings.

## Current State

### Application image

`/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/Dockerfile` currently:

- uses `dhi.io/golang:* -dev` for the build stage
- uses `dhi.io/static:*` for the runtime stage
- builds a static `wakapi` binary and a static healthcheck binary
- publishes to `ghcr.io/yaelmoshi/wakapi-dhi` through Woodpecker

This is already hardened, but it is still Dockerfile-driven rather than native DHI-definition-driven.

### Chart

`/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi` currently:

- publishes as an OCI chart on `ghcr.io/yaelmoshi/charts/wakapi-dhi`
- defaults to `ghcr.io/yaelmoshi/wakapi-dhi`
- supports digest-pinned images
- already exposes security context and `imagePullSecrets`
- defaults `nameOverride` and `fullnameOverride` to `wakapi` to preserve runtime object names

### Infra wrapper

`/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi` currently:

- consumes the OCI chart from `ghcr.io/yaelmoshi/charts`
- supplies cluster-specific database, ingress, Gateway API, monitoring, and secret wiring
- pins the application image separately from the public chart defaults

## Recommended Approach

Use a native DHI image definition as the source of truth in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi`, but keep final image and chart distribution on GHCR.

This keeps the hardening and build model aligned with DHI guidance, without depending on `dhi.io` write access. It also preserves the current operating model:

- Forgejo as canonical source control
- Woodpecker as canonical builder
- GHCR as public runtime and chart registry
- wrapper chart in infra as the deploy contract

## Design

### 1. Native DHI image definition in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi`

Add a native DHI definition that becomes the source of truth for the runtime image.

The definition should:

- define the final runtime filesystem declaratively
- produce a minimal runtime image
- run as non-root
- carry OCI metadata
- support multi-platform output for at least `linux/amd64` and `linux/arm64` if the toolchain and application build remain compatible

The design keeps the existing application behavior:

- `wakapi` binary
- `healthcheck` binary
- `config.default.yml` installed as runtime config seed
- writable `/data`
- support for current environment-variable configuration model

Because Wakapi is a compiled Go application, the build should:

- build binaries in a builder path using DHI-supported build inputs
- assemble the final runtime image through the native DHI definition

The definition may rely on OCI artifacts or staged build outputs if a pure package-only image definition is not sufficient for the application build.

### 2. Woodpecker remains the build authority

`/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml` should be updated so the publish path uses the native DHI definition rather than `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/Dockerfile`.

The pipeline should:

- authenticate to `dhi.io` for build inputs
- authenticate to `ghcr.io` for publishing
- publish public tags:
  - short commit SHA
  - maintained semantic or release tag for the fork
  - `latest`
- preserve provenance and SBOM generation where the DHI toolchain supports it

The pipeline should continue to be triggered from Forgejo/Woodpecker rather than introducing a second release system.

### 3. DHI-style chart contract on GHCR

`/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi` should remain a public OCI chart on GHCR, but the chart documentation and defaults should clearly express the actual distribution model:

- native DHI-built image
- GHCR-hosted public runtime image
- GHCR-hosted public OCI Helm chart

Best-practice chart expectations:

- default image pinned by digest
- `imagePullSecrets` optional and empty by default
- security defaults remain strict
- chart remains generic and reusable outside the homelab
- runtime object names remain `wakapi`

The chart should not pretend to be hosted on `dhi.io`. It should describe itself as a DHI-built distribution delivered via GHCR.

### 4. Infra wrapper remains the cluster contract

`/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi` should not be structurally redesigned.

It should continue to own:

- CNPG integration
- OIDC and SMTP secret wiring
- Gateway API and ingress wiring
- ServiceMonitor integration
- per-cluster image pinning and operational resources

This avoids blending public chart responsibilities with private deployment concerns.

## DHI Best-Practice Mapping

### Image-side best practices to adopt

- Native DHI definition as source of truth
- non-root runtime user
- minimal runtime contents
- explicit metadata and versioned outputs
- multi-platform build where practical
- specific version tags and digest-aware downstream usage

### Chart-side best practices to adopt

- OCI chart distribution
- public default pulls where intended
- optional image pull secrets rather than mandatory secret wiring
- digest-pinned image defaults
- chart documentation that explains private-registry override paths without forcing them
- client-side validation of values and render shape

### Best practices not adopted

- publishing to `dhi.io`
- requiring authenticated pulls by default
- moving cluster-specific settings into the public chart

These are intentionally excluded because they do not fit the current ownership model or available credentials.

## Migration Plan

### Phase 1: Native image definition

In `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi`:

- add the native DHI build definition
- adapt the build flow to produce the current runtime payload
- keep the current Dockerfile only long enough to compare output and validate behavior

### Phase 2: CI migration

In `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml`:

- switch image publishing from Dockerfile input to the DHI-native path
- keep GHCR tag outputs stable
- keep DHI and GHCR logins

### Phase 3: Chart refresh

In `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi`:

- refresh README wording to describe the DHI-native build correctly
- pin the new default image digest from the native DHI build
- keep public pull defaults and security settings intact
- bump chart version

### Phase 4: Infra wrapper refresh

In `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi`:

- update the dependency version if the public chart version changes
- pin the tested application image tag or digest
- regenerate lock data and vendor chart tarball

## Validation

### `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi`

Required validation:

- native DHI build completes successfully
- produced image starts successfully
- `/api/health` responds successfully
- current Go tests still pass
- database-backed test flow still passes
- upgraded DHI DB images are verified in:
  - `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/compose.yml`
  - `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/testing/compose.yml`

### `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi`

Required validation:

- `helm lint`
- `helm template`
- rendered image reference is correct
- rendered security context remains correct
- rendered object names remain `wakapi`

### `/Users/yaelmeya/git/m0sh1.cc/infra/apps/user/wakapi`

Required validation:

- `helm dependency update`
- `helm lint`
- `helm template` with existing values files

No cluster writes are part of this design.

## Risks

### Native DHI build mismatch

The native DHI definition may not map one-to-one with the current multi-stage Dockerfile. The risk is mostly around how the build outputs are staged into the final image and how auxiliary files such as CA certs, zoneinfo, config, and healthcheck are carried over.

Mitigation:

- validate the produced runtime image locally before CI cutover
- compare startup behavior and health endpoint behavior directly

### CI toolchain mismatch

The existing Woodpecker plugin path is Dockerfile-oriented. The DHI-native path may need a different invocation model or a shell-based build command in the pipeline.

Mitigation:

- prove the local build commands first
- migrate CI only after the command line is stable

### Downstream digest drift

The public chart and infra wrapper both rely on stable, tested image references.

Mitigation:

- publish the native DHI-built image first
- update chart defaults only after the digest is real and tested
- update infra only after chart validation completes

## Decision

Proceed with option 1:

- native DHI image definition in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi`
- Woodpecker-based publish flow in `/Users/yaelmeya/git/m0sh1.cc/wakapi-dhi/.woodpecker/build.yaml`
- public image on `ghcr.io/yaelmoshi/wakapi-dhi`
- public OCI chart on `ghcr.io/yaelmoshi/charts/wakapi-dhi`
- DHI-style hardening and documentation in `/Users/yaelmeya/git/m0sh1.cc/helm-charts/charts/wakapi-dhi`
- no runtime rename and no `dhi.io` publication requirement
