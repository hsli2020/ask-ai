# 写一个Windows Powershell脚本，删除当前目录下的空文件(文件大小为0)，和超过一周的旧文件

以下是安全且高效的 PowerShell 脚本。为了防误删，**首次运行默认开启 `-WhatIf` 模拟模式**，确认无误后再移除该参数执行实际删除。

```powershell
# 定义时间阈值：7天前
$cutoffDate = (Get-Date).AddDays(-7)

# 获取当前目录下的文件，筛选空文件或最后修改时间超过7天的文件，并删除
Get-ChildItem -Path . -File | Where-Object {
    $_.Length -eq 0 -or $_.LastWriteTime -lt $cutoffDate
} | Remove-Item -Force -WhatIf  # ⚠️ 首次运行请保留 -WhatIf 进行安全测试
```

### 📖 脚本说明
| 代码片段 | 作用 |
|:---|:---|
| `Get-ChildItem -Path . -File` | 仅获取当前目录下的**文件**（排除文件夹） |
| `$_.Length -eq 0` | 匹配文件大小为 0 字节的空文件 |
| `$_.LastWriteTime -lt $cutoffDate` | 匹配最后修改时间早于7天的文件 |
| `Remove-Item -Force` | 强制删除（包含隐藏文件/只读文件） |
| `-WhatIf` | **安全预览**：显示将被删除的文件，但不实际执行删除 |

### ⚠️ 重要注意事项
1. **正式删除**：测试确认无误后，将 `-WhatIf` 删除或注释掉再运行：
   ```powershell
   } | Remove-Item -Force
   ```
2. **包含子目录**：如需同时清理当前目录及其所有子目录，在 `Get-ChildItem` 后添加 `-Recurse`：
   ```powershell
   Get-ChildItem -Path . -File -Recurse | ...
   ```
3. **跳过占用文件**：若某些文件正被其他程序占用，会报错。可追加 `-ErrorAction SilentlyContinue` 静默跳过：
   ```powershell
   } | Remove-Item -Force -ErrorAction SilentlyContinue
   ```
4. **时间判定基准**：脚本默认使用 `LastWriteTime`（最后修改时间）。若需按**创建时间**判断，将 `$_.LastWriteTime` 改为 `$_.CreationTime` 即可。
5. **权限要求**：删除系统保护目录或需要管理员权限的文件时，请以**管理员身份**运行 PowerShell。

### 💡 进阶：带日志记录的版本（推荐生产环境使用）
如需记录删除了哪些文件，可改为：
```powershell
$cutoffDate = (Get-Date).AddDays(-7)
$targetFiles = Get-ChildItem -Path . -File | Where-Object { $_.Length -eq 0 -or $_.LastWriteTime -lt $cutoffDate }

if ($targetFiles) {
    $targetFiles | Remove-Item -Force
    $targetFiles.FullName | Out-File -FilePath ".\deleted_files_$(Get-Date -Format 'yyyyMMdd_HHmmss').log" -Encoding UTF8
    Write-Host "✅ 已删除 $($targetFiles.Count) 个文件，详情见生成的日志文件。" -ForegroundColor Green
} else {
    Write-Host "🔍 未找到符合条件的文件。" -ForegroundColor Yellow
}
```

按需调整参数后直接粘贴到 PowerShell 终端执行即可。如有特定目录路径或过滤需求，可进一步修改 `-Path` 参数。


# 我要对多个目录进行上述清理操作，请改进上述脚本，循环处理多个目录

以下是优化后的多目录循环清理脚本。采用**顶部配置区**设计，方便你直接修改路径和参数，内置安全预览、日志记录和异常跳过机制。

