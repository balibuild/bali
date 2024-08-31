#!/usr/bin/env pwsh

Write-Host -ForegroundColor Green "bali: compiling bali ..."
$SOURCE_DIR = Split-Path -Path $PSScriptRoot
$BALI_SOURCE_DIR = Join-Path $SOURCE_DIR -ChildPath "cmd/bali"

$BALI_EXE = Join-Path $BALI_SOURCE_DIR -ChildPath "bali"
$BALI_STAGE0_EXE = Join-Path -Path $SOURCE_DIR -ChildPath "bali"
if ($PSEdition -eq "Desktop" -or $IsWindows) {
    $BALI_EXE += ".exe"
    $BALI_STAGE0_EXE += ".exe"
}

$ps = Start-Process -FilePath "go" -WorkingDirectory $BALI_SOURCE_DIR -ArgumentList "build" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Copy-Item -Force -Path $BALI_EXE -Destination $BALI_STAGE0_EXE

Write-Host -ForegroundColor Green "bali: create zip package ..."

$ps = Start-Process -FilePath $BALI_STAGE0_EXE -WorkingDirectory $SOURCE_DIR -ArgumentList "--pack=zip" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Write-Host -ForegroundColor Green "bali: bootstrap success"
