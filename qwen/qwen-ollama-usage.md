# 介绍一下 Ollama 的用法

# Ollama 使用指南

Ollama 是一个开源工具，让你可以在本地轻松运行和管理大型语言模型（LLM），支持 macOS、Windows、Linux 和 Docker。

---

## 🔧 安装方法

### macOS / Linux
```bash
curl -fsSL https://ollama.com/install.sh | sh
```

### Windows (PowerShell)
```powershell
irm https://ollama.com/install.ps1 | iex
```

### Docker
```bash
docker run -d -v ollama:/root/.ollama -p 11434:11434 --name ollama ollama/ollama
```

> 安装完成后，Ollama 服务会自动在后台运行，监听 `http://localhost:11434`

---

## 🚀 基础命令

### 1. 运行模型
```bash
# 下载并运行模型（首次自动下载）
ollama run llama3.2
ollama run qwen2.5
ollama run gemma3

# 运行后进入交互对话模式，输入 /? 查看帮助
```

### 2. 模型管理
```bash
# 查看已下载的模型
ollama list

# 下载模型（不运行）
ollama pull llama3.2

# 删除模型
ollama rm llama3.2

# 复制/重命名模型
ollama cp llama3.2 my-llama

# 查看模型详细信息
ollama show llama3.2
ollama show --modelfile llama3.2  # 查看 Modelfile
```

### 3. 高级运行选项
```bash
# 指定参数运行
ollama run llama3.2 --temperature 0.7 --num-predict 512

# 从文件读取提示词
ollama run llama3.2 < prompt.txt

# 管道输入
echo "解释量子力学" | ollama run llama3.2
```

---

## 🎯 启动集成应用

Ollama 支持一键启动与开发工具集成：

```bash
# 启动 Claude Code（代码助手）
ollama launch claude

# 启动 OpenClaw（多平台 AI 助手）
ollama launch openclaw

# 启动 Codex 等（需先安装对应插件）
ollama launch codex
```

---

## ⚙️ REST API 使用

Ollama 提供本地 HTTP API，方便集成到其他应用：

### Chat 接口
```bash
curl http://localhost:11434/api/chat -d '{
  "model": "llama3.2",
  "messages": [
    {"role": "user", "content": "你好，请介绍一下自己"}
  ],
  "stream": false
}'
```

### Generate 接口（纯文本生成）
```bash
curl http://localhost:11434/api/generate -d '{
  "model": "llama3.2",
  "prompt": "写一首关于春天的诗",
  "stream": false
}'
```

### 流式响应
设置 `"stream": true` 可获取逐词输出的响应，适合实时显示。

> 📚 完整 API 文档：https://github.com/ollama/ollama/blob/main/docs/api.md

---

## 🛠️ 自定义模型（Modelfile）

通过 Modelfile 可以创建定制化的模型配置：

### 示例 Modelfile
```modelfile
FROM llama3.2

# 设置系统提示词
SYSTEM """
你是一个专业的 Python 编程助手，回答简洁准确，优先提供可运行代码。
"""

# 调整参数
PARAMETER temperature 0.3
PARAMETER num_ctx 4096

# 设置元数据
LICENSE MIT
```

### 构建和使用
```bash
# 创建自定义模型
ollama create my-coder -f Modelfile

# 运行自定义模型
ollama run my-coder
```

---

## 📦 热门模型推荐

| 模型 | 用途 | 命令 |
|------|------|------|
| `llama3.2` | 通用对话，平衡性能 | `ollama run llama3.2` |
| `qwen2.5` | 中文友好，多语言 | `ollama run qwen2.5` |
| `gemma3` | 轻量高效，Google 出品 | `ollama run gemma3` |
| `codellama` | 代码生成专用 | `ollama run codellama` |
| `mistral` | 推理能力强 | `ollama run mistral` |

> 完整模型库：https://ollama.com/library

---

## 🔍 实用技巧

1. **后台服务管理**
   ```bash
   # 查看服务状态
   ollama serve  # 手动启动（通常无需）
   
   # macOS 重启服务
   launchctl stop com.ollama.ollama
   launchctl start com.ollama.ollama
   ```

2. **环境变量配置**
   ```bash
   # 修改模型存储路径
   export OLLAMA_MODELS=/path/to/models
   
   # 设置 API 监听地址
   export OLLAMA_HOST=0.0.0.0:11434
   ```

3. **与前端/应用集成**
   - 使用 Open WebUI、Continue.dev、LangChain 等工具连接 Ollama
   - 通过 API 将 Ollama 嵌入自己的应用

4. **多模型切换**
   ```bash
   # 在对话中按 Ctrl+D 退出当前模型
   # 然后运行另一个：ollama run qwen2.5
   ```

