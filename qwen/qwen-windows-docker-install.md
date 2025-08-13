# 在windows系统中如何使用docker?

在 Windows 系统中使用 Docker 的完整指南如下：

---

## 一、系统要求

- **操作系统**：Windows 10 64 位（专业版、企业版或教育版，版本 1903 及以上）或 Windows 11
- **硬件要求**：
  - 至少 4GB 内存（推荐 8GB 以上）
  - 支持并启用虚拟化（Intel VT-x / AMD-V）
  - 至少 10GB 可用磁盘空间
- **BIOS 设置**：确保虚拟化功能已启用
- **WSL 2 支持**：推荐使用 WSL 2 后端以获得最佳性能

---

## 二、安装步骤

### 1. 启用 WSL 2（推荐）

以管理员身份打开 PowerShell，依次执行：

```powershell
# 启用 WSL 功能
dism.exe /online /enable-feature /featurename:Microsoft-Windows-Subsystem-Linux /all /norestart

# 启用虚拟机平台
dism.exe /online /enable-feature /featurename:VirtualMachinePlatform /all /norestart

# 重启电脑
```

安装 WSL 2 内核更新包（从微软官网下载）：
🔗 https://aka.ms/wsl2kernel

设置 WSL 默认版本为 2：

```powershell
wsl --set-default-version 2
```

### 2. 安装 Docker Desktop

- 访问官网下载：https://www.docker.com/products/docker-desktop
- 运行安装程序，按提示完成安装
- 启动 Docker Desktop，首次运行时选择使用 WSL 2 后端
- 登录 Docker Hub 账户（可选）

---

## 三、验证安装

打开 PowerShell 或命令提示符，运行：

```powershell
docker --version
docker run hello-world
```

若看到欢迎信息，说明 Docker 安装成功。

---

## 四、基本使用

### 常用命令

```powershell
# 查看 Docker 信息
docker info

# 拉取镜像
docker pull nginx

# 运行容器
docker run -d -p 80:80 --name my-nginx nginx

# 查看运行中的容器
docker ps

# 查看所有容器
docker ps -a

# 停止/启动/删除容器
docker stop my-nginx
docker start my-nginx
docker rm my-nginx

# 查看镜像
docker images

# 删除镜像
docker rmi nginx

# 进入容器
docker exec -it my-nginx bash

# 查看容器日志
docker logs my-nginx
```

### 挂载本地目录

```powershell
docker run -d -p 80:80 -v C:/my-app:/usr/share/nginx/html nginx
```

### 使用 Docker Compose

创建 `docker-compose.yml` 文件，例如：

```yaml
version: '3'
services:
  web:
    image: nginx
    ports:
      - "80:80"
```

运行：

```powershell
docker-compose up -d
docker-compose down
```

---

## 五、常见问题与解决方案

| 问题 | 解决方案 |
|------|----------|
| Docker 启动失败 | 检查虚拟化是否启用；以管理员身份运行 Docker Desktop |
| WSL 2 未启用 | 运行 `wsl --set-default-version 2` |
| 权限错误 | 以管理员身份运行终端；检查文件路径权限 |
| 端口被占用 | 使用 `netstat -ano | findstr :80` 查找占用进程 |
| 镜像拉取慢 | 配置国内镜像加速器（如阿里云） |

### 配置镜像加速器（以阿里云为例）

1. 打开 Docker Desktop → Settings → Docker Engine
2. 添加如下配置：

```json
{
  "registry-mirrors": [
    "https://<your-id>.mirror.aliyuncs.com"
  ]
}
```

3. 点击 "Apply & Restart"

---

## 六、最佳实践

- 使用 `.dockerignore` 文件排除不必要文件
- 使用多阶段构建优化镜像大小
- 避免在容器内存储持久数据，使用卷挂载
- 定期清理无用资源：

```powershell
docker system prune -a
```

- 使用 Docker Compose 管理多容器应用
- 为容器设置资源限制（CPU、内存）

---

## 七、WSL 2 集成（推荐）

在 Docker Desktop 中启用 WSL 2 集成：

1. Settings → Resources → WSL Integration
2. 启用你需要的 Linux 发行版（如 Ubuntu）

在 WSL 终端中可直接使用 Docker 命令：

```bash
docker ps
docker run hello-world
```

---

## 八、卸载 Docker

1. 在“设置 → 应用”中卸载 Docker Desktop
2. 删除相关目录：

```powershell
Remove-Item -Recurse -Force $env:LOCALAPPDATA\Docker
Remove-Item -Recurse -Force $env:PROGRAMFILES\Docker
```

3. 可选：重置 WSL

```powershell
wsl --unregister docker-desktop
wsl --unregister docker-desktop-data
```

---

通过以上步骤，你就可以在 Windows 系统中顺利安装并使用 Docker。推荐始终使用 WSL 2 后端以获得更好的性能和兼容性。如有进一步问题，欢迎继续提问！
