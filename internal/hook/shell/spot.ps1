# spot shell hook for PowerShell. Add this line to your $PROFILE:
#   Invoke-Expression (& spot init pwsh | Out-String)
# Then use $global:SPOT_FACE in your prompt, or `spot init pwsh --prompt`.

$global:_SpotHistId = (Get-History -Count 1).Id
$global:SPOT_FACE = & '{{SPOTPS}}' prompt --shell pwsh
$global:_SpotPrompt = $function:prompt
$global:_SpotShowFace = {{SHOWFACE}}

function global:prompt {
    $ok = $?
    $code = $global:LASTEXITCODE
    $h = Get-History -Count 1
    if ($h -and $h.Id -ne $global:_SpotHistId) {
        $global:_SpotHistId = $h.Id
        $exit = 0
        if (-not $ok) { $exit = 1; if ($code) { $exit = $code } }
        $secs = ($h.EndExecutionTime - $h.StartExecutionTime).TotalSeconds.ToString([cultureinfo]::InvariantCulture)
        $global:SPOT_FACE = & '{{SPOTPS}}' _event --shell pwsh --exit $exit --dur $secs -- $h.CommandLine
    } else {
        $global:SPOT_FACE = & '{{SPOTPS}}' prompt --shell pwsh
    }
    $global:LASTEXITCODE = $code
    $rest = & $global:_SpotPrompt
    if ($global:_SpotShowFace) { return "$($global:SPOT_FACE) $rest" }
    return $rest
}
