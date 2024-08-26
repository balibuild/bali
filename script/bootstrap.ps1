#!/usr/bin/env pwsh

Write-Host -ForegroundColor Green "Bootstrap bali"
$TopLevel = Split-Path -Path $PSScriptRoot
$SOURCE_DIR = Join-Path $TopLevel -ChildPath "cmd/bali"

$BALI_EXE = "$SOURCE_DIR/bali"
$BALI_EXE_STAGE0 = "$TopLevel/bali.out"
if ($PSEdition -eq "Desktop" -or $IsWindows) {
    $BALI_EXE += ".exe"
    $BALI_EXE_STAGE0 += ".exe"
}

$ps = Start-Process -FilePath "go" -WorkingDirectory $SOURCE_DIR -ArgumentList "build" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Copy-Item -Force -Path $BALI_EXE -Destination $BALI_EXE_STAGE0

$ps = Start-Process -FilePath $BALI_EXE_STAGE0 -WorkingDirectory $TopLevel -ArgumentList "--pack=zip" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Write-Host -ForegroundColor Green "bootstrap bali success"