---

## ❓ 常见问题

- **模型下载慢？** 检查网络，或考虑使用镜像/代理
- **显存不足？** 选择更小参数模型（如 7B 而非 70B），或设置 `OLLAMA_NUM_GPU=0` 使用 CPU
- **如何更新 Ollama？** 重新运行安装脚本即可自动升级

---

> 📚 官方资源：
> - 官网：https://ollama.com
> - GitHub：https://github.com/ollama/ollama
> - 文档：https://github.com/ollama/ollama/tree/main/docs

如果你有特定的使用场景（如本地开发、私有部署、API 集成等），可以告诉我，我帮你提供更针对性的建议！🚀


# 在ollama中，如何指定/上传文件进行处理

# Ollama 文件处理指南

Ollama 本身**不直接支持文件上传功能**，但有多种方式可以将文件内容传递给模型进行处理：

---

## 📄 文本文件处理方法

### 方法 1：命令行管道（最常用）
```bash
# 将文件内容通过管道传递给模型
cat document.txt | ollama run llama3.2

# 或使用重定向
ollama run llama3.2 < document.txt

# 处理特定行
head -n 50 report.md | ollama run llama3.2 "总结以下内容："
```

### 方法 2：在交互模式中粘贴
```bash
ollama run llama3.2
>>> 请分析以下内容：
>>> [粘贴文件内容]
>>> [按 Ctrl+D 结束输入]
```

### 方法 3：使用 API 发送文件内容
```bash
# 读取文件并作为 prompt 发送
curl http://localhost:11434/api/generate -d '{
  "model": "llama3.2",
  "prompt": "'"$(cat document.txt)"'",
  "stream": false
}'
```

### 方法 4：Python 脚本读取文件
```python
import requests

with open("document.txt", "r", encoding="utf-8") as f:
    content = f.read()

response = requests.post("http://localhost:11434/api/generate", json={
    "model": "llama3.2",
    "prompt": f"请总结：{content}",
    "stream": False
})
print(response.json()["response"])
```

---

## 🖼️ 图片/多模态文件处理

Ollama 支持**多模态模型**（如 `llava`、`bakllava`、`moondream`）处理图片：

### 命令行方式（需 base64 编码）
```bash
# 将图片转为 base64 后发送
ollama run llava "描述这张图片" --images image.jpg
```

> ⚠️ 注意：当前 CLI 对图片支持有限，推荐使用 API

### API 方式处理图片
```bash
# 图片转 base64（Linux/Mac）
IMAGE_BASE64=$(base64 -w 0 photo.jpg)

curl http://localhost:11434/api/generate -d '{
  "model": "llava",
  "prompt": "这张图片里有什么？",
  "images": ["'"$IMAGE_BASE64"'"],
  "stream": false
}'
```

### Python 处理图片示例
```python
import base64, requests

with open("photo.jpg", "rb") as f:
    image_b64 = base64.b64encode(f.read()).decode()

response = requests.post("http://localhost:11434/api/generate", json={
    "model": "llava",
    "prompt": "描述这张图片",
    "images": [image_b64],
    "stream": False
})
print(response.json()["response"])
```

---

## 📚 长文档/RAG 处理方案

对于长文档，Ollama 本身**不提供内置 RAG**，但可以结合其他工具：

### 方案 1：手动分块 + 上下文管理
```bash
# 分割大文件后逐段处理
split -l 1000 large_doc.txt chunk_
for file in chunk_*; do
    ollama run llama3.2 "分析：$(cat $file)"
done
```

### 方案 2：使用 RAG 框架集成
| 工具 | 特点 | 链接 |
|------|------|------|
| **LangChain + Ollama** | 成熟的 RAG 框架，支持向量存储 | https://python.langchain.com |
| **LlamaIndex + Ollama** | 专为 RAG 设计，文档索引强大 | https://docs.llamaindex.ai |
| **PrivateGPT** | 本地知识库问答，开箱即用 | https://privategpt.dev |
| **AnythingLLM** | 图形界面，支持多文档上传 | https://anythingllm.com |

### LangChain 简单示例
```python
from langchain_community.chat_models import ChatOllama
from langchain_community.document_loaders import TextLoader
from langchain_text_splitters import CharacterTextSplitter

# 加载并分割文档
loader = TextLoader("document.txt")
docs = loader.load()
splitter = CharacterTextSplitter(chunk_size=1000, chunk_overlap=200)
chunks = splitter.split_documents(docs)

# 创建聊天模型
llm = ChatOllama(model="llama3.2")

# 构建 RAG 查询（简化版）
context = "\n\n".join([d.page_content for d in chunks[:3]])
query = "文档中关于 XX 的结论是什么？"
response = llm.invoke(f"基于以下上下文回答问题：\n{context}\n\n问题：{query}")
print(response.content)
```

