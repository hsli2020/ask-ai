# 🤵 我想用php建一个简单的网站，实现如下功能：
1、网站的某个目录中保存有很多虚拟数字产品，如电子书、音频、视频文件等。
2、我在其它的网上商城如esty等销售这些虚拟产品。
3、用户购买后，获得一个“提取码”。
4、用户来到我的网站，输入“提取码”，下载相应虚拟产品。

Qwen3-235B-A22B-2507

这是一个非常实用的场景，使用 PHP 可以轻松实现一个基于“提取码”下载虚拟产品的系统。
下面我为你提供一个**简单、安全、可运行**的实现方案，包含基本功能和安全建议。

---

## ✅ 功能需求总结

1. 存放虚拟产品文件（如 PDF、MP3、MP4 等）在一个服务器目录中。
2. 每个文件对应一个唯一的“提取码”。
3. 用户在其他平台（如 Etsy）购买后，你手动或自动发送提取码。
4. 用户访问你的网站，输入提取码，验证后可下载对应文件。

---

## 🗂️ 目录结构建议

```
/your-website/
├── index.php           # 用户输入提取码的页面
├── download.php        # 验证提取码并触发下载
├── config.php          # 配置文件（数据库连接、路径等）
├── products/           # 存放实际的虚拟产品文件（禁止直接访问）
│   ├── ebook1.pdf
│   ├── audio1.mp3
│   └── video1.mp4
└── .htaccess           # 保护 products 目录（Apache）
```

---

## 🔐 第一步：保护 products 目录

防止用户直接访问 `/products/` 下的文件。

### `.htaccess` 文件（放在 products 目录内）

```apache
Deny from all
```

> 或者使用 Nginx 的话，在配置中禁止访问该目录。

---

## 🛠️ 第二步：数据库设计（简单用数组或 SQLite/MySQL）

### 方案 A：简单版（用 PHP 数组，适合少量产品）

#### `config.php`

```php
<?php
// 产品列表：提取码 => 文件信息
$products = [
    'ABC123' => [
        'filename' => 'ebook1.pdf',
        'filepath' => __DIR__ . '/products/ebook1.pdf',
        'download_limit' => 5,    // 允许下载次数
        'used_count' => 0         // 当前已下载次数（简单模拟）
    ],
    'XYZ789' => [
        'filename' => 'audio1.mp3',
        'filepath' => __DIR__ . '/products/audio1.mp3',
        'download_limit' => 3,
        'used_count' => 0
    ],
    // 更多产品...
];

// 保存状态的文件（模拟数据库）
$state_file = __DIR__ . '/storage/state.json';

function loadState() {
    global $state_file, $products;
    if (file_exists($state_file)) {
        $data = json_decode(file_get_contents($state_file), true);
        foreach ($data as $code => $info) {
            if (isset($products[$code])) {
                $products[$code]['used_count'] = $info['used_count'];
            }
        }
    }
}

function saveState() {
    global $state_file, $products;
    $data = [];
    foreach ($products as $code => $info) {
        $data[$code] = ['used_count' => $info['used_count']];
    }
    file_put_contents($state_file, json_encode($data, JSON_PRETTY_PRINT));
}

loadState();
?>
```

> ⚠️ 注意：生产环境建议使用 MySQL 或 SQLite。

---

## 🖼️ 第三步：前端输入页面 `index.php`

```php
<?php include 'config.php'; ?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>输入提取码下载</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 500px; margin: 0 auto; }
        input[type="text"] { width: 100%; padding: 10px; font-size: 16px; margin: 10px 0; }
        button { padding: 10px 20px; font-size: 16px; }
    </style>
</head>
<body>
<div class="container">
    <h2>输入提取码下载您的数字产品</h2>
    <form action="download.php" method="post">
        <label>提取码：</label>
        <input type="text" name="code" placeholder="例如：ABC123" required>
        <button type="submit">下载</button>
    </form>
</div>
</body>
</html>
```

---

## 🔽 第四步：处理下载 `download.php`

