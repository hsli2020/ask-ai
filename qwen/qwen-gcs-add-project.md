# 编写一个网页，网页中有一个form，接收用户输入的设备信息，每个设备包含name，code，type，group三个字段，应该允许用户一次输入多个设备，所有设备输入完毕，再提交给服务器

这是一个完整的单文件解决方案，包含了 HTML 结构、CSS 样式和 JavaScript 逻辑。

这个网页允许用户动态添加多行设备信息，并在最后一次性打包成 JSON 格式提交。

### 完整代码 (index.html)

你可以将以下代码保存为 `index.html` 文件，然后用浏览器打开即可。

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>批量设备信息录入</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f4f4f9;
            padding: 20px;
            max-width: 1000px;
            margin: 0 auto;
        }

        h1 {
            text-align: center;
            color: #333;
        }

        .form-container {
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        /* 表头样式 */
        .header-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr 50px; /* 对应 4 个字段 + 操作列 */
            gap: 10px;
            margin-bottom: 10px;
            font-weight: bold;
            color: #555;
        }

        /* 输入行样式 */
        .input-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr 50px;
            gap: 10px;
            margin-bottom: 10px;
            align-items: center;
            animation: fadeIn 0.3s ease;
        }

        input {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            width: 100%;
            box-sizing: border-box; /* 确保 padding 不撑大宽度 */
        }

        input:focus {
            border-color: #007bff;
            outline: none;
        }

        /* 按钮样式 */
        .btn {
            padding: 10px 20px;
            border: none;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            transition: background 0.2s;
        }

        .btn-add {
            background-color: #28a745;
            color: white;
            margin-bottom: 20px;
        }
        .btn-add:hover { background-color: #218838; }

        .btn-submit {
            background-color: #007bff;
            color: white;
            width: 100%;
            margin-top: 10px;
        }
        .btn-submit:hover { background-color: #0069d9; }

        .btn-remove {
            background-color: #dc3545;
            color: white;
            padding: 8px 12px;
            font-size: 14px;
        }
        .btn-remove:hover { background-color: #c82333; }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-10px); }
            to { opacity: 1; transform: translateY(0); }
        }
    </style>
</head>
<body>

    <h1>设备批量录入</h1>

    <div class="form-container">
        <form id="deviceForm">
            <!-- 表头 -->
            <div class="header-row">
                <div>设备名称 (Name)</div>
                <div>设备编码 (Code)</div>
                <div>设备类型 (Type)</div>
                <div>所属分组 (Group)</div>
                <div>操作</div>
            </div>

            <!-- 动态生成的设备列表容器 -->
            <div id="deviceList">
                <!-- 默认先给一行 -->
            </div>

            <!-- 添加按钮 -->
            <button type="button" class="btn btn-add" onclick="addDeviceRow()">+ 添加设备</button>

            <!-- 提交按钮 -->
            <button type="submit" class="btn btn-submit">提交所有数据</button>
        </form>
    </div>

    <script>
        const deviceList = document.getElementById('deviceList');

        // 初始化：页面加载时添加一行
        window.onload = function() {
            addDeviceRow();
        };

        // 1. 添加一行设备输入框的函数
        function addDeviceRow() {
            const row = document.createElement('div');
            row.className = 'input-row';
            
            row.innerHTML = `
                <input type="text" name="name" placeholder="输入名称" required>
                <input type="text" name="code" placeholder="输入编码" required>
                <input type="text" name="type" placeholder="输入类型">
                <input type="text" name="group" placeholder="输入分组">
                <button type="button" class="btn btn-remove" onclick="removeRow(this)">删除</button>
            `;
            
            deviceList.appendChild(row);
        }

        // 2. 删除一行的函数
        function removeRow(button) {
            // 找到按钮所在的行 (父元素) 并移除
            const row = button.parentElement;
            // 如果只剩一行，提示用户或者禁止删除（这里允许删除，删完会自动补一行或者留空）
            if (deviceList.children.length > 1) {
                row.remove();
            } else {
                alert("至少需要保留一行输入框");
            }
        }

        // 3. 处理表单提交
        document.getElementById('deviceForm').addEventListener('submit', function(e) {
            e.preventDefault(); // 阻止表单默认跳转提交

            const rows = document.querySelectorAll('.input-row');
            const devices = [];

            // 遍历每一行，收集数据
            rows.forEach(row => {
                const inputs = row.querySelectorAll('input');
                
                // 获取每个字段的值
                const name = inputs[0].value.trim();
                const code = inputs[1].value.trim();
                const type = inputs[2].value.trim();
                const group = inputs[3].value.trim();

                // 简单的非空校验（根据需求可调整）
                if (name && code) {
                    devices.push({
                        name: name,
                        code: code,
                        type: type,
                        group: group
                    });
                } else if (name || code || type || group) {
                    // 如果填了部分信息但没填必填项
                    alert("请确保每行都填写了名称和编码，或者清空整行。");
                    return; 
                }
            });

            if (devices.length === 0) {
                alert("没有有效的数据可提交");
                return;
            }

            // 4. 发送数据给服务器
            console.log("准备提交的数据:", devices);
            
            // 模拟发送请求 (实际使用时请取消下面的注释并替换 URL)
            /*
            fetch('/api/save-devices', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json'
                },
                body: JSON.stringify(devices)
            })
            .then(response => response.json())
            .then(data => {
                alert('提交成功！');
            })
            .catch(error => {
                console.error('Error:', error);
                alert('提交失败');
            });
            */

            alert(`提交成功！共收集到 ${devices.length} 条设备信息。\n请在控制台 (F12) 查看 JSON 数据。`);
        });
    </script>
</body>
</html>
```

### 代码功能说明：

1.  **动态添加/删除 (JavaScript)**:
    *   `addDeviceRow()` 函数会创建一个新的 `div`，里面包含 4 个 `input` 和一个删除按钮，并追加到列表中。
    *   `removeRow()` 函数允许用户删除填错的行。
2.  **数据收集 (JavaScript)**:
    *   点击提交时，脚本会遍历所有生成的行。
    *   它将每一行的 4 个字段提取出来，组装成一个对象。
    *   最后将所有对象放入一个数组 `devices` 中。
3.  **数据提交 (Fetch API)**:
    *   代码中注释掉了真实的 `fetch` 请求部分。
    *   当前演示模式下，它会弹出一个 Alert 提示框，并将生成的 JSON 数据打印在浏览器的控制台（按 F12 查看 Console 标签）。

### 如何对接后端？

在代码的 `// 4. 发送数据给服务器` 部分，你可以取消注释并修改 URL。后端接收到的数据格式将是这样的 JSON 数组：

```json
[
  {
    "name": "服务器 A",
    "code": "SRV-001",
    "type": "Linux",
    "group": "生产环境"
  },
  {
    "name": "数据库 B",
    "code": "DB-002",
    "type": "MySQL",
    "group": "数据层"
  }
]
```


# 请把网页修改成静态网页，无需动态添加，就像传统的form一样，form中默认放上三个设备的输入框就可以了

这是一个修改后的静态网页版本。去掉了动态添加/删除的功能，默认固定显示三个设备的输入框，界面更加简洁，类似传统的表格表单。

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>设备信息录入</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f4f4f9;
            padding: 20px;
            max-width: 900px;
            margin: 0 auto;
        }

        h1 {
            text-align: center;
            color: #333;
        }

        .form-container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .form-description {
            margin-bottom: 20px;
            color: #666;
            font-size: 14px;
        }

        /* 表头样式 */
        .header-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr; /* 4 个字段，无操作列 */
            gap: 15px;
            margin-bottom: 15px;
            font-weight: bold;
            color: #555;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }

        /* 输入行样式 */
        .input-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr;
            gap: 15px;
            margin-bottom: 15px;
            align-items: center;
        }

        /* 行号标签 */
        .row-label {
            position: absolute;
            left: -40px;
            color: #999;
            font-weight: bold;
            width: 30px;
            text-align: right;
        }
        
        .input-wrapper {
            position: relative;
        }

        input {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            width: 100%;
            box-sizing: border-box;
            background-color: #fafafa;
        }

        input:focus {
            border-color: #007bff;
            background-color: #fff;
            outline: none;
            box-shadow: 0 0 0 3px rgba(0,123,255,0.1);
        }

        /* 提交按钮样式 */
        .btn-submit {
            background-color: #007bff;
            color: white;
            border: none;
            padding: 12px 20px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            width: 100%;
            margin-top: 20px;
            transition: background 0.2s;
        }

        .btn-submit:hover {
            background-color: #0069d9;
        }

        /* 响应式调整 */
        @media (max-width: 600px) {
            .header-row, .input-row {
                grid-template-columns: 1fr;
                gap: 5px;
            }
            .header-row {
                display: none; /* 手机端隐藏表头 */
            }
            .input-row {
                border: 1px solid #eee;
                padding: 10px;
                border-radius: 4px;
                margin-bottom: 10px;
            }
            .row-label {
                position: static;
                text-align: left;
                margin-bottom: 5px;
                display: block;
                color: #007bff;
            }
        }
    </style>
