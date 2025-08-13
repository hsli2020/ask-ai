# 编写一个网页，其中包含一个table，每一行只显示了客户信息的部分字段，每一行的最后一列是个编辑按钮，用户点击按钮时，从服务器端取得客户的全部信息，弹出一个对话框进行编辑，编辑完毕将结果发送给服务器保存。服务器端用PHP编写

我来为您创建一个完整的客户信息管理网页，包含前端HTML/JavaScript和后端PHP代码。

## 前端HTML/JavaScript (index.html)

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>客户信息管理</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            margin: 20px;
            background-color: #f5f5f5;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background-color: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        
        h1 {
            text-align: center;
            color: #333;
            margin-bottom: 30px;
        }
        
        table {
            width: 100%;
            border-collapse: collapse;
            margin-bottom: 20px;
        }
        
        th, td {
            padding: 12px;
            text-align: left;
            border-bottom: 1px solid #ddd;
        }
        
        th {
            background-color: #4CAF50;
            color: white;
        }
        
        tr:hover {
            background-color: #f5f5f5;
        }
        
        .edit-btn {
            background-color: #008CBA;
            color: white;
            border: none;
            padding: 6px 12px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
        }
        
        .edit-btn:hover {
            background-color: #007B9A;
        }
        
        /* Modal Styles */
        .modal {
            display: none;
            position: fixed;
            z-index: 1000;
            left: 0;
            top: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0,0,0,0.5);
        }
        
        .modal-content {
            background-color: white;
            margin: 5% auto;
            padding: 20px;
            border-radius: 8px;
            width: 60%;
            max-width: 600px;
            box-shadow: 0 4px 20px rgba(0,0,0,0.3);
        }
        
        .modal-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding-bottom: 10px;
            border-bottom: 2px solid #4CAF50;
        }
        
        .modal-header h2 {
            margin: 0;
            color: #333;
        }
        
        .close {
            color: #aaa;
            font-size: 28px;
            font-weight: bold;
            cursor: pointer;
        }
        
        .close:hover {
            color: #000;
        }
        
        .form-group {
            margin-bottom: 15px;
        }
        
        .form-group label {
            display: block;
            margin-bottom: 5px;
            font-weight: bold;
            color: #555;
        }
        
        .form-group input, .form-group textarea {
            width: 100%;
            padding: 8px;
            border: 1px solid #ddd;
            border-radius: 4px;
            font-size: 14px;
            box-sizing: border-box;
        }
        
        .form-group textarea {
            height: 80px;
            resize: vertical;
        }
        
        .modal-footer {
            text-align: right;
            margin-top: 20px;
        }
        
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 14px;
            margin-left: 10px;
        }
        
        .btn-primary {
            background-color: #4CAF50;
            color: white;
        }
        
        .btn-primary:hover {
            background-color: #45a049;
        }
        
        .btn-secondary {
            background-color: #6c757d;
            color: white;
        }
        
        .btn-secondary:hover {
            background-color: #5a6268;
        }
        
        .loading {
            text-align: center;
            padding: 20px;
            color: #666;
        }
        
        .message {
            padding: 10px;
            margin: 10px 0;
            border-radius: 4px;
            display: none;
        }
        
        .message.success {
            background-color: #d4edda;
            color: #155724;
            border: 1px solid #c3e6cb;
        }
        
        .message.error {
            background-color: #f8d7da;
            color: #721c24;
            border: 1px solid #f5c6cb;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>客户信息管理</h1>
        <div id="message" class="message"></div>
        
        <table id="customerTable">
            <thead>
                <tr>
                    <th>ID</th>
                    <th>姓名</th>
                    <th>邮箱</th>
                    <th>电话</th>
                    <th>操作</th>
                </tr>
            </thead>
            <tbody id="customerTableBody">
                <!-- 客户数据将通过AJAX加载 -->
            </tbody>
        </table>
    </div>

    <!-- 编辑模态框 -->
    <div id="editModal" class="modal">
        <div class="modal-content">
            <div class="modal-header">
                <h2>编辑客户信息</h2>
                <span class="close">&times;</span>
            </div>
            <div id="editForm">
                <input type="hidden" id="customerId">
                <div class="form-group">
                    <label for="customerName">姓名:</label>
                    <input type="text" id="customerName" required>
                </div>
                <div class="form-group">
                    <label for="customerEmail">邮箱:</label>
                    <input type="email" id="customerEmail" required>
                </div>
                <div class="form-group">
                    <label for="customerPhone">电话:</label>
                    <input type="tel" id="customerPhone">
                </div>
                <div class="form-group">
                    <label for="customerAddress">地址:</label>
                    <textarea id="customerAddress"></textarea>
                </div>
                <div class="form-group">
                    <label for="customerCompany">公司:</label>
                    <input type="text" id="customerCompany">
                </div>
                <div class="form-group">
                    <label for="customerNotes">备注:</label>
                    <textarea id="customerNotes"></textarea>
                </div>
                <div class="modal-footer">
                    <button class="btn btn-secondary" id="cancelBtn">取消</button>
                    <button class="btn btn-primary" id="saveBtn">保存</button>
                </div>
            </div>
        </div>
    </div>

    <script>
        // 获取DOM元素
        const modal = document.getElementById('editModal');
        const closeBtn = document.querySelector('.close');
        const cancelBtn = document.getElementById('cancelBtn');
        const saveBtn = document.getElementById('saveBtn');
        const messageDiv = document.getElementById('message');

        // 显示消息
        function showMessage(text, type) {
            messageDiv.textContent = text;
            messageDiv.className = `message ${type}`;
            messageDiv.style.display = 'block';
            setTimeout(() => {
                messageDiv.style.display = 'none';
            }, 3000);
        }

        // 关闭模态框
        function closeModal() {
            modal.style.display = 'none';
        }

        // 打开模态框并加载客户数据
        async function openEditModal(customerId) {
            try {
                const response = await fetch(`get_customer.php?id=${customerId}`);
                const customer = await response.json();
                
                if (customer.error) {
                    showMessage(customer.error, 'error');
                    return;
                }
                
                // 填充表单数据
                document.getElementById('customerId').value = customer.id;
                document.getElementById('customerName').value = customer.name || '';
                document.getElementById('customerEmail').value = customer.email || '';
                document.getElementById('customerPhone').value = customer.phone || '';
                document.getElementById('customerAddress').value = customer.address || '';
                document.getElementById('customerCompany').value = customer.company || '';
                document.getElementById('customerNotes').value = customer.notes || '';
                
                modal.style.display = 'block';
            } catch (error) {
                console.error('加载客户数据失败:', error);
                showMessage('加载客户数据失败', 'error');
            }
        }

        // 保存客户数据
        async function saveCustomer() {
            const customerId = document.getElementById('customerId').value;
            const customerData = {
                id: customerId,
                name: document.getElementById('customerName').value,
                email: document.getElementById('customerEmail').value,
                phone: document.getElementById('customerPhone').value,
                address: document.getElementById('customerAddress').value,
                company: document.getElementById('customerCompany').value,
                notes: document.getElementById('customerNotes').value
            };

            try {
                const response = await fetch('update_customer.php', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                    body: JSON.stringify(customerData)
                });
                
                const result = await response.json();
                
                if (result.success) {
                    showMessage('客户信息保存成功！', 'success');
                    closeModal();
                    loadCustomers(); // 重新加载客户列表
                } else {
                    showMessage(result.error || '保存失败', 'error');
                }
            } catch (error) {
                console.error('保存客户数据失败:', error);
                showMessage('保存客户数据失败', 'error');
            }
        }

        // 加载客户列表
        async function loadCustomers() {
            const tableBody = document.getElementById('customerTableBody');
            tableBody.innerHTML = '<tr><td colspan="5" class="loading">加载中...</td></tr>';
            
            try {
                const response = await fetch('get_customers.php');
                const customers = await response.json();
                
                if (customers.error) {
                    tableBody.innerHTML = `<tr><td colspan="5" style="text-align: center; color: red;">${customers.error}</td></tr>`;
                    return;
                }
                
                if (customers.length === 0) {
                    tableBody.innerHTML = '<tr><td colspan="5" style="text-align: center;">暂无客户数据</td></tr>';
                    return;
                }
                
                tableBody.innerHTML = customers.map(customer => `
                    <tr>
                        <td>${customer.id}</td>
                        <td>${customer.name || 'N/A'}</td>
                        <td>${customer.email || 'N/A'}</td>
                        <td>${customer.phone || 'N/A'}</td>
                        <td>
                            <button class="edit-btn" onclick="openEditModal(${customer.id})">编辑</button>
                        </td>
                    </tr>
                `).join('');
            } catch (error) {
                console.error('加载客户列表失败:', error);
                tableBody.innerHTML = '<tr><td colspan="5" style="text-align: center; color: red;">加载失败，请刷新页面重试</td></tr>';
            }
        }

        // 事件监听器
        closeBtn.onclick = closeModal;
        cancelBtn.onclick = closeModal;
        saveBtn.onclick = saveCustomer;

        // 点击模态框外部关闭
        window.onclick = function(event) {
            if (event.target === modal) {
                closeModal();
            }
        }

        // 页面加载完成后加载客户数据
        document.addEventListener('DOMContentLoaded', loadCustomers);
    </script>
