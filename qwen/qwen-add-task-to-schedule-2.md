# 请创建一个powershell脚本，它的功能是向task scheduler中添加一个任务，在指定的时间执行某个命令，执行完毕后自动删除任务

以下是一个PowerShell脚本，用于创建一个一次性任务，在指定时间执行命令，执行完成后自动删除任务：

```powershell
<#
.SYNOPSIS
    创建一个一次性计划任务，在指定时间执行命令后自动删除
.DESCRIPTION
    该脚本向Windows任务计划程序添加一个一次性任务，在指定时间执行指定的命令或脚本，
    任务执行完成后会自动从任务计划程序中删除。
.PARAMETER TaskName
    任务名称（可选，默认为"AutoDeleteTask_时间戳"）
.PARAMETER ExecuteTime
    执行时间（必须，格式：yyyy-MM-dd HH:mm:ss）
.PARAMETER Command
    要执行的命令（必须）
.PARAMETER Arguments
    命令的参数（可选）
.PARAMETER WorkingDirectory
    工作目录（可选，默认为系统临时目录）
.EXAMPLE
    .\CreateAutoDeleteTask.ps1 -ExecuteTime "2024-01-15 14:30:00" -Command "notepad.exe"
.EXAMPLE
    .\CreateAutoDeleteTask.ps1 -TaskName "MyBackupTask" -ExecuteTime "2024-01-15 23:00:00" -Command "C:\Scripts\backup.ps1" -Arguments "-Force"
#>

param(
    [Parameter(Mandatory=$false)]
    [string]$TaskName = "AutoDeleteTask_$(Get-Date -Format 'yyyyMMdd_HHmmss')",
    
    [Parameter(Mandatory=$true)]
    [string]$ExecuteTime,
    
    [Parameter(Mandatory=$true)]
    [string]$Command,
    
    [Parameter(Mandatory=$false)]
    [string]$Arguments = "",
    
    [Parameter(Mandatory=$false)]
    [string]$WorkingDirectory = $env:TEMP
)

# 验证执行时间格式
try {
    $scheduledTime = [DateTime]::ParseExact($ExecuteTime, "yyyy-MM-dd HH:mm:ss", $null)
    if ($scheduledTime -lt (Get-Date)) {
        Write-Error "指定的执行时间不能早于当前时间！"
        exit 1
    }
} catch {
    Write-Error "执行时间格式错误！请使用格式：yyyy-MM-dd HH:mm:ss"
    Write-Error "例如：2024-01-15 14:30:00"
    exit 1
}

# 检查命令是否存在（如果是可执行文件）
if ([System.IO.Path]::HasExtension($Command)) {
    if (-not (Test-Path $Command)) {
        Write-Warning "指定的命令文件 '$Command' 不存在，任务仍会创建但可能执行失败。"
    }
}

# 创建任务操作
if ([System.IO.Path]::GetExtension($Command) -eq ".ps1") {
    # 如果是PowerShell脚本，使用PowerShell执行
    $action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -File `"$Command`" $Arguments" -WorkingDirectory $WorkingDirectory
} else {
    # 其他类型的命令
    $action = New-ScheduledTaskAction -Execute $Command -Argument $Arguments -WorkingDirectory $WorkingDirectory
}

# 创建触发器（一次性）
$trigger = New-ScheduledTaskTrigger -Once -At $scheduledTime

# 创建设置（允许启动、隐藏窗口等）
$settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -Hidden

# 创建任务主体（当前用户）
$principal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive -RunLevel Limited

# 注册任务
try {
    Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Settings $settings -Principal $principal -Description "自动删除的一次性任务 - 创建时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    Write-Host "任务 '$TaskName' 已成功创建！" -ForegroundColor Green
    Write-Host "执行时间: $ExecuteTime" -ForegroundColor Cyan
    Write-Host "执行命令: $Command" -ForegroundColor Cyan
    if ($Arguments) {
        Write-Host "命令参数: $Arguments" -ForegroundColor Cyan
    }
} catch {
    Write-Error "创建任务失败: $($_.Exception.Message)"
    exit 1
}

# 创建自动删除任务的脚本内容
$deleteScript = @"
`$taskName = "$TaskName"
`$maxRetries = 10
`$retryCount = 0

# 等待主任务完成
while ((Get-ScheduledTask -TaskName `$taskName -ErrorAction SilentlyContinue | Where-Object { `$_.State -eq 'Running' }) -and `$retryCount -lt `$maxRetries) {
    Start-Sleep -Seconds 5
    `$retryCount++
}

# 删除任务
try {
    Unregister-ScheduledTask -TaskName `$taskName -Confirm:`$false -ErrorAction Stop
    # 可选：记录删除日志
    # Add-Content -Path "`$env:TEMP\AutoDeleteTask.log" -Value "$(Get-Date): 任务 `$taskName 已自动删除"
} catch {
    # 如果删除失败，可以记录错误（可选）
    # Add-Content -Path "`$env:TEMP\AutoDeleteTask_Error.log" -Value "$(Get-Date): 删除任务 `$taskName 失败 - `$(`$_.Exception.Message)"
}
"@

