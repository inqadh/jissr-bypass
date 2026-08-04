[CmdletBinding()]
param(
    [string]$Version = "dev"
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist\desktop"
$targets = @(
    @{ OS = "windows"; Arch = "amd64"; Ext = ".exe" },
    @{ OS = "windows"; Arch = "arm64"; Ext = ".exe" },
    @{ OS = "linux"; Arch = "amd64"; Ext = "" },
    @{ OS = "linux"; Arch = "arm64"; Ext = "" },
    @{ OS = "darwin"; Arch = "amd64"; Ext = "" },
    @{ OS = "darwin"; Arch = "arm64"; Ext = "" }
)

New-Item -ItemType Directory -Force -Path $dist | Out-Null
Push-Location $root
try {
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed." }
    foreach ($target in $targets) {
        $name = "jissr-bypass-$Version-$($target.OS)-$($target.Arch)$($target.Ext)"
        $output = Join-Path $dist $name
        $env:GOOS = $target.OS
        $env:GOARCH = $target.Arch
        $env:CGO_ENABLED = "0"
        go build -trimpath -ldflags="-s -w" -o $output ./cmd/jissr-bypass
        if ($LASTEXITCODE -ne 0) { throw "Build failed for $($target.OS)/$($target.Arch)." }
        Write-Host "$name $((Get-FileHash -Algorithm SHA256 -LiteralPath $output).Hash.ToLowerInvariant())"
    }
} finally {
    Remove-Item Env:GOOS -ErrorAction SilentlyContinue
    Remove-Item Env:GOARCH -ErrorAction SilentlyContinue
    Remove-Item Env:CGO_ENABLED -ErrorAction SilentlyContinue
    Pop-Location
}
