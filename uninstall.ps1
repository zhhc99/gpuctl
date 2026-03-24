#Requires -RunAsAdministrator
$ErrorActionPreference = "Stop"

$svcName = "gpuctl"
$binDir  = Join-Path $env:ProgramFiles "gpuctl"
$confDir = Join-Path $env:ProgramData "gpuctl"

# Stop and remove service
$svc = Get-Service -Name $svcName -ErrorAction SilentlyContinue
if ($svc) {
    Write-Host "Stopping service..."
    Stop-Service -Name $svcName -Force -ErrorAction SilentlyContinue
    Write-Host "Removing service..."
    sc.exe delete $svcName | Out-Null
} else {
    Write-Host "Service not found, skipping."
}

# Remove from PATH
$syspath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$filtered = ($syspath -split ";" | Where-Object { $_ -ne $binDir }) -join ";"
if ($filtered -ne $syspath) {
    [Environment]::SetEnvironmentVariable("Path", $filtered, "Machine")
    Write-Host "Removed $binDir from system PATH."
}

# Remove binary
if (Test-Path $binDir) {
    Write-Host "Removing $binDir..."
    Remove-Item $binDir -Recurse -Force
}

# Optionally remove config
if (Test-Path $confDir) {
    $ans = Read-Host "Remove config directory $confDir? [y/N] (default N)"
    if ($ans -match "^[Yy]") {
        Remove-Item $confDir -Recurse -Force
        Write-Host "Removed $confDir."
    } else {
        Write-Host "Keeping $confDir."
    }
}

Write-Host "Done. gpuctl has been uninstalled."
