Set oShell = CreateObject("WScript.Shell")
Set oEnv = oShell.Environment("USER")
Set oFS = CreateObject("Scripting.FileSystemObject")

Dim baseUrl
baseUrl = "http://127.0.0.1:8181"

MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

oEnv("AGENT_DEBUG_OVERRIDE_PROXY_URL") = baseUrl
oEnv("GITHUB_COPILOT_OVERRIDE_PROXY_URL") = baseUrl
oEnv("AGENT_DEBUG_OVERRIDE_CAPI_URL") = baseUrl & "/v1"
oEnv("GITHUB_COPILOT_OVERRIDE_CAPI_URL") = baseUrl & "/v1"

MsgBox "Done."
