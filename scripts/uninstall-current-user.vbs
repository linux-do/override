Set oShell = CreateObject("WScript.Shell")
Set oEnv = oShell.Environment("USER")

MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

oEnv.Remove("AGENT_DEBUG_OVERRIDE_PROXY_URL")
oEnv.Remove("GITHUB_COPILOT_OVERRIDE_PROXY_URL")
oEnv.Remove("AGENT_DEBUG_OVERRIDE_CAPI_URL")
oEnv.Remove("GITHUB_COPILOT_OVERRIDE_CAPI_URL")

MsgBox "Done."
