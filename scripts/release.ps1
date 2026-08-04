[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$')]
    [string]$Version,

    [switch]$SkipAndroid,
    [switch]$SkipDesktop
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist"

Push-Location $root
try {
    go mod tidy
    if ($LASTEXITCODE -ne 0) { throw "go mod tidy failed." }
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed." }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw "Go vet failed." }

    if (-not $SkipAndroid) {
        & (Join-Path $PSScriptRoot "build-android.ps1")
        if ($LASTEXITCODE -ne 0) { throw "Android build failed." }
    }
    if (-not $SkipDesktop) {
        & (Join-Path $PSScriptRoot "build-desktop.ps1") -Version $Version
        if ($LASTEXITCODE -ne 0) { throw "Desktop build failed." }
    }

    $checksumPath = Join-Path $dist "SHA256SUMS"
    $artifacts = Get-ChildItem -LiteralPath $dist -Recurse -File |
        Where-Object { $_.FullName -ne $checksumPath } |
        Sort-Object FullName
    $lines = foreach ($artifact in $artifacts) {
        $relative = [IO.Path]::GetRelativePath($dist, $artifact.FullName).Replace('\', '/')
        $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $artifact.FullName).Hash.ToLowerInvariant()
        "$hash  $relative"
    }
    [IO.File]::WriteAllLines($checksumPath, $lines, [Text.UTF8Encoding]::new($false))

    Write-Host ""
    Write-Host "Jissr Bypass $Version candidate is ready in $dist"
    Write-Host "Apple artifacts must be produced on macOS with scripts/build-apple.sh."
    Write-Host "Do not tag until the final public module path is confirmed."
} finally {
    Pop-Location
}
