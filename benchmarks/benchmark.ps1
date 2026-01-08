# Surge Benchmark Suite - Simplified
# Run from benchmarks folder or surge root

param(
    [int]$Runs = 1,
    [switch]$Quick
)

$ErrorActionPreference = "Continue"

# Find surge.exe
$SurgePath = if (Test-Path "./surge.exe") { "./surge.exe" } 
             elseif (Test-Path "../surge.exe") { "../surge.exe" } 
             else { $null }

# Test files
$TestFiles = @(
    @{Name="1GB"; URL="https://ash-speed.hetzner.com/1GB.bin"; Size=1073741824}
)
if (-not $Quick) {
    $TestFiles += @{Name="10GB"; URL="https://ash-speed.hetzner.com/10GB.bin"; Size=10737418240}
}

$Results = @()

Write-Host "`n========================================"  -ForegroundColor Cyan
Write-Host "       SURGE BENCHMARK SUITE            " -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

Write-Host "System: $([System.Environment]::OSVersion.VersionString)"
Write-Host "CPU: $((Get-CimInstance Win32_Processor).Name)"
Write-Host "Date: $(Get-Date -Format 'yyyy-MM-dd HH:mm')`n"

# Check tools
Write-Host "Available tools:" -ForegroundColor Yellow
$Tools = @()

if ($SurgePath) { 
    Write-Host "  [OK] surge" -ForegroundColor Green
    $Tools += @{Name="surge"; Available=$true}
} else {
    Write-Host "  [MISSING] surge (run 'go build' first)" -ForegroundColor Red
}

if (Get-Command aria2c -ErrorAction SilentlyContinue) {
    Write-Host "  [OK] aria2c" -ForegroundColor Green
    $Tools += @{Name="aria2c"; Available=$true}
}

if (Get-Command curl.exe -ErrorAction SilentlyContinue) {
    Write-Host "  [OK] curl" -ForegroundColor Green
    $Tools += @{Name="curl"; Available=$true}
}

Write-Host ""

foreach ($file in $TestFiles) {
    Write-Host "Testing: $($file.Name)" -ForegroundColor Cyan
    Write-Host "URL: $($file.URL)"
    Write-Host ""

    foreach ($tool in $Tools) {
        if (-not $tool.Available) { continue }
        
        Write-Host "  $($tool.Name):" -ForegroundColor Yellow
        $times = @()
        
        for ($i = 1; $i -le $Runs; $i++) {
            $outFile = "$($file.Name)_$($tool.Name).bin"
            Remove-Item $outFile -ErrorAction SilentlyContinue
            
            Write-Host "    Run $i/$Runs... " -NoNewline
            
            $sw = [System.Diagnostics.Stopwatch]::StartNew()
            
            try {
                switch ($tool.Name) {
                    "surge" {
                        $proc = Start-Process -FilePath $SurgePath -ArgumentList "get","--headless","-o",$outFile,$file.URL -Wait -PassThru -NoNewWindow
                    }
                    "aria2c" {
                        $proc = Start-Process -FilePath "aria2c" -ArgumentList "-x","16","-s","16","-o",$outFile,$file.URL -Wait -PassThru -NoNewWindow
                    }
                    "curl" {
                        $proc = Start-Process -FilePath "curl.exe" -ArgumentList "-s","-o",$outFile,$file.URL -Wait -PassThru -NoNewWindow
                    }
                }
            } catch {
                Write-Host "ERROR: $_" -ForegroundColor Red
                continue
            }
            
            $sw.Stop()
            $elapsed = $sw.Elapsed.TotalSeconds
            
            # Check if file exists and has content
            $fileInfo = Get-Item $outFile -ErrorAction SilentlyContinue
            $valid = $fileInfo -and $fileInfo.Length -gt 0
            
            if ($valid) {
                $speed = $fileInfo.Length / $elapsed / 1MB
                Write-Host "$([math]::Round($elapsed, 2))s ($([math]::Round($speed, 2)) MB/s)" -ForegroundColor Green
                $times += @{Time=$elapsed; Speed=$speed; Size=$fileInfo.Length}
            } else {
                Write-Host "FAILED" -ForegroundColor Red
            }
            
            Remove-Item $outFile -ErrorAction SilentlyContinue
        }
        
        if ($times.Count -gt 0) {
            $avgTime = ($times | ForEach-Object { $_.Time } | Measure-Object -Average).Average
            $avgSpeed = ($times | ForEach-Object { $_.Speed } | Measure-Object -Average).Average
            
            Write-Host "    Average: $([math]::Round($avgTime, 2))s @ $([math]::Round($avgSpeed, 2)) MB/s" -ForegroundColor Cyan
            
            $Results += [PSCustomObject]@{
                File = $file.Name
                Tool = $tool.Name
                AvgTimeSec = [math]::Round($avgTime, 2)
                AvgSpeedMBs = [math]::Round($avgSpeed, 2)
                Runs = $times.Count
            }
        }
        Write-Host ""
    }
}

# Summary
Write-Host "========================================"  -ForegroundColor Cyan
Write-Host "              SUMMARY                   " -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

if ($Results.Count -gt 0) {
    $Results | Format-Table -AutoSize
    
    $csv = "benchmark_$(Get-Date -Format 'yyyyMMdd_HHmm').csv"
    $Results | Export-Csv $csv -NoTypeInformation
    Write-Host "Results saved to: $csv" -ForegroundColor Green
} else {
    Write-Host "No results collected." -ForegroundColor Red
}
