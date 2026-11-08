Set-Location 'C:\whats4linux'
Remove-Item .\fyne-run.log,.\fyne-run.err.log -ErrorAction SilentlyContinue
$env:WHATS4LINUX_OPEN_JID='6282134544728@s.whatsapp.net'
$p = Start-Process -FilePath 'C:\whats4linux\whats4linux-fyne-win.exe' -WorkingDirectory 'C:\whats4linux' -RedirectStandardOutput 'C:\whats4linux\fyne-run.log' -RedirectStandardError 'C:\whats4linux\fyne-run.err.log' -PassThru
Write-Output ("STARTED PID=" + $p.Id)
Start-Sleep 6
Get-Process -Name 'whats4linux-fyne-win' -ErrorAction SilentlyContinue | Select-Object Id,Responding,MainWindowTitle | Format-List
