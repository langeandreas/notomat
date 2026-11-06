$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$root = Split-Path $PSScriptRoot -Parent
$output = Join-Path $root 'dist'
$licenses = Join-Path $output 'licenses'
$oldGOOS = $env:GOOS
$oldGOARCH = $env:GOARCH
$oldCGO = $env:CGO_ENABLED

Push-Location $root
try {
    $env:GOOS = 'windows'
    $env:GOARCH = 'amd64'
    $env:CGO_ENABLED = '0'
    New-Item -ItemType Directory -Path $output, $licenses -Force | Out-Null

    go build -trimpath -ldflags '-s -w -H=windowsgui' -o (Join-Path $output 'notomat.exe') .
    if ($LASTEXITCODE -ne 0) { throw 'Release build failed.' }

    $modules = go list -deps -f '{{with .Module}}{{if not .Main}}{{.Path}}|{{.Dir}}{{end}}{{end}}' .
    if ($LASTEXITCODE -ne 0) { throw 'Dependency license discovery failed.' }
    foreach ($module in ($modules | Sort-Object -Unique)) {
        if ([string]::IsNullOrWhiteSpace($module)) { continue }
        $parts = $module -split '\|', 2
        $licenseFiles = @(Get-ChildItem -LiteralPath $parts[1] -File |
            Where-Object { $_.Name -match '^(LICENSE|COPYING)' })
        if ($licenseFiles.Count -eq 0) { throw "No license found for $($parts[0])." }
        foreach ($file in $licenseFiles) {
            $name = $parts[0].Replace('/', '__') + '__' + $file.Name
            Copy-Item -LiteralPath $file.FullName -Destination (Join-Path $licenses $name) -Force
        }
    }
    $goroot = go env GOROOT
    if ($LASTEXITCODE -ne 0) { throw 'Go license discovery failed.' }
    Copy-Item -LiteralPath (Join-Path $goroot 'LICENSE') -Destination (Join-Path $licenses 'Go-LICENSE') -Force
    Copy-Item -LiteralPath (Join-Path $root 'README.md') -Destination $output -Force

    $archive = Join-Path $output 'notomat-windows-amd64.zip'
    Compress-Archive -Path (Join-Path $output 'notomat.exe'), (Join-Path $output 'README.md'), $licenses -DestinationPath $archive -Force
    $checksums = foreach ($file in @((Join-Path $output 'notomat.exe'), $archive)) {
        $hash = Get-FileHash -LiteralPath $file -Algorithm SHA256
        "$($hash.Hash.ToLowerInvariant())  $([System.IO.Path]::GetFileName($file))"
    }
    $checksums | Set-Content -LiteralPath (Join-Path $output 'SHA256SUMS.txt') -Encoding ascii
    Write-Output "Release package: $archive"
} finally {
    $env:GOOS = $oldGOOS
    $env:GOARCH = $oldGOARCH
    $env:CGO_ENABLED = $oldCGO
    Pop-Location
}
