#!/usr/bin/env pwsh

Write-Host -ForegroundColor Green "Bootstrap baligo"
$TopLevel = Split-Path -Path $PSScriptRoot
$SrcDir = Join-Path $TopLevel -ChildPath "cmd/bali"

$BaliFile = "$SrcDir/bali"
$BailBin = "$TopLevel/bali.out"
if ($PSEdition -eq "Desktop" -or $IsWindows) {
    $BaliFile += ".exe"
    $BailBin += ".exe"
}

$ps = Start-Process -FilePath "go" -WorkingDirectory $SrcDir -ArgumentList "build" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Copy-Item -Force -Path $BaliFile -Destination $BailBin

$ps = Start-Process -FilePath $BailBin -WorkingDirectory $TopLevel -ArgumentList "-z" -PassThru -Wait -NoNewWindow
if ($ps.ExitCode -ne 0) {
    Exit $ps.ExitCode
}

Write-Host -ForegroundColor Green "bootstrap bali success"
