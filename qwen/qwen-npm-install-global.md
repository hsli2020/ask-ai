# 我用命令 npm install -g puppeteer 安装了一个库，在代码中 import("puppeteer")，运行的时候node却报错，找不到模块 puppeteer，是什么原因？

这是一个常见的 Node.js 模块解析问题。主要原因和解决方案如下：

## 问题原因

**全局安装的 npm 包不能被项目代码直接 require/import**

当你使用 `npm install -g puppeteer` 全局安装时，puppeteer 被安装在全局目录（如 `/usr/local/lib/node_modules/` 或 `%APPDATA%\npm\node_modules/`），而 Node.js 在运行时只会从以下位置查找模块：
1. 当前项目的 `node_modules` 目录
2. 父级目录的 `node_modules`（逐级向上）
3. 内置模块
4. **不会**自动查找全局安装的包

## 解决方案

### 方案 1：本地安装（推荐）
```bash
# 在你的项目根目录下执行
npm install puppeteer

# 或者使用 yarn
yarn add puppeteer
```

这样 puppeteer 会被安装到项目的 `node_modules` 目录中，代码就能正常导入了。

### 方案 2：如果必须使用全局安装
如果你确实需要使用全局安装的包，可以通过以下方式：

**方法 A：设置 NODE_PATH 环境变量**
```bash
# Linux/macOS
export NODE_PATH=$(npm root -g)
node your-script.js

# Windows (PowerShell)
$env:NODE_PATH = npm root -g
node your-script.js
```

**方法 B：在代码中动态添加全局模块路径**
```javascript
const globalModulesPath = require('child_process')
  .execSync('npm root -g', { encoding: 'utf8' })
  .trim();

require('module').Module.globalPaths.push(globalModulesPath);

// 现在可以导入全局模块
const puppeteer = await import('puppeteer');
```

### 方案 3：使用 npx（适用于 CLI 工具）
如果 puppeteer 提供了 CLI 命令，你可以使用：
```bash
npx puppeteer [command]
```

## 最佳实践

- **开发项目依赖**：始终使用本地安装 (`npm install package-name`)
- **CLI 工具**：可以全局安装 (`npm install -g package-name`)，但不要在项目代码中导入它们
- **特殊情况**：如果多个项目共享同一个大体积依赖（如 puppeteer 的 Chromium），可以考虑使用符号链接或工作区管理

对于 puppeteer 这种库，强烈建议使用**方案 1（本地安装）**，这是 Node.js 生态的标准做法。
