#Requires -RunAsAdministrator
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo    = "zhhc99/gpuctl"
$binDir  = Join-Path $env:ProgramFiles "gpuctl"
$binPath = Join-Path $binDir "gpuctl.exe"
$svcName = "gpuctl"
$arch    = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }
$url     = "https://github.com/$repo/releases/latest/download/gpuctl_Windows_${arch}.zip"
$tmp     = Join-Path $env:TEMP "gpuctl_install"

Write-Host "Downloading gpuctl for Windows/${arch}..."
if (Test-Path $tmp) { Remove-Item $tmp -Recurse -Force }
New-Item $tmp -ItemType Directory -Force | Out-Null

try {
    Invoke-WebRequest $url -OutFile "$tmp\g.zip" -UseBasicParsing
    Expand-Archive "$tmp\g.zip" $tmp -Force

    Write-Host "Installing binary to $binDir..."
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
    Copy-Item "$tmp\gpuctl.exe" $binPath -Force

    # Add to system PATH if not already present
    $syspath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    if ($syspath -notlike "*$binDir*") {
        [Environment]::SetEnvironmentVariable("Path", "$syspath;$binDir", "Machine")
        Write-Host "Added $binDir to system PATH."
    }

    # Register or update Windows Service
    $existing = Get-Service -Name $svcName -ErrorAction SilentlyContinue
    if ($existing) {
        Write-Host "Service already registered — updating..."
        Stop-Service -Name $svcName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 1
        sc.exe config $svcName binPath= "`"$binPath`" daemon" | Out-Null
    } else {
        Write-Host "Registering service..."
        New-Service -Name $svcName `
            -BinaryPathName "`"$binPath`" daemon" `
            -DisplayName "gpuctl GPU controller" `
            -Description "gpuctl — apply profiles and control fans" `
            -StartupType Automatic | Out-Null
    }

    Write-Host "Starting service..."
    Start-Service -Name $svcName
    Write-Host "Done. Service is running."
    Write-Host "Restart your terminal for PATH changes to take effect."
} finally {
    Remove-Item $tmp -Recurse -Force
}