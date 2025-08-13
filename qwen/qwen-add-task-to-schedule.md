# 如何通过编程方式向Windows的Task Schedule中添加任务，并设定执行时间，执行完毕后自动删除任务

在Windows中，可以通过多种方式编程添加计划任务，最常用的是使用PowerShell命令或调用Windows Task Scheduler的COM接口。以下是几种主要方法：

## 方法一：使用PowerShell命令（推荐）

### 1. 基本语法
```powershell
# 创建一次性任务，执行后自动删除
$taskName = "MyOneTimeTask"
$scriptPath = "C:\path\to\your\script.ps1"
$executionTime = (Get-Date).AddMinutes(5)  # 5分钟后执行

# 创建任务动作
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-File `"$scriptPath`""

# 创建触发器（一次性）
$trigger = New-ScheduledTaskTrigger -Once -At $executionTime

# 创建设置（执行后删除）
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1)

# 注册任务
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings
```

### 2. 完整示例函数
```powershell
function Create-OneTimeTask {
    param(
        [string]$TaskName,
        [string]$ScriptPath,
        [DateTime]$ExecutionTime,
        [string]$Description = "One-time task that auto-deletes after execution"
    )
    
    try {
        # 验证脚本路径是否存在
        if (-not (Test-Path $ScriptPath)) {
            throw "Script file not found: $ScriptPath"
        }
        
        # 创建任务动作
        $action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -File `"$ScriptPath`""
        
        # 创建一次性触发器
        $trigger = New-ScheduledTaskTrigger -Once -At $ExecutionTime
        
        # 创建设置：执行后立即删除
        $settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
        
        # 注册任务
        $task = Register-ScheduledTask -TaskName $TaskName -Action $action -Trigger $trigger -Settings $settings -Description $Description
        
        Write-Host "Task '$TaskName' created successfully. Will execute at: $($ExecutionTime.ToString('yyyy-MM-dd HH:mm:ss'))"
        return $task
    }
    catch {
        Write-Error "Failed to create task: $($_.Exception.Message)"
    }
}

# 使用示例
Create-OneTimeTask -TaskName "BackupTask" -ScriptPath "C:\Scripts\backup.ps1" -ExecutionTime (Get-Date).AddHours(1)
```

## 方法二：使用C#调用Task Scheduler COM接口

```csharp
using System;
using System.IO;
using Microsoft.Win32.TaskScheduler;

public class TaskSchedulerHelper
{
    public static void CreateOneTimeTask(string taskName, string scriptPath, DateTime executionTime)
    {
        try
        {
            // 验证文件存在
            if (!File.Exists(scriptPath))
                throw new FileNotFoundException($"Script file not found: {scriptPath}");

            // 创建任务服务实例
            using (TaskService ts = new TaskService())
            {
                // 创建任务定义
                TaskDefinition td = ts.NewTask();
                td.RegistrationInfo.Description = "One-time task that auto-deletes after execution";
                
                // 设置触发器（一次性）
                td.Triggers.Add(new TimeTrigger(executionTime));
                
                // 设置执行动作
                td.Actions.Add(new ExecAction("PowerShell.exe", 
                    $"-ExecutionPolicy Bypass -File \"{scriptPath}\"", null));
                
                // 设置任务在过期后删除（1秒后）
                td.Settings.DeleteExpiredTaskAfter = TimeSpan.FromSeconds(1);
                td.Settings.AllowStartIfOnBatteries = true;
                td.Settings.DontStopIfGoingOnBatteries = true;
                
                // 注册任务
                ts.RootFolder.RegisterTaskDefinition(taskName, td);
                
                Console.WriteLine($"Task '{taskName}' created successfully. Will execute at: {executionTime}");
            }
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Error creating task: {ex.Message}");
            throw;
        }
    }
}

// 使用示例
// 需要先安装 TaskScheduler NuGet 包: Install-Package TaskScheduler
TaskSchedulerHelper.CreateOneTimeTask("MyTask", @"C:\Scripts\mytask.ps1", DateTime.Now.AddMinutes(10));
```

## 方法三：使用schtasks.exe命令行工具

```powershell
# 创建一次性任务
$taskName = "MyOneTimeTask"
$scriptPath = "C:\path\to\script.ps1"
$executionTime = (Get-Date).AddMinutes(5).ToString("HH:mm")

