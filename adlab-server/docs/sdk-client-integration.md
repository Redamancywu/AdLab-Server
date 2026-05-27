# SDK Client Integration Guide

## Overview

This guide is for client SDK developers integrating against the current AdLab / AeroBid `sdkapi` runtime.

The current SDK-facing API is complete enough for a minimum mediation loop:

- fetch app + placement config
- report adapter init result
- request ads from a single unified endpoint
- run optional C2S bidding locally and report the result
- track impression / click / video progress
- detect config refresh through heartbeat

## Recommended Flow

### 1. Initialize on app launch

Call:

```http
POST /api/v1/sdk/init
```

Purpose:

- get app-level network initialization params
- get placement instances and compatibility waterfall
- get `config_version` and `config_hash` for later heartbeat comparison

### 2. Report adapter initialization result

After local network SDK initialization completes, call:

```http
POST /api/v1/sdk/init_complete
```

Purpose:

- tell the server which adapters succeeded or failed
- receive adjusted placement waterfall if some adapters failed

### 3. Request an ad

Use the unified ad entry:

```http
POST /api/v1/ad/request
```

Possible server behaviors:

- `bid_mode = s2s`: server-side bidding already completed
- `bid_mode = waterfall`: sequential server-side fill already completed
- `bid_mode = mock`: mock fallback used
- `bid_mode = c2s` or response includes `c2s_sources`: client should complete local bidding

### 4. If client-side bidding is used, report result

After local bidding finishes:

```http
POST /api/v1/c2s/result
```

If the winning ad is displayed later:

```http
POST /api/v1/c2s/display
```

### 5. Track exposure and playback lifecycle

Use:

```http
GET /api/v1/track
POST /api/v1/track
```

Common events:

- `impression`
- `click`
- `start`
- `firstQuartile`
- `midpoint`
- `thirdQuartile`
- `complete`

### 6. Periodic heartbeat

Call:

```http
POST /api/v1/sdk/heartbeat
```

If `config_updated = true`, rerun:

```text
sdk/init -> sdk/init_complete
```

## Core Endpoints

### SDK lifecycle

- `POST /api/v1/sdk/init`
- `POST /api/v1/sdk/init_complete`
- `POST /api/v1/sdk/heartbeat`
- `POST /api/v1/sdk/ecpm`

### Ad serving

- `POST /api/v1/ad/request`
- `POST /api/v1/s2s/bid`
- `POST /api/v1/waterfall/bid`
- `POST /api/v1/mock/fill`

### C2S reporting

- `POST /api/v1/c2s/result`
- `POST /api/v1/c2s/display`

### Tracking

- `GET /api/v1/track`
- `POST /api/v1/track`

### Public docs and debug surfaces

- `GET /docs`
- `GET /docs/openapi.json`
- `GET /api/v1/docs/{key}`
- `GET /health`
- `GET /ready`
- `GET /version`
- `GET /metrics`

## Important Response Fields

### `sdk/init`

Focus on these fields:

- `global.default_timeout_ms`
- `global.max_retries`
- `global.enable_mock_fallback`
- `global.heartbeat_interval_s`
- `networks[]`
- `placements[]`
- `placements[].instances[]`
- `placements[].waterfall[]`
- `config_version`
- `config_hash`

### `ad/request`

Focus on these fields:

- `request_id`
- `bid_mode`
- `winner_dsp_id`
- `winner_price`
- `vast_xml`
- `image_url`
- `splash_url`
- `native_ad`
- `track_urls`
- `c2s_sources`

## Completeness Assessment

### What is already complete for client integration

- Unified init flow is present
- Adapter init feedback flow is present
- Unified ad request entry is present
- C2S result reporting is present
- Tracking callbacks are present
- Heartbeat refresh detection is present
- Mock fallback is present
- Health / readiness / version / metrics are present

### Current gaps to be aware of

- Some public routes exist in runtime but were previously missing from the checked-in OpenAPI copy
- The default developer docs were previously outdated and did not reflect the current `sdk/init -> init_complete -> ad/request` flow
- Error-code guidance is still lighter than ideal for a production SDK contract
- Public route documentation for logs and stats is still more operational than SDK-oriented

### Practical conclusion

For an internal SDK, prototype SDK, or learning project, the current SDKAPI surface is complete enough to support real client integration.

For a stronger production-facing contract, the next recommended improvements are:

1. extend OpenAPI coverage for all public runtime routes
2. add a dedicated error-code appendix
3. add per-ad-format response examples for banner / splash / native / video
4. document retry, timeout, and idempotency rules in more detail

## Examples

### `sdk/init`

```json
{
  "app_id": "app_ios_001",
  "sdk_version": "1.0.0",
  "platform": "ios",
  "device": {
    "os_version": "17.5",
    "device_model": "iPhone15,2",
    "language": "zh-CN",
    "ifa": "00000000-0000-0000-0000-000000000000",
    "ifa_type": "idfa"
  }
}
```

### `ad/request`

```json
{
  "placement_id": "plc_rewarded_001",
  "app": {
    "app_id": "app_ios_001",
    "bundle_id": "com.example.game",
    "name": "Example Game",
    "version": "1.0.0"
  },
  "device": {
    "platform": "ios",
    "os_version": "17.5",
    "device_model": "iPhone15,2",
    "language": "zh-CN",
    "ifa": "00000000-0000-0000-0000-000000000000",
    "ifa_type": "idfa"
  }
}
```

### `c2s/result`

```json
{
  "request_id": "sdk-c2s-001",
  "placement_id": "plc_rewarded_001",
  "winner_dsp_id": "admob",
  "winner_price": 1.52,
  "displayed": false,
  "bidding_details": [
    {
      "dsp_id": "admob",
      "bid_price": 1.52,
      "status": "win"
    },
    {
      "dsp_id": "pangle",
      "bid_price": 1.12,
      "status": "lose"
    }
  ]
}
```
