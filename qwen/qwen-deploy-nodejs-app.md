# 我有一台Ubuntu服务器，已经安装了nginx，我想部署我用nodejs开发的web应用，请告诉我详细的操作步骤

以下是部署 Node.js Web 应用到 Ubuntu 服务器的完整生产级步骤。整个流程分为 **环境准备 → 代码部署 → 进程守护 → Nginx 反向代理 → 安全加固** 五个阶段，命令已适配主流 Ubuntu 版本（20.04/22.04/24.04）。

---
### 📦 准备工作
- 确保已拥有 `sudo` 权限的 SSH 访问
- 域名已解析到服务器公网 IP（若暂用 IP 访问可跳过域名相关配置）
- 记录你的 Node 应用启动端口（默认为 `3000`，请根据实际情况替换）

---
### 🔹 步骤 1：安装 Node.js 运行环境
推荐使用 NodeSource 官方源安装 LTS 版本（以 Node 20 为例）：
```bash
# 1. 更新系统包索引
sudo apt update && sudo apt upgrade -y

# 2. 安装 Node.js 20 LTS
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs build-essential

# 3. 验证安装
node -v   # 应输出 v20.x.x
npm -v    # 应输出 10.x.x
```

---
### 🔹 步骤 2：部署应用代码
```bash
# 1. 创建应用目录（建议放在 /var/www 或当前用户家目录）
sudo mkdir -p /var/www/myapp
sudo chown -R $USER:$USER /var/www/myapp

# 2. 上传代码（任选其一）
# 方式 A：SCP 本地上传
scp -r ./your-project/* user@your_server_ip:/var/www/myapp/

# 方式 B：Git 克隆
cd /var/www/myapp && git clone https://github.com/yourname/your-repo.git .

# 3. 安装生产依赖
cd /var/www/myapp
npm ci --omit=dev   # 推荐：使用 package-lock.json 精确安装
# 或 npm install --production

# 4. 配置环境变量（如有）
nano .env   # 填入 PORT=3000, DATABASE_URL 等敏感配置
```
✅ **本地验证**：先手动运行 `node app.js`（或 `npm start`），确保能正常访问 `http://localhost:3000`，按 `Ctrl+C` 退出。

---
### 🔹 步骤 3：使用 PM2 管理进程（生产必备）
PM2 可保证应用崩溃自启、开机自启、日志管理、热更新。
```bash
# 1. 全局安装 PM2
sudo npm install -g pm2

# 2. 启动应用（替换 app.js 为你的入口文件）
pm2 start app.js --name myapp --env production

# 3. 保存进程列表 & 配置开机自启
pm2 save
pm2 startup   # ⚠️ 执行后会输出一行命令，复制粘贴执行即可
sudo systemctl enable pm2-$(whoami)

# 4. 常用命令
pm2 status      # 查看状态
pm2 logs myapp  # 查看实时日志
pm2 restart myapp
pm2 stop myapp
```

---
### 🔹 步骤 4：配置 Nginx 反向代理
```bash
# 1. 创建站点配置
sudo nano /etc/nginx/sites-available/myapp
```
粘贴以下内容（**注意替换 `your_domain.com` 和端口 `3000`**）：
```nginx
server {
    listen 80;
    server_name your_domain.com;  # 若用 IP 访问可写 _;

    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # 可选：静态资源由 Nginx 直接处理（提升性能）
    location /static/ {
        alias /var/www/myapp/public/;
        expires 30d;
    }
}
```
启用并重启 Nginx：
```bash
# 2. 创建软链接启用站点
sudo ln -s /etc/nginx/sites-available/myapp /etc/nginx/sites-enabled/

# 3. 移除默认配置（避免端口冲突，按需执行）
sudo rm /etc/nginx/sites-enabled/default

# 4. 测试配置并重载
sudo nginx -t
sudo systemctl reload nginx
```
此时访问 `http://your_domain.com` 或 `http://服务器IP` 即可看到你的应用。

---
### 🔹 步骤 5：安全与生产加固（强烈推荐）
#### 🔒 配置防火墙
```bash
sudo ufw allow OpenSSH
sudo ufw allow 'Nginx Full'  # 开放 80 和 443
sudo ufw enable
sudo ufw status
```
> ⚠️ **不要开放 Node 端口（如 3000）到公网**，仅允许 `127.0.0.1` 访问即可。

#### 🔐 配置 HTTPS（Let's Encrypt）
```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your_domain.com
```
按提示输入邮箱、同意协议、选择是否强制跳转 HTTPS。证书会自动续期。

#### 🛡️ 其他生产建议
- 使用非 root 用户运行 PM2（当前步骤已默认使用 `$USER`）
- 敏感配置勿写死代码，使用 `.env` 或云厂商 Secret Manager
- 定期更新：`sudo apt update && sudo npm update -g pm2`
- 监控日志：`pm2 logs`、`sudo journalctl -u nginx -f`

---
### 🐛 常见问题排查
| 现象 | 排查命令 |
|------|----------|
| Nginx 502 Bad Gateway | `pm2 status` 查看应用是否运行；检查 `proxy_pass` 端口是否匹配 |
| Nginx 403/404 | `sudo nginx -t` 检查语法；确认 `proxy_pass` 地址正确 |
| 应用启动失败 | `pm2 logs myapp` 查看报错；检查 `.env` 或依赖缺失 |
| 权限拒绝 | `sudo chown -R $USER:$USER /var/www/myapp` |

---
按此流程部署后，你的 Node.js 应用将具备：**进程守护、反向代理、HTTPS 加密、开机自启、日志追踪** 等生产级特性。如需配置域名多站点、负载均衡、Docker 化部署或 CI/CD 自动化，可告知具体需求，我将提供对应方案。