---

## 🔧 实用技巧

### 1. 处理编码问题
```bash
# 指定 UTF-8 编码读取文件
iconv -f gbk -t utf-8 file.txt | ollama run llama3.2
```

### 2. 限制输入长度（避免超出上下文）
```bash
# 只取前 4000 字符（约 1000 tokens）
head -c 4000 document.txt | ollama run llama3.2
```

### 3. 批量处理多个文件
```bash
for file in *.txt; do
    echo "=== 处理 $file ==="
    ollama run llama3.2 "总结 $file 的内容：$(cat $file)"
done
```

### 4. 使用环境变量配置上下文长度
```bash
# 增加上下文窗口（需模型支持）
export OLLAMA_NUM_CTX=8192
ollama run llama3.2
```

---

## ⚠️ 注意事项

1. **上下文长度限制**：大多数模型上下文为 4K-32K tokens，超长文件需分块
2. **图片大小**：多模态模型通常限制图片分辨率，建议压缩到 1024px 以内
3. **二进制文件**：Ollama 仅处理文本/图片，PDF/Word 需先转换为纯文本
4. **隐私安全**：本地运行 Ollama 时文件不会上传云端，但使用 API 时注意网络配置

---

## 🔄 推荐的完整工作流

```bash
# 1. 转换 PDF 为文本（需安装 pdftotext）
pdftotext report.pdf - | tee report.txt | \

# 2. 预处理（去空行、限制长度）
sed '/^$/d' | head -c 8000 | \

# 3. 发送给 Ollama
ollama run llama3.2 "请提取这份报告的关键结论和数据来源："
```

如果你需要处理特定格式的文件（如 PDF、Word、Excel），建议先用 `pandoc`、`python-docx`、`pandas` 等工具转换为纯文本，再交给 Ollama 处理。

有具体的文件类型或使用场景吗？我可以提供更针对性的方案！🚀

# 在windows下，ollama安装的模型存放在哪里？如果我的磁盘空间不够了，想将模型移动到更大空间的磁盘上，应该如何做？

# Ollama Windows 模型存储与迁移指南

## 📁 默认模型存储位置

在 Windows 系统中，Ollama 下载的模型默认存放在：

```
C:\Users\<你的用户名>\.ollama\models
```

### 快速访问方法：
```powershell
# 方法1：直接在文件资源管理器地址栏输入
%USERPROFILE%\.ollama\models

# 方法2：PowerShell 打开
explorer "$env:USERPROFILE\.ollama\models"

# 方法3：CMD 打开
start %USERPROFILE%\.ollama\models
```

### 目录结构说明：
```
.ollama/
├── models/          # 模型文件存放位置（占用空间最大）
│   ├── blobs/       # 模型权重文件（sha256 命名的大文件）
│   └── manifests/   # 模型配置清单（小文件）
├── history/         # 对话历史记录
└── config.json      # 配置文件
```

> 💡 单个 7B 模型约占用 4-8GB 空间，70B 模型可能超过 40GB

---

## 🔄 迁移模型到其他磁盘（详细步骤）

假设你要将模型从 `C:` 盘迁移到 `D:\ollama-models`

### 方法一：使用环境变量（推荐 ✅）

#### 步骤 1：停止 Ollama 服务
```powershell
# 任务栏右键 Ollama 图标 → Quit
# 或使用 PowerShell 强制结束
Stop-Process -Name "ollama" -Force -ErrorAction SilentlyContinue
```

#### 步骤 2：移动模型文件
```powershell
# 创建新目录
New-Item -ItemType Directory -Path "D:\ollama-models" -Force

# 移动整个 .ollama 目录（或只移动 models 子目录）
Move-Item -Path "$env:USERPROFILE\.ollama\models" -Destination "D:\ollama-models\" -Force

# 确认移动成功
Get-ChildItem "D:\ollama-models\models"
```

#### 步骤 3：设置环境变量 `OLLAMA_MODELS`

##### 🔹 临时生效（当前 PowerShell 会话）
```powershell
$env:OLLAMA_MODELS="D:\ollama-models"
ollama serve  # 手动启动测试
```

##### 🔹 永久生效（系统级）

**方式 A：通过图形界面设置**
1. 按 `Win + R`，输入 `sysdm.cpl` 回车
2. 点击「高级」→「环境变量」
3. 在「用户变量」或「系统变量」中点击「新建」
4. 变量名：`OLLAMA_MODELS`
5. 变量值：`D:\ollama-models`
6. 确定保存，重启电脑或重启 Ollama

