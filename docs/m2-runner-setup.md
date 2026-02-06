# M2 Self-Hosted Runner Setup Guide

**Author:** Eric (Researcher)  
**Date:** 2026-02-06  
**Related Issues:** #253, #254, #224

## Overview

This guide documents how to configure **marin-6** and **marin-10** as self-hosted GitHub Actions runners for M2Sim benchmark execution.

## Available Hardware

| Machine | Specs | Status |
|---------|-------|--------|
| **marin-6** | Mac Mini, Apple M2, 24GB RAM | Available |
| **marin-10** | Mac Mini, Apple M2, 24GB RAM | Available |

Both machines are typically idle and suitable for:
- M2 baseline timing capture
- PolyBench/Embench benchmark runs
- SPEC CPU 2017 validation
- Long-running accuracy tests

## Self-Hosted Runner Setup

### Prerequisites

1. macOS with admin access
2. GitHub account with repo write access
3. Network connectivity to GitHub

### Step 1: Generate Runner Token

1. Go to https://github.com/sarchlab/m2sim/settings/actions/runners
2. Click **"New self-hosted runner"**
3. Select **macOS** and **ARM64**
4. Copy the token from the configuration command

### Step 2: Install Runner (on marin-6 or marin-10)

```bash
# Create runner directory
mkdir -p ~/actions-runner && cd ~/actions-runner

# Download latest runner (check GitHub for current version)
curl -o actions-runner-osx-arm64-2.314.1.tar.gz -L https://github.com/actions/runner/releases/download/v2.314.1/actions-runner-osx-arm64-2.314.1.tar.gz

# Extract
tar xzf ./actions-runner-osx-arm64-2.314.1.tar.gz
```

### Step 3: Configure Runner

```bash
# Configure with custom labels for M2 identification
./config.sh \
  --url https://github.com/sarchlab/m2sim \
  --token <YOUR_TOKEN> \
  --name "marin-6-m2" \
  --labels "self-hosted,macOS,ARM64,m2-chip" \
  --work _work
```

**Important labels:**
- `self-hosted` - Required for self-hosted runner selection
- `macOS` - OS identification
- `ARM64` - Architecture
- `m2-chip` - Custom label for M2-specific workflows

### Step 4: Run as Service

```bash
# Install as launch daemon (runs on boot)
sudo ./svc.sh install

# Start the service
sudo ./svc.sh start

# Check status
sudo ./svc.sh status
```

### Step 5: Verify Installation

Go to https://github.com/sarchlab/m2sim/settings/actions/runners

The runner should appear as **Idle** (green) or **Active** (yellow).

## Workflow Configuration

### Using the M2 Runner

Workflows should specify the custom labels:

```yaml
jobs:
  benchmark:
    runs-on: [self-hosted, m2-chip]
    # or more specific:
    # runs-on: [self-hosted, macOS, ARM64, m2-chip]
```

### Example: PolyBench Workflow

```yaml
name: Run PolyBench
on:
  workflow_dispatch:
    inputs:
      benchmark:
        type: choice
        options: [all, gemm, atax, 2mm, mvt, jacobi-1d, 3mm, bicg]

jobs:
  polybench:
    runs-on: [self-hosted, m2-chip]
    timeout-minutes: 360
    steps:
      - uses: actions/checkout@v4
      - name: Run benchmark
        run: ./scripts/capture-m2-baselines.sh ${{ inputs.benchmark }}
      - uses: actions/upload-artifact@v4
        with:
          name: polybench-results
          path: results/
```

## Agent Integration

Once configured, agents can trigger and monitor runs:

```bash
# Trigger a benchmark run
gh workflow run benchmark.yml -f suite=polybench -f benchmark=all

# Check run status
gh run list --workflow=benchmark.yml --limit=1

# Download results when complete
gh run download <run-id> -D results/
```

## Maintenance

### Updating the Runner

```bash
cd ~/actions-runner
sudo ./svc.sh stop
# Download new version, extract
sudo ./svc.sh start
```

### Checking Logs

```bash
# View service logs
cat ~/actions-runner/_diag/Runner_*.log | tail -100

# View job logs
cat ~/actions-runner/_diag/Worker_*.log | tail -100
```

### Troubleshooting

| Issue | Solution |
|-------|----------|
| Runner offline | Check `sudo ./svc.sh status`, restart if needed |
| Job stuck | Check machine isn't sleeping (disable sleep in System Settings) |
| Token expired | Re-run `./config.sh` with new token |
| Disk full | Clean old `_work` directories |

## Security Notes

- Runners have access to repository secrets during workflow runs
- Only run workflows from trusted sources
- Consider network isolation for sensitive workloads

## Dual Runner Strategy

For reliability, configure **both** marin-6 and marin-10:

- If one is busy, jobs auto-queue to the other
- Provides redundancy for nightly runs
- Labels allow targeting specific machines if needed:
  - `marin-6`: `self-hosted, m2-chip, marin-6`
  - `marin-10`: `self-hosted, m2-chip, marin-10`
