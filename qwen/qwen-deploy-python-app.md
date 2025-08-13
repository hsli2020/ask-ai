# 我有一台Ubuntu服务器，已经安装了nginx，我想部署我用python+typescript+react+tailwindcss开发的web应用，请告诉我详细的操作步骤

部署一个前后端分离的 Web 应用到 Ubuntu + Nginx 环境，标准架构是：**Nginx 作为反向代理和静态文件服务器 → 前端构建产物由 Nginx 直接托管 → 后端 Python 服务通过 systemd 守护运行 → Nginx 将 API 请求代理到后端**。

以下是详细、可落地的操作步骤（以 Ubuntu 22.04/24.04 为例，命令需使用 `sudo` 或 root 权限）：

---
### 📦 0. 前置准备：确认目录结构规划
建议统一放在 `/var/www/yourapp/` 下：
```
/var/www/yourapp/
├── backend/      # Python 后端代码
├── frontend/     # React 前端构建产物（静态文件）
└── logs/         # 可选：应用日志目录
```

---
### 🔧 步骤 1：安装运行环境
```bash
sudo apt update && sudo apt upgrade -y

# 安装 Node.js (LTS 版本，用于构建前端)
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs build-essential

# 安装 Python 虚拟环境工具
sudo apt install -y python3 python3-pip python3-venv
```

---
### 🐍 步骤 2：部署 Python 后端
#### 2.1 上传代码并配置虚拟环境
```bash
sudo mkdir -p /var/www/yourapp/backend
# 使用 git clone 或 scp 将后端代码上传至此目录
cd /var/www/yourapp/backend

python3 -m venv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
```
> 💡 确保 `requirements.txt` 包含生产级 WSGI/ASGI 服务器：`fastapi`+`uvicorn`+`gunicorn`，或 `flask`+`gunicorn`，或 `django`+`gunicorn`。

#### 2.2 测试后端是否可启动
```bash
# 以 FastAPI 为例（根据你的框架调整命令）
uvicorn main:app --host 127.0.0.1 --port 8000
# 访问 http://127.0.0.1:8000/docs 或 /health 验证后 Ctrl+C 退出
```

#### 2.3 创建 systemd 服务（开机自启+崩溃重启）
```bash
sudo nano /etc/systemd/system/yourapp-backend.service
```
填入以下内容（**根据你的框架修改 `ExecStart`**）：
```ini
[Unit]
Description=YourApp Python Backend
After=network.target

[Service]
User=www-data
Group=www-data
WorkingDirectory=/var/www/yourapp/backend
Environment="PATH=/var/www/yourapp/backend/venv/bin"
# 如有 .env 文件可取消注释下一行
# EnvironmentFile=/var/www/yourapp/backend/.env
Restart=always
RestartSec=5

# ⬇️ 根据你的后端框架选择一行（删除其他）：
# FastAPI / Starlette:
ExecStart=/var/www/yourapp/backend/venv/bin/gunicorn -w 4 -k uvicorn.workers.UvicornWorker main:app --bind 127.0.0.1:8000
# Flask:
# ExecStart=/var/www/yourapp/backend/venv/bin/gunicorn -w 4 -b 127.0.0.1:8000 app:app
# Django:
# ExecStart=/var/www/yourapp/backend/venv/bin/gunicorn -w 4 -b 127.0.0.1:8000 yourproject.wsgi:application

[Install]
WantedBy=multi-user.target
```

启动并设为开机自启：
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now yourapp-backend
sudo systemctl status yourapp-backend  # 确认 active (running)
```

---
### ⚛️ 步骤 3：构建并部署 React 前端
```bash
sudo mkdir -p /var/www/yourapp/frontend
cd /path/to/your/react/project  # 本地或服务器上的源码目录

# 生产依赖安装（推荐用 ci 保证锁文件一致性）
npm ci

# 构建（Vite 默认输出 dist/，Create React App 默认输出 build/）
npm run build
```

将构建产物复制到 Nginx 目录：
```bash
# 根据你的构建工具替换 dist 或 build
sudo cp -r dist/* /var/www/yourapp/frontend/
sudo chown -R www-data:www-data /var/www/yourapp/frontend
```

---
### 🌐 步骤 4：配置 Nginx
#### 4.1 创建站点配置
```bash
sudo nano /etc/nginx/sites-available/yourapp
```
填入：
```nginx
server {
    listen 80;
    server_name your_domain.com 或 你的服务器公网IP;

    root /var/www/yourapp/frontend;
    index index.html;

    # 1. 前端静态文件 & React Router 路由支持（SPA 必备）
    location / {
        try_files $uri $uri/ /index.html;
    }

    # 2. 后端 API 反向代理（路径前缀根据你的实际接口调整，如 /api/ 或 /v1/）
    location /api/ {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # 支持 WebSocket（如果你的后端有）
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    # 3. 静态资源长期缓存
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    # 4. 安全头（可选）
    add_header X-Content-Type-Options nosniff;
    add_header X-Frame-Options SAMEORIGIN;
}
```

#### 4.2 启用配置并测试
```bash
sudo ln -s /etc/nginx/sites-available/yourapp /etc/nginx/sites-enabled/
sudo nginx -t   # 必须显示 syntax is ok 和 test is successful
sudo systemctl reload nginx
```

---
### 🔒 步骤 5：配置防火墙与 HTTPS（强烈建议）
```bash
# 开放 HTTP/HTTPS
sudo ufw allow 'Nginx Full'
sudo ufw status

# 自动申请 Let's Encrypt 证书
sudo apt install -y certbot python3-certbot-nginx
sudo certbot --nginx -d your_domain.com
# 按提示输入邮箱，选择 2（自动重定向到 HTTPS）
# 证书会每 90 天自动续期
```

---
### ✅ 步骤 6：验证与日常维护
| 操作 | 命令 |
|------|------|
| 查看后端日志 | `sudo journalctl -u yourapp-backend -f` |
| 查看 Nginx 访问/错误日志 | `sudo tail -f /var/log/nginx/access.log`<br>`sudo tail -f /var/log/nginx/error.log` |
| 重新构建前端后部署 | `npm run build && sudo cp -r dist/* /var/www/yourapp/frontend/` |
| 后端代码更新后重启 | `sudo systemctl restart yourapp-backend` |
| Nginx 重载配置 | `sudo nginx -t && sudo systemctl reload nginx` |

---
### ⚠️ 关键注意事项
1. **CORS 问题**：由于前端通过 Nginx 代理访问 `/api/`，浏览器认为请求是**同源**的，通常**不需要**在后端配置 CORS。如果必须直连后端端口（开发调试用），才需在后端配置 `CORS_ALLOW_ORIGINS`。
2. **路径前缀匹配**：`location /api/` 后的 `proxy_pass` **不要**加末尾斜杠，这样 `/api/users` 会完整转发到 `http://127.0.0.1:8000/api/users`。如果你的后端接口没有 `/api/` 前缀，需改用：
   ```nginx
   location /api/ {
       rewrite ^/api/(.*)$ /$1 break;
       proxy_pass http://127.0.0.1:8000;
       ...
   }
   ```
3. **环境变量**：生产环境敏感配置（数据库密码、密钥等）请放在 `.env` 文件，并在 systemd 中通过 `EnvironmentFile=` 加载，**不要**硬编码。
4. **权限安全**：示例中使用 `www-data` 运行后端是常见做法。如果后端需要写文件/上传，确保对应目录 `chown www-data:www-data`。

---
按照以上步骤，你的应用即可稳定运行。如果在某一步遇到报错，可提供 `journalctl` 或 `nginx error.log` 的具体输出，我会帮你精准定位。
