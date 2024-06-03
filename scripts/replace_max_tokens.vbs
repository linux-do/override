' VBScript to change max tokens to 2048

MsgBox "It may take a few seconds to execute this script." & vbCrLf & vbCrLf & "Click 'OK' button and wait for the prompt of 'Done.' to pop up!"

Const ForReading = 1
Const ForWriting = 2

' Subpath of the file to be replaced
subpath = "dist\extension.js"

pattern = "\.maxPromptCompletionTokens\(([a-zA-Z0-9_]+),([0-9]+)\)"
replacement = ".maxPromptCompletionTokens($1,2048)"

' Iterate over all github copilot directories
Set objFSO = CreateObject("Scripting.FileSystemObject")
Set objShell = CreateObject("WScript.Shell")
Set colExtensions = objFSO.GetFolder(objShell.ExpandEnvironmentStrings("%USERPROFILE%") & "\.vscode\extensions").SubFolders

For Each objExtension In colExtensions
    extension_path = objExtension.Path & "\" & subpath
    If objFSO.FileExists(extension_path) Then
        backupfile = extension_path & ".bak"
        
        ' Delete if backup file exists
        If objFSO.FileExists(backupfile) Then
            objFSO.DeleteFile backupfile, True
        End If
        
        ' Backup
        objFSO.CopyFile extension_path, backupfile
        
        ' Do search and replace with pattern
        Set objFile = objFSO.OpenTextFile(extension_path, ForReading)
        strContent = objFile.ReadAll
        objFile.Close
        
        Set objRegEx = New RegExp
        objRegEx.Global = True
        objRegEx.IgnoreCase = True
        objRegEx.Pattern = pattern
        strContent = objRegEx.Replace(strContent, replacement)
        
        Set objFile = objFSO.OpenTextFile(extension_path, ForWriting)
        objFile.Write strContent
        objFile.Close
    End If
Next

MsgBox "Max tokens modification completed"