</head>
<body>

    <h1>设备信息录入</h1>

    <div class="form-container">
        <p class="form-description">请填写以下三个设备的信息，完成后点击提交。</p>
        
        <form id="deviceForm">
            <!-- 表头 -->
            <div class="header-row">
                <div>设备名称 (Name)</div>
                <div>设备编码 (Code)</div>
                <div>设备类型 (Type)</div>
                <div>所属分组 (Group)</div>
            </div>

            <!-- 设备 1 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 1</span>
                    <input type="text" name="name" placeholder="例如：Web 服务器" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code" placeholder="例如：SRV-001" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type" placeholder="例如：Linux">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group" placeholder="例如：生产组">
                </div>
            </div>

            <!-- 设备 2 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 2</span>
                    <input type="text" name="name" placeholder="例如：数据库服务器" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code" placeholder="例如：DB-002" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type" placeholder="例如：MySQL">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group" placeholder="例如：数据组">
                </div>
            </div>

            <!-- 设备 3 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 3</span>
                    <input type="text" name="name" placeholder="例如：缓存服务器" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code" placeholder="例如：CACHE-003" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type" placeholder="例如：Redis">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group" placeholder="例如：中间件组">
                </div>
            </div>

            <!-- 提交按钮 -->
            <button type="submit" class="btn-submit">提交所有设备信息</button>
        </form>
    </div>

    <script>
        document.getElementById('deviceForm').addEventListener('submit', function(e) {
            e.preventDefault(); // 阻止表单默认刷新提交

            const rows = document.querySelectorAll('.input-row');
            const devices = [];

            // 遍历固定的三行
            rows.forEach((row, index) => {
                const inputs = row.querySelectorAll('input');
                
                const name = inputs[0].value.trim();
                const code = inputs[1].value.trim();
                const type = inputs[2].value.trim();
                const group = inputs[3].value.trim();

                // 校验：如果某一行填了部分内容，则必须填完必填项
                if (name || code || type || group) {
                    if (!name || !code) {
                        alert(`设备 ${index + 1} 缺少必填项（名称或编码）`);
                        throw new Error("Validation failed"); // 中断提交
                    }
                    devices.push({
                        name: name,
                        code: code,
                        type: type,
                        group: group
                    });
                }
            });

            if (devices.length === 0) {
                alert("请至少填写一个设备的信息");
                return;
            }

            // 准备提交的数据
            console.log("提交的数据 JSON:", JSON.stringify(devices, null, 2));

            // 模拟发送请求
            /*
            fetch('/api/save-devices', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(devices)
            })
            .then(res => res.json())
            .then(data => alert('提交成功！'))
            .catch(err => alert('提交失败'));
            */

            alert(`成功收集 ${devices.length} 条设备信息。\n请按 F12 打开控制台查看 JSON 数据结构。`);
        });
    </script>