# 创建任务（注意：schtasks不直接支持执行后自动删除，需要额外处理）
schtasks /create /tn $taskName /tr "PowerShell.exe -File `"$scriptPath`"" /sc once /st $executionTime /f

# 方法：在脚本末尾添加删除任务的命令
# 在你的脚本最后添加：schtasks /delete /tn "MyOneTimeTask" /f
```

## 方法四：完整的PowerShell解决方案（包含错误处理）

```powershell
function New-OneTimeScheduledTask {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory=$true)]
        [string]$TaskName,
        
        [Parameter(Mandatory=$true)]
        [string]$ScriptPath,
        
        [Parameter(Mandatory=$true)]
        [DateTime]$ExecutionTime,
        
        [string]$WorkingDirectory = $null,
        [string]$Description = "Auto-deleting one-time task"
    )
    
    # 验证参数
    if (-not (Test-Path $ScriptPath)) {
        throw "Script file does not exist: $ScriptPath"
    }
    
    if ($ExecutionTime -lt (Get-Date)) {
        throw "Execution time must be in the future"
    }
    
    try {
        # 检查是否已存在同名任务
        if (Get-ScheduledTask -TaskName $TaskName -ErrorAction SilentlyContinue) {
            Write-Warning "Task '$TaskName' already exists. Deleting existing task..."
            Unregister-ScheduledTask -TaskName $TaskName -Confirm:$false
        }
        
        # 准备执行参数
        $scriptArgs = "-ExecutionPolicy Bypass -WindowStyle Hidden -File `"$ScriptPath`""
        
        # 创建任务动作
        $action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument $scriptArgs
        
        if ($WorkingDirectory) {
            $action.WorkingDirectory = $WorkingDirectory
        }
        
        # 创建触发器
        $trigger = New-ScheduledTaskTrigger -Once -At $ExecutionTime
        
        # 创建设置
        $settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) `
                                                -AllowStartIfOnBatteries `
                                                -DontStopIfGoingOnBatteries `
                                                -ExecutionTimeLimit (New-TimeSpan -Hours 2)
        
        # 注册任务
        $task = Register-ScheduledTask -TaskName $TaskName `
                                      -Action $action `
                                      -Trigger $trigger `
                                      -Settings $settings `
                                      -Description $Description `
                                      -RunLevel Highest
        
        Write-Host "✅ Task '$TaskName' created successfully!" -ForegroundColor Green
        Write-Host "   Execution time: $($ExecutionTime.ToString('yyyy-MM-dd HH:mm:ss'))" -ForegroundColor Cyan
        Write-Host "   Script: $ScriptPath" -ForegroundColor Cyan
        
        return $task
    }
    catch {
        Write-Error "❌ Failed to create task '$TaskName': $($_.Exception.Message)"
        throw
    }
}

# 使用示例
try {
    New-OneTimeScheduledTask -TaskName "SystemCleanup" `
                            -ScriptPath "C:\Scripts\cleanup.ps1" `
                            -ExecutionTime (Get-Date).AddHours(2) `
                            -WorkingDirectory "C:\Scripts" `
                            -Description "Clean up temporary files"
}
catch {
    Write-Host "Task creation failed: $($_.Exception.Message)" -ForegroundColor Red
}
```

## 重要注意事项

1. **权限要求**：创建计划任务通常需要管理员权限
2. **路径处理**：确保脚本路径使用绝对路径，并正确转义引号
3. **执行策略**：PowerShell脚本可能需要绕过执行策略限制
4. **错误处理**：建议在脚本中添加适当的错误处理和日志记录
5. **测试**：在生产环境使用前先在测试环境中验证

## 验证任务创建

```powershell
# 查看任务
Get-ScheduledTask -TaskName "YourTaskName"

# 查看任务历史（需要启用历史记录）
Get-WinEvent -LogName "Microsoft-Windows-TaskScheduler/Operational" | 
    Where-Object {$_.Message -like "*YourTaskName*"} | 
    Select-Object TimeCreated, Message
```

推荐使用**方法一（PowerShell）**，因为它简单、可靠，并且是Windows原生支持的方式。如果需要在.NET应用程序中集成，可以使用**方法二（C#）**。

# 能用golang语言实现上述功能吗？

是的，可以使用Go语言实现向Windows Task Scheduler添加任务并设置执行后自动删除。以下是几种实现方法：

## 方法一：使用Windows API调用（推荐）

首先需要安装必要的依赖：

```bash
go get golang.org/x/sys/windows
go get github.com/go-ole/go-ole
go get github.com/go-ole/go-ole/oleutil
```

### 完整实现代码

```go
package main

