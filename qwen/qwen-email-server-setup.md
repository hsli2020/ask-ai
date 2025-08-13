# 在我的Ubuntu+Nginx服务器上，如何添加邮件服务功能，比如我的域名是www.mysite.com，我想拥有我的域名邮箱，如admin@mysite.com，请给出详细的设置步骤。


# Ubuntu + Nginx 服务器搭建域名邮箱完整指南

> ⚠️ **重要提示**：自建邮件服务器技术复杂，且容易被主流邮箱服务商（Gmail、QQ、163等）标记为垃圾邮件。如果是生产环境，建议优先考虑使用专业邮箱服务（腾讯企业邮、阿里企业邮、Google Workspace等）。如果坚持自建，请严格按以下步骤操作。

---

## 📋 前置条件

- Ubuntu 20.04/22.04/24.04 服务器（公网固定IP）
- 域名 `mysite.com` 已解析到服务器IP
- **确保云服务器已开放25端口**（多数云厂商默认关闭，需提交工单申请）[[56]]
- 已安装 Nginx（用于Webmail界面）

---

## 🔧 第一部分：基础邮件服务配置（Postfix + Dovecot）

### 1️⃣ 设置主机名

```bash
# 设置主机名为 mail.mysite.com
sudo hostnamectl set-hostname mail.mysite.com

# 编辑 /etc/hosts
sudo nano /etc/hosts
```
添加：
```
你的服务器公网IP  mail.mysite.com mail mysite.com
```

### 2️⃣ 安装 Postfix（SMTP发信服务）

```bash
sudo apt update
sudo apt install postfix -y
```
安装时会弹出配置界面：
- 选择 **"Internet Site"** → Enter
- 系统邮件名称输入：`mysite.com` → Enter [[1]]

### 3️⃣ 配置 Postfix

```bash
# 备份原配置
sudo cp /etc/postfix/main.cf /etc/postfix/main.cf.bak

# 编辑主配置
sudo nano /etc/postfix/main.cf
```

添加/修改以下关键配置：
```ini
# 基础设置
myhostname = mail.mysite.com
myorigin = /etc/mailname
mydestination = localhost.$mydomain, localhost
inet_interfaces = all
inet_protocols = all

# TLS加密设置
smtpd_use_tls = yes
smtpd_tls_cert_file = /etc/letsencrypt/live/mail.mysite.com/fullchain.pem
smtpd_tls_key_file = /etc/letsencrypt/live/mail.mysite.com/privkey.pem
smtpd_tls_security_level = may

# SASL认证（与Dovecot集成）
smtpd_sasl_auth_enable = yes
smtpd_sasl_type = dovecot
smtpd_sasl_path = private/auth

# 虚拟域名支持
virtual_transport = lmtp:unix:private/dovecot-lmtp
virtual_mailbox_domains = /etc/postfix/virtual_mailbox_domains

# 中继限制（防开放中继）
smtpd_relay_restrictions = permit_mynetworks, permit_sasl_authenticated, reject_unauth_destination
```

创建虚拟域名文件：
```bash
echo "mysite.com" | sudo tee /etc/postfix/virtual_mailbox_domains
sudo postmap /etc/postfix/virtual_mailbox_domains
```

编辑 `/etc/mailname`：
```
mail.mysite.com
```

### 4️⃣ 安装 Dovecot（IMAP/POP3收信服务）

```bash
sudo apt install dovecot-core dovecot-imapd dovecot-pop3d dovecot-lmtpd -y
```

#### 配置 Dovecot

**① 编辑 `/etc/dovecot/conf.d/10-mail.conf`**：
```ini
# 修改邮件存储位置
mail_location = maildir:/var/mail/vhosts/%d/%n
```

**② 创建邮件存储目录**：
```bash
sudo mkdir -p /var/mail/vhosts/mysite.com
sudo groupadd -g 5000 vmail
sudo useradd -r -g vmail -u 5000 vmail -d /var/mail/vhosts -c "virtual mail user"
sudo chown -R vmail:vmail /var/mail/vhosts/
```

**③ 编辑 `/etc/dovecot/conf.d/10-master.conf`**，启用与Postfix的集成：
```ini
# 找到 service auth 部分，添加：
service auth {
  unix_listener /var/spool/postfix/private/auth {
    mode = 0666
    user = postfix
    group = postfix
  }
}

# 找到 service lmtp 部分，添加：
service lmtp {
  unix_listener /var/spool/postfix/private/dovecot-lmtp {
    mode = 0600
    user = postfix
    group = postfix
  }
}
```

**④ 编辑 `/etc/dovecot/conf.d/10-auth.conf`**：
```ini
disable_plaintext_auth = yes
auth_mechanisms = plain login
# 注释掉系统认证
#!include auth-system.conf.ext
# 启用密码文件认证
!include auth-passwdfile.conf.ext
```

