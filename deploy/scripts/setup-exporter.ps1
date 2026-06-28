<#
.SYNOPSIS
    Install and configure Windows Exporter for Prometheus metrics collection.

.DESCRIPTION
    Downloads and installs Windows Exporter MSI package, configures it to
    expose system-level metrics (CPU, memory, disk, network) on port 9182
    for collection by Prometheus.

    Per PRD §10.4.2, infrastructure metrics are collected every 15s.

.PARAMETER Version
    Windows Exporter version to install. Default: 0.25.1

.PARAMETER Port
    Metrics exposition port. Default: 9182

.PARAMETER InstallDir
    Installation directory. Default: C:\Program Files\windows_exporter

.EXAMPLE
    .\setup-exporter.ps1
    .\setup-exporter.ps1 -Version "0.25.1" -Port 9182
#>

param(
    [string]$Version = "0.25.1",

    [int]$Port = 9182,

    [string]$InstallDir = "C:\Program Files\windows_exporter"
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

# ──────────────────────────────────────────────
# Check if already installed
# ──────────────────────────────────────────────
Write-Step "Checking existing installation"

$service = Get-Service -Name "windows_exporter" -ErrorAction SilentlyContinue
if ($service) {
    Write-Host "    Windows Exporter is already installed"
    Write-Host "    Current status: $($service.Status)"

    $currentVersion = & "C:\Program Files\windows_exporter\windows_exporter.exe" --version 2>&1
    Write-Host "    Current version: $currentVersion"

    $response = Read-Host "    Reinstall? (y/N)"
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Host "    Skipping installation"
        exit 0
    }
}

# ──────────────────────────────────────────────
# Download Windows Exporter MSI
# ──────────────────────────────────────────────
Write-Step "Downloading Windows Exporter v$Version"

$msiFile = "windows_exporter-${Version}-amd64.msi"
$downloadUrl = "https://github.com/prometheus-community/windows_exporter/releases/download/v${Version}/${msiFile}"
$downloadPath = Join-Path $env:TEMP $msiFile

try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $downloadUrl -OutFile $downloadPath -UseBasicParsing
    Write-Success "Downloaded to $downloadPath"
}
catch {
    Write-Fail "Download failed: $_"
    Write-Host "    Manual download: $downloadUrl"
    exit 1
}

# ──────────────────────────────────────────────
# Install MSI
# ──────────────────────────────────────────────
Write-Step "Installing Windows Exporter"

# Collectors to enable (covers all PRD §10.4.2 infrastructure metrics)
$collectors = @(
    "cpu",           # CPU usage (PRD: >80% for 10min → P2)
    "cs",            # Computer system info
    "logical_disk",  # Disk usage (PRD: >80% → P2)
    "memory",        # Memory usage (PRD: >85% for 10min → P2)
    "net",           # Network metrics
    "os",            # OS info
    "process",       # Process metrics
    "service",       # Service state
    "system",        # System metrics
    "textfile"       # Custom metrics from text files
)

$collectorList = $collectors -join ","

$msiArgs = @(
    "/i", $downloadPath,
    "/qn",  # Quiet install, no UI
    "ENABLED_COLLECTORS=$collectorList",
    "LISTEN_PORT=$Port",
    "INSTALLDIR=`"$InstallDir`""
)

try {
    $process = Start-Process msiexec.exe -ArgumentList $msiArgs -Wait -PassThru -NoNewWindow

    if ($process.ExitCode -ne 0) {
        Write-Fail "MSI installation failed with exit code: $($process.ExitCode)"
        exit 1
    }

    Write-Success "Windows Exporter installed"
}
catch {
    Write-Fail "Installation failed: $_"
    exit 1
}

# ──────────────────────────────────────────────
# Verify service is running
# ──────────────────────────────────────────────
Write-Step "Verifying service"

Start-Sleep -Seconds 3

$service = Get-Service -Name "windows_exporter" -ErrorAction SilentlyContinue
if ($service) {
    if ($service.Status -ne "Running") {
        Start-Service -Name "windows_exporter"
        Start-Sleep -Seconds 2
    }

    $service = Get-Service -Name "windows_exporter"
    Write-Success "Service status: $($service.Status)"
}
else {
    Write-Fail "Service not found after installation"
    exit 1
}

# ──────────────────────────────────────────────
# Verify metrics endpoint
# ──────────────────────────────────────────────
Write-Step "Verifying metrics endpoint at http://localhost:${Port}/metrics"

try {
    $response = Invoke-WebRequest -Uri "http://localhost:${Port}/metrics" -UseBasicParsing -TimeoutSec 10

    if ($response.StatusCode -eq 200 -and $response.Content -match "windows_cpu_time_total") {
        Write-Success "Metrics endpoint responding correctly"
    }
    else {
        Write-Fail "Metrics endpoint returned unexpected response"
    }
}
catch {
    Write-Fail "Cannot reach metrics endpoint: $_"
    Write-Host "    Check Windows Firewall rules for port $Port"
}

# ──────────────────────────────────────────────
# Windows Firewall rule
# ──────────────────────────────────────────────
Write-Step "Configuring Windows Firewall"

$ruleName = "Windows Exporter (TCP $Port)"

try {
    $existingRule = Get-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue
    if (-not $existingRule) {
        New-NetFirewallRule `
            -DisplayName $ruleName `
            -Direction Inbound `
            -Protocol TCP `
            -LocalPort $Port `
            -Action Allow `
            -Profile Domain,Private `
            -Description "Allow Prometheus to scrape Windows Exporter metrics" | Out-Null

        Write-Success "Firewall rule created"
    }
    else {
        Write-Host "    Firewall rule already exists"
    }
}
catch {
    Write-Host "    Could not configure firewall (may need admin privileges)"
    Write-Host "    Manual: New-NetFirewallRule -DisplayName '$ruleName' -Direction Inbound -Protocol TCP -LocalPort $Port -Action Allow"
}

# ──────────────────────────────────────────────
# Cleanup
# ──────────────────────────────────────────────
Remove-Item $downloadPath -Force -ErrorAction SilentlyContinue

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
Write-Host "`n" -NoNewline
Write-Host "========================================" -ForegroundColor Green
Write-Host " Windows Exporter Installed" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host " Version:     $Version"
Write-Host " Port:        $Port"
Write-Host " Endpoint:    http://localhost:${Port}/metrics"
Write-Host " Service:     windows_exporter"
Write-Host " Collectors:  $collectorList"
Write-Host ""
Write-Host " Add to prometheus.yml scrape_configs:" -ForegroundColor Yellow
Write-Host "   - job_name: 'windows'" -ForegroundColor Yellow
Write-Host "     static_configs:" -ForegroundColor Yellow
Write-Host "       - targets: ['localhost:${Port}']" -ForegroundColor Yellow
Write-Host "     scrape_interval: 15s" -ForegroundColor Yellow
Write-Host "========================================" -ForegroundColor Green
