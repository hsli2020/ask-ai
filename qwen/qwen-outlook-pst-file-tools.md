# 请推荐一些能够操作outlook的pst文件格式的工具

# 🔧 Outlook PST 文件操作工具推荐

PST（Personal Storage Table）是 Microsoft Outlook 用于存储邮件、联系人、日历等数据的专有文件格式。以下是按类别整理的实用工具推荐：

---

## 🆓 开源/免费工具（推荐优先尝试）

### 🔹 XstReader（跨平台查看器）
- **功能**：开源的 PST/OST 文件查看器，使用 C# 编写，仅需 .NET Framework 4，无需安装 Office 组件 [[56]]
- **亮点**：
  - 三栏式邮件浏览界面，支持搜索、导出邮件/附件/属性
  - 提供命令行工具 `XstExport` 批量导出为 HTML/RTF/TXT/CSV 格式
  - 支持加密/签名邮件查看（需证书）
- **获取**：[GitHub - Dijji/XstReader](https://github.com/Dijji/XstReader)

### 🔹 libpst（Linux/Unix 命令行工具）
- **功能**：C 语言库 + `readpst` 命令行工具，用于读取和转换 PST 文件 [[40]]
- **亮点**：
  - 可将 PST 转换为 mbox、MH 等标准邮件格式
  - 适合服务器端批量处理或脚本集成
- **安装**：`sudo apt install libpst-utils` 或从源码编译 [[39]]
- **获取**：[GitHub - pst-format/libpst](https://github.com/pst-format/libpst)

### 🔹 libpff / pypff（Python 开发库）
- **功能**：libpff 是访问 PST/OST 格式的 C 库，pypff 是其 Python 绑定 [[47]][[49]]
- **亮点**：
  - 支持程序化读取邮件、日历、联系人等对象
  - 适合开发自动化脚本或数据提取工具
- **安装**：`pip install libpff-python`
- **文档**：[PyPI - libpff-python](https://pypi.org/project/libpff-python/)

### 🔹 java-libpst（Java 开发库）
- **功能**：纯 Java 实现的 PST 读取库，无需外部依赖 [[44]]
- **适用**：Java 项目中集成 PST 解析功能
- **获取**：[GitHub - rjohnsondev/java-libpst](https://github.com/rjohnsondev/java-libpst)

### 🔹 Microsoft PST File Format SDK（官方开发工具）
- **功能**：微软官方发布的 C++ 跨平台库，提供对 PST 文件的底层读取访问 [[68]]
- **亮点**：
  - 附带 PST 数据结构查看工具，便于理解文件格式
  - 适合需要深度定制或合规审计的场景
- **获取**：[GitHub - enrondata/microsoft-pst-sdk](https://github.com/enrondata/microsoft-pst-sdk)

---

## 💼 商业/专业工具（功能更强大）

| 工具名称 | 主要功能 | 适用场景 |
|---------|---------|---------|
| **Softaken PST Management** | 拆分、合并、压缩、升级、提取、转换 PST 文件 [[1]] | 企业级批量管理 |
| **Shoviv PST Manager** | 5 合 1 工具集，修复损坏文件、迁移数据 [[2]] | 故障恢复与迁移 |
| **SysTools PST Viewer** | 免费查看邮件/联系人/日历/附件，支持搜索 [[6]] | 快速预览与取证 |
| **eSoftTools PST Viewer** | 无需 Outlook 即可打开读取 PST 文件 [[3]] | 个人用户轻量查看 |
| **DataNumen PST Repair** | 专业修复损坏/加密的 PST 文件（搜索结果提及） | 数据恢复 |

---

## 🐍 开发者快速上手示例（Python + pypff）

```python
import pypff

# 打开 PST 文件
pst = pypff.file()
pst.open("example.pst")

# 遍历根文件夹
root = pst.get_root_folder()
for folder in root.sub_folders:
    print(f"文件夹: {folder.name}")
    for message in folder.sub_messages:
        print(f"  邮件: {message.subject}")

pst.close()
```

> ⚠️ 注意：pypff 仅支持**读取**，不支持写入或修改 PST 文件 [[44]]

---

## ✅ 选型建议

| 需求场景 | 推荐工具 |
|---------|---------|
| 🔍 仅查看/搜索 PST 内容 | XstReader、SysTools PST Viewer |
| 🔄 格式转换（PST→MBOX/EML等） | libpst + readpst |
| 🧑‍💻 开发集成（Python/Java/C++） | pypff / java-libpst / PST SDK |
| 🛠️ 修复损坏/加密的 PST | Shoviv、DataNumen（商业工具） |
| 🖥️ 跨平台命令行处理 | libpst（Linux/macOS/WSL） |
| 📦 批量导出邮件/附件 | XstExport（XstReader 配套工具） |

---

> 💡 **提示**：操作重要 PST 文件前务必备份！部分商业工具提供免费试用版，建议先测试再购买。

如需了解某个工具的详细使用教程，欢迎告诉我，我可以为您进一步搜索具体文档。
