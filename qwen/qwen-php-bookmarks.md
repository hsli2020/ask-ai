# 🤵 我想用php做一个简单的web应用，管理我的bookmarks

Qwen3-235B-A22B-2507

太好了！用 PHP 做一个简单的书签管理 Web 应用是个非常实用且适合学习的项目。下面我会一步步帮你搭建一个基础但功能完整的书签管理应用。

---

## 🌟 功能需求（基础版）

- 添加书签（标题、URL、可选标签）
- 查看书签列表
- 编辑和删除书签
- 按标签或关键词搜索书签

---

## 🛠️ 技术栈

- PHP（后端）
- SQLite（轻量数据库，无需安装 MySQL）
- HTML/CSS/Bootstrap（前端界面）
- 可选：jQuery（简化操作）

---

## 📁 项目结构

```
/bookmark-app/
├── index.php           # 主页面：显示书签
├── add.php             # 添加书签
├── edit.php            # 编辑书签
├── delete.php          # 删除书签
├── db.php              # 数据库连接
└── style.css           # 简单样式
```

---

## 1. 创建数据库（SQLite）

创建文件：`db.php`

```php
<?php
// db.php - 数据库连接与初始化

$dbFile = 'bookmarks.db';
$db = new PDO("sqlite:$dbFile");
$db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

// 创建表（如果不存在）
$db->exec("CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)");
?>
```

> 运行时会自动创建 `bookmarks.db` 文件。

---

## 2. 添加书签：`add.php`

```php
<?php
require 'db.php';

if ($_POST) {
    $title = trim($_POST['title']);
    $url = trim($_POST['url']);
    $tags = trim($_POST['tags']);

    if ($title && $url) {
        $stmt = $db->prepare("INSERT INTO bookmarks (title, url, tags) VALUES (?, ?, ?)");
        $stmt->execute([$title, $url, $tags]);

        header('Location: index.php');
        exit;
    } else {
        $error = "标题和URL是必填的！";
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>添加书签</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="style.css">
</head>
<body>
<div class="container mt-5">
    <h2>添加书签</h2>
    <?php if (isset($error)): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>

    <form method="post">
        <div class="mb-3">
            <label>标题</label>
            <input type="text" name="title" class="form-control" required>
        </div>
        <div class="mb-3">
            <label>URL</label>
            <input type="url" name="url" class="form-control" required>
        </div>
        <div class="mb-3">
            <label>标签（可选，用逗号分隔）</label>
            <input type="text" name="tags" class="form-control">
        </div>
        <button type="submit" class="btn btn-primary">保存</button>
        <a href="index.php" class="btn btn-secondary">返回</a>
    </form>
</div>
</body>
</html>
```

---

## 3. 主页面：`index.php`（显示所有书签）

```php
<?php
require 'db.php';

// 搜索功能
$search = $_GET['search'] ?? '';
$tags = $_GET['tags'] ?? '';

$query = "SELECT * FROM bookmarks WHERE 1=1";
$params = [];

if ($search) {
    $query .= " AND (title LIKE ? OR tags LIKE ?)";
    $params[] = "%$search%";
    $params[] = "%$search%";
}

if ($tags) {
    $query .= " AND tags LIKE ?";
    $params[] = "%$tags%";
}

$query .= " ORDER BY created_at DESC";
$stmt = $db->prepare($query);
$stmt->execute($params);
$bookmarks = $stmt->fetchAll();
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>我的书签</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
    <link rel="stylesheet" href="style.css">
</head>
<body>
<div class="container mt-5">
    <h1>🔖 我的书签</h1>

    <!-- 搜索 -->
    <form method="get" class="mb-4">
        <div class="row">
            <div class="col-md-6">
                <input type="text" name="search" class="form-control" placeholder="搜索标题或标签..." value="<?= htmlspecialchars($search) ?>">
            </div>
            <div class="col-md-4">
                <input type="text" name="tags" class="form-control" placeholder="按标签过滤..." value="<?= htmlspecialchars($tags) ?>">
            </div>
            <div class="col">
                <button type="submit" class="btn btn-primary">搜索</button>
                <a href="index.php" class="btn btn-secondary">重置</a>
            </div>
        </div>
    </form>

    <!-- 添加按钮 -->
    <p><a href="add.php" class="btn btn-success mb-3">➕ 添加书签</a></p>

    <!-- 书签列表 -->
    <div class="list-group">
        <?php if (count($bookmarks) == 0): ?>
            <p class="text-muted">暂无书签。</p>
        <?php else: ?>
            <?php foreach ($bookmarks as $b): ?>
                <div class="list-group-item">
                    <h5><a href="<?= htmlspecialchars($b['url']) ?>" target="_blank"><?= htmlspecialchars($b['title']) ?></a></h5>
                    <small class="text-muted"><?= htmlspecialchars($b['url']) ?></small>
                    <?php if ($b['tags']): ?>
                        <div class="badge bg-secondary mt-1"><?= htmlspecialchars($b['tags']) ?></div>
                    <?php endif; ?>
                    <div class="mt-2">
                        <a href="edit.php?id=<?= $b['id'] ?>" class="btn btn-sm btn-outline-primary">编辑</a>
                        <a href="delete.php?id=<?= $b['id'] ?>" class="btn btn-sm btn-outline-danger" onclick="return confirm('确定删除？')">删除</a>
                    </div>
                </div>
            <?php endforeach; ?>
        <?php endif; ?>
    </div>
</div>
</body>
</html>
```

---

## 4. 编辑书签：`edit.php`

