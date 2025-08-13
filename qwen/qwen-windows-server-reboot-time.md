# 🤵 我有一台windows服务器，偶尔会自动重启，我如何才能知道服务器重启的时间？

Qwen3-235B-A22B-2507

要确定你的 Windows 服务器何时自动重启，你可以通过以下几种方法查看重启时间：

---

### ✅ 方法一：使用事件查看器（Event Viewer）——最推荐

1. **打开事件查看器**：
   - 按 `Win + R`，输入 `eventvwr.msc`，回车。
   - 或者在“开始”菜单搜索“事件查看器”。

2. **导航到系统日志**：
   - 左侧依次展开：**Windows 日志** → **系统**。

3. **筛选事件 ID**：
   - 在右侧点击“筛选当前日志”。
   - 在“事件ID”框中输入：`6005, 6006, 6008, 41`
   - 点击“确定”。

   常见事件说明：
   - **6005**：事件日志服务已启动 → **系统开机时间**
   - **6006**：事件日志服务已停止 → **正常关机时间**
   - **6008**：上一次系统关机异常 → **非正常关机（可能突然断电或崩溃）**
   - **41**：系统在未正常关机的情况下重新启动（Kernel-Power 41）→ **意外重启**

4. **查看事件时间**：
   - 查看这些事件的时间戳，即可知道服务器重启的具体时间。

> 🔍 特别注意：**事件 ID 41（Kernel-Power）** 表示系统未正常关机就重启了，可能是蓝屏、断电、强制重启等原因。

---

### ✅ 方法二：使用命令提示符或 PowerShell

#### 1. 查看上次启动时间（当前运行时间）
```cmd
systeminfo | find "System Boot Time"
```

或者用 PowerShell：
```powershell
Get-CimInstance -ClassName Win32_OperatingSystem | Select-Object LastBootUpTime
```

这会显示**本次开机时间**，结合历史记录可推断重启时间。

#### 2. 使用 `wevtutil` 查询重启事件（PowerShell）
```powershell
# 查询最近的系统启动事件（6005）
wevtutil qe System /c:10 /f:text /q:"Event[System[(EventID=6005)]]"
```

---

### ✅ 方法三：检查蓝屏日志（如果存在崩溃重启）

1. 打开 **事件查看器** → **Windows 日志** → **系统**
2. 查找 **事件级别为“错误”** 的事件，特别是来源为：
   - `BugCheck`
   - `Kernel-Power`（ID 41）
   - `Kernel-EventTracing`（有时伴随崩溃）