```php
<?php
include 'config.php';

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    die('非法请求');
}

$code = strtoupper(trim($_POST['code']));

if (!isset($products[$code])) {
    die('❌ 提取码无效，请检查后重试。');
}

$product = $products[$code];

if ($product['used_count'] >= $product['download_limit']) {
    die('❌ 该提取码已达到最大下载次数，无法继续下载。');
}

if (!file_exists($product['filepath'])) {
    die('❌ 文件不存在，请联系客服。');
}

// 增加下载次数
$product['used_count']++;
$products[$code]['used_count'] = $product['used_count'];
saveState();

// 强制浏览器下载文件
$filename = $product['filename'];
$file = $product['filepath'];

header('Content-Description: File Transfer');
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="' . basename($filename) . '"');
header('Expires: 0');
header('Cache-Control: must-revalidate');
header('Pragma: public');
header('Content-Length: ' . filesize($file));
readfile($file);
exit;
```

---

## ✅ 使用流程

1. 你上传 `ebook1.pdf` 到 `products/` 目录。
2. 在 `config.php` 中为它分配提取码 `'ABC123'`。
3. 用户在 Etsy 购买后，你把 `ABC123` 发给他。
4. 用户访问你的网站，输入 `ABC123`，点击下载。
5. 系统验证后下载文件，并记录次数。

---

## 🔒 安全增强建议（进阶）

| 项目 | 建议 |
|------|------|
| 数据库存储 | 使用 MySQL 存储提取码、使用次数、过期时间等 |
| 提取码生成 | 自动生成唯一码（如 `bin2hex(random_bytes(3))`） |
| 过期时间 | 添加 `expires_at` 字段，限制有效期 |
| 日志记录 | 记录 IP、时间、提取码使用情况 |
| 防刷机制 | 同一 IP 限制下载频率 |
| HTTPS | 必须启用，防止提取码被窃听 |

---

## 💡 可扩展功能

- 后台管理页面（生成提取码、查看使用记录）
- 批量生成提取码
- 邮件自动发送提取码（集成 SMTP）
- 支持二维码扫描提取码下载

---

## ✅ 总结

你现在可以：

1. 把文件放进 `products/`
2. 在 `config.php` 中配置提取码
3. 上传到你的 PHP 主机（如阿里云、腾讯云、Bluehost 等）
4. 用户访问网站输入提取码即可下载

> ✅ 代码简单、安全、可运行，适合个人数字产品分发。

---

如果你希望我帮你生成一个 **完整可下载的 ZIP 包**，或者升级为 **MySQL 版本** 或 **带后台管理的版本**，也可以告诉我，我可以继续帮你完善。

祝你项目顺利！🚀

--------------------------------------------------------------------------------

# 🤵 请升级为MySQL版本，带后台管理功能

当然可以！以下是升级后的 **MySQL + 后台管理功能** 的完整 PHP 虚拟产品分发系统，支持：

- ✅ MySQL 存储提取码、产品信息、下载记录
- ✅ 管理员后台：生成提取码、查看使用情况
- ✅ 用户前端：输入提取码下载文件
- ✅ 安全控制：下载次数限制、过期时间、防直接访问
- ✅ 响应式界面

---

## 📁 项目目录结构

```
/virtual-product-system/
├── index.php               # 用户输入提取码页面
├── download.php            # 下载处理
├── admin/                  # 管理后台
│   ├── login.php           # 登录页
│   ├── dashboard.php       # 后台首页（生成码、查看记录）
│   ├── logout.php
│   └── includes/
│       ├── db.php          # 数据库连接
│       └── auth.php        # 管理员认证
├── products/               # 存放虚拟产品（禁止直接访问）
│   └── .htaccess
├── config.php              # 全局配置
├── .htaccess               # 根目录保护
└── install.php             # 安装脚本（首次运行）
```

---

## 🛠️ 第一步：创建数据库表

### SQL 脚本（可导入 phpMyAdmin）

```sql
CREATE DATABASE IF NOT EXISTS `virtual_store`;
USE `virtual_store`;

-- 产品表
CREATE TABLE IF NOT EXISTS `products` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `title` VARCHAR(255) NOT NULL,
  `filename` VARCHAR(255) NOT NULL,
  `file_path` TEXT NOT NULL,
  `download_limit` INT DEFAULT 5,
  `expire_days` INT DEFAULT 7,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 提取码表
CREATE TABLE IF NOT EXISTS `download_codes` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `code` CHAR(8) NOT NULL UNIQUE,
  `product_id` INT NOT NULL,
  `used_count` INT DEFAULT 0,
  `max_downloads` INT DEFAULT 5,
  `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
  `expires_at` DATETIME,
  FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- 下载日志
