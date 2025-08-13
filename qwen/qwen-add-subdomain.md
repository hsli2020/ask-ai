# 在我的Ubuntu+Nginx服务器上，如何添加一个子域名？比如，我的站点是www.mysite.com，我想添加一个blog.mysite.com，请给出详细操作步骤

# 🚀 在 Ubuntu + Nginx 服务器上添加子域名的完整指南

假设您的主站是 `www.mysite.com`，现在要添加 `blog.mysite.com` 子域名，请按以下步骤操作：

---

## 📋 准备工作

- 服务器：Ubuntu + Nginx
- 域名管理权限（用于配置DNS）
- SSH 访问权限
- 已安装 Nginx（如未安装：`sudo apt install nginx`）

---

## 🔹 步骤 1：配置 DNS 记录

在您的域名注册商或 DNS 服务商处添加一条 **A 记录**：

| 类型 | 主机/名称 | 值 | TTL |
|------|----------|-----|-----|
| A | `blog` | `您的服务器IP` | 3600 |

> 示例：`blog.mysite.com → 192.0.2.100`

验证 DNS 是否生效：
```bash
ping blog.mysite.com
# 或
dig blog.mysite.com
```
> 等待直到解析到您的服务器IP [[1]]

---

## 🔹 步骤 2：创建网站目录

为子域名创建独立的文件目录：

```bash
# 创建目录
sudo mkdir -p /var/www/blog.mysite.com/html

# 设置权限（推荐使用非root用户运行）
sudo chown -R $USER:$USER /var/www/blog.mysite.com/html

# 设置目录权限
sudo chmod -R 755 /var/www/blog.mysite.com
```

创建测试页面：
```bash
nano /var/www/blog.mysite.com/html/index.html
```

```html
<!DOCTYPE html>
<html>
<head>
    <title>Blog - My Site</title>
</head>
<body>
    <h1>🎉 blog.mysite.com 配置成功！</h1>
</body>
</html>
```

---

## 🔹 步骤 3：创建 Nginx Server Block 配置

创建配置文件：
```bash
sudo nano /etc/nginx/sites-available/blog.mysite.com
```

### 📄 基础配置（HTTP）：
```nginx
server {
    listen 80;
    listen [::]:80;
    
    server_name blog.mysite.com www.blog.mysite.com;
    
    root /var/www/blog.mysite.com/html;
    index index.html index.htm;
    
    # 日志配置
    access_log /var/log/nginx/blog.mysite.com.access.log;
    error_log /var/log/nginx/blog.mysite.com.error.log;
    
    location / {
        try_files $uri $uri/ =404;
    }
}
```

### 🔁 如果是反向代理（如Node.js应用）：
```nginx
location / {
    proxy_pass http://127.0.0.1:3000;  # 替换为您的应用端口
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    
    # WebSocket 支持（如需要）
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
}
```
> WebSocket 配置对于实时应用非常重要 [[1]]

---

## 🔹 步骤 4：启用配置并测试

```bash
# 创建符号链接启用站点
sudo ln -s /etc/nginx/sites-available/blog.mysite.com /etc/nginx/sites-enabled/

# （可选）删除默认配置避免冲突
sudo rm /etc/nginx/sites-enabled/default

# 测试 Nginx 配置语法
sudo nginx -t

# 重载 Nginx 使配置生效
sudo systemctl reload nginx
```

> ✅ 如果 `nginx -t` 显示 `syntax is ok` 和 `test is successful`，说明配置正确 [[3]]

---

## 🔹 步骤 5：测试 HTTP 访问

在浏览器中访问：
```
http://blog.mysite.com
```
应该能看到您创建的测试页面。

---

## 🔹 步骤 6：🔐 配置 HTTPS（推荐 - 使用 Let's Encrypt）

### 安装 Certbot：
```bash
sudo apt update
sudo apt install certbot python3-certbot-nginx
```

### 自动获取并配置 SSL 证书：
```bash
sudo certbot --nginx -d blog.mysite.com -d www.blog.mysite.com
```

Certbot 会自动：
- ✅ 验证域名所有权
- ✅ 生成 SSL 证书
- ✅ 修改 Nginx 配置添加 HTTPS
- ✅ 设置 HTTP → HTTPS 重定向 [[1]]

### 验证自动续期：
```bash
sudo systemctl status certbot.timer
```

---

## 🔹 步骤 7：最终验证

```bash
# 检查 Nginx 状态
sudo systemctl status nginx

# 测试 HTTPS
curl -I https://blog.mysite.com

# 检查 SSL 证书
openssl s_client -connect blog.mysite.com:443 -servername blog.mysite.com
```

在浏览器访问 `https://blog.mysite.com`，确认：
- 🔒 地址栏显示安全锁图标
- 🔄 自动从 HTTP 跳转到 HTTPS
- 📄 正确显示您的博客内容

---

## 🛠️ 常见问题排查

| 问题 | 解决方案 |
|------|---------|
| **404 Not Found** | 检查 `root` 路径是否正确，文件权限是否为 755 |
| **502 Bad Gateway** | 反向代理时确认后端应用正在运行且端口正确 |
| **证书验证失败** | 确保 DNS 已生效，80 端口可公网访问 |
| **配置不生效** | 运行 `sudo nginx -t` 检查语法，然后 `sudo systemctl reload nginx` |
| **混合内容警告** | 确保页面内所有资源（图片、CSS、JS）都使用 HTTPS 链接 |

---

## 📦 附加：批量管理多个子域名

如果需要管理多个子域名，可以使用通配符证书：

```bash
# 申请通配符证书（需要 DNS 验证）
sudo certbot certonly --manual --preferred-challenges dns \
  -d mysite.com -d '*.mysite.com'
```

然后在每个子域名的 Nginx 配置中引用同一证书：
```nginx
ssl_certificate /etc/letsencrypt/live/mysite.com/fullchain.pem;
ssl_certificate_key /etc/letsencrypt/live/mysite.com/privkey.pem;
```

> ⚠️ 通配符证书续期需要自动化 DNS 验证，配置较复杂 [[27]]

---

## ✅ 配置完成清单

- [ ] DNS 记录已添加并生效
- [ ] 网站目录已创建且权限正确
- [ ] Nginx 配置文件已创建并启用
- [ ] `nginx -t` 测试通过
- [ ] HTTP 访问正常
- [ ] SSL 证书已安装（Certbot）
- [ ] HTTPS 访问正常且自动重定向
- [ ] 防火墙允许 80/443 端口：`sudo ufw allow 'Nginx Full'`

---

> 💡 **小贴士**：建议将子域名配置单独存放在 `/etc/nginx/sites-available/` 中，便于管理和备份。每次修改配置后务必先 `nginx -t` 测试，再重载服务，避免配置错误导致服务中断 [[3]]。

如有其他问题，欢迎继续提问！🎯