</body>
</html>
```

### 主要修改点：

1.  **静态结构**：HTML 中直接写死了三个 `.input-row`  div，不再依赖 JavaScript 生成。
2.  **移除操作列**：去掉了“添加”按钮、“删除”按钮以及表头中的“操作”列，界面更清爽。
3.  **行号标识**：在 CSS 中添加了 `.row-label`，在每一行左侧显示“设备 1"、“设备 2"等，方便用户区分。
4.  **移动端适配**：添加了简单的媒体查询，在手机屏幕上会自动变为卡片式布局，隐藏表头，体验更好。
5.  **数据提交**：保留了 JavaScript 的提交拦截逻辑，依然会将三个设备的数据打包成一个 JSON 数组发送给服务器（模拟）。


# 避免使用javascript

这是一个完全不含 JavaScript 的纯 HTML/CSS 静态网页版本。

### 主要特点：
1.  **纯 HTML/CSS**：没有任何 `<script>` 标签或事件处理函数。
2.  **静态行**：默认硬编码了 3 行设备输入框（如需更多，需手动复制 HTML 代码）。
3.  **标准提交**：使用传统的 `<form>`  POST 提交方式，点击提交后页面会刷新/跳转。
4.  **数组命名**：输入框的 `name` 属性使用了 `[]` 后缀（例如 `name="name[]"`），这样服务器接收到的数据会自动成为数组，方便后端处理多行数据。
5.  **HTML5 验证**：利用 `required` 属性进行基本的必填项验证，无需 JS。

### 代码 (index.html)

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>设备信息录入 (纯静态版)</title>
    <style>
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background-color: #f4f4f9;
            padding: 20px;
            max-width: 900px;
            margin: 0 auto;
        }

        h1 {
            text-align: center;
            color: #333;
        }

        .form-container {
            background: white;
            padding: 30px;
            border-radius: 8px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }

        .form-description {
            margin-bottom: 20px;
            color: #666;
            font-size: 14px;
            border-left: 4px solid #007bff;
            padding-left: 10px;
        }

        /* 表头样式 */
        .header-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr;
            gap: 15px;
            margin-bottom: 15px;
            font-weight: bold;
            color: #555;
            border-bottom: 2px solid #eee;
            padding-bottom: 10px;
        }

        /* 输入行样式 */
        .input-row {
            display: grid;
            grid-template-columns: 2fr 2fr 2fr 2fr;
            gap: 15px;
            margin-bottom: 15px;
            align-items: center;
        }

        /* 行号标签 */
        .row-label {
            position: absolute;
            left: -40px;
            color: #999;
            font-weight: bold;
            width: 30px;
            text-align: right;
        }
        
        .input-wrapper {
            position: relative;
        }

        input {
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 4px;
            width: 100%;
            box-sizing: border-box;
            background-color: #fafafa;
            transition: border-color 0.2s;
        }

        /* 利用 CSS 伪类实现简单的交互反馈，无需 JS */
        input:focus {
            border-color: #007bff;
            background-color: #fff;
            outline: none;
            box-shadow: 0 0 0 3px rgba(0,123,255,0.1);
        }

        /* 必填项验证失败时的样式 (当用户尝试提交但字段为空且被聚焦时) */
        input:invalid:not(:placeholder-shown) {
            border-color: #dc3545;
            background-color: #fff5f5;
        }

        /* 提交按钮样式 */
        .btn-submit {
            background-color: #007bff;
            color: white;
            border: none;
            padding: 12px 20px;
            border-radius: 4px;
            cursor: pointer;
            font-size: 16px;
            width: 100%;
            margin-top: 20px;
            transition: background 0.2s;
        }

        .btn-submit:hover {
            background-color: #0069d9;
        }

        /* 响应式调整 */
        @media (max-width: 600px) {
            .header-row {
                display: none;
            }
            .input-row {
                grid-template-columns: 1fr;
                border: 1px solid #eee;
                padding: 15px;
                border-radius: 4px;
                margin-bottom: 15px;
            }
            .row-label {
                position: static;
                text-align: left;
                margin-bottom: 5px;
                display: block;
                color: #007bff;
            }
        }
    </style>
</head>
<body>

    <h1>设备信息录入</h1>

    <div class="form-container">
        <p class="form-description">
            提示：这是一个纯静态表单。请填写下方设备信息，点击提交后页面将刷新并将数据发送至服务器。
        </p>
        
        <!-- 
            action: 替换为你的服务器接收地址，例如 "/api/devices"
            method: 使用 POST 方法提交数据 
        -->
        <form action="/submit-devices" method="POST">
            
            <!-- 表头 -->
            <div class="header-row">
                <div>设备名称 (Name)</div>
                <div>设备编码 (Code)</div>
                <div>设备类型 (Type)</div>
                <div>所属分组 (Group)</div>
            </div>

            <!-- 设备 1 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 1</span>
                    <!-- name 使用 [] 表示数组，后端会收到一个列表 -->
                    <input type="text" name="name[]" placeholder="输入名称" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code[]" placeholder="输入编码" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type[]" placeholder="输入类型">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group[]" placeholder="输入分组">
                </div>
            </div>

            <!-- 设备 2 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 2</span>
                    <input type="text" name="name[]" placeholder="输入名称" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code[]" placeholder="输入编码" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type[]" placeholder="输入类型">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group[]" placeholder="输入分组">
                </div>
            </div>

            <!-- 设备 3 -->
            <div class="input-row">
                <div class="input-wrapper">
                    <span class="row-label">设备 3</span>
                    <input type="text" name="name[]" placeholder="输入名称" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="code[]" placeholder="输入编码" required>
                </div>
                <div class="input-wrapper">
                    <input type="text" name="type[]" placeholder="输入类型">
                </div>
                <div class="input-wrapper">
                    <input type="text" name="group[]" placeholder="输入分组">
                </div>
            </div>

            <!-- 提交按钮 -->
            <button type="submit" class="btn-submit">提交所有设备信息</button>
        </form>
    </div>

</body>
</html>
```