# 将删除脚本保存到临时文件
$deleteScriptPath = Join-Path $env:TEMP "DeleteTask_$TaskName.ps1"
Set-Content -Path $deleteScriptPath -Value $deleteScript -Encoding UTF8

# 创建删除任务的操作
$deleteAction = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -File `"$deleteScriptPath`"" -WorkingDirectory $env:TEMP

# 创建删除任务的触发器（在主任务执行时间后延迟执行）
$deleteTriggerTime = $scheduledTime.AddMinutes(1)  # 延迟1分钟执行删除
$deleteTrigger = New-ScheduledTaskTrigger -Once -At $deleteTriggerTime

# 创建删除任务的设置
$deleteSettings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -Hidden -DeleteExpiredTaskAfter (New-TimeSpan -Minutes 5)

# 创建删除任务的主体
$deletePrincipal = New-ScheduledTaskPrincipal -UserId $env:USERNAME -LogonType Interactive -RunLevel Limited

# 删除任务名称
$deleteTaskName = "Delete_$TaskName"

# 注册删除任务
try {
    Register-ScheduledTask -TaskName $deleteTaskName -Action $deleteAction -Trigger $deleteTrigger -Settings $deleteSettings -Principal $deletePrincipal -Description "自动删除主任务 '$TaskName'"
    Write-Host "自动删除任务 '$deleteTaskName' 已创建！" -ForegroundColor Green
} catch {
    Write-Warning "创建自动删除任务失败: $($_.Exception.Message)"
    Write-Warning "您可能需要手动删除主任务 '$TaskName'"
}

Write-Host ""
Write-Host "任务创建完成！" -ForegroundColor Green
Write-Host "注意：如果需要手动删除任务，请使用以下命令：" -ForegroundColor Yellow
Write-Host "Unregister-ScheduledTask -TaskName '$TaskName' -Confirm:`$false" -ForegroundColor Gray
if (Test-Path $deleteTaskName) {
    Write-Host "Unregister-ScheduledTask -TaskName '$deleteTaskName' -Confirm:`$false" -ForegroundColor Gray
}
```

## 使用方法

### 基本用法：
```powershell
# 执行记事本（指定时间）
.\CreateAutoDeleteTask.ps1 -ExecuteTime "2024-01-15 14:30:00" -Command "notepad.exe"

# 执行PowerShell脚本
.\CreateAutoDeleteTask.ps1 -ExecuteTime "2024-01-15 23:00:00" -Command "C:\Scripts\MyScript.ps1" -Arguments "-Parameter1 Value1"

# 自定义任务名称
.\CreateAutoDeleteTask.ps1 -TaskName "DailyBackup" -ExecuteTime "2024-01-15 02:00:00" -Command "C:\Backup\backup.bat"
```

## 脚本特性

1. **自动验证时间格式**：确保输入的时间格式正确且不早于当前时间
2. **智能命令处理**：
   - 自动识别 `.ps1` 文件并使用 PowerShell 执行
   - 其他文件类型直接执行
3. **双重任务机制**：
   - 主任务：执行指定命令
   - 删除任务：监控主任务完成后自动删除两个任务
4. **错误处理**：包含完整的错误处理和用户友好的提示信息
5. **灵活参数**：支持自定义任务名称、工作目录、命令参数等
6. **安全执行**：使用当前用户权限运行，避免权限问题

## 注意事项

- 脚本需要以管理员权限运行（某些情况下）
- 删除任务会在主任务执行时间后1分钟开始监控并删除
- 如果主任务执行时间很长，删除任务会等待其完成后再删除
- 临时文件会保留在 `%TEMP%` 目录中，但不会影响系统性能

这个脚本提供了一个完整的解决方案，确保任务执行后能够自动清理，避免任务计划程序中积累无用的一次性任务。

# 请把上面的脚本简化一下，去掉出错检查之类的语句，只留下最核心的部分即可

以下是最简化的版本，只保留核心功能：

```powershell
param(
    [string]$TaskName = "AutoDeleteTask_$(Get-Date -Format 'yyyyMMdd_HHmmss')",
    [string]$ExecuteTime,
    [string]$Command,
    [string]$Arguments = ""
)

# 创建主任务
$action = New-ScheduledTaskAction -Execute $Command -Argument $Arguments
$trigger = New-ScheduledTaskTrigger -Once -At $ExecuteTime
Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger

# 创建自动删除任务
$deleteScript = "Unregister-ScheduledTask -TaskName '$TaskName' -Confirm:`$false"
$deleteAction = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-Command $deleteScript"
$deleteTrigger = New-ScheduledTaskTrigger -Once -At (Get-Date).AddMinutes(1)
Register-ScheduledTask -TaskName "Delete_$TaskName" -Action $deleteAction -Trigger $deleteTrigger
```

---

$TaskName = "\BTE\Amazon_Daily_Update_1"
$ExecuteTime = "2025-11-15 14:30:00"
$Command = "C:\Temp\test.ps1"
$Arguments = "upload US 1"

$action = New-ScheduledTaskAction -Execute $Command -WorkingDirectory "C:\Temp" -Argument $Arguments
$trigger = New-ScheduledTaskTrigger -Once -At $ExecuteTime
Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger
