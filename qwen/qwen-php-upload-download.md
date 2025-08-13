# php中如何实现文件的上传和下载。

在 PHP 中实现文件的上传和下载是 Web 开发中的常见需求。下面分别详细介绍**文件上传**和**文件下载**的完整实现方法，包括前端表单、后端处理、安全注意事项和最佳实践。

---

## ✅ 一、PHP 文件上传

### 1. HTML 表单（前端）

```html
<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <title>文件上传</title>
</head>
<body>
    <h2>上传文件</h2>
    <form action="upload.php" method="POST" enctype="multipart/form-data">
        <input type="file" name="uploaded_file" required>
        <button type="submit">上传文件</button>
    </form>

    <?php if (isset($uploadMessage)): ?>
        <p style="color: <?= $uploadSuccess ? 'green' : 'red' ?>;">
            <?= htmlspecialchars($uploadMessage) ?>
        </p>
    <?php endif; ?>
</body>
</html>
```

> ⚠️ 注意：必须设置 `enctype="multipart/form-data"`，否则无法上传文件。

---

### 2. PHP 处理上传（upload.php）

```php
<?php
$uploadMessage = '';
$uploadSuccess = false;

// 上传目录（确保可写）
$uploadDir = 'uploads/';
if (!is_dir($uploadDir)) {
    mkdir($uploadDir, 0755, true);
}

// 检查是否提交了文件
if ($_SERVER['REQUEST_METHOD'] === 'POST' && isset($_FILES['uploaded_file'])) {
    $file = $_FILES['uploaded_file'];

    // 检查上传错误
    if ($file['error'] !== UPLOAD_ERR_OK) {
        $uploadMessage = '上传失败：错误代码 ' . $file['error'];
    } else {
        $fileName = basename($file['name']); // 获取原始文件名
        $fileTmpPath = $file['tmp_name'];
        $fileSize = $file['size'];
        $fileType = $file['type'];

        // 安全检查：限制文件类型（白名单）
        $allowedTypes = ['image/jpeg', 'image/png', 'application/pdf', 'text/plain'];
        if (!in_array($fileType, $allowedTypes)) {
            $uploadMessage = '不支持的文件类型！仅允许：JPG, PNG, PDF, TXT';
        }
        // 限制文件大小（例如：5MB）
        elseif ($fileSize > 5 * 1024 * 1024) {
            $uploadMessage = '文件太大！最大支持 5MB。';
        }
        // 防止路径遍历攻击
        elseif (preg_match('/[^a-zA-Z0-9._-]/', $fileName)) {
            $uploadMessage = '文件名包含非法字符！';
        }
        else {
            // 生成唯一文件名，避免重名和安全问题
            $fileExtension = strtolower(pathinfo($fileName, PATHINFO_EXTENSION));
            $uniqueName = uniqid() . '.' . $fileExtension;
            $targetPath = $uploadDir . $uniqueName;

            if (move_uploaded_file($fileTmpPath, $targetPath)) {
                $uploadMessage = "文件上传成功！<br>保存为: " . htmlspecialchars($uniqueName);
                $uploadSuccess = true;
            } else {
                $uploadMessage = '文件移动失败，请检查权限。';
            }
        }
    }
}
?>
```

### 🔐 安全建议（重要！）

| 风险 | 解决方案 |
|------|----------|
| 文件类型伪造 | 使用 `mime_content_type()` 或 `finfo` 检查真实 MIME 类型 |
| 路径遍历 | 使用 `basename()` 过滤文件名，不要使用用户输入的路径 |
| 执行脚本 | 禁止上传目录执行 PHP 脚本（`.htaccess` 或服务器配置） |
| 文件名冲突 | 使用 `uniqid()` 或时间戳生成唯一文件名 |
| 大小限制 | 同时设置 `post_max_size` 和 `upload_max_filesize`（php.ini） |

#### ✅ 推荐：使用 `finfo` 检测真实 MIME 类型：