### 后端接收数据说明

由于没有 JavaScript 打包 JSON，数据将以标准的 `application/x-www-form-urlencoded` 或 `multipart/form-data` 格式提交。后端接收到的数据结构大致如下（以 Python/Flask 或 PHP 为例）：

**提交的数据流：**
```text
name[]=设备 A&code[]=CODE001&type[]=TypeA&group[]=Group1&
name[]=设备 B&code[]=CODE002&type[]=TypeB&group[]=Group2&
...
```

**后端处理逻辑示例 (Python Flask):**
```python
@app.route('/submit-devices', methods=['POST'])
def submit_devices():
    # 获取所有名为 name[] 的字段，形成一个列表
    names = request.form.getlist('name[]')
    codes = request.form.getlist('code[]')
    types = request.form.getlist('type[]')
    groups = request.form.getlist('group[]')
    
    devices = []
    # 通过索引将各个列表组合成对象
    for i in range(len(names)):
        devices.append({
            'name': names[i],
            'code': codes[i],
            'type': types[i],
            'group': groups[i]
        })
    
    # 处理 devices 列表...
    return "提交成功"
```

### 注意事项
1.  **修改 Action**：请将 `<form action="/submit-devices" ...>` 中的 `/submit-devices` 替换为你实际的服务器接口地址。
2.  **页面刷新**：点击提交后，浏览器会跳转到 action 指定的页面。如果需要提交后留在当前页，通常需要后端重定向回当前页面，或者必须使用 JavaScript (AJAX)，但本方案严格遵守了无 JS 的要求。
3.  **增加行数**：如果需要支持更多设备，只需在 HTML 中复制粘贴 `<div class="input-row">...</div>` 代码块即可。

