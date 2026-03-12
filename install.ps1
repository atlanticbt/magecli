# install.ps1 — Download and install the latest magecli binary on Windows.
#
# Usage:
#   irm https://raw.githubusercontent.com/atlanticbt/magecli/main/install.ps1 | iex
#
#   # Or with options:
#   & { param($Dir, $Version) irm https://raw.githubusercontent.com/atlanticbt/magecli/main/install.ps1 | iex } -Dir "$HOME\bin" -Version "v1.2.0"
#
[CmdletBinding()]
param(
    [string]$Dir = "$env:LOCALAPPDATA\magecli\bin",
    [string]$Version = ""
)

$ErrorActionPreference = "Stop"
$Repo = "atlanticbt/magecli"
$Binary = "magecli"

# ---------- Detect architecture ----------

$Arch = switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
    "X64"   { "amd64" }
    "Arm64" { "arm64" }
    default { throw "Unsupported architecture: $_" }
}

Write-Host "Detected platform: windows/$Arch" -ForegroundColor Cyan

# ---------- Resolve version ----------

if (-not $Version) {
    Write-Host "Fetching latest release..."
    $release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
    $Version = $release.tag_name
    if (-not $Version) {
        throw "Could not determine latest version. Check https://github.com/$Repo/releases"
    }
}

$VersionNum = $Version.TrimStart("v")
Write-Host "Installing $Binary $Version..." -ForegroundColor Cyan

# ---------- Download ----------

$Archive = "${Binary}_${VersionNum}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Version/$Archive"

$TmpDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null

try {
    $ArchivePath = Join-Path $TmpDir $Archive
    Write-Host "Downloading $Url..."
    Invoke-WebRequest -Uri $Url -OutFile $ArchivePath -UseBasicParsing

    # Verify checksum if available
    $ChecksumUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"
    try {
        $ChecksumPath = Join-Path $TmpDir "checksums.txt"
        Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
        $checksumLine = Get-Content $ChecksumPath | Where-Object { $_ -match $Archive }
        if ($checksumLine) {
            $expected = ($checksumLine -split '\s+')[0]
            $actual = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
            if ($expected -ne $actual) {
                throw "Checksum mismatch!`n  expected: $expected`n  actual:   $actual"
            }
            Write-Host "Checksum verified." -ForegroundColor Green
        }
    } catch [System.Net.WebException] {
        # Checksums file not available, skip verification
    }

    # Extract
    Expand-Archive -Path $ArchivePath -DestinationPath $TmpDir -Force

    # Install
    if (-not (Test-Path $Dir)) {
        New-Item -ItemType Directory -Path $Dir -Force | Out-Null
    }

    $DestPath = Join-Path $Dir "$Binary.exe"
    Move-Item -Path (Join-Path $TmpDir "$Binary.exe") -Destination $DestPath -Force

    Write-Host ""
    Write-Host "$Binary $Version installed to $DestPath" -ForegroundColor Green
    Write-Host ""
    Write-Host "Run '$Binary --version' to verify."

    # Check PATH
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$Dir*") {
        Write-Host ""
        Write-Host "Adding $Dir to your user PATH..." -ForegroundColor Yellow
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$Dir", "User")
        $env:Path = "$env:Path;$Dir"
        Write-Host "Done. Restart your terminal for the PATH change to take effect in new sessions."
    }

} finally {
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
