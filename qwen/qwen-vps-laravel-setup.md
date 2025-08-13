# 我有一台VPS服务器，运行Ubuntu 24.04，已经安装好了PHP+NGINX，并能从IP地址访问默认页面。我想安装一个Laravel开发的应用，请告诉我操作步骤

# 在 Ubuntu 24.04 VPS 上部署 Laravel 应用的完整指南

以下是基于你已有 PHP + Nginx 环境的完整部署流程，适用于 Laravel 10/11 项目。

---

## 一、检查 PHP 版本与扩展

Laravel 10/11 要求 PHP ≥ 8.1，建议 8.2+。

```bash
php -v
php -m | grep -E "mbstring|xml|curl|zip|gd|mysql|pdo|bcmath|intl"
```

如缺少扩展，请安装：

```bash
sudo apt update
sudo apt install php8.3-mbstring php8.3-xml php8.3-curl php8.3-zip php8.3-gd php8.3-mysql php8.3-pdo php8.3-bcmath php8.3-intl -y
sudo systemctl restart php8.3-fpm  # 或对应版本
```

---

## 二、安装 Composer

```bash
curl -sS https://getcomposer.org/installer | php
sudo mv composer.phar /usr/local/bin/composer
composer --version
```

---

## 三、创建 Laravel 项目

```bash
cd /var/www
sudo composer create-project laravel/laravel your-project-name
sudo chown -R www-data:www-data your-project-name
```

---

## 四、配置 Nginx

创建站点配置文件：

```bash
sudo nano /etc/nginx/sites-available/your-project-name
```

示例配置（替换域名与路径）：

```nginx
server {
    listen 80;
    server_name your-domain.com www.your-domain.com;
    root /var/www/your-project-name/public;

    index index.php;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;  # 根据实际版本调整
        fastcgi_param SCRIPT_FILENAME $realpath_root$fastcgi_script_name;
        include fastcgi_params;
    }

    location ~ /\.ht {
        deny all;
    }

    # 可选：禁止访问敏感文件
    location ~ /(\.env|\.git|composer\.(json|lock)) {
        deny all;
    }
}
```

启用站点并测试：

```bash
sudo ln -s /etc/nginx/sites-available/your-project-name /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

---

## 五、设置权限与环境变量

```bash
cd /var/www/your-project-name

# 设置权限
sudo chown -R www-data:www-data storage bootstrap/cache
sudo chmod -R 775 storage bootstrap/cache

# 创建 .env 文件
sudo -u www-data cp .env.example .env

# 生成应用密钥
sudo -u www-data php artisan key:generate

# 编辑 .env 配置（如数据库、APP_URL 等）
sudo nano .env
```

---

## 六、配置数据库（如使用 MySQL）

```bash
sudo apt install mysql-server -y
sudo mysql_secure_installation

sudo mysql -u root -p
```

在 MySQL 中执行：

```sql
CREATE DATABASE laravel_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'laravel_user'@'localhost' IDENTIFIED BY 'strong_password';
GRANT ALL PRIVILEGES ON laravel_db.* TO 'laravel_user'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

在 `.env` 中填写对应数据库信息，然后运行迁移：

```bash
sudo -u www-data php artisan migrate
```

---

## 七、可选优化配置

### 1. 启用 OPcache（提升性能）

```bash
sudo nano /etc/php/8.3/fpm/conf.d/10-opcache.ini
```

添加：

```ini
opcache.enable=1
opcache.memory_consumption=128
opcache.interned_strings_buffer=8
opcache.max_accelerated_files=4000
opcache.revalidate_freq=60
```

重启 PHP-FPM：

```bash
sudo systemctl restart php8.3-fpm
```

### 2. 配置队列（如使用 Redis）

```bash
sudo apt install redis-server -y
sudo systemctl enable redis-server
sudo systemctl start redis-server

# 在 .env 中设置
QUEUE_CONNECTION=redis
```

### 3. 设置定时任务

```bash
sudo crontab -u www-data -e
```

添加：

```bash
* * * * * cd /var/www/your-project-name && php artisan schedule:run >> /dev/null 2>&1
```

### 4. 配置日志轮转

```bash
sudo nano /etc/logrotate.d/laravel
```

内容：

```
/var/www/your-project-name/storage/logs/*.log {
    daily
    rotate 7
    compress
    delaycompress
    notifempty
    create 0664 www-data www-data
}
```

---

## 八、安全建议

- 设置防火墙：

```bash
sudo ufw allow 'Nginx Full'
sudo ufw allow OpenSSH
sudo ufw enable
```

- 配置 SSL（推荐 Let's Encrypt）：

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d your-domain.com -d www.your-domain.com
```

- 隐藏敏感文件（已在 Nginx 示例中配置）
- 定期更新系统：

```bash
sudo apt update && sudo apt upgrade -y
```

---

## 九、验证部署

```bash
# 检查权限
ls -la storage/
ls -la bootstrap/cache/

# 清除缓存
sudo -u www-data php artisan optimize:clear

# 测试访问
curl -I http://your-domain.com
```

访问 `http://your-domain.com` 应能看到 Laravel 欢迎页面。

---

## 十、常用命令速查

```bash
# 进入项目目录
cd /var/www/your-project-name

# 清除缓存
php artisan optimize:clear

# 查看日志
tail -f storage/logs/laravel.log

# 进入 Tinker
php artisan tinker

# 运行队列
php artisan queue:work

# 查看路由
php artisan route:list
```

---

## 故障排查

- 500 错误：查看 `storage/logs/laravel.log`
- 403 错误：检查目录权限与 Nginx 配置
- 502 错误：确认 PHP-FPM 是否运行
- 白屏：检查 `.env` 配置与 APP_DEBUG 设置

---

完成以上步骤后，你的 Laravel 应用应可正常运行。如需进一步部署生产环境（如使用 Horizon、队列、缓存优化等），可继续扩展配置。

