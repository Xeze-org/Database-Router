# Downloads mTLS client certificates from the db-router server.
# Usage: .\fetch-certs.ps1 -Host root@168.144.22.57 [-Key ~/.ssh/id_ed25519]
param(
    [Parameter(Mandatory=$true)]
    [string]$RemoteHost,
    [string]$Key
)

$CertDir = Join-Path $PSScriptRoot "certs"
New-Item -ItemType Directory -Force -Path $CertDir | Out-Null

$keyArgs = @()
if ($Key) { $keyArgs = @("-i", $Key) }

Write-Host "Downloading client certificates from $RemoteHost ..." -ForegroundColor Cyan

$files = @("ca.crt", "client.crt", "client.key")
foreach ($f in $files) {
    scp -o StrictHostKeyChecking=no @keyArgs "${RemoteHost}:/opt/db-router/certs/$f" "$CertDir/$f"
    if ($LASTEXITCODE -ne 0) { Write-Error "Failed to download $f"; exit 1 }
}

Write-Host ""
Write-Host "Certificates saved to: $CertDir" -ForegroundColor Green
Get-ChildItem $CertDir | Format-Table Name, Length, LastWriteTime
Write-Host "Ready to test. Run:"
Write-Host "  .\test-grpc.ps1 -Host grpc.db.0.xeze.org"
