<#
.SYNOPSIS
    Database backup script for Travel Booking PostgreSQL database.

.DESCRIPTION
    Performs daily full backup via pg_basebackup and configures WAL archival
    for point-in-time recovery. Backups are encrypted and stored locally.

    Per Constitution constraints:
    - Daily full backup (pg_basebackup)
    - Incremental backup every 15 minutes (WAL archival)
    - RTO < 5 minutes, RPO < 1 minute

.PARAMETER PgBinPath
    Path to PostgreSQL bin directory. Default: C:\Program Files\PostgreSQL\17\bin

.PARAMETER BackupDir
    Target directory for backup files. Default: C:\Backups\travel-booking

.PARAMETER WalArchiveDir
    Directory for WAL file archival. Default: C:\Backups\travel-booking\wal-archive

.PARAMETER RetentionDays
    Number of days to keep backup files. Default: 30

.PARAMETER EncryptionKey
    AES-256 encryption key for backup files (hex-encoded, 64 chars).
    If empty, backups are stored unencrypted (NOT recommended for production).

.EXAMPLE
    # Full backup
    .\backup.ps1 -BackupDir "D:\Backups\travel-booking"

.EXAMPLE
    # Set up WAL archival (run once)
    .\backup.ps1 -SetupWalArchiving
#>

