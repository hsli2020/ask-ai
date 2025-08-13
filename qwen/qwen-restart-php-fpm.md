# 在ubuntu+nginx+php搭建的web服务器上，修改了php.ini之后，如何让它生效

在 Ubuntu + Nginx + PHP (通常是 PHP-FPM) 的环境下，修改 `php.ini` 后，**必须重启 PHP-FPM 服务**才能生效。仅仅重启 Nginx 是无效的，因为 Nginx 只是反向代理，不处理 PHP 代码。

以下是详细的操作步骤和注意事项：

### 1. 确认 PHP 版本
Ubuntu 上可能安装了多个 PHP 版本（如 7.4, 8.0, 8.1 等）。你需要知道当前 Nginx 正在使用哪个版本。

```bash
php -v
```
或者查看 `/etc/php/` 目录：
```bash
ls /etc/php/
```
假设你的版本是 **8.1**（请根据实际情况替换）。

### 2. 确认修改了正确的 php.ini 文件
PHP 有两个主要的配置文件，**修改 Web 服务配置必须修改 FPM 的那个**：

*   **CLI 模式 (命令行):** `/etc/php/8.1/cli/php.ini` (修改这个对网页无效)
*   **FPM 模式 (Web 服务):** `/etc/php/8.1/fpm/php.ini` (**必须修改这个**)

*提示：你可以创建一个 `<?php phpinfo(); ?>` 文件在网页中查看 "Loaded Configuration File" 路径，确认到底是哪个文件。*

### 3. 重启 PHP-FPM 服务
使用 `systemctl` 重启对应的 PHP-FPM 服务。

```bash
# 语法：sudo systemctl restart php<版本号>-fpm
sudo systemctl restart php8.1-fpm
```

如果不确定服务名称，可以尝试：
```bash
sudo systemctl restart php-fpm
```
*(注：在较新的 Ubuntu 版本中，通常需要带上版本号，如 `php8.1-fpm`)*

### 4. (可选) 重载 Nginx
虽然修改 `php.ini` 不需要重启 Nginx，但为了保险起见，或者如果你同时修改了 Nginx 的配置（如上传大小限制），可以重载 Nginx：

```bash
sudo systemctl reload nginx
```

### 5. 验证生效
1.  在网站根目录创建一个 `info.php` 文件，写入：
    ```php
    <?php phpinfo(); ?>
    ```
2.  在浏览器访问 `http://你的域名/info.php`。
3.  搜索你修改的配置项（例如 `memory_limit` 或 `upload_max_filesize`），查看 **Local Value** 和 **Master Value** 是否已更新。
4.  **重要：** 验证完成后，**删除** `info.php` 文件，以免泄露服务器信息。

---

### 常见坑点与补充

#### 1. 开启了 Opcache
如果你启用了 PHP Opcache (`opcache.enable=1`)，有时候修改 `php.ini` 后重启服务可能不会立即生效，因为脚本被缓存了。
*   **解决方法：** 确保重启了 PHP-FPM（通常重启服务会清除 Opcache 缓存）。如果仍然不行，检查 `opcache.ini` 中的 `opcache.revalidate_freq` 设置，或者临时关闭 Opcache 测试。

#### 2. 上传文件大小限制 (Nginx 端)
如果你在 `php.ini` 中修改了 `upload_max_filesize` 和 `post_max_size`，但上传大文件仍然失败，可能是因为 **Nginx** 也有限制。
*   **检查 Nginx 配置** (`/etc/nginx/nginx.conf` 或站点配置)：
    ```nginx
    http {
        # ...
        client_max_body_size 100M; # 确保这个值大于等于 PHP 的设置
    }
    ```
*   修改 Nginx 配置后，需要 `sudo systemctl reload nginx`。

#### 3. 权限问题
确保修改后的 `php.ini` 文件所有者和权限正确（通常属于 `root:root`，权限 `644`），否则 PHP-FPM 可能无法读取导致服务启动失败。

### 总结命令
假设你使用的是 PHP 8.1，最快的一条龙命令是：

```bash
sudo systemctl restart php8.1-fpm && sudo systemctl reload nginx
```