```php
<?php
require 'db.php';

$id = $_GET['id'] ?? null;
if (!$id || !is_numeric($id)) {
    header('Location: index.php');
    exit;
}

// 获取书签
$stmt = $db->prepare("SELECT * FROM bookmarks WHERE id = ?");
$stmt->execute([$id]);
$bookmark = $stmt->fetch();

if (!$bookmark) {
    die("书签不存在！");
}

if ($_POST) {
    $title = trim($_POST['title']);
    $url = trim($_POST['url']);
    $tags = trim($_POST['tags']);

    if ($title && $url) {
        $stmt = $db->prepare("UPDATE bookmarks SET title = ?, url = ?, tags = ? WHERE id = ?");
        $stmt->execute([$title, $url, $tags, $id]);

        header('Location: index.php');
        exit;
    } else {
        $error = "标题和URL是必填的！";
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>编辑书签</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5">
    <h2>编辑书签</h2>
    <?php if (isset($error)): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>

    <form method="post">
        <div class="mb-3">
            <label>标题</label>
            <input type="text" name="title" class="form-control" value="<?= htmlspecialchars($bookmark['title']) ?>" required>
        </div>
        <div class="mb-3">
            <label>URL</label>
            <input type="url" name="url" class="form-control" value="<?= htmlspecialchars($bookmark['url']) ?>" required>
        </div>
        <div class="mb-3">
            <label>标签（可选）</label>
            <input type="text" name="tags" class="form-control" value="<?= htmlspecialchars($bookmark['tags']) ?>">
        </div>
        <button type="submit" class="btn btn-primary">更新</button>
        <a href="index.php" class="btn btn-secondary">取消</a>
    </form>
</div>
</body>
</html>
```

---

## 5. 删除书签：`delete.php`

```php
<?php
require 'db.php';

$id = $_GET['id'] ?? null;
if ($id && is_numeric($id)) {
    $stmt = $db->prepare("DELETE FROM bookmarks WHERE id = ?");
    $stmt->execute([$id]);
}

header('Location: index.php');
exit;
?>
```

---

## 6. 可选：添加简单样式 `style.css`

```css
body {
    background-color: #f8f9fa;
}
.list-group-item {
    border-left: 4px solid #007bff;
}
```

---

## 🚀 如何运行？