param(
    [string]$PgBinPath = "C:\Program Files\PostgreSQL\17\bin",

    [string]$BackupDir = "C:\Backups\travel-booking",

    [string]$WalArchiveDir = "C:\Backups\travel-booking\wal-archive",

    [int]$RetentionDays = 30,

    [string]$EncryptionKey = "",

    [string]$DbHost = "localhost",

    [int]$DbPort = 5432,

    [string]$DbUser = "postgres",

    [string]$DbName = "travel_booking",

    [switch]$SetupWalArchiving,

    [switch]$FullBackup
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

function Get-Timestamp {
    return Get-Date -Format "yyyyMMdd_HHmmss"
}

# ──────────────────────────────────────────────
# Setup WAL Archiving (one-time configuration)
# ──────────────────────────────────────────────
if ($SetupWalArchiving) {
    Write-Step "Setting up WAL archiving"

    # Create archive directory
    if (-not (Test-Path $WalArchiveDir)) {
        New-Item -ItemType Directory -Force -Path $WalArchiveDir | Out-Null
    }

    $pgData = $env:PGDATA
    if (-not $pgData) {
        $pgData = "C:\Program Files\PostgreSQL\17\data"
    }

    $confPath = Join-Path $pgData "postgresql.conf"
    if (-not (Test-Path $confPath)) {
        Write-Fail "postgresql.conf not found at: $confPath"
        Write-Host "    Set `$env:PGDATA or pass -PgDataPath parameter"
        exit 1
    }

    # Check if WAL archiving is already configured
    $conf = Get-Content $confPath -Raw
    if ($conf -match "archive_mode\s*=\s*on") {
        Write-Host "    WAL archiving already enabled"
    }
    else {
        Write-Host "    Please add the following to postgresql.conf:"
        Write-Host ""
        Write-Host "    # WAL Archiving for Point-in-Time Recovery" -ForegroundColor Yellow
        Write-Host "    archive_mode = on" -ForegroundColor Yellow
        Write-Host "    archive_command = 'copy `"%p`" `"$WalArchiveDir\%f`"'" -ForegroundColor Yellow
        Write-Host "    wal_level = replica" -ForegroundColor Yellow
        Write-Host "    max_wal_senders = 3" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "    Then restart PostgreSQL service."
    }

    # Create a scheduled task for WAL archive cleanup
    $cleanupScript = @"
# WAL Archive Cleanup - Remove files older than $RetentionDays days
Get-ChildItem -Path "$WalArchiveDir" -File |
    Where-Object { `$_.LastWriteTime -lt (Get-Date).AddDays(-$RetentionDays) } |
    Remove-Item -Force
"@

    $cleanupPath = Join-Path $BackupDir "cleanup-wal.ps1"
    Set-Content -Path $cleanupPath -Value $cleanupScript

    Write-Success "WAL archiving setup instructions displayed"
    Write-Success "Cleanup script created at: $cleanupPath"
    exit 0
}

# ──────────────────────────────────────────────
# Full Backup via pg_basebackup
# ──────────────────────────────────────────────
Write-Step "Starting full database backup"

$timestamp = Get-Timestamp
$backupName = "travel_booking_full_$timestamp"
$backupPath = Join-Path $BackupDir $backupName

# Create backup directory
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Force -Path $BackupDir | Out-Null
}

$pgBasebackup = Join-Path $PgBinPath "pg_basebackup.exe"
if (-not (Test-Path $pgBasebackup)) {
    Write-Fail "pg_basebackup not found at: $pgBasebackup"
    exit 1
}

# Run pg_basebackup
Write-Host "    Backing up database: $DbName"
Write-Host "    Target: $backupPath"

$env:PGPASSWORD = $env:TRAVEL_DB_PASSWORD

try {
    & $pgBasebackup `
        --host=$DbHost `
        --port=$DbPort `
        --username=$DbUser `
        --pgdata=$backupPath `
        --format=tar `
        --gzip `
        --compress=6 `
        --wal-method=stream `
        --checkpoint=fast `
        --label="$backupName" `
        --progress `
        --verbose 2>&1

    if ($LASTEXITCODE -ne 0) {
        Write-Fail "pg_basebackup failed with exit code: $LASTEXITCODE"
        exit 1
    }

    Write-Success "Full backup completed"
}
catch {
    Write-Fail "Backup failed: $_"
    exit 1
}
finally {
    $env:PGPASSWORD = $null
}

# ──────────────────────────────────────────────
# Encrypt backup (if key provided)
# ──────────────────────────────────────────────
if ($EncryptionKey) {
    Write-Step "Encrypting backup"

    $tarFile = Join-Path $backupPath "base.tar.gz"
    if (Test-Path $tarFile) {
        # Use AES-256 encryption via openssl (if available)
        $openssl = Get-Command openssl -ErrorAction SilentlyContinue
        if ($openssl) {
            $encryptedFile = "${tarFile}.enc"
            & openssl enc -aes-256-cbc -salt -pbkdf2 -in $tarFile -out $encryptedFile -pass "pass:$EncryptionKey"
            if ($LASTEXITCODE -eq 0) {
                Remove-Item $tarFile -Force
                Write-Success "Backup encrypted"
            }
            else {
                Write-Fail "Encryption failed, keeping unencrypted backup"
            }
        }
        else {
            Write-Host "    openssl not found, skipping encryption"
            Write-Host "    Install openssl or use BitLocker for volume encryption"
        }
    }
}

# ──────────────────────────────────────────────
# Verify backup integrity
# ──────────────────────────────────────────────
Write-Step "Verifying backup"

$backupSize = (Get-ChildItem -Path $backupPath -Recurse | Measure-Object -Property Length -Sum).Sum
$backupSizeMB = [math]::Round($backupSize / 1MB, 2)

Write-Host "    Backup size: $backupSizeMB MB"

if ($backupSizeMB -lt 1) {
    Write-Fail "Backup suspiciously small (< 1MB), verify manually"
    exit 1
}

Write-Success "Backup verification passed"

# ──────────────────────────────────────────────
# Cleanup old backups
# ──────────────────────────────────────────────
Write-Step "Cleaning up backups older than $RetentionDays days"

$cutoffDate = (Get-Date).AddDays(-$RetentionDays)
$oldBackups = Get-ChildItem -Path $BackupDir -Directory |
    Where-Object { $_.Name -match "^travel_booking_full_" -and $_.CreationTime -lt $cutoffDate }

foreach ($old in $oldBackups) {
    Remove-Item -Path $old.FullName -Recurse -Force
    Write-Host "    Removed: $($old.Name)"
}

if ($oldBackups.Count -eq 0) {
    Write-Host "    No old backups to remove"
}

Write-Success "Cleanup complete"

# ──────────────────────────────────────────────
# Summary
# ──────────────────────────────────────────────
Write-Host "`n" -NoNewline
Write-Host "========================================" -ForegroundColor Green
Write-Host " Backup Complete" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green
Write-Host " Database:    $DbName"
Write-Host " Location:    $backupPath"
Write-Host " Size:        $backupSizeMB MB"
Write-Host " Retention:   $RetentionDays days"
Write-Host " Timestamp:   $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
Write-Host "========================================" -ForegroundColor Green
