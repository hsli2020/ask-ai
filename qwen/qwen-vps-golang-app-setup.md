# 我有一个VPS服务器运行Ubuntu+Nginx，已经有一个PHP开发的网站在运行，我想再增加一个网站，但这个网站是用golang开发的，我应该如何设置才能让这个golang编写的网站运行在同一台服务器上 

这是一个非常经典的部署场景。核心思路是利用 **Nginx 作为反向代理（Reverse Proxy）**。

Nginx 会根据访问的**域名**（`server_name`）来区分流量：
*   访问 **PHP 网站域名** -> Nginx 转发给 **PHP-FPM** 处理。
*   访问 **Go 网站域名** -> Nginx 转发给 **Go 程序监听的端口** 处理。

以下是详细的操作步骤：

### 第一步：准备 Go 应用程序

1.  **修改 Go 代码监听端口**
    确保你的 Go 程序**不要**监听 80 或 443 端口（因为这些端口被 Nginx 占用了）。通常使用 8080, 8081, 3000 等。
    为了安全，建议只监听本地回环地址 `127.0.0.1`，这样外网无法直接访问 Go 端口，必须经过 Nginx。

    ```go
    // 示例代码
    package main

    import (
        "fmt"
        "net/http"
    )

    func main() {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
            fmt.Fprintln(w, "Hello from Golang!")
        })

        // 关键点：绑定 127.0.0.1 和自定义端口 (例如 8081)
        err := http.ListenAndServe("127.0.0.1:8081", nil)
        if err != nil {
            panic(err)
        }
    }
    ```

2.  **编译并上传**
    在本地交叉编译或在服务器上编译：
    ```bash
    go build -o mygoapp main.go
    ```
    将编译好的二进制文件 `mygoapp` 上传到服务器，例如放在 `/var/www/go-site/` 目录下。
    ```bash
    sudo mkdir -p /var/www/go-site
    sudo chown -R $USER:$USER /var/www/go-site
    # 上传文件后...
    chmod +x /var/www/go-site/mygoapp
    ```

### 第二步：配置 Systemd 服务（让 Go 程序后台运行）

不要直接在终端运行 `./mygoapp`，否则关闭终端程序就停了。我们需要创建一个系统服务，让它开机自启、崩溃自动重启。

1.  **创建服务文件**
    ```bash
    sudo nano /etc/systemd/system/go-site.service
    ```

2.  **填入以下内容**
    ```ini
    [Unit]
    Description=My Golang Website
    After=network.target

    [Service]
    # 运行程序的用户，建议用 www-data 或专门创建的用户，这里为了简单先用当前用户或 www-data
    User=www-data
    Group=www-data
    # 程序所在目录
    WorkingDirectory=/var/www/go-site
    # 启动命令
    ExecStart=/var/www/go-site/mygoapp
    # 失败后自动重启
    Restart=always
    # 标准输出和错误日志记录到 journal
    StandardOutput=journal
    StandardError=journal
    # 环境变量 (如果有需要)
    Environment=PORT=8081

    [Install]
    WantedBy=multi-user.target
    ```

3.  **启动服务**
    ```bash
    # 重新加载配置
    sudo systemctl daemon-reload
    # 启动服务
    sudo systemctl start go-site
    # 设置开机自启
    sudo systemctl enable go-site
    # 查看状态 (确保是 active (running))
    sudo systemctl status go-site
    ```
    *注意：如果提示权限错误，确保 `/var/www/go-site/mygoapp` 对 `www-data` 用户可执行。*

### 第三步：配置 Nginx 反向代理

1.  **创建新的 Nginx 配置文件**
    在 `sites-available` 中为新域名创建配置：
    ```bash
    sudo nano /etc/nginx/sites-available/go-site.conf
    ```

2.  **填入配置内容**
    假设你的新域名是 `go.example.com`，Go 程序运行在 `8081` 端口。

    ```nginx
    server {
        listen 80;
        server_name go.example.com; # 你的新域名

        # 日志路径
        access_log /var/log/nginx/go-site-access.log;
        error_log /var/log/nginx/go-site-error.log;

        location / {
            # 关键：反向代理到 Go 程序监听的地址
            proxy_pass http://127.0.0.1:8081;

            # 传递必要的 Header 信息，否则 Go 程序可能获取不到真实 IP 或 Host
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;

            # 如果是 WebSocket 需要额外配置，普通 HTTP 不需要
            # proxy_http_version 1.1;
            # proxy_set_header Upgrade $http_upgrade;
            # proxy_set_header Connection "upgrade";
        }
    }
    ```

3.  **启用配置并重启 Nginx**
    ```bash
    # 创建软链接到 sites-enabled
    sudo ln -s /etc/nginx/sites-available/go-site.conf /etc/nginx/sites-enabled/

    # 测试配置是否正确
    sudo nginx -t

    # 如果测试通过，重载 Nginx
    sudo systemctl reload nginx
    ```

### 第四步：配置 HTTPS (SSL)

既然你已经有 PHP 网站在运行，大概率已经配置了 SSL。为新网站配置 SSL 推荐使用 **Certbot**。

```bash
# 安装 certbot (如果未安装)
sudo apt install certbot python3-certbot-nginx

# 自动获取证书并修改 Nginx 配置
sudo certbot --nginx -d go.example.com
```
Certbot 会自动修改你的 Nginx 配置文件，将 `listen 80` 改为 `listen 443 ssl` 并添加证书路径，同时配置 80 到 443 的重定向。

### 第五步：防火墙设置 (UFW)

确保服务器防火墙只开放了必要的端口。Go 程序绑定在 `127.0.0.1`，所以**不需要**在防火墙开放 8081 端口。只需要开放 80 和 443。

```bash
sudo ufw allow 'Nginx Full'
sudo ufw allow OpenSSH
sudo ufw enable
```

### 第六步：验证与排查

1.  **DNS 解析**：确保 `go.example.com` 的 A 记录已经指向了你的 VPS IP 地址。
2.  **访问测试**：在浏览器访问 `http://go.example.com` (或 https)。
3.  **如果无法访问**：
    *   检查 Go 服务状态：`sudo systemctl status go-site`
    *   查看 Go 应用日志：`sudo journalctl -u go-site -f`
    *   查看 Nginx 错误日志：`sudo tail -f /var/log/nginx/go-site-error.log`
    *   检查端口监听：`sudo netstat -tulpn | grep 8081` (应该显示 127.0.0.1:8081)

### 总结架构图

```text
用户浏览器
    |
    v
[ Nginx (Port 80/443) ] <--- 根据域名分流
    |-----------------------|
    |                       |
    v                       v
[ PHP-FPM ]           [ Go App (127.0.0.1:8081) ]
(旧网站)                (新网站，由 Systemd 守护)
```

### 常见注意事项

1.  **端口冲突**：确保 Go 程序使用的端口（如 8081）没有被其他程序占用。
2.  **静态文件**：如果你的 Go 程序需要服务于静态文件（CSS/JS/图片），最好在 Go 代码里处理，或者在 Nginx 里配置 `location /static/ { root ...; }` 直接由 Nginx 处理（性能更好）。
3.  **权限问题**：Systemd 中指定的 `User` 必须对二进制文件和日志目录有读取/执行权限。使用 `www-data` 通常比较安全且方便与 Nginx 配合。
4.  **资源限制**：如果 Go 程序内存占用高，注意 VPS 内存是否足够同时运行 PHP-FPM 和 Go 进程。

按照以上步骤，你就可以在同一台 Ubuntu + Nginx 服务器上完美共存 PHP 和 Golang 两个网站了。