import (
	"fmt"
	"log"
	"time"
	"unsafe"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// TaskScheduler 用于管理Windows计划任务
type TaskScheduler struct {
	scheduler *ole.IDispatch
}

// NewTaskScheduler 创建新的TaskScheduler实例
func NewTaskScheduler() (*TaskScheduler, error) {
	ole.CoInitialize(0)
	defer ole.CoUninitialize()

	unknown, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return nil, fmt.Errorf("failed to create Schedule.Service object: %v", err)
	}

	scheduler, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return nil, fmt.Errorf("failed to query interface: %v", err)
	}

	// 连接到本地计算机
	_, err = oleutil.CallMethod(scheduler, "Connect")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to task scheduler: %v", err)
	}

	return &TaskScheduler{scheduler: scheduler}, nil
}

// CreateOneTimeTask 创建一次性任务，执行后自动删除
func (ts *TaskScheduler) CreateOneTimeTask(taskName, scriptPath string, executionTime time.Time) error {
	// 获取根文件夹
	rootFolder, err := oleutil.GetProperty(ts.scheduler, "GetFolder", "\\")
	if err != nil {
		return fmt.Errorf("failed to get root folder: %v", err)
	}
	defer rootFolder.Clear()

	// 创建任务定义
	taskDefinition, err := oleutil.CallMethod(ts.scheduler, "NewTask", 0)
	if err != nil {
		return fmt.Errorf("failed to create new task: %v", err)
	}
	defer taskDefinition.Clear()

	taskDef := taskDefinition.ToIDispatch()

	// 设置注册信息
	regInfo, err := oleutil.GetProperty(taskDef, "RegistrationInfo")
	if err != nil {
		return fmt.Errorf("failed to get registration info: %v", err)
	}
	defer regInfo.Clear()

	_, err = oleutil.PutProperty(regInfo.ToIDispatch(), "Description", "Go-created one-time task that auto-deletes")
	if err != nil {
		return fmt.Errorf("failed to set description: %v", err)
	}

	// 设置触发器
	triggers, err := oleutil.GetProperty(taskDef, "Triggers")
	if err != nil {
		return fmt.Errorf("failed to get triggers: %v", err)
	}
	defer triggers.Clear()

	// 创建一次性触发器
	trigger, err := oleutil.CallMethod(triggers.ToIDispatch(), "Create", 1) // TASK_TRIGGER_TIME = 1
	if err != nil {
		return fmt.Errorf("failed to create trigger: %v", err)
	}
	defer trigger.Clear()

	triggerDisp := trigger.ToIDispatch()

	// 设置执行时间（ISO 8601格式）
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	_, err = oleutil.PutProperty(triggerDisp, "StartBoundary", execTimeStr)
	if err != nil {
		return fmt.Errorf("failed to set start boundary: %v", err)
	}

	// 设置动作
	actions, err := oleutil.GetProperty(taskDef, "Actions")
	if err != nil {
		return fmt.Errorf("failed to get actions: %v", err)
	}
	defer actions.Clear()

	// 创建执行动作
	action, err := oleutil.CallMethod(actions.ToIDispatch(), "Create", 0) // TASK_ACTION_EXEC = 0
	if err != nil {
		return fmt.Errorf("failed to create action: %v", err)
	}
	defer action.Clear()

	actionDisp := action.ToIDispatch()

	// 设置执行程序和参数
	_, err = oleutil.PutProperty(actionDisp, "Path", "PowerShell.exe")
	if err != nil {
		return fmt.Errorf("failed to set path: %v", err)
	}

	scriptArgs := fmt.Sprintf("-ExecutionPolicy Bypass -WindowStyle Hidden -File \"%s\"", scriptPath)
	_, err = oleutil.PutProperty(actionDisp, "Arguments", scriptArgs)
	if err != nil {
		return fmt.Errorf("failed to set arguments: %v", err)
	}

	// 设置任务设置
	settings, err := oleutil.GetProperty(taskDef, "Settings")
	if err != nil {
		return fmt.Errorf("failed to get settings: %v", err)
	}
	defer settings.Clear()

	settingsDisp := settings.ToIDispatch()

	// 设置执行后删除（1秒后过期）
	deleteTime := "PT1S" // ISO 8601 duration format: 1 second
	_, err = oleutil.PutProperty(settingsDisp, "DeleteExpiredTaskAfter", deleteTime)
	if err != nil {
		return fmt.Errorf("failed to set delete expired task after: %v", err)
	}

	// 允许在电池模式下运行
	_, err = oleutil.PutProperty(settingsDisp, "AllowStartIfOnBatteries", true)
	if err != nil {
		return fmt.Errorf("failed to set allow start if on batteries: %v", err)
	}

	_, err = oleutil.PutProperty(settingsDisp, "DisallowStartIfOnBatteries", false)
	if err != nil {
		return fmt.Errorf("failed to set disallow start if on batteries: %v", err)
	}

	// 注册任务
	_, err = oleutil.CallMethod(rootFolder.ToIDispatch(), "RegisterTaskDefinition",
		taskName,
		taskDef,
		6, // TASK_CREATE_OR_UPDATE = 6
		"",
		"",
		3, // TASK_LOGON_INTERACTIVE_TOKEN = 3
		nil)
	if err != nil {
		return fmt.Errorf("failed to register task: %v", err)
	}

	return nil
}

