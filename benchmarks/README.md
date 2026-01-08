# Surge Benchmarks

## Quick Start

```powershell
# Quick test (100MB only, 3 runs)
.\benchmark.ps1 -Quick

# Full benchmark (100MB + 1GB, 3 runs each)
.\benchmark.ps1

# Extended benchmark (5 runs)
.\benchmark.ps1 -Runs 5
```

## Requirements

- **PowerShell 5.1+** (Windows) or PowerShell Core (cross-platform)
- **surge.exe** built in parent directory
- Optional: aria2c, wget, curl for comparison

## Install Comparison Tools (Windows)

```powershell
# Using Chocolatey
choco install aria2 wget curl

# Or using Scoop
scoop install aria2 wget curl
```

## Metrics Collected

| Metric | Description |
|--------|-------------|
| Throughput (MB/s) | Total bytes / elapsed time |
| Time to First Byte | Connection setup overhead |
| Total Time | Start to completion |
| Retry Rate | Number of retries during download |
| Connection Count | Max/avg concurrent connections |
| Memory Usage | Peak RAM consumption |

## Test Files

| Size | Source |
|------|--------|
| 100MB | speed.hetzner.de |
| 1GB | speed.hetzner.de |

## Output

Results are:
1. Displayed in terminal as formatted table
2. Exported to `benchmark_results_YYYYMMDD_HHMMSS.csv`

## Example Results

```
File  Tool   AvgTime AvgSpeed MinTime MaxTime Runs
----- ------ ------- -------- ------- ------- ----
100MB surge  1.52    67.5     1.45    1.61    3
100MB aria2  1.87    54.8     1.82    1.95    3
100MB wget   4.53    22.6     4.41    4.72    3
100MB curl   4.41    23.2     4.35    4.51    3
```
