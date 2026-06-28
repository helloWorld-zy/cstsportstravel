<#
.SYNOPSIS
    Deployment script for Travel Booking API Server on Windows Server.

.DESCRIPTION
    Downloads the build artifact, stops the WinSW service, replaces the binary,
    starts the service, and verifies health.

.PARAMETER ArtifactPath
    Path to the deployment artifact directory (contains travel-api.exe).

.PARAMETER InstallPath
    Target installation directory. Default: C:\Services\travel-api

.PARAMETER ServiceName
    WinSW service name. Default: travel-api

.PARAMETER HealthUrl
    Health check URL. Default: http://localhost:8088/health

.PARAMETER HealthRetries
    Number of health check retries. Default: 30

.PARAMETER HealthInterval
    Seconds between health checks. Default: 2

.EXAMPLE
    .\deploy.ps1 -ArtifactPath ".\artifact" -InstallPath "C:\Services\travel-api"
#>

param(
    [Parameter(Mandatory = $true)]
    [string]$ArtifactPath,

    [string]$InstallPath = "C:\Services\travel-api",

    [string]$ServiceName = "travel-api",

    [string]$HealthUrl = "http://localhost:8088/health",

    [int]$HealthRetries = 30,

    [int]$HealthInterval = 2
)

$ErrorActionPreference = "Stop"

function Write-Step {
    param([string]$Message)
    Write-Host "`n==> $Message" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "    [OK] $Message" -ForegroundColor Green
}

function Write-Fail {
    param([string]$Message)
    Write-Host "    [FAIL] $Message" -ForegroundColor Red
}

function Test-Health {
    param(
        [string]$Url,
        [int]$Retries,
        [int]$Interval
    )

    for ($i = 1; $i -le $Retries; $i++) {
        try {
            $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 5
            if ($response.StatusCode -eq 200) {
                return $true
            }
        }
        catch {
            # Service not ready yet
        }

        if ($i -lt $Retries) {
            Write-Host "    Health check attempt $i/$Retries failed, retrying in ${Interval}s..."
            Start-Sleep -Seconds $Interval
        }
    }

    return $false
}

# ──────────────────────────────────────────────
# Pre-flight checks
# ──────────────────────────────────────────────
Write-Step "Pre-flight checks"

if (-not (Test-Path $ArtifactPath)) {
    Write-Fail "Artifact path not found: $ArtifactPath"
    exit 1
}

$binaryPath = Join-Path $ArtifactPath "travel-api.exe"
if (-not (Test-Path $binaryPath)) {
    Write-Fail "Binary not found: $binaryPath"
    exit 1
}

Write-Success "Artifact validated"

# ──────────────────────────────────────────────
# Create backup of current deployment
# ──────────────────────────────────────────────
Write-Step "Creating backup"

if (Test-Path $InstallPath) {
    $backupPath = "${InstallPath}_backup_$(Get-Date -Format 'yyyyMMdd_HHmmss')"
    Copy-Item -Path $InstallPath -Destination $backupPath -Recurse -Force
    Write-Success "Backup created at $backupPath"
}
else {
    Write-Host "    No existing installation found, skipping backup"
}

# ──────────────────────────────────────────────
# Stop the service
# ──────────────────────────────────────────────
Write-Step "Stopping service: $ServiceName"

$service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
if ($service -and $service.Status -eq "Running") {
    Stop-Service -Name $ServiceName -Force -Timeout 30
    Write-Success "Service stopped"
}
else {
    Write-Host "    Service not running, skipping stop"
}

# ──────────────────────────────────────────────
# Deploy new binary and files
# ──────────────────────────────────────────────
Write-Step "Deploying to $InstallPath"

# Create installation directory if needed
if (-not (Test-Path $InstallPath)) {
    New-Item -ItemType Directory -Force -Path $InstallPath | Out-Null
}

# Copy binary
Copy-Item -Path $binaryPath -Destination (Join-Path $InstallPath "travel-api.exe") -Force
Write-Success "Binary deployed"

# Copy configs
$configSrc = Join-Path $ArtifactPath "configs"
if (Test-Path $configSrc) {
    $configDst = Join-Path $InstallPath "configs"
    if (-not (Test-Path $configDst)) {
        New-Item -ItemType Directory -Force -Path $configDst | Out-Null
    }
    Copy-Item -Path "$configSrc\*" -Destination $configDst -Recurse -Force
    Write-Success "Configs deployed"
}

# Copy migrations
$migrationSrc = Join-Path $ArtifactPath "migrations"
if (Test-Path $migrationSrc) {
    $migrationDst = Join-Path $InstallPath "migrations"
    if (-not (Test-Path $migrationDst)) {
        New-Item -ItemType Directory -Force -Path $migrationDst | Out-Null
    }
    Copy-Item -Path "$migrationSrc\*" -Destination $migrationDst -Recurse -Force
    Write-Success "Migrations deployed"
}

# Copy WinSW config
$winswSrc = Join-Path $ArtifactPath "winsw"
if (Test-Path $winswSrc) {
    Copy-Item -Path "$winswSrc\*" -Destination $InstallPath -Recurse -Force
    Write-Success "WinSW config deployed"
}

# Copy deploy scripts
$scriptsSrc = Join-Path $ArtifactPath "scripts"
if (Test-Path $scriptsSrc) {
    $scriptsDst = Join-Path $InstallPath "scripts"
    if (-not (Test-Path $scriptsDst)) {
        New-Item -ItemType Directory -Force -Path $scriptsDst | Out-Null
    }
    Copy-Item -Path "$scriptsSrc\*" -Destination $scriptsDst -Recurse -Force
    Write-Success "Deploy scripts deployed"
}

# ──────────────────────────────────────────────
# Start the service
# ──────────────────────────────────────────────
Write-Step "Starting service: $ServiceName"

Start-Service -Name $ServiceName
Write-Success "Service start command sent"

# ──────────────────────────────────────────────
# Health check verification
# ──────────────────────────────────────────────
Write-Step "Verifying health at $HealthUrl"

if (Test-Health -Url $HealthUrl -Retries $HealthRetries -Interval $HealthInterval) {
    Write-Success "Health check passed!"
}
else {
    Write-Fail "Health check failed after $HealthRetries attempts"

    # Rollback
    Write-Step "Rolling back to backup"
    if (Test-Path $backupPath) {
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $InstallPath -Recurse -Force
        Rename-Item -Path $backupPath -NewName $InstallPath
        Start-Service -Name $ServiceName
        Write-Success "Rollback complete"
    }

    exit 1
}

# ──────────────────────────────────────────────
# Cleanup old backups (keep last 5)
# ──────────────────────────────────────────────
Write-Step "Cleaning up old backups"

$parentDir = Split-Path $InstallPath -Parent
$backups = Get-ChildItem -Path $parentDir -Filter "${ServiceName}_backup_*" -Directory |
    Sort-Object CreationTime -Descending

if ($backups.Count -gt 5) {
    $backups | Select-Object -Skip 5 | ForEach-Object {
        Remove-Item -Path $_.FullName -Recurse -Force
        Write-Host "    Removed old backup: $($_.Name)"
    }
}

Write-Success "Cleanup complete"

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
Write-Host "`n" -NoNewline
Write-Host "========================================" -ForegroundColor Green
Write-Host " Deployment Complete" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host " Service:     $ServiceName"
Write-Host " Location:    $InstallPath"
Write-Host " Health:      $HealthUrl"
Write-Host " Timestamp:   $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Host "========================================" -ForegroundColor Green
