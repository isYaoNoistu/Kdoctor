param(
    [string]$GOOS = "windows",
    [string]$GOARCH = "amd64"
)

$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$workspaceRoot = Split-Path -Parent $repoRoot
$distDir = Join-Path $workspaceRoot "dist"
if (-not (Test-Path $distDir)) {
    New-Item -ItemType Directory -Path $distDir | Out-Null
}

$rawDir = Join-Path $distDir "raw"
$targetDir = Join-Path $rawDir "$GOOS-$GOARCH"
if (-not (Test-Path $targetDir)) {
    New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
}

$cacheRoot = Join-Path $distDir ".build-cache"
$goCacheDir = Join-Path $cacheRoot "gocache"
$goTmpDir = Join-Path $cacheRoot "gotmp"
if (-not (Test-Path $goCacheDir)) {
    New-Item -ItemType Directory -Path $goCacheDir -Force | Out-Null
}
if (-not (Test-Path $goTmpDir)) {
    New-Item -ItemType Directory -Path $goTmpDir -Force | Out-Null
}

$binaryName = "kdoctor"
if ($GOOS -eq "windows") {
    $binaryName += ".exe"
}
$output = Join-Path $targetDir $binaryName

Push-Location $repoRoot
try {
    $commit = (git rev-parse --short HEAD).Trim()
    $version = "v2.0.0"
    $ldflags = "-s -w -X kdoctor/pkg/buildinfo.Version=$version -X kdoctor/pkg/buildinfo.Commit=$commit"
    $env:GOOS = $GOOS
    $env:GOARCH = $GOARCH
    $env:GOCACHE = $goCacheDir
    $env:GOTMPDIR = $goTmpDir
    go build -trimpath -ldflags $ldflags -o $output ./cmd/kdoctor
    Write-Host "Built $output"
}
finally {
    Pop-Location
}
