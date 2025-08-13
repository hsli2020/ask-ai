# 在Dos的批处理文件中，怎样才能并行启动多个程序，而不是顺序执行

在 DOS/Windows 批处理文件中实现并行启动多个程序（而非顺序执行），**核心方法是使用 `start` 命令**。以下是详细说明：

### 1. 基本原理
- 默认情况下，批处理命令是**顺序执行**的（前一个程序结束后才执行下一个）
- `start` 命令会**立即启动新进程并返回**，不等待程序结束，从而实现并行 [[1]]
- 注意：纯 MS-DOS（单任务系统）无法真正并行，但 Windows 95/98/NT 及以后的 `cmd.exe` 可利用 Windows 多任务能力实现真正的并行执行

### 2. 基本语法
```batch
start ["title"] [/options] command [parameters]
```
- **`"title"`**：窗口标题（**必须提供**，若不需要可设为空字符串 `""`）
- **关键参数**：
  - `/B`：后台启动，不创建新窗口（在当前窗口运行）[[1]]
  - `/MIN`：最小化窗口启动
  - `/D path`：指定工作目录
  - `/WAIT`：**等待程序结束**（这是顺序执行，非并行，慎用）

### 3. 实用示例
```batch
@echo off
REM 并行启动多个程序（每个都立即返回）
start "" notepad.exe
start "" calc.exe
start "" "C:\Program Files\YourApp\app.exe"  :: 路径含空格需加引号

REM 后台启动（不创建新窗口）
start /B ping 127.0.0.1 -t

REM 指定工作目录启动
start /D "G:\MyApp" MyApp.exe

REM 启动另一个批处理文件（并行）
start "" "sub_script.bat"
```

### 4. 重要注意事项
1. **路径含空格时**：必须用引号包裹路径，且**第一个参数必须是标题**（即使为空字符串 `""`）[[19]]
   ```batch
   :: 正确
   start "" "C:\Program Files\App\app.exe"
   
   :: 错误（会把路径第一部分当作标题）
   start "C:\Program Files\App\app.exe"
   ```

2. **避免使用 `call`**：`call` 会等待被调用的批处理执行完毕，是顺序执行 [[11]]

3. **等待所有并行进程完成**：`start` 本身不提供等待机制，如需同步需额外处理（如用 `tasklist` 检测进程或创建临时锁文件）[[2]]

### 5. 与顺序执行对比
```batch
:: 顺序执行（等待每个程序结束）
notepad.exe
calc.exe

:: 并行执行（立即启动所有程序）
start "" notepad.exe
start "" calc.exe
```

> 💡 提示：在现代 Windows 系统中，`start` 是实现批处理并行执行的标准且可靠的方法，适用于启动 EXE 程序、BAT 脚本或文档文件。

