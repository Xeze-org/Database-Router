# Quick gRPC test using grpcurl.
# Usage: .\test-grpc.ps1 -GrpcHost grpc.db.0.xeze.org [-NoTLS]
param(
    [Parameter(Mandatory=$true)]
    [string]$GrpcHost,
    [int]$Port = 443,
    [switch]$NoTLS
)

$CertDir = Join-Path $PSScriptRoot "certs"

if (-not (Get-Command grpcurl -ErrorAction SilentlyContinue)) {
    Write-Error "grpcurl is not installed. Install: https://github.com/fullstorydev/grpcurl#installation"
    exit 1
}

$tlsArgs = @()
if (-not $NoTLS -and (Test-Path "$CertDir/client.crt")) {
    Write-Host "Using mTLS certificates from $CertDir" -ForegroundColor Cyan
    $tlsArgs = @("-cacert", "$CertDir/ca.crt", "-cert", "$CertDir/client.crt", "-key", "$CertDir/client.key")
} elseif ($NoTLS) {
    Write-Host "Skipping client certificates (plain TLS)" -ForegroundColor Yellow
} else {
    Write-Host "WARNING: No client certs found. Run fetch-certs.ps1 first." -ForegroundColor Yellow
}

$endpoint = "${GrpcHost}:${Port}"

$tests = @(
    @{ Name = "List Services";           Args = @($endpoint, "list") },
    @{ Name = "Health Check All";        Args = @($endpoint, "dbrouter.HealthService/CheckAll") },
    @{ Name = "Test PostgreSQL";         Args = @($endpoint, "dbrouter.HealthService/CheckPostgres") },
    @{ Name = "Test MongoDB";            Args = @($endpoint, "dbrouter.HealthService/CheckMongo") },
    @{ Name = "Test Redis";              Args = @($endpoint, "dbrouter.HealthService/CheckRedis") },
    @{ Name = "List Postgres Databases"; Args = @($endpoint, "dbrouter.PostgresService/ListDatabases") },
    @{ Name = "List Redis Keys";         Args = @("-d", '{"pattern":"*"}', $endpoint, "dbrouter.RedisService/ListKeys") }
)

foreach ($t in $tests) {
    Write-Host "`n=== $($t.Name) ===" -ForegroundColor Green
    & grpcurl @tlsArgs @($t.Args) 2>&1
}

Write-Host "`nDone." -ForegroundColor Green
