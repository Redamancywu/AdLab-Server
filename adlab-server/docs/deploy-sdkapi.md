# SDKAPI Deployment Guide

## Overview

`cmd/sdkapi` is the lightweight runtime for SDK-facing traffic.

Use it when you want:

- isolated SDK/public routes
- independent rate limiting
- separate scaling from the full admin/backend runtime
- simpler smoke checks and operational debugging

## Entry Modes

### Local

```bash
cd adlab-server
make run-sdkapi
```

### Docker Compose

Minimal runtime:

```bash
cd adlab-server
docker compose -f docker-compose.sdkapi.yml up -d
```

Shared stack runtime:

```bash
cd adlab-server
docker compose up -d postgres sdkapi
```

## Ports

- `sdkapi`: `8090` by default
- `backend`: `8080` by default

## Endpoints

### Health

- `GET /health`
- `GET /ready`

### Observability

- `GET /metrics`
- `GET /version`

### SDK APIs

- `POST /api/v1/sdk/init`
- `POST /api/v1/sdk/init_complete`
- `POST /api/v1/sdk/heartbeat`
- `POST /api/v1/sdk/ecpm`
- `POST /api/v1/ad/request`

## Runtime Config

Configured under `sdkapi` in `config/config.yaml`:

- `port`
- `enable_docs`
- `enable_lab`
- `enable_health`
- `rate_limit_enabled`
- `rate_limit_rps`
- `rate_limit_burst`

## Smoke Check

```bash
cd adlab-server
make smoke-sdkapi
```

## Build Metadata

`/version` reads build metadata from:

- `Version`
- `GitSHA`
- `Component`

These are injected through `Makefile` `LDFLAGS`.

Set before running:

```bash
export ADLAB_VERSION=0.1.0
export ADLAB_GIT_SHA=$(git rev-parse --short HEAD)
```

## Logs

Access logs include:

- `component=sdkapi`
- request method
- route path
- status
- latency
- request id

This makes it easier to isolate SDK traffic from full backend/admin traffic.
