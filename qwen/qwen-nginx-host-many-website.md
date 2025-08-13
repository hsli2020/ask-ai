# 我的一台web服务器，运行ubuntu和nginx，已经有一个网站在运行，比如www.site1.com，我想运行第二给网站，如www.site2.com，如何配置nginx

# 在Ubuntu+Nginx上配置第二个网站（虚拟主机）

您已经在运行 `www.site1.com`，现在要添加 `www.site2.com`，这是通过 **Nginx Server Blocks（服务器块）** 实现的。以下是详细步骤：

---

## 📋 前置条件

- 域名 `site2.com` 的DNS已解析到服务器公网IP [[4]]
- 您有 `sudo` 权限或root访问权限
- Nginx已安装并正常运行

---

## 🔧 配置步骤

### 1️⃣ 创建网站目录结构

```bash
# 创建site2.com的网站根目录
sudo mkdir -p /var/www/site2.com/public_html

# 创建测试页面
sudo nano /var/www/site2.com/public_html/index.html
```

添加以下内容作为测试：
```html
<!DOCTYPE html>
<html>
<head>
    <title>Welcome to site2.com</title>
</head>
<body>
    <h1>✅ site2.com 配置成功！</h1>
</body>
</html>
```

### 2️⃣ 设置目录权限

```bash
# 将目录所有权交给Nginx用户(www-data)
sudo chown -R www-data:www-data /var/www/site2.com

# 设置合理的读取权限
sudo chmod -R 755 /var/www/site2.com
```

### 3️⃣ 创建Server Block配置文件

```bash
sudo nano /etc/nginx/sites-available/site2.com
```

添加以下配置 [[4]]：

```nginx
server {
    listen 80;
    listen [::]:80;
    
    # 域名配置（支持带www和不带www）
    server_name site2.com www.site2.com;
    
    # 网站根目录
    root /var/www/site2.com/public_html;
    
    # 默认首页文件
    index index.html index.htm;
    
    # 请求处理逻辑
    location / {
        try_files $uri $uri/ =404;
    }
    
    # 日志文件（便于排查问题）
    access_log /var/log/nginx/site2.com.access.log;
    error_log /var/log/nginx/site2.com.error.log;
}
```

> 💡 如果您的网站使用PHP，需要在`location /`块中添加PHP-FPM配置，详见后文补充。

### 4️⃣ 启用网站配置

```bash
# 创建符号链接到sites-enabled目录
sudo ln -s /etc/nginx/sites-available/site2.com /etc/nginx/sites-enabled/

# （可选）如果默认配置会干扰，可禁用default
sudo rm /etc/nginx/sites-enabled/default
```

### 5️⃣ 测试并重载Nginx

```bash
# 测试配置语法是否正确
sudo nginx -t

# 如果显示 "syntax is ok" 和 "test is successful"，则重载配置
sudo systemctl reload nginx
```

---

## 🌐 验证配置

1. 确保 `site2.com` 的DNS A记录指向您的服务器IP
2. 在浏览器访问 `http://site2.com`，应看到测试页面
3. 检查日志：`sudo tail -f /var/log/nginx/site2.com.access.log`

---

## 🔐 补充：启用HTTPS（推荐）

使用Certbot免费获取Let's Encrypt证书 [[2]]：

```bash
# 安装Certbot
sudo apt install certbot python3-certbot-nginx

# 自动配置并启用HTTPS
sudo certbot --nginx -d site2.com -d www.site2.com

# 设置自动续期（通常已自动配置）
sudo systemctl status certbot.timer
```

---

## 🐘 补充：如果网站需要PHP支持

在server block的`location /`块后添加 [[10]]：

```nginx
location ~ \.php$ {
    include snippets/fastcgi-php.conf;
    fastcgi_pass unix:/run/php/php8.1-fpm.sock;  # 根据实际PHP版本调整
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
}
```

> ⚠️ 请确认PHP-FPM服务已安装并运行：`sudo systemctl status php8.1-fpm`

---

## 🛠️ 常见问题排查

| 问题 | 解决方案 |
|------|---------|
| 访问显示默认Nginx页面 | 检查`server_name`是否匹配域名；确认符号链接在`sites-enabled`中 |
| 403 Forbidden | 检查目录权限：`sudo chown -R www-data:www-data /var/www/site2.com` |
| 配置测试失败 | 运行`sudo nginx -t`查看具体错误行，常见原因是缺少分号或括号不匹配 |
| 端口被占用 | 运行`sudo ss -tlnp \| grep :80`查看哪个进程占用了80端口 |

---

## 📁 目录结构参考

```
/etc/nginx/
├── sites-available/          # 所有站点配置（未启用）
│   ├── site1.com
│   └── site2.com ← 新建
├── sites-enabled/            # 实际启用的配置（符号链接）
│   ├── site1.com → ../sites-available/site1.com
│   └── site2.com → ../sites-available/site2.com ← 新建链接
└── nginx.conf                # 主配置文件（通常无需修改）

/var/www/
├── site1.com/
│   └── public_html/
└── site2.com/ ← 新建
    └── public_html/
        └── index.html
```

---

✅ 完成以上步骤后，您的Nginx服务器就能同时托管 `site1.com` 和 `site2.com` 两个独立网站了！每个网站有独立的根目录、日志和配置，互不干扰 [[4]][[2]]。

如需添加更多网站，只需重复步骤1-5即可。如有其他问题（如反向代理、负载均衡等），欢迎继续提问！
