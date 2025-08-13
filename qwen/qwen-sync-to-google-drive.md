# 我想编写一个程序，自动将我的电脑上的某些文件周期性地同步到google drive上的某个文件夹中，需要如何开始呢？

要实现“将本地文件周期性同步到 Google Drive 指定文件夹”，你可以通过 Google Drive API + 定时任务（如 cron 或 Windows 任务计划）来实现。以下是详细的步骤指南：

---

## ✅ 第一步：准备工作

### 1. 创建 Google Cloud 项目并启用 Drive API
- 访问 [Google Cloud Console](https://console.cloud.google.com/)
- 创建新项目（或使用现有项目）
- 启用 **Google Drive API**
  - 导航到 “API和服务” > “库” > 搜索 “Google Drive API” > 启用

### 2. 创建 OAuth 2.0 凭据
- 在 “API和服务” > “凭据” > “创建凭据” > 选择 “OAuth 2.0 客户端 ID”
- 应用类型：选择“桌面应用”（Desktop Application）
- 下载生成的 `credentials.json` 文件，保存到你的项目目录中

> ⚠️ 第一次运行程序时，会弹出浏览器让你授权，之后会生成 `token.json` 用于后续自动认证。

---

## ✅ 第二步：安装依赖

推荐使用 Python（简单易用），安装 Google 客户端库：

```bash
pip install --upgrade google-api-python-client google-auth-httplib2 google-auth-oauthlib
```

---

## ✅ 第三步：编写同步脚本（Python 示例）

下面是一个基础脚本，它会：

- 认证 Google Drive
- 上传本地文件夹中所有文件到指定 Drive 文件夹（按文件名去重更新）
- 支持创建目标文件夹（如果不存在）

📁 文件名：`sync_to_drive.py`

```python
import os
import pickle
from google.auth.transport.requests import Request
from google_auth_oauthlib.flow import InstalledAppFlow
from googleapiclient.discovery import build
from googleapiclient.http import MediaFileUpload

# 权限范围（读写文件）
SCOPES = ['https://www.googleapis.com/auth/drive']

def authenticate():
    creds = None
    if os.path.exists('token.pickle'):
        with open('token.pickle', 'rb') as token:
            creds = pickle.load(token)
    if not creds or not creds.valid:
        if creds and creds.expired and creds.refresh_token:
            creds.refresh(Request())
        else:
            flow = InstalledAppFlow.from_client_secrets_file('credentials.json', SCOPES)
            creds = flow.run_local_server(port=0)
        with open('token.pickle', 'wb') as token:
            pickle.dump(creds, token)
    return build('drive', 'v3', credentials=creds)

def get_or_create_folder(service, folder_name, parent_id=None):
    query = f"name='{folder_name}' and mimeType='application/vnd.google-apps.folder'"
    if parent_id:
        query += f" and '{parent_id}' in parents"
    results = service.files().list(q=query, fields="files(id, name)").execute()
    items = results.get('files', [])
    if items:
        return items[0]['id']
    else:
        folder_metadata = {
            'name': folder_name,
            'mimeType': 'application/vnd.google-apps.folder'
        }
        if parent_id:
            folder_metadata['parents'] = [parent_id]
        folder = service.files().create(body=folder_metadata, fields='id').execute()
        return folder.get('id')

def upload_file(service, file_path, folder_id):
    file_name = os.path.basename(file_path)
    # 检查是否已存在同名文件
    query = f"name='{file_name}' and '{folder_id}' in parents and trashed=false"
    results = service.files().list(q=query, fields="files(id)").execute()
    existing_files = results.get('files', [])
    
    file_metadata = {'name': file_name, 'parents': [folder_id]}
    media = MediaFileUpload(file_path, resumable=True)
    
    if existing_files:
        # 更新现有文件
        file_id = existing_files[0]['id']
        service.files().update(fileId=file_id, media_body=media).execute()
        print(f"✅ Updated: {file_name}")
    else:
        # 上传新文件
        service.files().create(body=file_metadata, media_body=media, fields='id').execute()
        print(f"✅ Uploaded: {file_name}")

def sync_folder(local_folder, drive_folder_name):
    service = authenticate()
    # 获取或创建目标文件夹
    folder_id = get_or_create_folder(service, drive_folder_name)
    # 遍历本地文件夹
    for filename in os.listdir(local_folder):
        file_path = os.path.join(local_folder, filename)
        if os.path.isfile(file_path):
            upload_file(service, file_path, folder_id)

if __name__ == '__main__':
    LOCAL_FOLDER = r"/path/to/your/local/folder"  # 👈 修改为你的本地路径
    DRIVE_FOLDER_NAME = "MySyncFolder"           # 👈 修改为你想在Drive中创建的文件夹名
    sync_folder(LOCAL_FOLDER, DRIVE_FOLDER_NAME)
```

---

## ✅ 第四步：设置定时任务

### 🐧 Linux/macOS：使用 `cron`

编辑 crontab：

```bash
crontab -e
```

添加一行（例如每天凌晨2点执行）：

```bash
0 2 * * * cd /path/to/your/script && python3 sync_to_drive.py >> sync.log 2>&1
```

### 🪟 Windows：使用“任务计划程序”

1. 打开“任务计划程序”
2. 创建基本任务 → 设置触发器（如每天）
3. 操作 → 启动程序 → 选择 `python.exe`
4. 参数：填写脚本路径，如 `"C:\path\to\sync_to_drive.py"`
5. 起始于：填写脚本所在目录

---

## ✅ 第五步（可选）：增强功能

- ✅ **只同步新增或修改的文件**（比较修改时间或哈希值）
- ✅ **日志记录 & 错误通知**（邮件或 Telegram 通知）
- ✅ **排除某些文件**（如临时文件、隐藏文件）
- ✅ **双向同步**（更复杂，建议用 rclone 或 Insync 等工具）

---

## ✅ 替代方案（如果你不想写代码）

### 使用 `rclone`（推荐给非开发者）

1. 安装 rclone：https://rclone.org/
2. 配置 Google Drive：

```bash
rclone config
# 按提示添加 Google Drive remote
```

3. 同步命令：

```bash
rclone sync /local/folder remote:DriveFolderName --update --verbose
```

4. 加入定时任务（同上）

> ✅ 优点：稳定、支持增量同步、断点续传、双向同步、无需写代码

---

## 📌 注意事项

- 第一次运行需手动授权（浏览器弹窗）
- 保存好 `credentials.json` 和 `token.pickle`
- 不要将凭据文件上传到 GitHub！
- 免费版 Google Drive 有 API 配额限制（通常个人使用足够）

---

## ✅ 总结

| 步骤 | 内容 |
|------|------|
| 1️⃣ | 创建 Google Cloud 项目 + 启用 Drive API |
| 2️⃣ | 下载凭据 + 安装依赖 |
| 3️⃣ | 编写同步脚本（或使用 rclone） |
| 4️⃣ | 设置定时任务（cron / 任务计划） |
| 5️⃣ | 测试 + 日志监控 |

---

如果你希望我帮你定制脚本（比如指定路径、排除文件、日志邮件等），请提供：

- 本地文件夹路径
- Google Drive 目标文件夹名
- 同步频率
- 是否需要跳过某些文件类型

我可以为你生成完整可运行的代码 👍

祝你同步顺利！🚀
