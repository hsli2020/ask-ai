# 在nginx中，同一域名下，如何添加多端口

# Nginx 同一域名配置多端口方法

在 Nginx 中，可以通过为同一域名配置多个 `server` 块，每个块监听不同端口，实现同一域名下多端口服务。以下是完整配置指南。

---

## 一、基础配置示例

```nginx
# 80 端口
server {
    listen 80;
    server_name example.com www.example.com;
    
    location / {
        root /var/www/html;
        index index.html;
    }
}

# 8080 端口
server {
    listen 8080;
    server_name example.com www.example.com;
    
    location / {
        root /var/www/app;
        index index.html;
    }
}

# 443 端口 (HTTPS)
server {
    listen 443 ssl;
    server_name example.com www.example.com;
    
    ssl_certificate /etc/nginx/ssl/example.com.crt;
    ssl_certificate_key /etc/nginx/ssl/example.com.key;
    
    location / {
        root /var/www/secure;
        index index.html;
    }
}
```

---

## 二、多端口反向代理示例

```nginx
# 80 端口 - 主站
server {
    listen 80;
    server_name example.com;
    
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# 8080 端口 - API 服务
server {
    listen 8080;
    server_name example.com;
    
    location / {
        proxy_pass http://127.0.0.1:3001;
        proxy_set_header Host $host;
    }
}

# 8888 端口 - 管理后台
server {
    listen 8888;
    server_name example.com;
    
    location / {
        proxy_pass http://127.0.0.1:3002;
        # 可添加访问控制
        allow 192.168.1.0/24;
        deny all;
    }
}
```

---

## 三、完整配置示例（含常用优化）

```nginx
http {
    # 80 端口
    server {
        listen 80;
        server_name example.com;
        
        # 强制跳转 HTTPS（可选）
        # return 301 https://$server_name$request_uri;
        
        location / {
            root /var/www/main;
            index index.html index.htm;
            try_files $uri $uri/ =404;
        }
        
        # 静态资源缓存
        location ~* \.(jpg|jpeg|png|gif|ico|css|js)$ {
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
    
    # 8080 端口 - 测试环境
    server {
        listen 8080;
        server_name example.com;
        
        location / {
            root /var/www/test;
            index index.html;
        }
        
        # 访问日志独立
        access_log /var/log/nginx/test_access.log;
        error_log /var/log/nginx/test_error.log;
    }
    
    # 443 端口 - HTTPS
    server {
        listen 443 ssl http2;
        server_name example.com;
        
        ssl_certificate /etc/nginx/ssl/example.com.crt;
        ssl_certificate_key /etc/nginx/ssl/example.com.key;
        ssl_protocols TLSv1.2 TLSv1.3;
        ssl_ciphers HIGH:!aNULL:!MD5;
        
        location / {
            proxy_pass http://127.0.0.1:3000;
            proxy_set_header Host $host;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

---

## 四、配置验证与生效

```bash
# 1. 检查配置语法
nginx -t

# 2. 重新加载配置（不中断服务）
nginx -s reload

# 3. 或重启服务
systemctl restart nginx

# 4. 查看监听端口
netstat -tlnp | grep nginx
# 或
ss -tlnp | grep nginx
```

---

## 五、防火墙配置（如使用 firewalld）

```bash
# 开放多个端口
firewall-cmd --permanent --add-port=80/tcp
firewall-cmd --permanent --add-port=8080/tcp
firewall-cmd --permanent --add-port=443/tcp
firewall-cmd --reload

# 查看已开放端口
firewall-cmd --list-ports
```

---

## 六、注意事项

| 项目 | 说明 |
|------|------|
| **端口占用** | 确保端口未被其他程序占用 |
| **权限问题** | 1024 以下端口需要 root 权限 |
| **防火墙** | 记得开放对应端口 |
| **SSL证书** | 不同端口可共用同一证书 |
| **日志分离** | 建议不同端口使用独立日志文件 |
| **安全加固** | 管理端口建议限制访问 IP |

---

## 七、测试访问

```bash
# 测试不同端口
curl http://example.com          # 80 端口
curl http://example.com:8080     # 8080 端口
curl https://example.com         # 443 端口

# 查看响应头
curl -I http://example.com:8080
```

---

通过以上配置，您可以在同一域名下灵活地监听多个端口，分别承载不同服务或环境，满足多样化部署需求。