**方式 B：PowerShell 命令设置（需管理员权限）**
```powershell
# 设置用户级环境变量（推荐）
[Environment]::SetEnvironmentVariable("OLLAMA_MODELS", "D:\ollama-models", "User")

# 或设置系统级环境变量（需管理员）
# [Environment]::SetEnvironmentVariable("OLLAMA_MODELS", "D:\ollama-models", "Machine")
```

#### 步骤 4：验证配置
```powershell
# 重启 Ollama（从开始菜单启动）
# 然后检查模型是否被正确识别
ollama list

# 查看环境变量是否生效
echo $env:OLLAMA_MODELS
```

---

### 方法二：使用符号链接（高级用户）

如果不想修改环境变量，可以创建目录联结（Junction）：

```powershell
# 1. 停止 Ollama
Stop-Process -Name "ollama" -Force -ErrorAction SilentlyContinue

# 2. 移动 models 目录
Move-Item "$env:USERPROFILE\.ollama\models" "D:\ollama-models\"

# 3. 创建符号链接（需要管理员权限的 PowerShell）
cmd /c mklink /J "$env:USERPROFILE\.ollama\models" "D:\ollama-models\models"

# 4. 验证链接
dir "$env:USERPROFILE\.ollama\models"
```

> ✅ 优点：对 Ollama 透明，无需配置环境变量  
> ⚠️ 注意：删除链接时用 `rmdir` 而非 `del`，避免误删原文件

---

## 🔍 验证迁移是否成功

```powershell
# 1. 检查环境变量
echo $env:OLLAMA_MODELS
# 应输出：D:\ollama-models

# 2. 列出模型（应正常显示）
ollama list

# 3. 运行模型测试
ollama run llama3.2 "hello"

# 4. 查看新位置是否有文件写入
Get-ChildItem "D:\ollama-models\models" -Recurse | Measure-Object
```

---

## ⚙️ 其他相关配置（可选）

| 环境变量 | 作用 | 示例值 |
|---------|------|--------|
| `OLLAMA_MODELS` | 模型存储路径 | `D:\ollama-models` |
| `OLLAMA_HOST` | API 监听地址 | `0.0.0.0:11434` |
| `OLLAMA_ORIGINS` | CORS 允许的来源 | `*` 或 `http://localhost:3000` |
| `OLLAMA_NUM_GPU` | 使用 GPU 层数 | `0`=纯CPU, `-1`=全部GPU |

### 设置多个环境变量（PowerShell）：
```powershell
[Environment]::SetEnvironmentVariable("OLLAMA_MODELS", "D:\ollama-models", "User")
[Environment]::SetEnvironmentVariable("OLLAMA_NUM_GPU", "40", "User")
```

---

## ❓ 常见问题解答

### Q：迁移后 `ollama list` 显示为空？
**A**：可能原因：
1. 环境变量未生效 → 重启终端或电脑
2. 路径权限问题 → 确保当前用户有读写权限
3. 只移动了 blobs 没移动 manifests → 建议移动整个 `models` 目录

### Q：如何确认模型实际占用空间？
```powershell
# 计算 models 目录总大小
Get-ChildItem "$env:OLLAMA_MODELS\models" -Recurse | 
  Measure-Object -Property Length -Sum | 
  Select-Object @{N="GB";E={[math]::Round($_.Sum / 1GB, 2)}}
```

### Q：迁移过程中可以继续使用 Ollama 吗？
**A**：❌ 不可以！迁移前务必完全退出 Ollama，否则文件可能被锁定导致移动失败或数据损坏。

### Q：D 盘空间也不够了怎么办？
**A**：可以：
1. 删除不用的模型：`ollama rm 模型名`
2. 使用更小参数的模型（如 7B 代替 70B）
3. 将 `OLLAMA_MODELS` 指向外接硬盘或 NAS（需保持连接）

---

## 🧹 附加：清理无用模型释放空间

```powershell
# 查看所有模型及大小
ollama list

# 删除指定模型
ollama rm llama2
ollama rm mistral

# 批量删除（谨慎操作！）
"model1", "model2" | ForEach-Object { ollama rm $_ }
```

> 💡 提示：删除模型后，`blobs` 中未被其他模型引用的文件不会自动清理，如需彻底清理可手动检查（但建议保留，避免误删共享层）

---

✅ **总结**：  
1. Windows 模型默认在 `%USERPROFILE%\.ollama\models`  
2. 通过设置 `OLLAMA_MODELS` 环境变量可轻松迁移到大容量磁盘  
3. 迁移前务必备份 + 停止服务，避免数据损坏  

如果迁移过程中遇到具体问题，欢迎告诉我你的 Windows 版本和操作步骤，我帮你进一步排查！🛠️