</body>
</html>
```

## 后端PHP文件

### 1. 数据库配置 (config.php)

```php
<?php
// 数据库配置
$host = 'localhost';
$dbname = 'customer_db';
$username = 'your_username';
$password = 'your_password';

try {
    $pdo = new PDO("mysql:host=$host;dbname=$dbname;charset=utf8", $username, $password);
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);
} catch(PDOException $e) {
    die("数据库连接失败: " . $e->getMessage());
}
?>
```

### 2. 获取客户列表 (get_customers.php)

```php
<?php
header('Content-Type: application/json; charset=utf-8');

// 包含数据库配置
require_once 'config.php';

try {
    // 查询客户基本信息（部分字段）
    $stmt = $pdo->query("
        SELECT id, name, email, phone 
        FROM customers 
        ORDER BY id DESC
    ");
    $customers = $stmt->fetchAll(PDO::FETCH_ASSOC);
    
    echo json_encode($customers, JSON_UNESCAPED_UNICODE);
    
} catch(PDOException $e) {
    echo json_encode(['error' => '获取客户列表失败: ' . $e->getMessage()]);
}
?>
```

### 3. 获取单个客户详细信息 (get_customer.php)

```php
<?php
header('Content-Type: application/json; charset=utf-8');

// 包含数据库配置
require_once 'config.php';

// 获取客户ID
if (!isset($_GET['id']) || !is_numeric($_GET['id'])) {
    echo json_encode(['error' => '无效的客户ID']);
    exit;
}

$customerId = (int)$_GET['id'];

try {
    // 查询客户详细信息
    $stmt = $pdo->prepare("
        SELECT id, name, email, phone, address, company, notes 
        FROM customers 
        WHERE id = ?
    ");
    $stmt->execute([$customerId]);
    $customer = $stmt->fetch(PDO::FETCH_ASSOC);
    
    if (!$customer) {
        echo json_encode(['error' => '未找到指定客户']);
    } else {
        echo json_encode($customer, JSON_UNESCAPED_UNICODE);
    }
    
} catch(PDOException $e) {
    echo json_encode(['error' => '获取客户信息失败: ' . $e->getMessage()]);
}
?>
```

### 4. 更新客户信息 (update_customer.php)

```php
<?php
header('Content-Type: application/json; charset=utf-8');

// 包含数据库配置
require_once 'config.php';

// 获取POST数据
$input = json_decode(file_get_contents('php://input'), true);

// 验证必需字段
if (!isset($input['id']) || !is_numeric($input['id'])) {
    echo json_encode(['success' => false, 'error' => '无效的客户ID']);
    exit;
}

if (empty($input['name']) || empty($input['email'])) {
    echo json_encode(['success' => false, 'error' => '姓名和邮箱为必填项']);
    exit;
}

// 验证邮箱格式
if (!filter_var($input['email'], FILTER_VALIDATE_EMAIL)) {
    echo json_encode(['success' => false, 'error' => '邮箱格式不正确']);
    exit;
}

$customerId = (int)$input['id'];
$name = trim($input['name']);
$email = trim($input['email']);
$phone = isset($input['phone']) ? trim($input['phone']) : '';
$address = isset($input['address']) ? trim($input['address']) : '';
$company = isset($input['company']) ? trim($input['company']) : '';
$notes = isset($input['notes']) ? trim($input['notes']) : '';

try {
    // 检查邮箱是否已存在（排除当前客户）
    $stmt = $pdo->prepare("
        SELECT id FROM customers 
        WHERE email = ? AND id != ?
    ");
    $stmt->execute([$email, $customerId]);
    if ($stmt->fetch()) {
        echo json_encode(['success' => false, 'error' => '该邮箱已被其他客户使用']);
        exit;
    }
    
    // 更新客户信息
    $stmt = $pdo->prepare("
        UPDATE customers 
        SET name = ?, email = ?, phone = ?, address = ?, company = ?, notes = ?, 
            updated_at = NOW()
        WHERE id = ?
    ");
    
    $result = $stmt->execute([
        $name, $email, $phone, $address, $company, $notes, $customerId
    ]);
    
    if ($result) {
        echo json_encode(['success' => true]);
    } else {
        echo json_encode(['success' => false, 'error' => '更新失败']);
    }
    
} catch(PDOException $e) {
    echo json_encode(['success' => false, 'error' => '数据库错误: ' . $e->getMessage()]);
}
?>
```

### 5. 创建数据库表的SQL脚本

```sql
-- 创建数据库
CREATE DATABASE IF NOT EXISTS customer_db CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 使用数据库
USE customer_db;

-- 创建客户表
CREATE TABLE IF NOT EXISTS customers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL UNIQUE,
    phone VARCHAR(20),
    address TEXT,
    company VARCHAR(100),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 插入示例数据
INSERT INTO customers (name, email, phone, address, company, notes) VALUES
('张三', 'zhangsan@example.com', '13800138001', '北京市朝阳区xxx街道', 'ABC公司', '重要客户'),
('李四', 'lisi@example.com', '13800138002', '上海市浦东新区xxx路', 'XYZ公司', '新客户'),
('王五', 'wangwu@example.com', '13800138003', '广州市天河区xxx大厦', 'DEF公司', '需要跟进');
```

## 使用说明

1. **数据库配置**：
   - 修改 `config.php` 中的数据库连接信息
   - 执行SQL脚本创建数据库和表

2. **文件结构**：
   ```
   /your-project-folder
   ├── index.html
   ├── config.php
   ├── get_customers.php
   ├── get_customer.php
   └── update_customer.php
   ```

3. **功能特点**：
   - 响应式设计，适配不同屏幕尺寸
   - 表单验证（前端和后端）
   - 错误处理和用户友好提示
   - 模态框编辑体验
   - AJAX异步操作，无需页面刷新
   - 安全的数据库操作（使用PDO预处理语句）

4. **安全考虑**：
   - 使用PDO预处理语句防止SQL注入
   - 邮箱格式验证
   - 邮箱唯一性检查
   - 输入数据过滤和验证

这个完整的解决方案提供了客户信息的展示、编辑和保存功能，具有良好的用户体验和安全性。