// DeleteTask 删除指定任务
func (ts *TaskScheduler) DeleteTask(taskName string) error {
	rootFolder, err := oleutil.GetProperty(ts.scheduler, "GetFolder", "\\")
	if err != nil {
		return fmt.Errorf("failed to get root folder: %v", err)
	}
	defer rootFolder.Clear()

	_, err = oleutil.CallMethod(rootFolder.ToIDispatch(), "DeleteTask", taskName, 0)
	if err != nil {
		return fmt.Errorf("failed to delete task: %v", err)
	}

	return nil
}

// Close 释放资源
func (ts *TaskScheduler) Close() {
	if ts.scheduler != nil {
		ts.scheduler.Release()
	}
	ole.CoUninitialize()
}

func main() {
	// 创建TaskScheduler实例
	scheduler, err := NewTaskScheduler()
	if err != nil {
		log.Fatalf("Failed to create task scheduler: %v", err)
	}
	defer scheduler.Close()

	// 设置任务参数
	taskName := "GoOneTimeTask"
	scriptPath := `C:\temp\test.ps1` // 替换为你的脚本路径
	executionTime := time.Now().Add(2 * time.Minute) // 2分钟后执行

	// 创建任务
	err = scheduler.CreateOneTimeTask(taskName, scriptPath, executionTime)
	if err != nil {
		log.Fatalf("Failed to create task: %v", err)
	}

	fmt.Printf("Task '%s' created successfully!\n", taskName)
	fmt.Printf("Execution time: %s\n", executionTime.Format("2006-01-02 15:04:05"))
	fmt.Printf("Script: %s\n", scriptPath)
}
```

## 方法二：使用PowerShell命令（更简单）

如果不想直接调用COM接口，可以使用Go执行PowerShell命令：

```go
package main

import (
	"fmt"
	"os/exec"
	"time"
)

// CreateOneTimeTaskUsingPowerShell 使用PowerShell创建一次性任务
func CreateOneTimeTaskUsingPowerShell(taskName, scriptPath string, executionTime time.Time) error {
	// 格式化执行时间
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 构建PowerShell命令
	psCommand := fmt.Sprintf(`
$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -File \`"%s\`""
$trigger = New-ScheduledTaskTrigger -Once -At "%s"
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName "%s" -Action $action -Trigger $trigger -Settings $settings -Description "Go-created one-time task"
`, scriptPath, execTimeStr, taskName)

	// 执行PowerShell命令
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}

// DeleteTaskUsingPowerShell 删除任务
func DeleteTaskUsingPowerShell(taskName string) error {
	cmd := exec.Command("powershell", "-Command", fmt.Sprintf("Unregister-ScheduledTask -TaskName \"%s\" -Confirm:$false", taskName))
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("Failed to delete task: %v, output: %s", err, string(output))
	}
	
	return nil
}

func main() {
	taskName := "GoPowerShellTask"
	scriptPath := `C:\temp\test.ps1`
	executionTime := time.Now().Add(1 * time.Minute)

	err := CreateOneTimeTaskUsingPowerShell(taskName, scriptPath, executionTime)
	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		return
	}

	fmt.Printf("Task '%s' created successfully using PowerShell!\n", taskName)
}
```

## 方法三：使用schtasks.exe命令行

```go
package main

import (
	"fmt"
	"os/exec"
	"time"
)

