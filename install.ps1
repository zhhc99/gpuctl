<#
.SYNOPSIS
    Installs gpuctl to $HOME\.gpuctl
#>
$ErrorActionPreference = "Stop"

# os & arch check
# os check fails on PS 5.1; skip it
# if (!$IsWindows) {
#     $current_os = $PSVersionTable.OS -ifEmpty ""
#     Write-Host "❗ Error: unsupported os: $current_os" -F Red
#     exit 1
# }
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "x86_64" } else { "arm64" }

# find url for latest release
$platform = "Windows_$arch"
$repo = "zhhc99/gpuctl"
$url = "https://github.com/$repo/releases/latest/download/gpuctl_$platform.zip"


# download and install
$dir  = "$env:LOCALAPPDATA\Programs\gpuctl"
$tmp  = "$env:TEMP\gpuctl_tmp"

Write-Host "🚀 Downloading gpuctl for Windows_$arch..." -F Cyan
if (Test-Path $tmp) { rm -rf $tmp }
mkdir $tmp, $dir -Force | Out-Null

try {
    iwr $url -OutFile "$tmp\g.zip" -UseBasicParsing
    Expand-Archive "$tmp\g.zip" $tmp -Force
    Write-Host "⚙️  Installing to $dir..." -F Cyan
    mv "$tmp\gpuctl.exe" "$dir\gpuctl.exe" -Force
} catch {
    Write-Host "❗ Error: failed to download. something must be wrong... 🤔" -F Red; exit 1
} finally {
    rm -rf $tmp
}

# add to path
$user_path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($user_path -split ';' -notcontains $dir) {
    [Environment]::SetEnvironmentVariable("Path", "$user_path;$dir", "User")
    $env:Path += ";$dir"
}

Write-Host "🎉 Done. Try run 'gpuctl'!" -F Green
