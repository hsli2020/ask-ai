# 在windows下，如何查看某个端口被哪个程序占用？

在 Windows 下查看某个端口被哪个程序占用，有几种常用方法：

## 方法一：使用 netstat 命令（推荐）

### 1. 打开命令提示符（CMD）或 PowerShell
- 按 `Win + R`，输入 `cmd` 或 `powershell`
- 或者在开始菜单搜索 "命令提示符" 或 "PowerShell"

### 2. 查看特定端口
```cmd
netstat -ano | findstr :端口号
```

例如查看 8080 端口：
```cmd
netstat -ano | findstr :8080
```

### 3. 根据 PID 查找程序名称
从上一步得到 PID（进程ID）后，使用以下命令：
```cmd
tasklist | findstr PID号
```

例如：
```cmd
tasklist | findstr 12345
```

### 4. 一步到位的方法
```cmd
netstat -ano | findstr :8080 && tasklist | findstr 对应的PID
```

## 方法二：使用 PowerShell（更直观）

```powershell
Get-NetTCPConnection -LocalPort 端口号 | Select-Object OwningProcess
```

然后根据返回的进程ID查找程序：
```powershell
Get-Process -Id 进程ID
```

或者直接一行命令：
```powershell
Get-Process -Id (Get-NetTCPConnection -LocalPort 8080).OwningProcess
```

## 方法三：使用资源监视器

1. 按 `Ctrl + Shift + Esc` 打开任务管理器
2. 切换到"性能"选项卡
3. 点击底部的"打开资源监视器"
4. 在"网络"选项卡中，展开"监听端口"
5. 在列表中查找对应的端口

## 方法四：使用 TCPView 工具（第三方）

微软官方提供的免费工具：
1. 下载 [TCPView](https://docs.microsoft.com/en-us/sysinternals/downloads/tcpview)
2. 运行后可以看到所有网络连接和对应的进程
3. 可以按端口号排序查找

## 常用 netstat 参数说明

- `-a`：显示所有连接和监听端口
- `-n`：以数字形式显示地址和端口号
- `-o`：显示进程ID（PID）
- `-b`：显示创建每个连接或监听端口的可执行程序（需要管理员权限）

**注意**：使用 `-b` 参数需要以管理员身份运行命令提示符。

## 示例完整流程

```cmd
# 1. 查找占用8080端口的进程
C:\> netstat -ano | findstr :8080
  TCP    0.0.0.0:8080           0.0.0.0:0              LISTENING       12345

# 2. 根据PID查找程序名称
C:\> tasklist | findstr 12345
java.exe                     12345 Console                    1     45,236 K
```

这样就能知道是 `java.exe` 进程占用了 8080 端口。
