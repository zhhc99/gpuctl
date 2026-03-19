#Requires -RunAsAdministrator
$ErrorActionPreference = "Stop"
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

$repo = "zhhc99/gpuctl"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }
$url  = "https://github.com/$repo/releases/latest/download/gpuctl_Windows_${arch}.zip"
$tmp  = Join-Path $env:TEMP "gpuctl_install"

Write-Host "Downloading gpuctl for Windows_${arch}..."
if (Test-Path $tmp) { Remove-Item $tmp -Recurse -Force }
New-Item $tmp -ItemType Directory -Force | Out-Null

try {
    Invoke-WebRequest $url -OutFile "$tmp\g.zip" -UseBasicParsing
    Expand-Archive "$tmp\g.zip" $tmp -Force
    Write-Host "Installing..."
    & "$tmp\gpuctl.exe" install
} finally {
    Remove-Item $tmp -Recurse -Force
}

Write-Host "Done. Restart your terminal and try `gpuctl`."