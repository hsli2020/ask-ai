# javascript中提供了alert和confirm等内置的对话框，但比较丑陋，请实现自定义的对话框替代alert和confirm的功能，但要漂亮一些

我来为你实现一个美观的自定义对话框系统，替代原生的 `alert` 和 `confirm`。

```html
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>自定义对话框</title>
    <style>
        /* 遮罩层样式 */
        .custom-dialog-overlay {
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background-color: rgba(0, 0, 0, 0.5);
            display: flex;
            justify-content: center;
            align-items: center;
            z-index: 9999;
            opacity: 0;
            visibility: hidden;
            transition: opacity 0.3s ease, visibility 0.3s ease;
        }

        .custom-dialog-overlay.show {
            opacity: 1;
            visibility: visible;
        }

        /* 对话框主体样式 */
        .custom-dialog {
            background: white;
            border-radius: 12px;
            box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
            max-width: 400px;
            width: 90%;
            transform: scale(0.8);
            transition: transform 0.3s ease;
            overflow: hidden;
        }

        .custom-dialog-overlay.show .custom-dialog {
            transform: scale(1);
        }

        /* 对话框头部 */
        .custom-dialog-header {
            padding: 20px 24px 12px;
            text-align: center;
        }

        .custom-dialog-title {
            margin: 0;
            font-size: 18px;
            font-weight: 600;
            color: #333;
        }

        /* 对话框内容 */
        .custom-dialog-content {
            padding: 16px 24px;
            text-align: center;
            min-height: 60px;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .custom-dialog-message {
            margin: 0;
            font-size: 15px;
            color: #555;
            line-height: 1.5;
        }

        /* 对话框按钮区域 */
        .custom-dialog-footer {
            padding: 16px 24px;
            display: flex;
            justify-content: center;
            gap: 12px;
            border-top: 1px solid #eee;
        }

        /* 按钮样式 */
        .custom-dialog-btn {
            padding: 10px 24px;
            border: none;
            border-radius: 6px;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s ease;
            min-width: 80px;
        }

        .custom-dialog-btn:hover {
            transform: translateY(-1px);
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
        }

        .custom-dialog-btn:active {
            transform: translateY(0);
        }

        /* 确认按钮 - 主要按钮 */
        .custom-dialog-btn.confirm {
            background: #007bff;
            color: white;
        }

        .custom-dialog-btn.confirm:hover {
            background: #0069d9;
        }

        /* 取消按钮 - 次要按钮 */
        .custom-dialog-btn.cancel {
            background: #f8f9fa;
            color: #6c757d;
            border: 1px solid #dee2e6;
        }

        .custom-dialog-btn.cancel:hover {
            background: #e9ecef;
            border-color: #adb5bd;
        }

        /* 仅有一个按钮时的样式 */
        .custom-dialog-footer.single-btn {
            justify-content: center;
        }

        .custom-dialog-btn.single {
            background: #28a745;
            color: white;
        }

        .custom-dialog-btn.single:hover {
            background: #218838;
        }

        /* 响应式设计 */
        @media (max-width: 480px) {
            .custom-dialog {
                width: 95%;
                margin: 20px;
            }
            
            .custom-dialog-header {
                padding: 16px 20px 10px;
            }
            
            .custom-dialog-content {
                padding: 12px 20px;
            }
            
            .custom-dialog-footer {
                padding: 12px 20px;
                flex-direction: column;
                gap: 8px;
            }
            
            .custom-dialog-btn {
                width: 100%;
                padding: 12px;
            }
        }
    </style>
</head>
<body>
    <!-- 测试按钮 -->
    <div style="padding: 50px; text-align: center;">
        <button onclick="showCustomAlert('这是一个自定义的alert对话框！')" style="margin: 10px; padding: 10px 20px;">显示Alert</button>
        <button onclick="showCustomConfirm('确定要删除这个项目吗？')" style="margin: 10px; padding: 10px 20px;">显示Confirm</button>
    </div>

    <script>
        // 全局变量存储当前对话框的Promise resolve函数
        let currentResolve = null;

        /**
         * 创建对话框DOM元素
         */
        function createDialog() {
            const overlay = document.createElement('div');
            overlay.className = 'custom-dialog-overlay';
            
            const dialog = document.createElement('div');
            dialog.className = 'custom-dialog';
            
            const header = document.createElement('div');
            header.className = 'custom-dialog-header';
            const title = document.createElement('h3');
            title.className = 'custom-dialog-title';
            header.appendChild(title);
            
            const content = document.createElement('div');
            content.className = 'custom-dialog-content';
            const message = document.createElement('p');
            message.className = 'custom-dialog-message';
            content.appendChild(message);
            
            const footer = document.createElement('div');
            footer.className = 'custom-dialog-footer';
            
            dialog.appendChild(header);
            dialog.appendChild(content);
            dialog.appendChild(footer);
            overlay.appendChild(dialog);
            
            document.body.appendChild(overlay);
            
            return { overlay, dialog, title, message, footer };
        }

        /**
         * 显示自定义Alert对话框
         * @param {string} message - 显示的消息
         * @param {string} title - 对话框标题（可选）
         * @returns {Promise<void>} - Promise对象
         */
        function showCustomAlert(message, title = '提示') {
            return new Promise((resolve) => {
                const { overlay, title: titleEl, message: messageEl, footer } = createDialog();
                
                // 设置内容
                titleEl.textContent = title;
                messageEl.textContent = message;
                
                // 创建确定按钮
                const okBtn = document.createElement('button');
                okBtn.className = 'custom-dialog-btn single';
                okBtn.textContent = '确定';
                okBtn.onclick = () => {
                    closeDialog(overlay, resolve);
                };
                
                footer.appendChild(okBtn);
                footer.className = 'custom-dialog-footer single-btn';
                
                // 显示对话框
                overlay.classList.add('show');
                
                // 绑定ESC键关闭
                const handleEsc = (e) => {
                    if (e.key === 'Escape') {
                        closeDialog(overlay, resolve);
                        document.removeEventListener('keydown', handleEsc);
                    }
                };
                document.addEventListener('keydown', handleEsc);
            });
        }

        /**
         * 显示自定义Confirm对话框
         * @param {string} message - 显示的消息
         * @param {string} title - 对话框标题（可选）
         * @returns {Promise<boolean>} - Promise对象，resolve(true)表示确认，resolve(false)表示取消
         */
        function showCustomConfirm(message, title = '确认') {
            return new Promise((resolve) => {
                const { overlay, title: titleEl, message: messageEl, footer } = createDialog();
                
                // 设置内容
                titleEl.textContent = title;
                messageEl.textContent = message;
                
                // 创建取消按钮
                const cancelBtn = document.createElement('button');
                cancelBtn.className = 'custom-dialog-btn cancel';
                cancelBtn.textContent = '取消';
                cancelBtn.onclick = () => {
                    closeDialog(overlay, () => resolve(false));
                };
                
                // 创建确定按钮
                const confirmBtn = document.createElement('button');
                confirmBtn.className = 'custom-dialog-btn confirm';
                confirmBtn.textContent = '确定';
                confirmBtn.onclick = () => {
                    closeDialog(overlay, () => resolve(true));
                };
                
                footer.appendChild(cancelBtn);
                footer.appendChild(confirmBtn);
                
                // 显示对话框
                overlay.classList.add('show');
                
                // 绑定ESC键关闭（按ESC等同于取消）
                const handleEsc = (e) => {
                    if (e.key === 'Escape') {
                        closeDialog(overlay, () => resolve(false));
                        document.removeEventListener('keydown', handleEsc);
                    }
                };
                document.addEventListener('keydown', handleEsc);
                
                // 默认聚焦确定按钮
                setTimeout(() => confirmBtn.focus(), 100);
            });
        }

        /**
         * 关闭对话框
         * @param {HTMLElement} overlay - 遮罩层元素
         * @param {Function} callback - 关闭后的回调函数
         */
        function closeDialog(overlay, callback) {
            // 移除遮罩层的显示类
            overlay.classList.remove('show');
            
            // 延迟移除DOM元素，等待动画完成
            setTimeout(() => {
                if (overlay.parentNode) {
                    overlay.parentNode.removeChild(overlay);
                }
                if (callback) {
                    callback();
                }
            }, 300);
        }

        // 为方便使用，可以覆盖原生方法（可选）
        // window.alert = showCustomAlert;
        // window.confirm = showCustomConfirm;

        // 测试用的确认回调函数
        function handleConfirm(result) {
            if (result) {
                console.log('用户点击了确定');
                showCustomAlert('操作已确认！');
            } else {
                console.log('用户点击了取消');
            }
        }

        // 使用示例（异步/await方式）
        async function exampleUsage() {
            // 显示alert
            await showCustomAlert('操作成功完成！');
            
            // 显示confirm
            const result = await showCustomConfirm('确定要继续吗？');
            if (result) {
                await showCustomAlert('用户选择了确定');
            } else {
                await showCustomAlert('用户选择了取消');
            }
        }
    </script>
</body>
</html>
```

