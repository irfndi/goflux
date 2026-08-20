# Migration from Techan

## Overview

GoFlux is a maintained fork of [sdcoffey/techan](https://github.com/sdcoffey/techan).

## What Changed

- **Module path**: `github.com/sdcoffey/techan` → `github.com/irfndi/goflux`
- **Package name**: `techan` → `goflux`
- **Tooling**: standardized Make targets (and Docker-based workflows)

## Migration Guide

Update imports:

```go
// Before
import "github.com/sdcoffey/techan"

// After
import "github.com/irfndi/goflux/pkg"
```

The legacy root facade preserves the v0.0.5 factory-variable API and the historical database scaffolds. New functionality is organized in subpackages (`analysis`, `backtest`, `database`, `decimal`, `indicators`, `series`, `telemetry`, and `trading`). Database backends remain scaffolds and return an explicit not-implemented error until a client is configured.

## Project Layout

GoFlux is organized as multiple Go packages under `pkg/`, with usage examples under `example/`.

## Next Steps

Use `bd ready` for the current issue-tracking roadmap and see `AGENTS.md` for the project workflow.