**⑤ 编辑 `/etc/dovecot/conf.d/auth-passwdfile.conf.ext`**：
```ini
passdb {
  driver = passwd-file
  args = scheme=PLAIN username_format=%u /etc/dovecot/dovecot-users
}
userdb {
  driver = static
  args = uid=vmail gid=vmail home=/var/mail/vhosts/%d/%n
}
```

**⑥ 创建用户密码文件**：
```bash
# 生成加密密码（推荐使用 doveadm pw 命令）
sudo doveadm pw -s SHA512-CRYPT
# 输入密码后会输出加密串，如：{SHA512-CRYPT}$6$xxx...

# 编辑用户文件
sudo nano /etc/dovecot/dovecot-users
```
添加用户（格式：`邮箱:加密密码`）：
```
admin@mysite.com:{SHA512-CRYPT}$6$xxx...
```

### 5️⃣ 配置 SSL/TLS 证书

```bash
# 安装 Certbot
sudo apt install certbot python3-certbot-nginx -y

# 申请 Let's Encrypt 证书
sudo certbot certonly --nginx -d mail.mysite.com -d mysite.com
```

**配置 Dovecot 使用证书**（`/etc/dovecot/conf.d/10-ssl.conf`）：
```ini
ssl = required
ssl_cert = </etc/letsencrypt/live/mail.mysite.com/fullchain.pem
ssl_key = </etc/letsencrypt/live/mail.mysite.com/privkey.pem
```

### 6️⃣ 启动服务

```bash
sudo systemctl restart postfix dovecot
sudo systemctl enable postfix dovecot
```

---

## 🌐 第二部分：配置 Webmail（Roundcube + Nginx）

### 1️⃣ 安装 Roundcube 及依赖

```bash
# 安装PHP及扩展
sudo apt install php-fpm php-mysql php-imap php-mbstring php-xml php-zip php-curl -y

# 下载 Roundcube
cd /tmp
wget https://github.com/roundcube/roundcubemail/releases/download/1.6.5/roundcubemail-1.6.5-complete.tar.gz
tar xzf roundcubemail-1.6.5-complete.tar.gz
sudo mv roundcubemail-1.6.5 /var/www/roundcube

# 设置权限
sudo chown -R www-data:www-data /var/www/roundcube
sudo chmod -R 755 /var/www/roundcube/temp /var/www/roundcube/logs
```

### 2️⃣ 创建 MariaDB 数据库

```bash
sudo apt install mariadb-server -y
sudo mysql -u root
```

```sql
CREATE DATABASE roundcubemail DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'roundcube'@'localhost' IDENTIFIED BY '你的强密码';
GRANT ALL PRIVILEGES ON roundcubemail.* TO 'roundcube'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

导入初始表结构：
```bash
sudo mysql roundcubemail < /var/www/roundcube/SQL/mysql.initial.sql
```

### 3️⃣ 配置 Nginx

```bash
sudo nano /etc/nginx/sites-available/mail.mysite.com
```

```nginx
server {
    listen 80;
    listen [::]:80;
    server_name mail.mysite.com;
    root /var/www/roundcube;
    index index.php;

    # 安全限制
    location ~ ^/(bin|SQL|logs|temp|installer) { deny all; }
    location ~ /\. { deny all; }

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include snippets/fastcgi-php.conf;
        fastcgi_pass unix:/run/php/php8.1-fpm.sock;  # 根据实际PHP版本调整
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    # 静态文件缓存
    location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff|woff2)$ {
        expires 30d;
        access_log off;
    }
}
```

启用配置：
```bash
sudo ln -s /etc/nginx/sites-available/mail.mysite.com /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

### 4️⃣ 配置 Roundcube

```bash
cd /var/www/roundcube/config
cp config.inc.php.sample config.inc.php
sudo nano config.inc.php
```

关键配置项：
```php
$config['db_dsnw'] = 'mysql://roundcube:你的密码@localhost/roundcubemail';
$config['imap_host'] = 'tls://mail.mysite.com:143';
$config['smtp_host'] = 'tls://mail.mysite.com:587';
$config['des_key'] = '生成一个32位随机字符串';  // 用于加密
$config['plugins'] = ['archive', 'zipdownload', 'managesieve', 'password'];
```

⚠️ **完成后务必删除安装目录**：
```bash
sudo rm -rf /var/www/roundcube/installer
```

---

## 📬 第三部分：DNS 记录配置（关键！）

在域名注册商或DNS服务商处添加以下记录：

| 记录类型 | 主机记录 | 记录值 | 说明 |
|---------|---------|--------|------|
| **A** | `mail` | `你的服务器公网IP` | 邮件服务器地址 |
| **MX** | `@` | `10 mail.mysite.com` | 邮件交换记录，优先级10 |
| **TXT (SPF)** | `@` | `v=spf1 ip4:你的IP ~all` | 防伪造，指定合法发信IP [[41]] |
| **TXT (DKIM)** | `dkim._domainkey` | `v=DKIM1; k=rsa; p=MIIBIjANBgkq...` | 邮件数字签名（需配置OpenDKIM）[[47]] |
| **TXT (DMARC)** | `_dmarc` | `v=DMARC1; p=quarantine; rua=mailto:admin@mysite.com` | 邮件认证策略 [[42]] |