// CreateOneTimeTaskUsingSchTasks 使用schtasks创建任务
func CreateOneTimeTaskUsingSchTasks(taskName, scriptPath string, executionTime time.Time) error {
	// 格式化时间和日期
	dateStr := executionTime.Format("2006/01/02")
	timeStr := executionTime.Format("15:04")

	// 构建命令
	cmd := exec.Command("schtasks", "/create", 
		"/tn", taskName,
		"/tr", fmt.Sprintf("PowerShell.exe -ExecutionPolicy Bypass -File \"%s\"", scriptPath),
		"/sc", "once",
		"/sd", dateStr,
		"/st", timeStr,
		"/f") // /f 表示强制覆盖

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("schtasks command failed: %v, output: %s", err, string(output))
	}

	// 注意：schtasks不支持直接设置执行后删除，需要在脚本末尾添加删除命令
	// 或者使用单独的清理机制

	return nil
}
```

## 辅助函数：创建测试脚本

```go
// CreateTestScript 创建一个简单的测试脚本
func CreateTestScript(scriptPath string) error {
	scriptContent := `
# 测试脚本
Write-Host "Hello from Go-created scheduled task!" -ForegroundColor Green
Write-Host "Current time: $(Get-Date)" -ForegroundColor Cyan

# 可选：在脚本末尾添加任务自删除（如果使用schtasks方法）
# schtasks /delete /tn "YourTaskName" /f
`
	
	return os.WriteFile(scriptPath, []byte(scriptContent), 0644)
}
```

## 使用建议

1. **推荐使用方法一（COM接口）**：功能最完整，可以直接设置`DeleteExpiredTaskAfter`属性
2. **方法二（PowerShell）**：代码更简洁，适合快速实现
3. **方法三（schtasks）**：兼容性最好，但在Windows 10/11中功能有限

## 注意事项

- 需要以管理员权限运行程序
- 确保脚本路径存在且可访问
- 时间格式必须正确（ISO 8601格式用于COM接口）
- 在生产环境中添加适当的错误处理和日志记录
- 测试时建议先使用短时间间隔（如1-2分钟）

选择哪种方法取决于你的具体需求和环境限制。COM接口方法提供了最完整的控制，而PowerShell方法更简单易用。


# 运行时出现以下错误信息，如何修复
Error creating task: PowerShell command failed: exit status 1, output: The string is missing the terminator: ".
    + CategoryInfo          : ParserError: (:) [], ParentContainsErrorRecordException
    + FullyQualifiedErrorId : TerminatorExpectedAtEndOfString


这个错误是由于PowerShell命令中的引号没有正确转义导致的。问题出现在构建PowerShell命令字符串时，引号嵌套和转义不正确。

## 问题分析

错误信息 `The string is missing the terminator: "."` 表明PowerShell在解析字符串时找不到结束的引号。这通常是因为：

1. 路径中包含反斜杠 `\`，在PowerShell中被当作转义字符
2. 引号嵌套层次过多，转义不正确
3. 字符串拼接时引号没有正确闭合

## 解决方案

### 方案一：修复PowerShell命令的引号转义

```go
package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// escapePowerShellString 转义PowerShell字符串中的特殊字符
func escapePowerShellString(s string) string {
	// 将反斜杠替换为双反斜杠
	s = strings.ReplaceAll(s, "\\", "\\\\")
	// 转义双引号
	s = strings.ReplaceAll(s, `"`, `\"`)
	return s
}