1. 安装 PHP 和 Web 服务器（推荐使用 [XAMPP](https://www.apachefriends.org/) 或 [PHP 内置服务器](https://www.php.net/manual/en/features.commandline.webserver.php)）
2. 将所有文件放入 `htdocs/bookmark-app`（XAMPP）或任意目录
3. 启动服务器
4. 访问：`http://localhost/bookmark-app/index.php`

---

## ✅ 后续可扩展功能

- 用户登录系统（用 session）
- 标签自动提取与分类
- 导出书签为 HTML 或 JSON
- 自动获取网站标题（用 `get_meta_tags()`）
- 使用 AJAX 实现无刷新操作

---

需要我帮你打包成 ZIP 或加上自动获取网站标题的功能吗？欢迎继续提问 😊

--------------------------------------------------------------------------------

# 🤵 加上用户登录功能

太好了！我们来为书签管理应用加上 **用户登录功能**，让每个用户拥有自己的书签数据。

我们将使用 PHP 的 `$_SESSION` 来管理登录状态，并添加一个简单的用户表。

---

## ✅ 新增功能

- 用户注册
- 用户登录 / 登出
- 每个用户只能看到和管理自己的书签
- 登录保护（未登录不能访问书签）

---

## 🛠️ 修改项目结构

```
/bookmark-app/
├── index.php        # 书签列表（需登录）
├── add.php          # 添加书签（需登录）
├── edit.php         # 编辑书签（需登录）
├── delete.php       # 删除书签（需登录）
├── login.php        # 登录页面（新增）
├── register.php     # 注册页面（新增）
├── logout.php       # 登出（新增）
├── db.php           # 数据库连接（更新）
└── style.css
```

---

## 1. 更新 `db.php`：添加用户表

```php
<?php
// db.php

session_start(); // 启动 Session

$dbFile = 'bookmarks.db';
$db = new PDO("sqlite:$dbFile");
$db->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

// 创建 bookmarks 表
$db->exec("CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
)");

// 创建 users 表
$db->exec("CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)");

// 初始管理员用户（可选，只运行一次）
/*
$stmt = $db->prepare("SELECT COUNT(*) FROM users");
$count = $stmt->fetchColumn();
if ($count == 0) {
    $stmt = $db->prepare("INSERT INTO users (username, password) VALUES (?, ?)");
    $stmt->execute(['admin', password_hash('123456', PASSWORD_DEFAULT)]);
    echo "初始用户创建成功：admin / 123456<br>";
}
*/
?>
```

> ⚠️ 首次运行后，建议注释掉创建初始用户的代码。

---

## 2. 创建 `register.php`：用户注册

```php
<?php
require 'db.php';

if ($_SESSION['user_id']) {
    header('Location: index.php');
    exit;
}

$error = $success = '';

if ($_POST) {
    $username = trim($_POST['username']);
    $password = $_POST['password'];
    $confirm_password = $_POST['confirm_password'];

    if (empty($username) || empty($password)) {
        $error = "用户名和密码不能为空。";
    } elseif ($password !== $confirm_password) {
        $error = "两次密码不一致。";
    } elseif (strlen($password) < 6) {
        $error = "密码至少6位。";
    } else {
        // 检查用户名是否已存在
        $stmt = $db->prepare("SELECT id FROM users WHERE username = ?");
        $stmt->execute([$username]);
        if ($stmt->fetch()) {
            $error = "用户名已存在。";
        } else {
            $hashed = password_hash($password, PASSWORD_DEFAULT);
            $stmt = $db->prepare("INSERT INTO users (username, password) VALUES (?, ?)");
            $stmt->execute([$username, $hashed]);
            $success = "注册成功！<a href='login.php'>去登录</a>";
        }
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>注册账号</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5" style="max-width: 400px;">
    <h2>📝 注册</h2>
    <?php if ($error): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>
    <?php if ($success): ?>
        <div class="alert alert-success"><?= $success ?></div>
    <?php else: ?>
        <form method="post">
            <div class="mb-3">
                <label>用户名</label>
                <input type="text" name="username" class="form-control" required>
            </div>
            <div class="mb-3">
                <label>密码（至少6位）</label>
                <input type="password" name="password" class="form-control" required>
            </div>
            <div class="mb-3">
                <label>确认密码</label>
                <input type="password" name="confirm_password" class="form-control" required>
            </div>
            <button type="submit" class="btn btn-primary">注册</button>
            <a href="login.php" class="btn btn-secondary">已有账号？去登录</a>
        </form>
    <?php endif; ?>
</div>
</body>
</html>
```

---

## 3. 创建 `login.php`：用户登录

```php
<?php
require 'db.php';

if ($_SESSION['user_id']) {
    header('Location: index.php');
    exit;
}

$error = '';

if ($_POST) {
    $username = trim($_POST['username']);
    $password = $_POST['password'];

    $stmt = $db->prepare("SELECT id, password FROM users WHERE username = ?");
    $stmt->execute([$username]);
    $user = $stmt->fetch();

    if ($user && password_verify($password, $user['password'])) {
        $_SESSION['user_id'] = $user['id'];
        $_SESSION['username'] = $username;
        header('Location: index.php');
        exit;
    } else {
        $error = "用户名或密码错误。";
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>登录</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5" style="max-width: 400px;">
    <h2>🔐 登录</h2>
    <?php if ($error): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>
    <form method="post">
        <div class="mb-3">
            <label>用户名</label>
            <input type="text" name="username" class="form-control" required>
        </div>
        <div class="mb-3">
            <label>密码</label>
            <input type="password" name="password" class="form-control" required>
        </div>
        <button type="submit" class="btn btn-primary">登录</button>
        <a href="register.php" class="btn btn-secondary">注册账号</a>
    </form>
</div>
</body>
</html>
```

---

## 4. 创建 `logout.php`：登出

```php
<?php
session_start();
session_destroy();
header('Location: login.php');
exit;
?>
```

---

## 5. 在所有需要登录的页面顶部添加权限检查

在 `index.php`, `add.php`, `edit.php`, `delete.php` 的最顶部（`require 'db.php';` 之后）添加：

```php
// 检查是否登录
if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}
```

例如 `index.php` 开头变成：

```php
<?php
require 'db.php';

// 检查是否登录
if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}

// 后续代码...
```

---

## 6. 修改数据库操作：绑定 user_id

### 在 `add.php` 中插入时加上 `user_id`

```php
$stmt = $db->prepare("INSERT INTO bookmarks (user_id, title, url, tags) VALUES (?, ?, ?, ?)");
$stmt->execute([$_SESSION['user_id'], $title, $url, $tags]);
```

### 在 `index.php` 中只查当前用户的书签

```php
$query = "SELECT * FROM bookmarks WHERE user_id = ?";
$params = [$_SESSION['user_id']];

if ($search) {
    $query .= " AND (title LIKE ? OR tags LIKE ?)";
    $params[] = "%$search%";
    $params[] = "%$search%";
}
if ($tags) {
    $query .= " AND tags LIKE ?";
    $params[] = "%$tags%";
}
$query .= " ORDER BY created_at DESC";

$stmt = $db->prepare($query);
$stmt->execute($params);
$bookmarks = $stmt->fetchAll();
```

### 在 `edit.php` 和 `delete.php` 中检查书签归属

在获取书签前加：

```php
$stmt = $db->prepare("SELECT * FROM bookmarks WHERE id = ? AND user_id = ?");
$stmt->execute([$id, $_SESSION['user_id']]);
$bookmark = $stmt->fetch();

if (!$bookmark) {
    die("书签不存在或无权访问！");
}
```

更新操作也要加 `user_id` 条件：

```php
$stmt = $db->prepare("UPDATE bookmarks SET title = ?, url = ?, tags = ? WHERE id = ? AND user_id = ?");
$stmt->execute([$title, $url, $tags, $id, $_SESSION['user_id']]);
```

---

## 7. 在 `index.php` 加上用户信息和登出按钮

在页面顶部加：

```php
<p class="text-end">
    欢迎，<?= htmlspecialchars($_SESSION['username']) ?> |
    <a href="logout.php" class="btn btn-sm btn-outline-danger">登出</a>
</p>
```

---

## ✅ 测试流程

1. 访问 `register.php` → 注册一个账号
2. 访问 `login.php` → 登录
3. 进入 `index.php` → 添加、编辑、删除书签
4. 登出测试权限控制

---

## 🔐 安全提醒

- 使用了 `password_hash()` 和 `password_verify()` 保证密码安全
- 所有输出使用 `htmlspecialchars()` 防止 XSS
- 所有用户操作都检查了 `user_id` 权限，防止越权
- 使用了 PDO 预处理防止 SQL 注入

---

需要我帮你：
- 添加“忘记密码”功能？
- 支持多用户共享书签？
- 响应式设计优化？
- 打包成可下载的 ZIP？

欢迎继续告诉我！ 😊

--------------------------------------------------------------------------------

# 🤵 添加“忘记密码”功能

太好了！我们来为你的书签管理应用添加 **“忘记密码”** 功能。

由于这是一个轻量级 PHP + SQLite 应用，我们不会使用邮件发送（需要 SMTP 配置），
而是采用 **安全令牌 + 重置链接** 的方式，模拟完整流程。你也可以后续集成邮件服务（如 PHPMailer）。

---

## ✅ 功能设计

1. 用户输入用户名 → 系统生成一个临时重置令牌
2. 生成重置链接（如：`reset.php?token=abc123`）
3. 用户点击链接 → 设置新密码
4. 令牌使用后失效，防止重复使用

---

## 🛠️ 新增文件

- `forgot.php` —— 输入用户名，请求重置
- `reset.php`   —— 使用令牌，设置新密码

---

## 1. 修改数据库：添加 `reset_token` 字段

更新 `db.php` 中的 `users` 表结构（加在创建表语句中）：

```php
$db->exec("CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    reset_token TEXT,
    reset_expires DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)");
```

> ✅ 字段说明：
> - `reset_token`: 临时令牌
> - `reset_expires`: 令牌有效期（比如 1 小时）

---

## 2. 创建 `forgot.php`：请求重置密码

```php
<?php
require 'db.php';

if ($_SESSION['user_id']) {
    header('Location: index.php');
    exit;
}

$message = $error = '';

if ($_POST) {
    $username = trim($_POST['username']);

    $stmt = $db->prepare("SELECT id FROM users WHERE username = ?");
    $stmt->execute([$username]);
    $user = $stmt->fetch();

    if (!$user) {
        $error = "该用户名不存在。";
    } else {
        // 生成随机令牌
        $token = bin2hex(random_bytes(32));
        $expires = date('Y-m-d H:i:s', time() + 3600); // 1小时后过期

        // 存入数据库
        $stmt = $db->prepare("UPDATE users SET reset_token = ?, reset_expires = ? WHERE id = ?");
        $stmt->execute([$token, $expires, $user['id']]);

        // 模拟“发送邮件”——实际项目中这里调用 PHPMailer
        $resetLink = "http://$_SERVER[HTTP_HOST]" . dirname($_SERVER['SCRIPT_NAME']);
        $resetLink = rtrim($resetLink, '/') . '/reset.php?token=' . $token;

        // 显示链接（生产环境应发送邮件）
        $message = "
        <div class='alert alert-success'>
            密码重置请求成功！<br><br>
            请在 1 小时内点击以下链接重置密码：<br>
            <a href='$resetLink' class='btn btn-sm btn-primary mt-2'>重置密码</a><br><br>
            <small>（实际应用中此链接会通过邮件发送）</small>
        </div>";
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>忘记密码</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5" style="max-width: 400px;">
    <h2>🔑 忘记密码</h2>
    <p>请输入你的用户名，我们将发送重置链接。</p>

    <?php if ($error): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>

    <?php if ($message): ?>
        <?= $message ?>
    <?php else: ?>
        <form method="post">
            <div class="mb-3">
                <label>用户名</label>
                <input type="text" name="username" class="form-control" required>
            </div>
            <button type="submit" class="btn btn-primary">发送重置链接</button>
            <a href="login.php" class="btn btn-secondary">返回登录</a>
        </form>
    <?php endif; ?>
</div>
</body>
</html>
```

---

## 3. 创建 `reset.php`：设置新密码

```php
<?php
require 'db.php';

$message = $error = '';
$token = $_GET['token'] ?? '';

if (!$token) {
    $error = "无效的重置链接。";
} else {
    // 查找有效令牌
    $stmt = $db->prepare("SELECT id FROM users WHERE reset_token = ? AND reset_expires > datetime('now')");
    $stmt->execute([$token]);
    $user = $stmt->fetch();

    if (!$user) {
        $error = "令牌无效或已过期。";
    }

    if ($_POST) {
        $password = $_POST['password'];
        $confirm = $_POST['confirm'];

        if ($password !== $confirm) {
            $error = "两次密码不一致。";
        } elseif (strlen($password) < 6) {
            $error = "密码至少6位。";
        } else {
            $hashed = password_hash($password, PASSWORD_DEFAULT);

            // 更新密码并清除令牌
            $stmt = $db->prepare("UPDATE users SET password = ?, reset_token = NULL, reset_expires = NULL WHERE reset_token = ?");
            $stmt->execute([$hashed, $token]);

            $message = "
            <div class='alert alert-success'>
                密码已重置！<br>
                <a href='login.php' class='btn btn-primary mt-2'>去登录</a>
            </div>";
        }
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>重置密码</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5" style="max-width: 400px;">
    <h2>🔄 重置密码</h2>

    <?php if ($error): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>

    <?php if ($message): ?>
        <?= $message ?>
    <?php elseif ($token && !$error): ?>
        <form method="post">
            <div class="mb-3">
                <label>新密码（至少6位）</label>
                <input type="password" name="password" class="form-control" required>
            </div>
            <div class="mb-3">
                <label>确认新密码</label>
                <input type="password" name="confirm" class="form-control" required>
            </div>
            <button type="submit" class="btn btn-primary">重置密码</button>
            <a href="login.php" class="btn btn-secondary">返回登录</a>
        </form>
    <?php else: ?>
        <p class="text-muted">无效请求。</p>
        <a href="login.php" class="btn btn-secondary">返回登录</a>
    <?php endif; ?>
</div>
</body>
</html>
```

---

## ✅ 使用流程

1. 访问 `forgot.php`
2. 输入用户名 → 点击“发送重置链接”
3. 页面显示一个重置链接（模拟邮件）
4. 点击链接进入 `reset.php?token=...`
5. 输入新密码 → 重置成功
6. 令牌自动失效，无法再次使用

---

## 🔐 安全特性

- 令牌随机生成（`random_bytes`）
- 令牌有效期 1 小时
- 令牌使用后立即清除
- 防止暴力破解（无失败次数限制，但可后续加）
- 所有操作绑定数据库，防止越权

---

## 🚀 后续增强建议

| 功能 | 说明 |
|------|------|
| ✉️ 集成 PHPMailer | 把重置链接通过邮件发送 |
| 📱 记录日志 | 记录密码重置行为 |
| ⏳ 限制请求频率 | 防止滥用 `forgot.php` |
| 🔒 HTTPS | 生产环境必须启用 |

---

需要我帮你：
- 集成 PHPMailer 发邮件？
- 添加“记住我”功能？
- 增加密码强度检测？
- 打包整个项目为 ZIP 下载？

欢迎继续告诉我，我可以一步步帮你完善！ 😊

--------------------------------------------------------------------------------

# 🤵 集成 PHPMailer 发邮件
添加“记住我”功能
增加密码强度检测

太棒了！我们来为你的书签管理应用集成三大增强功能：

1. ✉️ **PHPMailer**：真正发送邮件（替代模拟链接）
2. 🍪 **“记住我”功能**：用户关闭浏览器后仍保持登录
3. 🔐 **密码强度检测**：前端 + 后端双重校验

---

## 📦 第一步：安装 PHPMailer

### 方法一：使用 Composer（推荐）

在项目根目录打开终端：

```bash
composer require phpmailer/phpmailer
```

> 如果没有安装 Composer，请先下载：https://getcomposer.org/

### 方法二：手动下载

将 `PHPMailer` 文件夹放入项目根目录，结构如下：

```
/bookmark-app/
├── PHPMailer/
│   ├── src/
│   │   ├── PHPMailer.php
│   │   ├── SMTP.php
│   │   └── ...
```

---

## ✉️ 1. 集成 PHPMailer 发送重置邮件

### 创建 `mail.php` 配置文件

```php
<?php
// mail.php - 邮件配置

use PHPMailer\PHPMailer\PHPMailer;
use PHPMailer\PHPMailer\SMTP;

require 'PHPMailer/src/PHPMailer.php';
require 'PHPMailer/src/SMTP.php';

function sendPasswordResetEmail($email, $username, $resetLink) {
    $mail = new PHPMailer(true);

    try {
        // 服务器设置
        $mail->isSMTP();
        $mail->Host       = 'smtp.gmail.com';           // Gmail SMTP
        $mail->SMTPAuth   = true;
        $mail->Username   = 'your-email@gmail.com';     // 你的邮箱
        $mail->Password   = 'your-app-password';        // Gmail App 密码（不是登录密码！）
        $mail->SMTPSecure = PHPMailer::ENCRYPTION_TLS;
        $mail->Port       = 587;

        // 收件人
        $mail->setFrom('no-reply@bookmarks.local', '书签系统');
        $mail->addAddress($email);

        // 内容
        $mail->isHTML(true);
        $mail->Subject = '【书签系统】密码重置请求';
        $mail->Body    = "
        <h3>你好，{$username}！</h3>
        <p>你请求了密码重置。请在 1 小时内点击下面的链接：</p>
        <p><a href='{$resetLink}' target='_blank' style='color:#007bff;'>👉 重置密码</a></p>
        <p>如果这不是你本人操作，请忽略此邮件。</p>
        <hr>
        <small>此邮件由系统自动发送，请勿回复。</small>
        ";

        $mail->send();
        return true;
    } catch (Exception $e) {
        error_log("邮件发送失败: {$mail->ErrorInfo}");
        return false;
    }
}
?>
```

> 🔐 **Gmail App 密码设置**：
> 1. 开启 Gmail 两步验证
> 2. 生成“应用专用密码”（16位）
> 3. 填入 `your-app-password`

---

### 修改 `forgot.php`：使用真实邮件

在文件顶部加入：

```php
require 'mail.php'; // 引入邮件功能
```

替换“生成链接后”的逻辑（保留前面代码不变）：

```php
if (!$user) {
    $error = "该用户名不存在。";
} else {
    // 生成令牌
    $token = bin2hex(random_bytes(32));
    $expires = date('Y-m-d H:i:s', time() + 3600);

    $stmt = $db->prepare("UPDATE users SET reset_token = ?, reset_expires = ? WHERE id = ?");
    $stmt->execute([$token, $expires, $user['id']]);

    // 构建重置链接
    $resetLink = "http://$_SERVER[HTTP_HOST]" . dirname($_SERVER['SCRIPT_NAME']);
    $resetLink = rtrim($resetLink, '/') . '/reset.php?token=' . $token;

    // 👇 发送邮件
    $email = $username . '@example.com'; // 实际项目中 users 表应有 email 字段
    if (sendPasswordResetEmail($email, $username, $resetLink)) {
        $message = "<div class='alert alert-success'>密码重置链接已发送至 <strong>$email</strong>，请查收邮件。</div>";
    } else {
        $error = "邮件发送失败，请联系管理员。";
    }
}
```

> 💡 建议：在 `users` 表中添加 `email TEXT` 字段，让用户注册时填写真实邮箱。

---

## 🍪 2. 添加“记住我”功能

### 修改 `login.php` 表单

```html
<div class="mb-3 form-check">
    <input type="checkbox" name="remember" class="form-check-input" id="remember">
    <label class="form-check-label" for="remember">记住我（30天）</label>
</div>
```

### 在登录逻辑中处理“记住我”

```php
if ($_POST) {
    $username = trim($_POST['username']);
    $password = $_POST['password'];
    $remember = isset($_POST['remember']);

    $stmt = $db->prepare("SELECT id, password FROM users WHERE username = ?");
    $stmt->execute([$username]);
    $user = $stmt->fetch();

    if ($user && password_verify($password, $user['password'])) {
        $_SESSION['user_id'] = $user['id'];
        $_SESSION['username'] = $username;

        // 👇 处理“记住我”
        if ($remember) {
            $token = bin2hex(random_bytes(16));
            // 存入数据库（实际项目应加密）
            $stmt = $db->prepare("UPDATE users SET remember_token = ? WHERE id = ?");
            $stmt->execute([$token, $user['id']]);

            // 设置 Cookie（30天）
            setcookie('remember_token', $token, time() + 30*24*3600, '/', '', false, true);
        }

        header('Location: index.php');
        exit;
    } else {
        $error = "用户名或密码错误。";
    }
}
```

### 修改 `db.php`：在 `users` 表加字段

```php
$db->exec("CREATE TABLE IF NOT EXISTS users (
    ...
    remember_token TEXT,
    ...
)");
```

### 在 `db.php` 开头添加自动登录逻辑

```php
session_start();

// 自动登录（如果 Cookie 存在）
if (!isset($_SESSION['user_id']) && !isset($_GET['logout'])) {
    if (isset($_COOKIE['remember_token'])) {
        $token = $_COOKIE['remember_token'];
        $stmt = $db->prepare("SELECT id, username FROM users WHERE remember_token = ?");
        $stmt->execute([$token]);
        $user = $stmt->fetch();

        if ($user) {
            $_SESSION['user_id'] = $user['id'];
            $_SESSION['username'] = $user['username'];
        }
    }
}
```

> ✅ 用户登出时清除 Cookie：

在 `logout.php` 中：

```php
<?php
session_start();
session_destroy();

// 清除 remember me cookie
if (isset($_COOKIE['remember_token'])) {
    setcookie('remember_token', '', time() - 3600, '/');
}

header('Location: login.php');
exit;
?>
```

---

## 🔐 3. 增加密码强度检测

### 前端：在 `register.php` 和 `reset.php` 中添加 JS 检测

```html
<script>
function checkPassword() {
    const pwd = document.getElementById('password').value;
    const hint = document.getElementById('pwd-hint');
    let msgs = [];

    if (pwd.length < 8) msgs.push('至少8位');
    if (!/[A-Z]/.test(pwd)) msgs.push('包含大写字母');
    if (!/[a-z]/.test(pwd)) msgs.push('包含小写字母');
    if (!/[0-9]/.test(pwd)) msgs.push('包含数字');
    if (!/[^A-Za-z0-9]/.test(pwd)) msgs.push('包含特殊字符');

    if (msgs.length === 0) {
        hint.innerHTML = '<span style="color:green">✅ 密码强度合格</span>';
    } else {
        hint.innerHTML = '⚠️ ' + msgs.join('，');
    }
}
</script>

<!-- 在密码输入框后添加 -->
<input type="password" name="password" id="password" oninput="checkPassword()" class="form-control" required>
<small id="pwd-hint" class="form-text text-muted">请输入强密码</small>
```

### 后端：在 `register.php` 和 `reset.php` 中添加校验函数

```php
function isStrongPassword($password) {
    $errors = [];
    if (strlen($password) < 8) $errors[] = "至少8位";
    if (!preg_match('/[A-Z]/', $password)) $errors[] = "包含大写字母";
    if (!preg_match('/[a-z]/', $password)) $errors[] = "包含小写字母";
    if (!preg_match('/[0-9]/', $password)) $errors[] = "包含数字";
    if (!preg_match('/[^A-Za-z0-9]/', $password)) $errors[] = "包含特殊字符";
    return $errors;
}
```

使用示例（在注册或重置时）：

```php
$pwdErrors = isStrongPassword($password);
if (!empty($pwdErrors)) {
    $error = "密码不安全：" . implode("、", $pwdErrors);
} else {
    // 继续处理
}
```

---

## ✅ 总结：你现在拥有了

| 功能 | 状态 |
|------|------|
| ✉️ PHPMailer 邮件发送 | ✅ |
| 🍪 “记住我”自动登录 | ✅ |
| 🔐 前后端密码强度检测 | ✅ |
| 🔐 会话安全 + 令牌过期 | ✅ |
| 📦 用户体验优化 | ✅ |

---

## 📦 项目打包建议

你可以将整个项目打包为 ZIP，结构如下：

```
bookmark-app/
├── index.php
├── add.php
├── edit.php
├── delete.php
├── login.php
├── register.php
├── forgot.php
├── reset.php
├── logout.php
├── db.php
├── mail.php
├── style.css
├── PHPMailer/
└── bookmarks.db (可选)
```

---

需要我：
- 提供完整 ZIP 下载？
- 添加“用户资料修改”功能？
- 支持导出书签为 HTML（浏览器兼容格式）？
- 部署到线上环境的建议？

欢迎继续告诉我，我可以帮你一步步上线这个应用！ 😊

--------------------------------------------------------------------------------

# 🤵 导出书签为 HTML


太好了！我们来为你的书签管理应用添加 **“导出书签为 HTML”** 功能。

这个功能会生成一个标准的 HTML 书签文件（兼容 Chrome、Firefox 等浏览器的“导入书签”功能），
让你可以备份或迁移到其他浏览器。

---

## ✅ 功能说明

- 用户点击“导出书签”
- 系统生成标准格式的 HTML 书签文件
- 自动下载，文件名为：`bookmarks_用户名_日期.html`
- 格式兼容主流浏览器导入

---

## 📁 新增文件

- `export.php` —— 生成并下载 HTML 书签文件

---

## 1. 创建 `export.php`

```php
<?php
require 'db.php';

// 检查登录
if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}

$username = $_SESSION['username'];

// 获取当前用户的书签
$stmt = $db->prepare("SELECT title, url, tags, created_at FROM bookmarks WHERE user_id = ? ORDER BY created_at DESC");
$stmt->execute([$_SESSION['user_id']]);
$bookmarks = $stmt->fetchAll();

// 生成 HTML 内容（标准书签格式）
$html = <<<HTML
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<!-- This is an automatically generated file.
     It will be read and overwritten.
     DO NOT EDIT! -->
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>书签</TITLE>
<H1>我的书签</H1>
<DL><p>
HTML;

foreach ($bookmarks as $b) {
    $title = htmlspecialchars($b['title']);
    $url = htmlspecialchars($b['url']);
    $tags = $b['tags'] ? ' #' . htmlspecialchars($b['tags']) : '';
    $addDate = strtotime($b['created_at']);

    $html .= "    <DT><A HREF=\"$url\" ADD_DATE=\"$addDate\" LAST_MODIFIED=\"\">" . $title . "</A>$tags\n";
}

$html .= "</DL><p>\n";

// 设置下载头
$filename = "bookmarks_{$username}_" . date('Y-m-d') . ".html";
header('Content-Type: application/octet-stream');
header('Content-Disposition: attachment; filename="' . $filename . '"');
header('Content-Length: ' . strlen($html));
header('Connection: close');

echo $html;
exit;
?>
```

---

## 2. 在 `index.php` 添加导出按钮

在页面顶部“添加书签”按钮旁边加上：

```php
<a href="export.php" class="btn btn-outline-info btn-sm">
    📥 导出书签
</a>
```

你也可以放在更显眼的位置，比如：

```php
<p class="text-end">
    欢迎，<?= htmlspecialchars($_SESSION['username']) ?> |
    <a href="export.php" class="btn btn-outline-info btn-sm">📥 导出书签</a>
    <a href="logout.php" class="btn btn-sm btn-outline-danger">登出</a>
</p>
```

---

## 🌐 生成的 HTML 示例

```html
<!DOCTYPE NETSCAPE-Bookmark-file-1>
<META HTTP-EQUIV="Content-Type" CONTENT="text/html; charset=UTF-8">
<TITLE>书签</TITLE>
<H1>我的书签</H1>
<DL><p>
    <DT><A HREF="https://google.com" ADD_DATE="1717000000">谷歌</A> #搜索引擎
    <DT><A HREF="https://github.com" ADD_DATE="1717000100">GitHub</A> #开发
</DL><p>
```

> ✅ 这个格式是 **Netscape Bookmark Format**，被 Chrome、Firefox、Edge 等广泛支持。

---

## ✅ 使用方法

1. 登录后进入 `index.php`
2. 点击“导出书签”
3. 浏览器自动下载 `.html` 文件
4. 在浏览器中：
   - Chrome：设置 → 书签 → 导入书签
   - Firefox：菜单 → 书签 → 显示所有书签 → 导入与备份 → 从 HTML 文件导入

---

## 🔐 安全提示

- 只导出当前用户的数据（已通过 `user_id` 限制）
- 输出内容使用 `htmlspecialchars()` 防止 XSS
- 不包含敏感信息（如密码、token）

---

## 🚀 可选增强功能

| 功能 | 说明 |
|------|------|
| 📁 按标签分组导出 | 在 HTML 中用 `<DT><H3>` 创建文件夹 |
| ⏱️ 添加“最后访问时间” | 记录点击次数和时间 |
| 🔄 导入书签功能 | 上传 HTML 文件并解析导入 |
| ☁️ 导出为 JSON | 用于程序化备份 |

---

### 示例：按标签分组导出（进阶）

如果你希望导出时按标签分文件夹，可以这样修改循环部分：

```php
$groups = [];

// 按标签分组
foreach ($bookmarks as $b) {
    $tag = $b['tags'] ? trim(explode(',', $b['tags'])[0]) : '未分类';
    $groups[$tag][] = $b;
}

foreach ($groups as $tag => $items) {
    $html .= "    <DT><H3>$tag</H3>\n    <DL><p>\n";
    foreach ($items as $b) {
        $title = htmlspecialchars($b['title']);
        $url = htmlspecialchars($b['url']);
        $addDate = strtotime($b['created_at']);
        $html .= "        <DT><A HREF=\"$url\" ADD_DATE=\"$addDate\">$title</A>\n";
    }
    $html .= "    </DL><p>\n";
}
```

---

## 📦 完整项目结构更新

```
/bookmark-app/
├── export.php        ← 新增
├── import.php        ← 可选（后续可加）
├── ...
```

---

需要我帮你：
- 添加“导入书签”功能？
- 支持导出为 JSON 或 Markdown？
- 添加“书签点击统计”？
- 生成一个完整的可下载 ZIP 包？

告诉我，我可以打包发你！ 😊

--------------------------------------------------------------------------------

# 🤵 添加“导入书签”功能
支持导出为 JSON 或 Markdown
添加“书签点击统计”


太棒了！我们来为你的书签管理应用一次性升级三大功能：

---

## ✅ 新增功能

1. 📥 **导入书签**（支持 HTML / JSON）
2. 💾 **导出为 JSON 和 Markdown**
3. 👁️ **书签点击统计**

---

## 📁 新增文件

```
/import.php      → 导入书签
/export_json.php → 导出为 JSON
/export_md.php   → 导出为 Markdown
```

---

## 1. 添加“导入书签”功能（`import.php`）

```php
<?php
require 'db.php';

if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}

$message = $error = '';
$successCount = 0;

if ($_POST && isset($_FILES['bookmark_file'])) {
    $file = $_FILES['bookmark_file'];
    
    if ($file['error'] !== UPLOAD_ERR_OK) {
        $error = "文件上传失败。";
    } else {
        $content = file_get_contents($file['tmp_name']);
        $filename = $file['name'];
        
        if (str_ends_with(strtolower($filename), '.html')) {
            // 导入 HTML
            $dom = new DOMDocument();
            libxml_use_internal_errors(true); // 忽略 HTML 警告
            $dom->loadHTML($content);
            libxml_clear_errors();
            
            $links = $dom->getElementsByTagName('a');
            foreach ($links as $link) {
                $title = $link->textContent;
                $url = $link->getAttribute('href');
                if ($url && filter_var($url, FILTER_VALIDATE_URL)) {
                    // 尝试提取标签（如果格式是 "标题 #tag1 #tag2"）
                    $tags = '';
                    if (preg_match('/^(.+?)\s+#([^\s]+)/', $title, $m)) {
                        $title = $m[1];
                        $tags = $m[2]; // 简单取第一个标签
                    }
                    
                    $stmt = $db->prepare("INSERT OR IGNORE INTO bookmarks (user_id, title, url, tags) VALUES (?, ?, ?, ?)");
                    $stmt->execute([$_SESSION['user_id'], $title, $url, $tags]);
                    $successCount++;
                }
            }
            $message = "✅ 成功导入 $successCount 个书签（HTML 格式）。";
            
        } elseif (str_ends_with(strtolower($filename), '.json')) {
            // 导入 JSON
            $data = json_decode($content, true);
            if (json_last_error() === JSON_ERROR_NONE) {
                foreach ($data as $item) {
                    if (!empty($item['url']) && filter_var($item['url'], FILTER_VALIDATE_URL)) {
                        $title = $item['title'] ?? '未命名';
                        $tags = $item['tags'] ?? '';
                        
                        $stmt = $db->prepare("INSERT OR IGNORE INTO bookmarks (user_id, title, url, tags) VALUES (?, ?, ?, ?)");
                        $stmt->execute([$_SESSION['user_id'], $title, $item['url'], $tags]);
                        $successCount++;
                    }
                }
                $message = "✅ 成功导入 $successCount 个书签（JSON 格式）。";
            } else {
                $error = "JSON 格式无效。";
            }
        } else {
            $error = "不支持的文件格式，请上传 .html 或 .json 文件。";
        }
    }
}
?>

<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>导入书签</title>
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css" rel="stylesheet">
</head>
<body>
<div class="container mt-5" style="max-width: 500px;">
    <h2>📤 导入书签</h2>
    <p>支持格式：</p>
    <ul>
        <li><strong>.html</strong>：浏览器导出的书签文件</li>
        <li><strong>.json</strong>：自定义 JSON 格式（<code>[{"title":"xxx","url":"https://...","tags":"dev"}]</code>）</li>
    </ul>

    <?php if ($error): ?>
        <div class="alert alert-danger"><?= $error ?></div>
    <?php endif; ?>
    <?php if ($message): ?>
        <div class="alert alert-success"><?= $message ?></div>
    <?php endif; ?>

    <form method="post" enctype="multipart/form-data">
        <div class="mb-3">
            <label>选择书签文件 (.html 或 .json)</label>
            <input type="file" name="bookmark_file" class="form-control" accept=".html,.json" required>
        </div>
        <button type="submit" class="btn btn-primary">导入</button>
        <a href="index.php" class="btn btn-secondary">返回</a>
    </form>
</div>
</body>
</html>
```

---

## 2. 导出为 JSON（`export_json.php`）

```php
<?php
require 'db.php';

if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}

$stmt = $db->prepare("SELECT title, url, tags, created_at FROM bookmarks WHERE user_id = ? ORDER BY created_at DESC");
$stmt->execute([$_SESSION['user_id']]);
$bookmarks = $stmt->fetchAll();

$data = [];
foreach ($bookmarks as $b) {
    $data[] = [
        'title' => $b['title'],
        'url' => $b['url'],
        'tags' => $b['tags'],
        'created_at' => $b['created_at']
    ];
}

$json = json_encode($data, JSON_UNESCAPED_UNICODE | JSON_PRETTY_PRINT);
$filename = "bookmarks_{$_SESSION['username']}_" . date('Y-m-d') . ".json";

header('Content-Type: application/json');
header("Content-Disposition: attachment; filename=\"$filename\"");
header('Content-Length: ' . strlen($json));
echo $json;
exit;
?>
```

---

## 3. 导出为 Markdown（`export_md.php`）

```php
<?php
require 'db.php';

if (!$_SESSION['user_id']) {
    header('Location: login.php');
    exit;
}

$stmt = $db->prepare("SELECT title, url, tags FROM bookmarks WHERE user_id = ? ORDER BY created_at DESC");
$stmt->execute([$_SESSION['user_id']]);
$bookmarks = $stmt->fetchAll();

$md = "# 我的书签\n\n";
$md .= "> 导出时间: " . date('Y-m-d H:i:s') . "\n\n";

foreach ($bookmarks as $b) {
    $tagStr = $b['tags'] ? " `{$b['tags']}`" : '';
    $md .= "- [{$b['title']}]({$b['url']}){$tagStr}\n";
}

$filename = "bookmarks_{$_SESSION['username']}_" . date('Y-m-d') . ".md";

header('Content-Type: text/markdown; charset=utf-8');
header("Content-Disposition: attachment; filename=\"$filename\"");
header('Content-Length: ' . strlen($md));
echo $md;
exit;
?>
```

---

## 4. 添加“书签点击统计”功能

### 1. 修改数据库：添加 `click_count` 字段

在 `db.php` 的 `bookmarks` 表创建语句中加入：

```php
$db->exec("CREATE TABLE IF NOT EXISTS bookmarks (
    ...
    click_count INTEGER DEFAULT 0,
    ...
)");
```

### 2. 创建 `go.php` 作为跳转中间页（用于统计）

```php
<?php
require 'db.php';

if (!$_SESSION['user_id']) {
    die("请先登录。");
}

$id = $_GET['id'] ?? null;
if (!$id || !is_numeric($id)) {
    die("无效链接。");
}

// 检查归属并更新点击数
$stmt = $db->prepare("UPDATE bookmarks SET click_count = click_count + 1 WHERE id = ? AND user_id = ?");
$stmt->execute([$id, $_SESSION['user_id']]);

// 获取 URL
$stmt = $db->prepare("SELECT url FROM bookmarks WHERE id = ? AND user_id = ?");
$stmt->execute([$id, $_SESSION['user_id']]);
$row = $stmt->fetch();

if (!$row) {
    die("书签不存在或无权访问。");
}

// 跳转
header('Location: ' . $row['url']);
exit;
?>
```

### 3. 修改 `index.php` 中的链接

把原来的：

```html
<a href="<?= htmlspecialchars($b['url']) ?>" target="_blank">
```

改成：

```html
<a href="go.php?id=<?= $b['id'] ?>" target="_blank">
```

### 4. 在书签列表显示点击次数

在 `index.php` 的书签项中添加：

```html
<small class="text-muted">点击: <?= $b['click_count'] ?> 次</small>
```

---

## ✅ 更新导航按钮（在 `index.php`）

```php
<p class="text-end">
    欢迎，<?= htmlspecialchars($_SESSION['username']) ?> |
    <div class="btn-group btn-group-sm">
        <a href="export.php" class="btn btn-outline-info">HTML</a>
        <a href="export_json.php" class="btn btn-outline-info">JSON</a>
        <a href="export_md.php" class="btn btn-outline-info">MD</a>
    </div>
    <a href="import.php" class="btn btn-outline-warning btn-sm">📥 导入</a>
    <a href="logout.php" class="btn btn-sm btn-outline-danger">登出</a>
</p>
```

---

## 📦 完整项目结构

```
/bookmark-app/
├── index.php
├── add.php
├── edit.php
├── delete.php
├── login.php
├── register.php
├── forgot.php
├── reset.php
├── logout.php
├── export.php
├── export_json.php
├── export_md.php
├── import.php
├── go.php
├── db.php
├── mail.php
└── style.css
```

---

## ✅ 功能总结

| 功能 | 说明 |
|------|------|
| 📥 导入 HTML/JSON | 支持从浏览器或其他系统迁移 |
| 💾 导出 HTML/JSON/MD | 多格式备份 |
| 👁️ 点击统计 | 了解常用书签 |
| 🔐 安全控制 | 用户隔离、XSS 防护 |

---

需要我：
- 把整个项目打包成 ZIP 供你下载？
- 添加“书签排序”功能（按点击数、时间等）？
- 增加“标签云”展示？
- 部署到服务器的详细指南？

告诉我，我可以立即为你生成完整压缩包或部署文档！ 😊

--------------------------------------------------------------------------------
