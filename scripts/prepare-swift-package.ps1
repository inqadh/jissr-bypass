[CmdletBinding()]
param(
    [Parameter(Mandatory)]
    [string]$Version,

    [Parameter(Mandatory)]
    [string]$XCFrameworkZip,

    [Parameter(Mandatory)]
    [string]$Output
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

if ($Version -notmatch '^\d+\.\d+\.\d+([.-][0-9A-Za-z.-]+)?$') {
    throw "Version must be an immutable semantic version without a v prefix."
}

$root = Split-Path -Parent $PSScriptRoot
$zipPath = [IO.Path]::GetFullPath($XCFrameworkZip)
if (-not (Test-Path -LiteralPath $zipPath -PathType Leaf)) {
    throw "XCFramework ZIP not found: $zipPath"
}

$outputPath = [IO.Path]::GetFullPath($Output)
if (Test-Path -LiteralPath $outputPath) {
    throw "Output already exists: $outputPath"
}

$checksum = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
$templatePath = Join-Path $root "swift-package\Package.swift.template"
$manifest = [IO.File]::ReadAllText($templatePath)
$manifest = $manifest.Replace("@VERSION@", $Version).Replace("@CHECKSUM@", $checksum)

New-Item -ItemType Directory -Path $outputPath | Out-Null
[IO.File]::WriteAllText((Join-Path $outputPath "Package.swift"), $manifest)
Copy-Item `
    -LiteralPath (Join-Path $root "adapters\apple\Sources") `
    -Destination (Join-Path $outputPath "Sources") `
    -Recurse
Copy-Item `
    -LiteralPath (Join-Path $root "adapters\apple\Tests") `
    -Destination (Join-Path $outputPath "Tests") `
    -Recurse
Copy-Item `
    -LiteralPath (Join-Path $root "LICENSE") `
    -Destination (Join-Path $outputPath "LICENSE")
Copy-Item `
    -LiteralPath (Join-Path $root "adapters\apple\README.md") `
    -Destination (Join-Path $outputPath "README.md")

Write-Host "Swift package source: $outputPath"
Write-Host "Version: $Version"
Write-Host "XCFramework checksum: $checksum"
