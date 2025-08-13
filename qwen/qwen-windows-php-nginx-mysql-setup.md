#-------------------------------------------------------------------------------
windows下面如何配置php+nginx+mysql环境

在 Windows 系统下配置 PHP + Nginx + MySQL 环境，可以通过手动安装各个组件来实现，也可以使用集成环境（如 XAMPP、WampServer、Laragon 等）。下面介绍 **手动配置 PHP + Nginx + MySQL** 的详细步骤，适合学习或定制化需求。

---

## ✅ 一、准备工作

1. **下载所需软件包**：
   - **PHP**：https://windows.php.net/download/
   - **Nginx**：http://nginx.org/en/download.html
   - **MySQL**：https://dev.mysql.com/downloads/mysql/ 或使用 MariaDB 替代
   - 推荐将所有文件解压/安装到一个统一目录，如：`D:\webserver\`

---

## ✅ 二、安装与配置步骤

### 1. 安装并配置 PHP

#### (1) 下载 PHP
- 建议选择 **Thread Safe (TS)** 版本（用于与 Nginx 配合）。
- 下载 ZIP 包（如 `php-8.3.0-Win32-vs16-x64.zip`）。
- 解压到 `D:\webserver\php`

#### (2) 配置 `php.ini`
- 复制 `php.ini-development` 为 `php.ini`。
- 修改 `php.ini` 中的关键配置：
```ini
; 启用扩展（去掉前面的分号）
extension_dir = "ext"
extension=mysqli
extension=pdo_mysql
cgi.fix_pathinfo=0
date.timezone = Asia/Shanghai
```

#### (3) 测试 PHP
打开命令行，运行：
```bash
D:\webserver\php\php.exe -v
```
应输出 PHP 版本信息。

---

### 2. 安装并配置 Nginx

#### (1) 下载并解压 Nginx
- 下载 `nginx-1.xx.x.zip`，解压到 `D:\webserver\nginx`

#### (2) 修改 `nginx.conf`
路径：`D:\webserver\nginx\conf\nginx.conf`

修改 `server` 块如下：
```nginx
server {
    listen       80;
    server_name  localhost;

    root   D:/webserver/www;  # 网站根目录
    index  index.php index.html index.htm;

    location / {
        try_files $uri $uri/ =404;
    }

    # 处理 PHP 请求
    location ~ \.php$ {
        fastcgi_pass   127.0.0.1:9000;
        fastcgi_index  index.php;
        fastcgi_param  SCRIPT_FILENAME  $document_root$fastcgi_script_name;
        include        fastcgi_params;
    }
}
```

> 💡 注意：`root` 路径使用正斜杠 `/` 或双反斜杠 `\\`，避免单反斜杠。

#### (3) 启动 Nginx
命令行进入 Nginx 目录：
```bash
cd D:\webserver\nginx
start nginx
```
或运行 `nginx.exe`

访问 `http://localhost` 应看到 "Welcome to nginx!" 页面。

停止 Nginx：
```bash
nginx -s stop
```

---

### 3. 安装并配置 MySQL

#### (1) 下载 MySQL Community Server
- 下载 ZIP 版或 Installer 版。
- 解压到 `D:\webserver\mysql`

#### (2) 初始化 MySQL
以管理员身份打开命令提示符：
```bash
cd D:\webserver\mysql\bin
mysqld --initialize --console
```
会输出临时 root 密码（记下来）。

#### (3) 安装为 Windows 服务
```bash
mysqld --install MySQL
net start MySQL
```

#### (4) 登录并修改密码
```bash
mysql -u root -p
```
输入临时密码后执行：
```sql
ALTER USER 'root'@'localhost' IDENTIFIED BY 'your_new_password';
exit;
```

---

### 4. 配置 PHP-FPM（关键！）

PHP 需通过 FastCGI 与 Nginx 通信，使用 PHP-FPM。

#### (1) 启用 PHP-FPM
编辑 `D:\webserver\php\php.ini`，取消注释：
```ini
extension=php_openssl.dll
extension=php_mbstring.dll
extension=php_curl.dll
extension=php_pdo_mysql.dll
extension=php_mysqli.dll
```
确保 `php-fpm` 配置存在（`php\php-fpm.conf` 和 `php\php-fpm.d\www.conf`）。

#### (2) 启动 PHP-FPM
命令行运行：
```bash
D:\webserver\php\php-fpm.exe
```
> 💡 默认监听 `127.0.0.1:9000`，与 Nginx 配置一致。

---

## ✅ 三、测试环境

### 1. 创建测试文件
在 `D:\webserver\www` 目录创建 `index.php`：
```php
<?php
phpinfo();
?>
```

### 2. 启动服务顺序
```bash
# 1. 启动 MySQL
net start MySQL

# 2. 启动 PHP-FPM
D:\webserver\php\php-fpm.exe

# 3. 启动 Nginx
D:\webserver\nginx\nginx.exe
```

### 3. 访问测试
浏览器打开：`http://localhost`

应看到 PHP 信息页面，确认 Nginx + PHP + MySQL 正常工作。

---

## ✅ 四、常见问题解决

| 问题 | 解决方法 |
|------|----------|
| 502 Bad Gateway | 检查 PHP-FPM 是否运行，端口是否为 9000 |
| 页面不解析 PHP | 检查 `SCRIPT_FILENAME` 路径是否正确 |
| MySQL 无法启动 | 以管理员身份运行命令，检查端口 3306 是否被占用 |
| 中文乱码 | 在 PHP 文件中添加 `header("Content-Type: text/html; charset=utf-8");` |

---

## ✅ 五、可选：使用集成环境（推荐新手）

如果你不想手动配置，推荐使用：

- **Laragon**（轻量、支持一键切换版本）✅ 推荐
- **XAMPP**（包含 Apache，但也可切换为 Nginx）
- **WampServer**（默认 Apache，不推荐 Nginx）

> Laragon 特别适合开发 Laravel 或现代 PHP 项目。

---

## ✅ 六、开机自启（可选）

可将 Nginx、MySQL、PHP-FPM 注册为 Windows 服务，或使用批处理脚本启动。

例如创建 `start.bat`：
```bat
@echo off
net start MySQL
start "" "D:\webserver\php\php-fpm.exe"
start "" "D:\webserver\nginx\nginx.exe"
```

---

✅ 完成！你现在拥有了一个完整的 PHP + Nginx + MySQL 开发环境。

如需部署项目，只需将代码放入 `D:\webserver\www`，并配置数据库即可。

如有具体问题（如连接数据库失败、伪静态等），欢迎继续提问！
#-------------------------------------------------------------------------------

