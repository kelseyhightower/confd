$DebugPreference = "Continue"

$tempDir = "$env:TEMP\confd"
$confDir = "$tempDir\confdir"

function Compare-TestResult {
    param(
        [string]$Expected,
        [string]$Actual
    )

    $diff = Compare-Object `
        (Get-Content "$PSScriptRoot\expect\$Expected") `
        (Get-Content "$tempDir\$Actual")
    if ($diff) {
        throw $diff
    }
}

Write-Debug "Creating temp dir [$tempDir]"
New-Item -ItemType Directory -Path $tempDir -Force
Write-Debug "Creating conf dir [$confDir]"
New-Item -ItemType Directory -Path $confDir -Force

Copy-Item -Path "$PSScriptRoot\confdir\templates" `
    -Destination "$confDir" -Recurse -Force
New-Item -ItemType Directory -Path "$confDir\conf.d" -Force
foreach ($conf in (Get-ChildItem "$PSScriptRoot\confdir\conf.d")) {
    Get-Content -Path $conf.FullName |
        ForEach-Object {$_ -replace '/tmp/', "$($tempDir -replace '\\', '/')/" } |
        Set-Content -Path "$confDir\conf.d\$conf"
}

try {
    Write-Debug "Running integration tests..."
    foreach ($test in (Get-ChildItem "$PSScriptRoot\*\test.ps1")) {
        Write-Debug "Running $($test.FullName)"
        & $test.FullName

        Compare-TestResult -Expected "basic.conf" -Actual "confd-basic-test.conf"
        Compare-TestResult -Expected "exists-test.conf" -Actual "confd-exists-test.conf"
        Compare-TestResult -Expected "iteration.conf" -Actual "confd-iteration-test.conf"
    }
}
finally {
    Remove-Item -Recurse -Force $tempDir
}