这个自定义对话框系统具有以下特点：

## 主要功能

1. **`showCustomAlert(message, title)`** - 替代原生 `alert`
   - 返回 Promise，可以使用 async/await
   - 支持自定义标题

2. **`showCustomConfirm(message, title)`** - 替代原生 `confirm`
   - 返回 Promise，resolve 为 boolean 值
   - true 表示确认，false 表示取消

## 美观特性

- **现代化设计**：圆角、阴影、平滑动画
- **响应式布局**：在移动设备上自动调整为垂直按钮布局
- **交互反馈**：按钮悬停和点击效果
- **键盘支持**：ESC 键可以关闭对话框
- **焦点管理**：自动聚焦到主要按钮

## 使用方式

### 方式1：Promise 链式调用
```javascript
showCustomAlert('操作成功！').then(() => {
    console.log('用户关闭了对话框');
});

showCustomConfirm('确定删除吗？').then(result => {
    if (result) {
        // 执行删除操作
    }
});
```

### 方式2：async/await
```javascript
async function handleDelete() {
    const confirmed = await showCustomConfirm('确定要删除这个项目吗？');
    if (confirmed) {
        await showCustomAlert('项目已删除！');
        // 执行删除逻辑
    }
}
```

### 方式3：覆盖原生方法（可选）
```javascript
// 如果你希望全局替换原生方法
window.alert = showCustomAlert;
window.confirm = showCustomConfirm;
```

## 自定义选项

你可以轻松修改 CSS 样式来匹配你的网站主题：
- 修改颜色方案（按钮颜色、文字颜色等）
- 调整尺寸和间距
- 更改动画效果
- 添加图标支持

这个实现完全独立，不依赖任何外部库，可以直接集成到任何项目中使用。
