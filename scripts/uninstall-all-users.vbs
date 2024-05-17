If Not WScript.Arguments.Named.Exists("elevate") Then
  CreateObject("Shell.Application").ShellExecute WScript.FullName, """" & WScript.ScriptFullName & """ /elevate", "", "runas", 10
  WScript.Quit
End If

MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

Sub RemoveEnv(env)
	On Error Resume Next

	env.Remove("AGENT_DEBUG_OVERRIDE_PROXY_URL")
	env.Remove("GITHUB_COPILOT_OVERRIDE_PROXY_URL")
    env.Remove("AGENT_DEBUG_OVERRIDE_CAPI_URL")
    env.Remove("GITHUB_COPILOT_OVERRIDE_CAPI_URL")
End Sub

Set oShell = CreateObject("WScript.Shell")

RemoveEnv oShell.Environment("USER")
RemoveEnv oShell.Environment("SYSTEM")

MsgBox "Done."