```powershell
# ================= 📝 配置区 =================
# 1. 需要清理的目录列表（支持绝对路径、相对路径、环境变量）
$TargetDirectories = @(
    "C:\Temp\Logs",
    "D:\Projects\Backups",
    ".\MySubFolder"
)

# 2. 文件年龄阈值（天）
$DaysThreshold = 7
$CutoffDate = (Get-Date).AddDays(-$DaysThreshold)

# 3. 是否递归子目录（$true 包含子文件夹，$false 仅处理当前层级）
$RecurseSubdirs = $true

# 4. 安全开关：$true = 预览模式（不实际删除），$false = 真实删除
$DryRunMode = $true

# 5. 日志保存路径（自动生成带时间戳的文件）
$LogFilePath = ".\CleanupLog_$(Get-Date -Format 'yyyyMMdd_HHmmss').txt"
# ===========================================

$AllTargetFiles = @()

foreach ($dir in $TargetDirectories) {
    # 解析完整路径并验证
    $resolvedDir = [System.IO.Path]::GetFullPath($dir)
    if (-not (Test-Path $resolvedDir -PathType Container)) {
        Write-Warning "⚠️ 目录不存在或无权限访问，已跳过: $dir"
        continue
    }

    Write-Host "`n📂 正在扫描: $resolvedDir" -ForegroundColor Cyan

    # 获取文件并过滤：空文件 或 超过指定天数的文件
    $files = Get-ChildItem -Path $resolvedDir -File -Recurse:$RecurseSubdirs -ErrorAction SilentlyContinue | Where-Object {
        $_.Length -eq 0 -or $_.LastWriteTime -lt $CutoffDate
    }

    if (-not $files) {
        Write-Host "  🔍 未找到符合条件的文件。" -ForegroundColor DarkGray
        continue
    }

    $AllTargetFiles += $files.FullName

    if ($DryRunMode) {
        # 预览模式：仅显示将被删除的文件
        $files | ForEach-Object { Write-Host "  🧪 [预览] $($_.FullName)" -ForegroundColor Yellow }
        Write-Host "  📊 本目录将处理 $($files.Count) 个文件。" -ForegroundColor Yellow
    } else {
        # 实际删除：强制删除，忽略被占用文件的报错
        $files | Remove-Item -Force -ErrorAction SilentlyContinue
        Write-Host "  🗑️ [已删除] 成功清理 $($files.Count) 个文件。" -ForegroundColor Green
    }
}

# 生成汇总日志
if ($AllTargetFiles.Count -gt 0) {
    $header = "📁 文件清理日志 | 生成时间: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss') | 模式: $(if($DryRunMode){'预览'}else{'删除'}) | 阈值: $DaysThreshold 天前`n`n"
    ($header + ($AllTargetFiles -join "`n")) | Set-Content -Path $LogFilePath -Encoding UTF8
    Write-Host "`n📝 详细清单已保存至: $LogFilePath" -ForegroundColor White
} else {
    Write-Host "`n✅ 所有目录处理完毕，无需清理的文件。" -ForegroundColor Green
}
```

### 🔑 核心改进说明
| 特性 | 说明 |
|:---|:---|
| **配置集中化** | 所有可变参数集中在顶部 `# 📝 配置区`，修改路径/天数/开关无需改动逻辑 |
| **路径安全校验** | 使用 `[System.IO.Path]::GetFullPath()` 自动补全相对路径，配合 `Test-Path` 跳过无效目录 |
| **双模式切换** | `$DryRunMode = $true` 时仅列出文件并统计；改为 `$false` 后执行真实删除 |
| **静默容错** | `-ErrorAction SilentlyContinue` 自动跳过无权限目录或被进程占用的文件，防止脚本中断 |
| **结构化日志** | 生成带时间戳的 `.txt` 日志，包含执行模式、时间阈值和完整文件路径清单 |

### 🛡️ 使用建议
1. **首次运行**：保持 `$DryRunMode = $true`，观察终端输出的 `[预览]` 列表是否符合预期。
2. **正式执行**：将配置区改为 `$DryRunMode = $false` 后重新运行脚本。
3. **排除特定目录**：若某些子文件夹需保留，可在 `Get-ChildItem` 后追加排除逻辑，例如：
   ```powershell
   | Where-Object { $_.FullName -notlike "*\node_modules\*" -and ($_.Length -eq 0 -or $_.LastWriteTime -lt $CutoffDate) }
   ```
4. **性能提示**：递归扫描超大目录（如数十万文件）时，首次运行可能需几秒。可添加 `-Filter` 或按扩展名过滤提升速度。
5. **定时任务**：配合 Windows 任务计划程序，可将此脚本设为每日/每周自动执行，实现无人值守清理。

直接保存为 `.ps1` 文件（如 `Clean-MultiDirs.ps1`），在 PowerShell 中运行 `.\Clean-MultiDirs.ps1` 即可。如有特定过滤规则或日志格式需求，可告知进一步定制。