```php
$finfo = finfo_open(FILEINFO_MIME_TYPE);
$mimeType = finfo_file($finfo, $fileTmpPath);
finfo_close($finfo);

$allowedMimeTypes = ['image/jpeg', 'image/png', 'application/pdf', 'text/plain'];
if (!in_array($mimeType, $allowedMimeTypes)) {
    die('非法文件类型！');
}
```

#### 📌 修改 php.ini（临时测试用）：
```ini
upload_max_filesize = 10M
post_max_size = 10M
```

---

## ✅ 二、PHP 文件下载

### 方法一：直接链接下载（简单但安全性低）

```html
<a href="uploads/example.pdf" download>下载示例文件</a>
```

> ⚠️ 缺点：暴露真实路径，任何人都能访问，不安全！

---

### 方法二：通过 PHP 脚本控制下载（推荐）

```php
<?php
// download.php

$file = $_GET['file'] ?? '';

// 基础安全校验：只允许下载 uploads/ 目录下的文件
$uploadDir = 'uploads/';
$filePath = $uploadDir . basename($file); // 防止路径遍历

// 检查文件是否存在且是合法文件
if (!$file || !file_exists($filePath) || strpos($filePath, $uploadDir) !== 0) {
    http_response_code(404);
    echo '文件不存在或访问被拒绝。';
    exit;
}

// 获取文件信息
$fileName = basename($filePath);
$fileSize = filesize($filePath);
$fileType = mime_content_type($filePath);

// 设置响应头，强制浏览器下载
header('Content-Type: ' . $fileType);
header('Content-Disposition: attachment; filename="' . urlencode($fileName) . '"');
header('Content-Length: ' . $fileSize);
header('Cache-Control: private');
header('Pragma: public');

// 输出文件内容
readfile($filePath);
exit;
?>
```

### 使用方式：

```html
<a href="download.php?file=abc123.jpg">下载图片</a>
```

### 🔐 下载安全增强建议：

- 不要让用户直接指定任意文件名 → 可用数据库映射 ID 到文件名
- 验证用户权限（如登录后才能下载）
- 记录下载日志
- 对敏感文件使用 `X-Sendfile`（Nginx/Apache 支持）提升性能

#### ✅ 更高级：使用数据库映射（推荐生产环境）

```sql
CREATE TABLE files (
    id INT AUTO_INCREMENT PRIMARY KEY,
    original_name VARCHAR(255),
    stored_name VARCHAR(255),
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

下载时根据 `id` 查询真实文件名，避免暴露物理路径：

```php
$id = (int)$_GET['id'];
$stmt = $pdo->prepare("SELECT stored_name FROM files WHERE id = ?");
$stmt->execute([$id]);
$row = $stmt->fetch();

if ($row) {
    $filePath = 'uploads/' . $row['stored_name'];
    // ... 后续同上
} else {
    die('文件不存在');
}
```

---

## ✅ 总结对比

| 功能 | 推荐做法 |
|------|----------|
| **上传** | 使用 `$_FILES` + 白名单 MIME 类型 + 重命名文件 + 限制大小 + 非执行目录 |
| **下载** | 通过 PHP 脚本控制，禁止直接访问上传目录，使用 `readfile()` + 正确响应头 |
| **安全** | 永远不要信任用户输入，使用 `basename()`、`uniqid()`、`finfo`、权限验证 |
| **性能** | 生产环境建议配合 `X-Sendfile`（Nginx/Apache）减少 PHP 内存占用 |

---

## 💡 额外提示

- 使用框架（如 Laravel、Symfony）内置的文件上传组件更安全便捷。
- 可结合 CDN 存储大文件（如阿里云 OSS、腾讯云 COS），减轻服务器压力。
- 对图片上传可增加缩略图生成（使用 GD 或 Imagick）。
- 使用 `is_uploaded_file()` 检查是否为真正上传的文件（比 `move_uploaded_file()` 更严格）。

---

✅ **完成！你现在已掌握 PHP 中安全可靠的文件上传与下载实现方法。**  
建议在本地测试后再部署到生产环境，并始终遵循“最小权限”和“输入过滤”原则。

