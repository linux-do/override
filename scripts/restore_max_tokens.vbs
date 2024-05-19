' VBScript to recovery max tokens
MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

Const ForReading = 1
Const ForWriting = 2

' Subpath of the file to be recovery
subpath = "dist\extension.js"

' Iterate over all github copilot directories
Set objFSO = CreateObject("Scripting.FileSystemObject")
Set objShell = CreateObject("WScript.Shell")
Set colExtensions = objFSO.GetFolder(objShell.ExpandEnvironmentStrings("%USERPROFILE%") & "\.vscode\extensions").SubFolders

For Each objExtension In colExtensions
    extension_path = objExtension.Path & "\" & subpath
    backupfile = extension_path & ".bak"
    
    If objFSO.FileExists(backupfile) Then
        ' Delete if exist extension file
        If objFSO.FileExists(extension_path) Then
            objFSO.DeleteFile extension_path, True
        End If
        
        ' Replace
        objFSO.MoveFile backupfile, extension_path
    End If
Next

MsgBox "Restore max tokens to default successed"