// CreateOneTimeTaskUsingPowerShell 修复后的版本
func CreateOneTimeTaskUsingPowerShell(taskName, scriptPath string, executionTime time.Time) error {
	// 转义脚本路径
	escapedScriptPath := escapePowerShellString(scriptPath)
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 使用Here-String语法避免复杂的引号嵌套
	psCommand := fmt.Sprintf(`
$taskName = "%s"
$scriptPath = "%s"
$executionTime = "%s"

$action = New-ScheduledTaskAction -Execute "PowerShell.exe" -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File ` + "`" + `"$scriptPath` + "`" + `""
$trigger = New-ScheduledTaskTrigger -Once -At $executionTime
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings -Description "Go-created one-time task" -Force
`, taskName, escapedScriptPath, execTimeStr)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
```

### 方案二：使用单引号避免转义问题（推荐）

```go
// CreateOneTimeTaskUsingPowerShellV2 使用单引号的版本
func CreateOneTimeTaskUsingPowerShellV2(taskName, scriptPath string, executionTime time.Time) error {
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 使用单引号包裹字符串，避免双引号转义问题
	// 在PowerShell中，单引号内的内容不会被解释
	psCommand := fmt.Sprintf(`
$taskName = '%s'
$scriptPath = '%s'
$executionTime = '%s'

$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File `+"`"+`$scriptPath`+"`"+`"
$trigger = New-ScheduledTaskTrigger -Once -At $executionTime
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings -Description 'Go-created one-time task' -Force
`, taskName, scriptPath, execTimeStr)

	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
```

### 方案三：使用Base64编码避免所有转义问题（最可靠）

```go
import (
	"encoding/base64"
)

// CreateOneTimeTaskUsingPowerShellBase64 使用Base64编码
func CreateOneTimeTaskUsingPowerShellBase64(taskName, scriptPath string, executionTime time.Time) error {
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 构建PowerShell脚本内容
	psScript := fmt.Sprintf(`
$taskName = '%s'
$scriptPath = '%s'
$executionTime = '%s'

$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File `+"`"+`$scriptPath`+"`"+`"
$trigger = New-ScheduledTaskTrigger -Once -At $executionTime
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings -Description 'Go-created one-time task' -Force
`, taskName, scriptPath, execTimeStr)
	
	// 编码为Base64
	encodedScript := base64.StdEncoding.EncodeToString([]byte(psScript))
	
	// 使用-EncodedCommand参数执行
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-EncodedCommand", encodedScript)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
```

### 方案四：简化命令，避免复杂嵌套

```go
// CreateOneTimeTaskSimple 简化版本
func CreateOneTimeTaskSimple(taskName, scriptPath string, executionTime time.Time) error {
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 直接构建完整的PowerShell命令行，避免多层嵌套
	// 使用反引号转义路径中的引号
	scriptArg := fmt.Sprintf(`-ExecutionPolicy Bypass -WindowStyle Hidden -File "%s"`, scriptPath)
	
	psCommand := fmt.Sprintf(
		"Register-ScheduledTask -TaskName '%s' -Action (New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument '%s') -Trigger (New-ScheduledTaskTrigger -Once -At '%s') -Settings (New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries) -Description 'Go-created one-time task' -Force",
		taskName, scriptArg, execTimeStr)
	
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psCommand)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
```

## 完整的测试代码

```go
package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
)

func main() {
	taskName := "GoTestTask"
	scriptPath := `C:\temp\test.ps1`
	executionTime := time.Now().Add(1 * time.Minute)

	// 确保脚本文件存在
	createTestScript(scriptPath)

	// 使用Base64编码的方法（最可靠）
	err := CreateOneTimeTaskUsingPowerShellBase64(taskName, scriptPath, executionTime)
	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Task '%s' created successfully!\n", taskName)
}

func createTestScript(scriptPath string) {
	scriptContent := `Write-Host "Task executed successfully at $(Get-Date)" -ForegroundColor Green`
	os.WriteFile(scriptPath, []byte(scriptContent), 0644)
}
```

## 推荐使用方案

**推荐使用方案三（Base64编码）**，因为：

1. 完全避免了引号转义问题
2. 不受路径中特殊字符影响
3. 在所有Windows版本中都能可靠工作
4. 代码清晰，易于维护

这样就能彻底解决引号转义导致的PowerShell解析错误问题。

# 我在执行最上面的Create-OneTimeTask时出现这个错误，如何修复
 The task XML is missing a required element or attribute.  (42,4):EndBoundary
 
这个错误表明任务XML缺少必需的元素或属性，具体是在第42行第4列缺少`EndBoundary`。这通常是因为一次性触发器需要设置结束边界（EndBoundary）或者触发器配置不完整。

## 问题分析

在Windows Task Scheduler中，一次性触发器（TimeTrigger）虽然理论上不需要`EndBoundary`，但在某些Windows版本中，如果只设置了`StartBoundary`而没有设置`EndBoundary`或`ExecutionTimeLimit`，可能会出现这个错误。

## 解决方案

### 方案一：为触发器添加EndBoundary

```go
// 在CreateOneTimeTask函数中，修改触发器设置部分
triggerDisp := trigger.ToIDispatch()

// 设置执行时间（StartBoundary）
execTimeStr := executionTime.Format("2006-01-02T15:04:05")
_, err = oleutil.PutProperty(triggerDisp, "StartBoundary", execTimeStr)
if err != nil {
    return fmt.Errorf("failed to set start boundary: %v", err)
}

// 添加EndBoundary（执行时间后1小时）
endTime := executionTime.Add(1 * time.Hour)
endTimeStr := endTime.Format("2006-01-02T15:04:05")
_, err = oleutil.PutProperty(triggerDisp, "EndBoundary", endTimeStr)
if err != nil {
    return fmt.Errorf("failed to set end boundary: %v", err)
}
```

### 方案二：设置ExecutionTimeLimit而不是EndBoundary

```go
// 在任务设置部分添加执行时间限制
settingsDisp := settings.ToIDispatch()

// 设置执行时间限制为2小时（PT2H = Period Time 2 Hours）
_, err = oleutil.PutProperty(settingsDisp, "ExecutionTimeLimit", "PT2H")
if err != nil {
    return fmt.Errorf("failed to set execution time limit: %v", err)
}

// 设置执行后删除（1秒后过期）
_, err = oleutil.PutProperty(settingsDisp, "DeleteExpiredTaskAfter", "PT1S")
if err != nil {
    return fmt.Errorf("failed to set delete expired task after: %v", err)
}
```

### 方案三：完整的修复版本

```go
// CreateOneTimeTask 创建一次性任务，执行后自动删除（修复版本）
func (ts *TaskScheduler) CreateOneTimeTask(taskName, scriptPath string, executionTime time.Time) error {
	// 获取根文件夹
	rootFolder, err := oleutil.GetProperty(ts.scheduler, "GetFolder", "\\")
	if err != nil {
		return fmt.Errorf("failed to get root folder: %v", err)
	}
	defer rootFolder.Clear()

	// 创建任务定义
	taskDefinition, err := oleutil.CallMethod(ts.scheduler, "NewTask", 0)
	if err != nil {
		return fmt.Errorf("failed to create new task: %v", err)
	}
	defer taskDefinition.Clear()

	taskDef := taskDefinition.ToIDispatch()

	// 设置注册信息
	regInfo, err := oleutil.GetProperty(taskDef, "RegistrationInfo")
	if err != nil {
		return fmt.Errorf("failed to get registration info: %v", err)
	}
	defer regInfo.Clear()

	_, err = oleutil.PutProperty(regInfo.ToIDispatch(), "Description", "Go-created one-time task that auto-deletes")
	if err != nil {
		return fmt.Errorf("failed to set description: %v", err)
	}

	// 设置触发器
	triggers, err := oleutil.GetProperty(taskDef, "Triggers")
	if err != nil {
		return fmt.Errorf("failed to get triggers: %v", err)
	}
	defer triggers.Clear()

	// 创建一次性触发器
	trigger, err := oleutil.CallMethod(triggers.ToIDispatch(), "Create", 1) // TASK_TRIGGER_TIME = 1
	if err != nil {
		return fmt.Errorf("failed to create trigger: %v", err)
	}
	defer trigger.Clear()

	triggerDisp := trigger.ToIDispatch()

	// 设置执行时间（StartBoundary）
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	_, err = oleutil.PutProperty(triggerDisp, "StartBoundary", execTimeStr)
	if err != nil {
		return fmt.Errorf("failed to set start boundary: %v", err)
	}

	// 关键修复：设置Enabled属性为true
	_, err = oleutil.PutProperty(triggerDisp, "Enabled", true)
	if err != nil {
		return fmt.Errorf("failed to set trigger enabled: %v", err)
	}

	// 设置动作
	actions, err := oleutil.GetProperty(taskDef, "Actions")
	if err != nil {
		return fmt.Errorf("failed to get actions: %v", err)
	}
	defer actions.Clear()

	// 创建执行动作
	action, err := oleutil.CallMethod(actions.ToIDispatch(), "Create", 0) // TASK_ACTION_EXEC = 0
	if err != nil {
		return fmt.Errorf("failed to create action: %v", err)
	}
	defer action.Clear()

	actionDisp := action.ToIDispatch()

	// 设置执行程序和参数
	_, err = oleutil.PutProperty(actionDisp, "Path", "PowerShell.exe")
	if err != nil {
		return fmt.Errorf("failed to set path: %v", err)
	}

	scriptArgs := fmt.Sprintf("-ExecutionPolicy Bypass -WindowStyle Hidden -File \"%s\"", scriptPath)
	_, err = oleutil.PutProperty(actionDisp, "Arguments", scriptArgs)
	if err != nil {
		return fmt.Errorf("failed to set arguments: %v", err)
	}

	// 设置工作目录（可选）
	_, err = oleutil.PutProperty(actionDisp, "WorkingDirectory", "C:\\")
	if err != nil {
		// 忽略工作目录设置错误
	}

	// 设置任务设置
	settings, err := oleutil.GetProperty(taskDef, "Settings")
	if err != nil {
		return fmt.Errorf("failed to get settings: %v", err)
	}
	defer settings.Clear()

	settingsDisp := settings.ToIDispatch()

	// 关键修复：设置所有必需的属性
	_, err = oleutil.PutProperty(settingsDisp, "AllowDemandStart", true)
	if err != nil {
		return fmt.Errorf("failed to set allow demand start: %v", err)
	}

	_, err = oleutil.PutProperty(settingsDisp, "RestartOnFailure", false)
	if err != nil {
		return fmt.Errorf("failed to set restart on failure: %v", err)
	}

	// 设置执行时间限制（重要）
	_, err = oleutil.PutProperty(settingsDisp, "ExecutionTimeLimit", "PT72H") // 72小时
	if err != nil {
		return fmt.Errorf("failed to set execution time limit: %v", err)
	}

	// 设置执行后删除（1秒后过期）
	_, err = oleutil.PutProperty(settingsDisp, "DeleteExpiredTaskAfter", "PT1S")
	if err != nil {
		return fmt.Errorf("failed to set delete expired task after: %v", err)
	}

	// 允许在电池模式下运行
	_, err = oleutil.PutProperty(settingsDisp, "AllowStartIfOnBatteries", true)
	if err != nil {
		return fmt.Errorf("failed to set allow start if on batteries: %v", err)
	}

	_, err = oleutil.PutProperty(settingsDisp, "DisallowStartIfOnBatteries", false)
	if err != nil {
		return fmt.Errorf("failed to set disallow start if on batteries: %v", err)
	}

	_, err = oleutil.PutProperty(settingsDisp, "StopIfGoingOnBatteries", false)
	if err != nil {
		return fmt.Errorf("failed to set stop if going on batteries: %v", err)
	}

	// 注册任务
	_, err = oleutil.CallMethod(rootFolder.ToIDispatch(), "RegisterTaskDefinition",
		taskName,
		taskDef,
		6, // TASK_CREATE_OR_UPDATE = 6
		"",
		"",
		3, // TASK_LOGON_INTERACTIVE_TOKEN = 3
		nil)
	if err != nil {
		return fmt.Errorf("failed to register task: %v", err)
	}

	return nil
}
```

### 方案四：使用更简单的PowerShell方法（推荐）

如果COM接口方法仍然有问题，建议使用修复后的PowerShell方法：

```go
func CreateOneTimeTaskUsingPowerShellFixed(taskName, scriptPath string, executionTime time.Time) error {
	execTimeStr := executionTime.Format("2006-01-02T15:04:05")
	
	// 使用单引号并确保所有必需参数都设置
	psScript := fmt.Sprintf(`
$ErrorActionPreference = "Stop"
$taskName = '%s'
$scriptPath = '%s'
$executionTime = '%s'

# 创建动作
$action = New-ScheduledTaskAction -Execute 'PowerShell.exe' -Argument "-ExecutionPolicy Bypass -WindowStyle Hidden -File `+"`"+`$scriptPath`+"`"+`"

# 创建触发器
$trigger = New-ScheduledTaskTrigger -Once -At $executionTime

# 创建设置（包含所有必需属性）
$settings = New-ScheduledTaskSettingsSet -DeleteExpiredTaskAfter (New-TimeSpan -Seconds 1) -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries -ExecutionTimeLimit (New-TimeSpan -Hours 72)

# 注册任务
Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger -Settings $settings -Description 'Go-created one-time task' -Force

Write-Host "Task created successfully"
`, taskName, scriptPath, execTimeStr)
	
	cmd := exec.Command("powershell", "-ExecutionPolicy", "Bypass", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	
	if err != nil {
		return fmt.Errorf("PowerShell command failed: %v, output: %s", err, string(output))
	}
	
	return nil
}
```

## 调试建议

1. **检查Windows版本**：不同Windows版本对任务XML的要求可能不同
2. **查看详细错误**：使用`Get-WinEvent`查看TaskScheduler的详细错误日志
3. **测试简单任务**：先创建一个最简单的任务，逐步添加属性

```powershell
# 查看TaskScheduler错误日志
Get-WinEvent -LogName "Microsoft-Windows-TaskScheduler/Operational" | 
    Where-Object {$_.Level -eq 2} | 
    Select-Object TimeCreated, Message -First 5
```

**推荐使用方案四（PowerShell方法）**，因为它更简单、更可靠，并且避免了COM接口的复杂性。

