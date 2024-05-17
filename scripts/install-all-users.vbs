If Not WScript.Arguments.Named.Exists("elevate") Then
  CreateObject("Shell.Application").ShellExecute WScript.FullName, """" & WScript.ScriptFullName & """ /elevate", "", "runas", 10
  WScript.Quit
End If

Set oShell = CreateObject("WScript.Shell")
Set oEnvSystem = oShell.Environment("SYSTEM")
Set oFS = CreateObject("Scripting.FileSystemObject")

MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

Dim baseUrl
baseUrl = "http://127.0.0.1:8181"

Sub RemoveEnv(env)
    On Error Resume Next

    env.Remove("AGENT_DEBUG_OVERRIDE_PROXY_URL")
    env.Remove("GITHUB_COPILOT_OVERRIDE_PROXY_URL")
    env.Remove("AGENT_DEBUG_OVERRIDE_CAPI_URL")
    env.Remove("GITHUB_COPILOT_OVERRIDE_CAPI_URL")
End Sub

RemoveEnv oShell.Environment("USER")

oEnvSystem("AGENT_DEBUG_OVERRIDE_PROXY_URL") = baseUrl
oEnvSystem("GITHUB_COPILOT_OVERRIDE_PROXY_URL") = baseUrl
oEnvSystem("AGENT_DEBUG_OVERRIDE_CAPI_URL") = baseUrl & "/v1"
oEnvSystem("GITHUB_COPILOT_OVERRIDE_CAPI_URL") = baseUrl & "/v1"

MsgBox "Done."
