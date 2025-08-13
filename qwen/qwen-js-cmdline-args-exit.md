# nodejs 如何取得命令行参数

 在 Node.js 中，获取命令行参数主要通过全局对象 `process` 的 `argv` 属性来实现。`process.argv` 返回一个数组，包含启动 Node.js 进程时的所有命令行参数。

### 基本用法

`process.argv` 数组的结构通常如下：
1. **索引 0**：Node.js 可执行文件的路径（例如 `/usr/bin/node`）。
2. **索引 1**：正在执行的 JavaScript 文件的路径。
3. **索引 2 及之后**：用户传入的实际命令行参数。

#### 示例代码

假设你有一个文件 `app.js`：

```javascript
// app.js
console.log('所有参数:', process.argv);
console.log('用户传入的参数:', process.argv.slice(2));
```

在终端运行：
```bash
node app.js hello world --port=3000
```

输出结果类似：
```text
所有参数: [
  '/usr/local/bin/node',
  '/Users/yourname/project/app.js',
  'hello',
  'world',
  '--port=3000'
]
用户传入的参数: [ 'hello', 'world', '--port=3000' ]
```

### 解析带选项的参数

如果命令行参数包含标志（如 `--name`）或键值对，手动解析会比较繁琐。推荐使用成熟的第三方库，如 **minimist** 或 **yargs**。

#### 方案 A：使用 `minimist` (轻量级)

首先安装：
```bash
npm install minimist
```

使用示例：
```javascript
const minimist = require('minimist');

// 截取用户参数并解析
const args = minimist(process.argv.slice(2));

console.log(args);
```

运行：
```bash
node app.js --name=Alice --port 3000 verbose
```

输出对象：
```javascript
{
  _: [ 'verbose' ],      // 非选项参数
  name: 'Alice',
  port: 3000
}
```

#### 方案 B：使用 `yargs` (功能更强大，支持帮助文档)

首先安装：
```bash
npm install yargs
```

使用示例：
```javascript
const yargs = require('yargs/yargs');
const { hideBin } = require('yargs/helpers');

const argv = yargs(hideBin(process.argv))
  .option('name', {
    alias: 'n',
    type: 'string',
    description: '用户名'
  })
  .option('port', {
    type: 'number',
    default: 8080,
    description: '端口号'
  })
  .help()
  .argv;

console.log(`你好, ${argv.name}! 服务运行在端口 ${argv.port}`);
```

### 总结

- **简单场景**：直接使用 `process.argv.slice(2)` 获取原始数组。
- **复杂场景**（需要解析 `--key=value` 或生成帮助文档）：建议使用 **minimist** 或 **yargs** 库，它们能自动处理类型转换、默认值和错误提示，使代码更健壮。


# 如何退出程序

在 Node.js 中，退出程序主要有以下几种方式，具体取决于你是想**正常结束**、**报错退出**还是**处理异步清理工作**。

### 1. 主动退出程序

#### A. `process.exit([code])` (最常用)
这是最直接的方法。调用它会立即终止 Node.js 进程。
- **参数 `code`**：退出码。
  - `0`：表示成功退出（默认值）。
  - 非 `0`（如 `1`）：表示发生错误退出。

```javascript
// 成功退出
process.exit(0);

// 失败退出
process.exit(1);
```

> **注意**：`process.exit()` 会强制立即停止进程，**不会等待**未完成的异步操作（如文件写入、数据库请求、定时器）完成。如果需要优雅退出，请看第 3 点。

#### B. `process.exitCode = code` (推荐用于正常流程)
如果你希望程序在当前事件循环结束后自然退出，但可以指定退出码，可以使用此属性。
- 它不会立即杀死进程，而是设置一个标志，当事件队列为空时，Node.js 会以该代码退出。
- 这允许当前的异步操作完成。

```javascript
// 设置退出码为 1，但允许当前异步任务完成后再退出
process.exitCode = 1;

// 模拟一个异步操作
setTimeout(() => {
  console.log('异步任务完成，程序即将以代码 1 退出');
  // 不需要调用 process.exit()，事件循环空了会自动退出
}, 1000);
```

---

### 2. 被动退出（自然结束）

如果主线程中没有待处理的异步操作（没有定时器、没有 pending 的 I/O 请求），Node.js 进程会**自动退出**。

```javascript
console.log('开始');
// 没有任何异步操作
console.log('结束');
// 程序运行到这里会自动退出，无需手动调用 exit
```

---

### 3. 优雅退出 (Graceful Shutdown)

在生产环境中，直接调用 `process.exit()` 可能导致数据丢失（例如日志没写完、数据库连接没关闭）。通常我们需要监听信号，先清理资源，再退出。

#### 场景：监听 `SIGINT` (Ctrl+C) 或 `SIGTERM`

```javascript
const server = require('http').createServer((req, res) => {
  res.end('Hello');
});

server.listen(3000, () => {
  console.log('服务器运行在端口 3000');
});

// 定义清理函数
const shutdown = (signal) => {
  console.log(`收到信号 ${signal}，正在关闭服务器...`);
  
  server.close(() => {
    console.log('HTTP 服务器已关闭');
    
    // 在这里执行其他清理工作：关闭数据库连接、刷新日志等
    // db.close(); 
    
    console.log('清理完成，退出进程');
    process.exit(0);
  });

  // 防止关闭超时，强制退出（例如 10 秒后）
  setTimeout(() => {
    console.error('未能优雅关闭，强制退出');
    process.exit(1);
  }, 10000);
};

// 监听终端中断 (Ctrl+C)
process.on('SIGINT', () => shutdown('SIGINT'));

// 监听终止信号 (常用于 docker stop 或 kill 命令)
process.on('SIGTERM', () => shutdown('SIGTERM'));
```

### 总结对比

| 方法 | 行为特点 | 适用场景 |
| :--- | :--- | :--- |
| **`process.exit(code)`** | **立即**终止，忽略未完成的异步任务。 | 脚本执行完毕、发生致命错误无法恢复时。 |
| **`process.exitCode = code`** | 设置退出码，**等待**当前事件循环完成后自然退出。 | 希望异步任务跑完再退出的常规逻辑。 |
| **自然结束** | 无待处理任务时自动退出。 | 简单的同步脚本。 |
| **信号监听 + 清理** | 先关闭服务/连接，再调用 `exit`。 | **生产环境服务器**，确保数据安全和资源释放。 |

**最佳实践建议**：
- 写简单脚本时，直接用 `process.exit()` 或让其自然结束。
- 写长期运行的服务（如 Web 服务器）时，务必实现**优雅退出**逻辑，不要直接在代码深处随意调用 `process.exit()`。