3. 或使用工具分析内存转储文件（如果有）：
   - 转储文件路径通常为：`C:\Windows\Memory.dmp` 或 `C:\Windows\Minidump\*.dmp`
   - 使用 [WinDbg](https://learn.microsoft.com/en-us/windows-hardware/drivers/debugger/) 或 [BlueScreenView](https://www.nirsoft.net/utils/blue_screen_view.html) 分析。

---

### ✅ 方法四：启用更详细的日志（可选）

如果你怀疑是计划任务、更新或第三方软件导致重启，可以：

- 检查 **Windows 更新** 是否设置了自动重启：
  - 组策略：`计算机配置 → 管理模板 → Windows 组件 → Windows 更新`
  - 查看是否启用了“自动安装更新后自动重启”

- 启用 **Shutdown Reason Tracker**（关机原因追踪）：
  ```cmd
  shutdown -r -t 0 -c "测试原因" -f
  ```
  或通过组策略启用关机理由记录。

---

### ✅ 总结：排查自动重启的步骤

| 步骤 | 操作 |
|------|------|
| 1 | 使用事件查看器查看事件 ID 6005/6006/6008/41 |
| 2 | 检查是否有蓝屏（BugCheck 事件或内存转储） |
| 3 | 查看 Windows 更新设置 |
| 4 | 检查是否有计划任务或第三方软件触发重启 |
| 5 | 分析硬件日志（如电源、温度、硬盘） |

---

如果你提供具体的事件日志（例如事件ID 41 或 BugCheck 信息），我可以进一步帮你分析重启原因。

需要我帮你写一个 PowerShell 脚本自动提取最近几次重启时间吗？

--------------------------------------------------------------------------------

# 🤵 写一个 PowerShell 脚本自动提取最近几次重启时间

当然可以！以下是一个 **PowerShell 脚本**，用于自动提取 Windows 服务器最近几次的 **重启时间**，
并显示详细信息（包括是否为异常重启）。

---

### ✅ PowerShell 脚本：提取最近 N 次重启时间

```powershell
# ----------------------------
# 获取服务器最近 N 次重启时间
# ----------------------------

param(
    [int]$Count = 5  # 可通过参数指定要显示的重启次数，默认为5
)

# 存储重启事件的数组
$rebootEvents = @()

# 查询事件ID为6005（系统启动）的事件
$events = Get-WinEvent -FilterHashtable @{
    LogName = 'System'
    ID = 6005
} -MaxEvents $Count -ErrorAction SilentlyContinue

if (-not $events) {
    Write-Host "未找到系统启动事件（事件ID 6005）。" -ForegroundColor Yellow
    exit
}

# 遍历每个 6005 事件（系统启动）
foreach ($event in $events) {
    $bootTime = $event.TimeCreated
    $message = $event.Message

    # 尝试查找对应的上一次关机事件（6006 正常关机 或 6008 异常关机）
    $shutdownEvent = Get-WinEvent -FilterHashtable @{
        LogName = 'System'
        ID = 6006, 6008
        StartTime = (Get-Date).AddDays(-30)  # 仅检查最近30天
    } -MaxEvents 100 | Where-Object { $_.TimeCreated -lt $bootTime } | Sort-Object TimeCreated -Descending | Select-Object -First 1

    $shutdownType = "未知"
    $shutdownTime = $null

    if ($shutdownEvent) {
        $shutdownTime = $shutdownEvent.TimeCreated
        if ($shutdownEvent.Id -eq 6006) {
            $shutdownType = "正常关机"
        } elseif ($shutdownEvent.Id -eq 6008) {
            $shutdownType = "异常关机"
        }
    }

    # 检查本次启动是否为意外重启（Kernel-Power 41）
    $kernelPowerEvent = Get-WinEvent -FilterHashtable @{
        LogName = 'System'
        ID = 41
        Level = 2  # 错误级别
        StartTime = $bootTime.AddMinutes(-1)
        EndTime = $bootTime.AddMinutes(5)
    } -MaxEvents 1 -ErrorAction SilentlyContinue

    $unexpected = $kernelPowerEvent -ne $null
    $reason = $unexpected ? "⚠️ 意外重启（可能蓝屏/断电）" : "✅ 正常启动"

    # 添加到结果
    $rebootEvents += [PSCustomObject]@{
        "重启时间"         = $bootTime
        "上次关机时间"     = $shutdownTime
        "关机类型"         = $shutdownType
        "是否意外重启"     = if ($unexpected) { "是" } else { "否" }
        "说明"            = $reason
    }
}

# 输出结果
Write-Host "`n📊 最近 $Count 次服务器重启记录：" -ForegroundColor Cyan
$rebootEvents | Format-Table -AutoSize

# 补充：显示当前系统运行时间
$os = Get-CimInstance -ClassName Win32_OperatingSystem
$uptime = (Get-Date) - $os.LastBootUpTime
Write-Host "`n⏱️ 当前系统已运行: $($uptime.Days)天 $($uptime.Hours)小时 $($uptime.Minutes)分钟" -ForegroundColor Green
```

---

### 📌 使用方法：

1. 打开 **PowerShell（以管理员身份运行）**
2. 将上述脚本保存为文件，例如：`Get-RebootHistory.ps1`
3. 运行脚本：

```powershell
.\Get-RebootHistory.ps1
```

或指定查看最近3次重启：

```powershell
.\Get-RebootHistory.ps1 -Count 3
```

---

### 🎯 脚本功能说明：

| 功能 | 说明 |
|------|------|
| ✅ 获取最近 N 次重启时间 | 基于事件 ID 6005 |
| ✅ 判断是否异常重启 | 检查是否存在 Kernel-Power 41 事件 |
| ✅ 显示上次关机时间与类型 | 区分正常/异常关机 |
| ✅ 显示当前系统运行时间 | 帮助判断上次重启距今多久 |

---

### 🔐 注意事项：

- 需要 **管理员权限** 才能读取部分系统事件。
- 如果事件日志被清除过，可能无法查到太久以前的记录。
- 若服务器频繁意外重启，建议结合 `C:\Windows\Minidump\` 中的蓝屏文件进一步分析。

---

如果你希望我添加 **导出到 CSV** 或 **邮件通知功能**，也可以告诉我，我可以帮你扩展这个脚本。

--------------------------------------------------------------------------------
