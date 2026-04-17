param(
    [string]$GOOS = "windows",
    [string]$GOARCH = "amd64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$distDir = Join-Path $repoRoot "dist"
if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir | Out-Null
}

$binaryName = "kdoctor-$GOOS-$GOARCH"
if ($GOOS -eq "windows") {
    $binaryName += ".exe"
}
$output = Join-Path $distDir $binaryName

Push-Location $repoRoot
try {
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    go build -o $output ./cmd/kdoctor
    Write-Host "Built $output"
}
finally {
    Pop-Location
}
