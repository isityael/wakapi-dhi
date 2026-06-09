# renovate: datasource=docker
ARG GO_BASE=dhi.io/golang:1.26.4-alpine3.23-dev@sha256:e5fdaf8c202914575667f343c52703d28d579454066138c682f5bb6f9d259854

FROM --platform=$BUILDPLATFORM ${GO_BASE} AS build-env
WORKDIR /src

RUN apk upgrade --no-cache && \
    apk add --no-cache ca-certificates tzdata && update-ca-certificates

COPY ./go.mod ./go.sum ./
COPY ./vendor ./vendor
COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 GOFLAGS=-mod=vendor GOEXPERIMENT=jsonv2 go build -ldflags "-s -w" -v -o wakapi main.go
# Need a statically linked healthcheck binary because the static runtime image does not include curl.
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 GOFLAGS=-mod=vendor go build -ldflags "-s -w" -v -o healthcheck scripts/healthcheck.go

WORKDIR /staging
RUN mkdir ./data ./app && \
    cp /src/wakapi app/ && \
    cp /src/healthcheck app/ && \
    cp /src/config.default.yml app/config.yml && \
    sed -i 's/listen_ipv6: ::1/listen_ipv6: "-"/g' app/config.yml

# Run Stage

# When running the application using `docker run`, you can pass environment variables
# to override config values using `-e` syntax.
# Available options can be found in [README.md#-configuration](README.md#-configuration)

# Note on the static runtime image:
# Wakapi is built with CGO_ENABLED=0, so the final image only needs a minimal runtime for static binaries.

FROM dhi.io/static:20250419-debian13@sha256:5561af898afa6015a84d39a2cefdfc7fc386f45026d7834c95d8dbc450061563
WORKDIR /app

# See README.md and config.default.yml for all config options
ENV ENVIRONMENT=prod \
    WAKAPI_DB_TYPE=sqlite3 \
    WAKAPI_DB_USER='' \
    WAKAPI_DB_PASSWORD='' \
    WAKAPI_DB_HOST='' \
    WAKAPI_DB_NAME=/data/wakapi.db \
    WAKAPI_PASSWORD_SALT='' \
    WAKAPI_LISTEN_IPV4='0.0.0.0' \
    WAKAPI_INSECURE_COOKIES='true' \
    WAKAPI_ALLOW_SIGNUP='true'

COPY --from=build-env --chown=nonroot:nonroot --chmod=0444 /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build-env --chown=nonroot:nonroot --chmod=0555 /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=build-env --chown=nonroot:nonroot /staging/app /app
COPY --from=build-env --chown=nonroot:nonroot /staging/data /data

LABEL org.opencontainers.image.url="https://github.com/yaelmoshi/wakapi-dhi" \
    org.opencontainers.image.documentation="https://github.com/muety/wakapi" \
    org.opencontainers.image.source="https://github.com/yaelmoshi/wakapi-dhi" \
    org.opencontainers.image.title="Wakapi (DHI-hardened)" \
    org.opencontainers.image.licenses="MIT" \
    org.opencontainers.image.description="Wakapi — DHI-hardened fork with bug fixes"

USER nonroot

EXPOSE 3000

# For long-running migrations, you might want to override `---health-start-period` as part of `docker run` or disable healthchecks entirely with `--no-healtcheck`
HEALTHCHECK --interval=60s --timeout=3s --start-period=120s --retries=3 CMD ["/app/healthcheck"]

ENTRYPOINT ["/app/wakapi"]
