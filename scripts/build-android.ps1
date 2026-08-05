[CmdletBinding()]
param(
    [string]$Output = ""
)

$ErrorActionPreference = "Stop"
Set-StrictMode -Version Latest

$root = Split-Path -Parent $PSScriptRoot
if (-not $Output) {
    $Output = Join-Path $root "dist\android\jissr-bypass.aar"
}

[array]$javaCandidates = @(
    $(if ($env:JAVA_HOME) { $env:JAVA_HOME }),
    "C:\Program Files\Android\Android Studio\jbr"
) | Where-Object {
    $_ -and (Test-Path -LiteralPath (Join-Path $_ "bin\javac.exe"))
}
if (-not $javaCandidates) {
    throw "A JDK was not found. Set JAVA_HOME or install Android Studio."
}
$env:JAVA_HOME = @($javaCandidates)[0]
$javaBin = Join-Path $env:JAVA_HOME "bin"
if (($env:PATH -split [IO.Path]::PathSeparator) -notcontains $javaBin) {
    $env:PATH = $javaBin + [IO.Path]::PathSeparator + $env:PATH
}

$Output = [IO.Path]::GetFullPath($Output)
$outputDir = Split-Path -Parent $Output
New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

Push-Location $root
try {
    go test ./...
    if ($LASTEXITCODE -ne 0) { throw "Go tests failed." }
    go tool gomobile init
    if ($LASTEXITCODE -ne 0) { throw "gomobile initialization failed." }
    go tool gomobile bind -target=android -androidapi 26 -o $Output ./mobile
    if ($LASTEXITCODE -ne 0) { throw "Android AAR build failed." }
    $hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $Output).Hash.ToLowerInvariant()
    Write-Host "Android artifact: $Output"
    Write-Host "SHA-256: $hash"
} finally {
    Pop-Location
}