### 🔐 配置 DKIM（推荐）

```bash
# 安装 OpenDKIM
sudo apt install opendkim opendkim-tools -y

# 生成密钥
sudo opendkim-genkey -b 2048 -d mysite.com -D /etc/opendkim/keys/mysite.com -s default -v

# 设置权限
sudo chown -R opendkim:opendkim /etc/opendkim/keys/mysite.com

# 配置 /etc/opendkim.conf
sudo nano /etc/opendkim.conf
```
添加：
```
Domain                  mysite.com
KeyFile                 /etc/opendkim/keys/mysite.com/default.private
Selector                default
Socket                  inet:8891@localhost
```

**将公钥添加到DNS**：
```bash
cat /etc/opendkim/keys/mysite.com/default.txt
# 复制输出内容作为 DKIM 的 TXT 记录值
```

---

## 🔥 第四部分：防火墙与安全配置

### 开放必要端口

```bash
# 如果使用 UFW
sudo ufw allow 25/tcp    # SMTP（发信，需云厂商开放）
sudo ufw allow 587/tcp   # SMTP Submission（客户端发信）
sudo ufw allow 465/tcp   # SMTPS（SSL加密发信）
sudo ufw allow 143/tcp   # IMAP
sudo ufw allow 993/tcp   # IMAPS（SSL加密收信）
sudo ufw allow 110/tcp   # POP3
sudo ufw allow 995/tcp   # POP3S
sudo ufw allow 80/tcp    # HTTP（Webmail）
sudo ufw allow 443/tcp   # HTTPS（Webmail）
```

> ⚠️ **25端口特别提示**：阿里云、腾讯云、AWS等默认屏蔽25端口，需提交工单申请解封，或使用中继服务（如SendGrid、Mailgun）[[56]][[61]]

### 基础安全加固

```bash
# 禁用 root 远程登录
sudo nano /etc/ssh/sshd_config
# 设置：PermitRootLogin no

# 安装 Fail2Ban 防暴力破解
sudo apt install fail2ban -y
sudo cp /etc/fail2ban/jail.conf /etc/fail2ban/jail.local
sudo systemctl enable --now fail2ban
```

---

## 🧪 测试验证

### 1. 本地测试发信
```bash
echo "Test email body" | mail -s "Test Subject" admin@mysite.com
```

### 2. 检查端口监听
```bash
sudo ss -tlnp | grep -E ':(25|143|993|587)\s'
```

### 3. Webmail 访问
浏览器访问：`https://mail.mysite.com`，使用 `admin@mysite.com` 登录

### 4. 外部测试工具
- [MX Toolbox](https://mxtoolbox.com/)：检查MX/SPF/DKIM/DMARC
- [Mail-Tester](https://www.mail-tester.com/)：发送测试邮件检测垃圾邮件评分

---

## 🚨 常见问题排查

| 问题 | 可能原因 | 解决方案 |
|-----|---------|---------|
| 邮件进垃圾箱 | 缺少SPF/DKIM/DMARC | 完善DNS认证记录 [[40]] |
| 无法收信 | Dovecot未启动/配置错误 | `sudo systemctl status dovecot` 检查日志 |
| Webmail 502 | PHP-FPM未运行 | `sudo systemctl status php8.1-fpm` |
| 25端口不通 | 云厂商屏蔽 | 提交工单申请或使用SMTP中继 |
| TLS证书错误 | 证书路径/权限问题 | 检查 `/etc/letsencrypt/` 文件权限 |

查看日志：
```bash
# Postfix日志
sudo tail -f /var/log/mail.log

# Dovecot日志
sudo doveconf -n | grep log_path
sudo tail -f /var/log/dovecot.log

# Nginx错误日志
sudo tail -f /var/log/nginx/error.log
```

---

## 💡 专业建议

1. **生产环境强烈建议使用专业邮箱服务**，自建维护成本高、送达率难保障
2. 定期检查服务器IP是否在 [Spamhaus](https://www.spamhaus.org/) 等黑名单
3. 配置自动备份：`/var/mail/vhosts/` 和 Roundcube 数据库
4. 考虑使用 [iRedMail](https://www.iredmail.org/) 等一键安装包简化管理

> 📚 更多参考：
> - [Server World Ubuntu Mail Guide](https://www.server-world.info/en/note?os=Ubuntu_24.04&p=mail) [[1]]
> - [Linux Terminal Roundcube教程](https://cn.linux-terminal.com/?p=3308) [[22]]
> - [SPF/DKIM/DMARC详解](https://www.voidking.com/dev-mail-spf-dkim-dmarc/) [[41]]

如有具体报错，请提供日志内容，我可帮你进一步分析！🔧