CREATE TABLE IF NOT EXISTS `download_logs` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `code` CHAR(8) NOT NULL,
  `product_id` INT NOT NULL,
  `ip_address` VARCHAR(45),
  `user_agent` TEXT,
  `downloaded_at` DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 管理员账号表
CREATE TABLE IF NOT EXISTS `admins` (
  `id` INT AUTO_INCREMENT PRIMARY KEY,
  `username` VARCHAR(50) NOT NULL UNIQUE,
  `password` VARCHAR(255) NOT NULL  -- 哈希存储
);

-- 插入默认管理员账号: 用户名 admin, 密码 123456
INSERT INTO `admins` (`username`, `password`) 
VALUES ('admin', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'); -- 123456 的哈希
```

> ⚠️ 默认密码是 `123456`，登录后请立即修改！

---

## 📄 `config.php`

```php
<?php
session_start();

define('ROOT_PATH', __DIR__);
define('PRODUCTS_DIR', ROOT_PATH . '/products/');

// 数据库配置
define('DB_HOST', 'localhost');
define('DB_NAME', 'virtual_store');
define('DB_USER', 'root');
define('DB_PASS', '');

// 网站标题
define('SITE_TITLE', '数字产品提取中心');
```

---

## 🔐 `admin/includes/db.php`

```php
<?php
require_once __DIR__ . '/../../config.php';

try {
    $pdo = new PDO(
        "mysql:host=" . DB_HOST . ";dbname=" . DB_NAME . ";charset=utf8mb4",
        DB_USER,
        DB_PASS,
        [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
    );
} catch (PDOException $e) {
    die("数据库连接失败: " . $e->getMessage());
}
```

---

## 🔐 `admin/includes/auth.php`

```php
<?php
require_once 'db.php';

function isAdminLoggedIn() {
    return isset($_SESSION['admin_logged_in']) && $_SESSION['admin_logged_in'] === true;
}

function loginAdmin($username, $password) {
    global $pdo;
    $stmt = $pdo->prepare("SELECT * FROM admins WHERE username = ?");
    $stmt->execute([$username]);
    $admin = $stmt->fetch();

    if ($admin && password_verify($password, $admin['password'])) {
        $_SESSION['admin_logged_in'] = true;
        $_SESSION['admin_username'] = $username;
        return true;
    }
    return false;
}

function logoutAdmin() {
    session_destroy();
}
```

---

## 🔐 `admin/login.php`

```php
<?php
require_once 'includes/auth.php';

if (isAdminLoggedIn()) {
    header('Location: dashboard.php');
    exit;
}

$error = '';

if ($_SERVER['REQUEST_METHOD'] === 'POST') {
    $username = trim($_POST['username']);
    $password = $_POST['password'];

    if (loginAdmin($username, $password)) {
        header('Location: dashboard.php');
        exit;
    } else {
        $error = '用户名或密码错误';
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>管理员登录 - <?= SITE_TITLE ?></title>
    <style>
        body { font-family: Arial; margin: 0; padding: 40px; background: #f4f4f4; }
        .login-box { max-width: 400px; margin: 50px auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 0 10px rgba(0,0,0,0.1); }
        input[type="text"], input[type="password"] { width: 100%; padding: 10px; margin: 10px 0; border: 1px solid #ddd; border-radius: 5px; }
        button { padding: 10px 20px; background: #007cba; color: white; border: none; border-radius: 5px; cursor: pointer; }
        .error { color: red; }
    </style>
</head>
<body>
<div class="login-box">
    <h2>管理员登录</h2>
    <?php if ($error): ?>
        <p class="error"><?= htmlspecialchars($error) ?></p>
    <?php endif; ?>
    <form method="post">
        <input type="text" name="username" placeholder="用户名" required><br>
        <input type="password" name="password" placeholder="密码" required><br>
        <button type="submit">登录</button>
    </form>
</div>
</body>
</html>
```

---

## 🖥️ `admin/dashboard.php`

```php
<?php
require_once 'includes/auth.php';
require_once 'includes/db.php';

if (!isAdminLoggedIn()) {
    header('Location: login.php');
    exit;
}

$success = $error = '';

// 生成提取码
if (isset($_POST['generate'])) {
    $product_id = (int)$_POST['product_id'];
    $count = (int)$_POST['count'] ?: 1;

    try {
        $pdo->beginTransaction();

        for ($i = 0; $i < $count; $i++) {
            $code = strtoupper(substr(bin2hex(random_bytes(4)), 0, 8)); // 8位随机码
            $stmt = $pdo->prepare("SELECT id FROM download_codes WHERE code = ?");
            $stmt->execute([$code]);
            if ($stmt->rowCount() > 0) continue; // 重复则跳过

            $stmt = $pdo->prepare("
                INSERT INTO download_codes (code, product_id, max_downloads, expires_at)
                VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL ? DAY))
            ");
            $stmt->execute([$code, $product_id, 5, 7]); // 默认5次，7天过期
        }

        $pdo->commit();
        $success = "成功生成 {$count} 个提取码！";
    } catch (Exception $e) {
        $pdo->rollback();
        $error = "生成失败: " . $e->getMessage();
    }
}

// 获取产品列表
$products = $pdo->query("SELECT id, title, filename FROM products ORDER BY title")->fetchAll();

// 获取最近提取码
$codes = $pdo->query("
    SELECT dc.*, p.title as product_title 
    FROM download_codes dc
    JOIN products p ON p.id = dc.product_id
    ORDER BY dc.created_at DESC
    LIMIT 50
")->fetchAll();

// 获取下载日志
$logs = $pdo->query("
    SELECT dl.*, p.title as product_title 
    FROM download_logs dl
    JOIN products p ON p.id = dl.product_id
    ORDER BY dl.downloaded_at DESC
    LIMIT 30
")->fetchAll();
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>管理后台 - <?= SITE_TITLE ?></title>
    <style>
        body { font-family: Arial; margin: 20px; }
        .container { max-width: 1200px; margin: 0 auto; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { padding: 10px; text-align: left; border: 1px solid #ddd; }
        th { background: #f0f0f0; }
        .btn { padding: 8px 12px; background: #007cba; color: white; border: none; border-radius: 4px; cursor: pointer; }
        .success { color: green; }
        .error { color: red; }
        .logout { float: right; }
    </style>
</head>
<body>
<div class="container">
    <h1>管理后台</h1>
    <a class="logout btn" href="logout.php">登出</a>
    <p>欢迎，<?= $_SESSION['admin_username'] ?>！</p>

    <?php if ($success) echo "<p class='success'>$success</p>"; ?>
    <?php if ($error) echo "<p class='error'>$error</p>"; ?>

    <h2>生成提取码</h2>
    <form method="post">
        <select name="product_id" required>
            <option value="">选择产品</option>
            <?php foreach ($products as $p): ?>
                <option value="<?= $p['id'] ?>"><?= htmlspecialchars($p['title']) ?></option>
            <?php endforeach; ?>
        </select>
        <input type="number" name="count" value="1" min="1" max="100" required>
        <button type="submit" name="generate" class="btn">生成提取码</button>
    </form>

    <h2>最近生成的提取码</h2>
    <table>
        <tr>
            <th>提取码</th>
            <th>产品</th>
            <th>已用/上限</th>
            <th>过期时间</th>
            <th>创建时间</th>
        </tr>
        <?php foreach ($codes as $c): ?>
        <tr>
            <td><strong><?= $c['code'] ?></strong></td>
            <td><?= htmlspecialchars($c['product_title']) ?></td>
            <td><?= $c['used_count'] ?>/<?= $c['max_downloads'] ?></td>
            <td><?= $c['expires_at'] ?></td>
            <td><?= $c['created_at'] ?></td>
        </tr>
        <?php endforeach; ?>
    </table>

    <h2>最近下载记录</h2>
    <table>
        <tr>
            <th>提取码</th>
            <th>产品</th>
            <th>IP地址</th>
            <th>设备</th>
            <th>时间</th>
        </tr>
        <?php foreach ($logs as $l): ?>
        <tr>
            <td><?= $l['code'] ?></td>
            <td><?= htmlspecialchars($l['product_title']) ?></td>
            <td><?= $l['ip_address'] ?></td>
            <td><?= htmlspecialchars(substr($l['user_agent'], 0, 30)) ?>...</td>
            <td><?= $l['downloaded_at'] ?></td>
        </tr>
        <?php endforeach; ?>
    </table>
</div>
</body>
</html>
```

---

## 🔐 `admin/logout.php`

```php
<?php
require_once 'includes/auth.php';
logoutAdmin();
header('Location: login.php');
exit;
```

---

## 🔽 `download.php`（MySQL 版）

```php
<?php
require 'config.php';
require 'admin/includes/db.php';

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    die('非法请求');
}

$code = strtoupper(trim($_POST['code']));

// 验证提取码
$stmt = $pdo->prepare("
    SELECT dc.*, p.filename, p.file_path 
    FROM download_codes dc
    JOIN products p ON p.id = dc.product_id
    WHERE dc.code = ?
");
$stmt->execute([$code]);
$row = $stmt->fetch();

if (!$row) {
    die('❌ 提取码无效。');
}

if ($row['used_count'] >= $row['max_downloads']) {
    die('❌ 此提取码已达到最大下载次数。');
}

if (new DateTime($row['expires_at']) < new DateTime()) {
    die('❌ 此提取码已过期。');
}

$file_path = $row['file_path'];
if (!file_exists($file_path) || !is_readable($file_path)) {
    die('❌ 文件不存在或无法读取。');
}

// 记录下载日志
$pdo->prepare("
    INSERT INTO download_logs (code, product_id, ip_address, user_agent)
    VALUES (?, ?, ?, ?)
")->execute([
    $code,
    $row['product_id'],
    $_SERVER['REMOTE_ADDR'],
    $_SERVER['HTTP_USER_AGENT']
]);

// 更新使用次数
$pdo->prepare("UPDATE download_codes SET used_count = used_count + 1 WHERE code = ?")
    ->execute([$code]);

// 强制下载
$filename = $row['filename'];
header('Content-Description: File Transfer');
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="' . basename($filename) . '"');
header('Content-Length: ' . filesize($file_path));
header('Expires: 0');
header('Cache-Control: must-revalidate');
readfile($file_path);
exit;
```

---

## 🖼️ `index.php`（用户页面）

```php
<?php require 'config.php'; ?>
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title><?= SITE_TITLE ?></title>
    <style>
        body { font-family: 'Segoe UI', sans-serif; margin: 0; padding: 40px; background: #f8f9fa; }
        .container { max-width: 500px; margin: 0 auto; background: white; padding: 40px; border-radius: 10px; box-shadow: 0 0 20px rgba(0,0,0,0.1); }
        h1 { text-align: center; color: #333; }
        input[type="text"] { width: 100%; padding: 12px; font-size: 16px; border: 1px solid #ddd; border-radius: 5px; margin: 15px 0; }
        button { padding: 12px 24px; background: #28a745; color: white; border: none; border-radius: 5px; font-size: 16px; cursor: pointer; }
        button:hover { background: #218838; }
    </style>
</head>
<body>
<div class="container">
    <h1>输入提取码下载</h1>
    <form action="download.php" method="post">
        <input type="text" name="code" placeholder="请输入您获得的提取码" required>
        <button type="submit">下载文件</button>
    </form>
    <p style="text-align: center; color: #666; margin-top: 20px;">
        提示：提取码不区分大小写
    </p>
</div>
</body>
</html>
```

---

## 🧪 `install.php`（首次运行安装）

```php
<?php
require 'config.php';
require 'admin/includes/db.php';

$sql = file_get_contents('schema.sql'); // 保存上面的 SQL 到 schema.sql
$statements = array_filter(array_map('trim', explode(';', $sql)));

try {
    foreach ($statements as $stmt) {
        if (!empty($stmt)) {
            $pdo->exec($stmt);
        }
    }
    echo "<h2 style='color:green'>✅ 安装成功！</h2>";
    echo "<p>请删除 install.php 文件以确保安全。</p>";
    echo "<a href='admin/login.php'>进入管理后台</a>";
} catch (Exception $e) {
    echo "<h2 style='color:red'>❌ 安装失败：</h2><pre>" . $e->getMessage() . "</pre>";
}
```

---

## ✅ 使用步骤

1. 上传所有文件到服务器
2. 创建数据库 `virtual_store` 并导入 SQL
3. 修改 `config.php` 中的数据库配置
4. 运行 `install.php`（可选，也可直接导入 SQL）
5. 访问 `/admin/login.php`，用 `admin` / `123456` 登录
6. 在后台添加产品（需先手动上传文件到 `products/` 目录）
7. 生成提取码并分发给用户
8. 用户访问首页输入提取码下载

---

## 🔒 安全建议

- 删除 `install.php` 和 `schema.sql` 文件
- 修改管理员密码（SQL 更新 `UPDATE admins SET password = '$2y$10$...' WHERE username='admin';`）
- 启用 HTTPS
- 定期备份数据库

---

## 💾 获取完整项目包

如果你希望我打包成 ZIP 文件（含所有文件 + SQL），可以告诉我邮箱或使用方式，我可以帮你生成。

或者你也可以直接复制以上代码结构部署。

需要 **添加产品上传功能** 或 **邮件自动发送提取码**，也可以继续告诉我！

祝你项目成功！🚀

--------------------------------------------------------------------------------

# 🤵 添加产品上传功能

太好了！我们来为系统添加 **产品上传功能**，让管理员可以在后台直接上传电子书、音频、视频等文件，
并自动创建产品记录。

---

## ✅ 新增功能

- 🖼️ 后台可上传文件到 `products/` 目录
- 📁 自动生成唯一文件名，防止覆盖
- 🛡️ 文件类型和大小限制
- ✅ 自动插入产品到数据库
- 📝 填写产品标题、下载次数限制、过期天数

---

### ✅ 修改步骤

我们将修改 `admin/dashboard.php`，添加「上传产品」表单和处理逻辑。

---

## ✏️ 更新 `admin/dashboard.php`（在顶部插入上传逻辑）

```php
<?php
require_once 'includes/auth.php';
require_once 'includes/db.php';

if (!isAdminLoggedIn()) {
    header('Location: login.php');
    exit;
}

$success = $error = '';

// ========== 新增：产品上传处理 ==========
if (isset($_POST['upload_product'])) {
    $title = trim($_POST['title']);
    $download_limit = (int)$_POST['download_limit'] ?: 5;
    $expire_days = (int)$_POST['expire_days'] ?: 7;

    if (empty($title)) {
        $error = "请输入产品标题";
    } elseif (!isset($_FILES['product_file']) || $_FILES['product_file']['error'] !== 0) {
        $error = "文件上传失败，请检查";
    } else {
        $file = $_FILES['product_file'];
        $allowed_types = [
            'pdf'  => 'application/pdf',
            'epub' => 'application/epub+zip',
            'mobi' => 'application/x-mobipocket-ebook',
            'mp3'  => 'audio/mpeg',
            'wav'  => 'audio/wav',
            'mp4'  => 'video/mp4',
            'zip'  => 'application/zip',
            'jpg'  => 'image/jpeg',
            'png'  => 'image/png'
        ];

        $finfo = finfo_open(FILEINFO_MIME_TYPE);
        $mime = finfo_file($finfo, $file['tmp_name']);
        finfo_close($finfo);

        $ext = strtolower(pathinfo($file['name'], PATHINFO_EXTENSION));

        if (!isset($allowed_types[$ext]) || $allowed_types[$ext] !== $mime) {
            $error = "不支持的文件类型：{$ext} ({$mime})";
        } elseif ($file['size'] > 500 * 1024 * 1024) { // 500MB 限制
            $error = "文件大小不能超过 500MB";
        } else {
            // 生成唯一文件名
            $new_filename = bin2hex(random_bytes(16)) . '.' . $ext;
            $destination = PRODUCTS_DIR . $new_filename;

            if (move_uploaded_file($file['tmp_name'], $destination)) {
                try {
                    $stmt = $pdo->prepare("
                        INSERT INTO products (title, filename, file_path, download_limit, expire_days)
                        VALUES (?, ?, ?, ?, ?)
                    ");
                    $stmt->execute([
                        $title,
                        $file['name'],           // 原始文件名（下载时显示）
                        $destination,            // 服务器存储路径
                        $download_limit,
                        $expire_days
                    ]);
                    $success = "✅ 产品上传成功！";
                } catch (Exception $e) {
                    unlink($destination); // 删除已上传文件
                    $error = "数据库保存失败: " . $e->getMessage();
                }
            } else {
                $error = "文件移动失败，请检查目录权限";
            }
        }
    }
}
// ========== 上传逻辑结束 ==========

// ========== 原有：生成提取码逻辑（保持不变）==========
if (isset($_POST['generate'])) {
    $product_id = (int)$_POST['product_id'];
    $count = (int)$_POST['count'] ?: 1;

    try {
        $pdo->beginTransaction();
        for ($i = 0; $i < $count; $i++) {
            $code = strtoupper(substr(bin2hex(random_bytes(4)), 0, 8));
            $stmt = $pdo->prepare("SELECT id FROM download_codes WHERE code = ?");
            $stmt->execute([$code]);
            if ($stmt->rowCount() > 0) continue;

            $stmt = $pdo->prepare("
                INSERT INTO download_codes (code, product_id, max_downloads, expires_at)
                VALUES (?, ?, ?, DATE_ADD(NOW(), INTERVAL ? DAY))
            ");
            $stmt->execute([$code, $product_id, 5, 7]);
        }
        $pdo->commit();
        $success = "成功生成 {$count} 个提取码！";
    } catch (Exception $e) {
        $pdo->rollback();
        $error = "生成失败: " . $e->getMessage();
    }
}
// ========== 生成逻辑结束 ==========

// 获取产品列表
$products = $pdo->query("SELECT id, title, filename FROM products ORDER BY title")->fetchAll();

// 获取最近提取码
$codes = $pdo->query("
    SELECT dc.*, p.title as product_title 
    FROM download_codes dc
    JOIN products p ON p.id = dc.product_id
    ORDER BY dc.created_at DESC
    LIMIT 50
")->fetchAll();

// 获取下载日志
$logs = $pdo->query("
    SELECT dl.*, p.title as product_title 
    FROM download_logs dl
    JOIN products p ON p.id = dl.product_id
    ORDER BY dl.downloaded_at DESC
    LIMIT 30
")->fetchAll();
?>
```

---

## ✅ 在 `dashboard.php` 中添加上传表单（放在生成提取码上方）

```php
<h2>上传新产品</h2>
<form method="post" enctype="multipart/form-data">
    <input type="text" name="title" placeholder="产品标题（如：Python入门电子书）" required style="width:100%; padding:10px; margin:10px 0;">
    
    <input type="file" name="product_file" required style="width:100%; padding:10px; margin:10px 0;">
    
    <select name="download_limit" style="padding:10px; margin:10px 0;">
        <option value="3">允许下载 3 次</option>
        <option value="5" selected>允许下载 5 次</option>
        <option value="10">允许下载 10 次</option>
        <option value="999">不限次数</option>
    </select>
    
    <select name="expire_days" style="padding:10px; margin:10px 0;">
        <option value="7">7天后过期</option>
        <option value="14">14天后过期</option>
        <option value="30">30天后过期</option>
        <option value="365">1年有效期</option>
    </select>
    
    <button type="submit" name="upload_product" class="btn">上传产品</button>
</form>
```

> 💡 提示：将这个表单放在页面最上方，方便使用。

---

## 🛠️ 确保 `products/` 目录可写

在终端执行（Linux/Mac）：

```bash
chmod 755 products/
```

或在主机控制面板设置目录权限为 `755` 或 `775`，确保 PHP 可写。

---

## ✅ 上传流程说明

1. 管理员登录后台
2. 填写产品标题
3. 选择文件（PDF/MP3/MP4等）
4. 设置下载次数和有效期
5. 点击“上传产品”
6. 系统保存文件为随机名（如 `a1b2c3d4...pdf`），并记录原始文件名用于下载
7. 上传成功后，可在“生成提取码”中选择该产品生成码

---

## 🔐 安全增强

- ✅ 文件类型白名单验证（MIME + 扩展名）
- ✅ 文件重命名，防止路径穿越
- ✅ 限制文件大小（500MB）
- ✅ 上传目录禁止执行 PHP（加 `.htaccess`）

### `products/.htaccess`（防止执行脚本）

```apache
<Files "*.php">
    Order Allow,Deny
    Deny from all
</Files>
```

---

## 🎯 效果示例

| 字段 | 值 |
|------|----|
| 用户上传文件 | `mybook.pdf` |
| 服务器存储为 | `3a7f...c2e9.pdf` |
| 数据库记录 | `filename="mybook.pdf"`（下载时显示这个名字） |
| 用户下载时 | 保存为 `mybook.pdf`，即使服务器上是乱码名 |

---

## ✅ 总结：你现在可以

- 📤 在后台直接上传电子书、音频、视频
- 🔐 安全存储，防止恶意文件执行
- 🧩 自动生成提取码分发
- 📊 查看下载记录

---

## 📦 下一步建议

如果你希望，我可以为你：

1. 打包完整的 ZIP 项目（含所有文件）
2. 添加「批量上传」功能
3. 添加「产品管理」页面（编辑/删除产品）
4. 添加「邮件自动发送提取码」功能（集成 SMTP）

只需告诉我你的需求！

祝你数字产品销售顺利！🚀📚🎧🎥

--------------------------------------------------------------------------------
