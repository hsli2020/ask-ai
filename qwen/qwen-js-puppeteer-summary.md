# 总结一下 js 库 puppeteer 的用法

# Puppeteer 用法总结

Puppeteer 是一个 Node.js 库，提供高级 API 来控制 Chrome 或 Chromium 浏览器，常用于网页自动化、截图、爬虫、测试等场景。

---

## 一、安装

```bash
# 安装完整版（含 Chromium）
npm install puppeteer

# 安装轻量版（需自行配置浏览器）
npm install puppeteer-core
```

---

## 二、基本使用示例

```javascript
const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch({
    headless: true,           // 无头模式
    args: ['--no-sandbox']    // 避免沙盒问题
  });
  const page = await browser.newPage();
  
  await page.goto('https://example.com', { waitUntil: 'networkidle2' });
  await page.screenshot({ path: 'example.png' });
  
  await browser.close();
})();
```

---

## 三、核心功能与常用 API

### 1. 页面导航
```javascript
await page.goto('https://example.com', {
  waitUntil: 'networkidle2',  // 等待网络空闲
  timeout: 30000
});
await page.reload();
await page.goBack();
```

### 2. 元素操作
```javascript
// 选择器
await page.$('.class');           // 单个元素
await page.$$('.class');         // 多个元素
await page.waitForSelector('.class');

// 点击与输入
await page.click('#button');
await page.type('#input', 'text');
await page.focus('#input');

// 获取内容
const text = await page.$eval('.class', el => el.textContent);
const html = await page.content();
```

### 3. 截图与 PDF
```javascript
// 截图
await page.screenshot({
  path: 'page.png',
  fullPage: true
});

// 生成 PDF
await page.pdf({
  path: 'page.pdf',
  format: 'A4',
  printBackground: true
});
```

### 4. 执行 JavaScript
```javascript
// 在页面上下文中执行
const result = await page.evaluate(() => {
  return document.title;
});

// 传入参数
const result = await page.evaluate((selector) => {
  return document.querySelector(selector).innerText;
}, '.title');
```

### 5. 网络请求拦截
```javascript
await page.setRequestInterception(true);

page.on('request', request => {
  if (request.url().endsWith('.png')) {
    request.abort();  // 阻止图片加载
  } else {
    request.continue();
  }
});
```

### 6. 文件上传与下载
```javascript
// 上传文件
await page.setInputFiles('#fileInput', 'path/to/file.txt');

// 下载文件（需配置下载路径）
const client = await page.target().createCDPSession();
await client.send('Page.setDownloadBehavior', {
  behavior: 'allow',
  downloadPath: './downloads'
});
```

### 7. 等待与监听
```javascript
// 等待选择器、导航、函数等
await page.waitForSelector('.loaded');
await page.waitForNavigation();
await page.waitForFunction(() => document.querySelector('.ready'));

// 监听事件
page.on('console', msg => console.log(msg.text()));
page.on('requestfailed', req => console.log(req.failure().errorText));
```

### 8. Cookie 管理
```javascript
// 获取与设置
const cookies = await page.cookies();
await page.setCookie({ name: 'token', value: 'abc123' });
await page.deleteCookie(...cookies);
```

### 9. 多页面与标签管理
```javascript
const pages = await browser.pages();
const newPage = await browser.newPage();
await page.close();
```

### 10. 设备模拟
```javascript
const iPhone = puppeteer.devices['iPhone 13'];
await page.emulate(iPhone);
```

---

## 四、高级用法

### 1. 绕过反爬机制
```javascript
await page.setExtraHTTPHeaders({
  'Accept-Language': 'zh-CN,zh;q=0.9'
});

await page.evaluateOnNewDocument(() => {
  Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
});
```

### 2. 使用代理
```javascript
const browser = await puppeteer.launch({
  args: ['--proxy-server=http://proxy-ip:port']
});
```

### 3. 性能分析
```javascript
await page.setCacheEnabled(false);
const metrics = await page.metrics();
const coverage = await page.coverage.startJSCoverage();
```

### 4. 移动端测试
```javascript
await page.emulateViewport({ width: 375, height: 667, isMobile: true });
```

---

## 五、最佳实践与注意事项

- ✅ 使用 `try...finally` 保证浏览器关闭
- ✅ 设置合理超时，避免卡死
- ✅ 使用 `waitForSelector` 替代硬编码 `sleep`
- ✅ 拦截不必要资源提升性能
- ✅ 生产环境建议使用 `puppeteer-core` + 自有浏览器
- ⚠️ 注意内存泄漏，及时关闭页面和浏览器
- ⚠️ 避免频繁启动浏览器，可复用实例
- ⚠️ 遵守目标网站 robots.txt 和使用条款

---

## 六、典型应用场景

- 网页自动化测试
- 生成截图或 PDF 报告
- 爬取动态渲染内容（SPA）
- 性能监控与分析
- 表单自动填写与提交
- SEO 预渲染

---

Puppeteer 功能强大且灵活，是前端自动化和测试领域的重要工具。建议结合具体业务需求，合理配置使用方式，以提升效率和稳定性。


# 如何根据链接上的文字点击该链接？

# Puppeteer 根据链接文字点击链接的方法

在 Puppeteer 中，如果你想根据链接（`<a>` 标签）上的可见文字来点击该链接，有多种实现方式。以下是常用方法及最佳实践：

---

## 方法一：使用 XPath（推荐）

```javascript
// 精确匹配链接文字
await page.click(`//a[text()='登录']`);

// 模糊匹配（包含文字）
await page.click(`//a[contains(text(), '登录')]`);

// 忽略大小写匹配
await page.click(`//a[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), '登录')]`);
```

✅ 优点：语义清晰，支持部分匹配  
⚠️ 注意：XPath 中文字需完全匹配，注意空格和大小写

---

## 方法二：使用 evaluate + querySelectorAll

```javascript
await page.evaluate(() => {
  const links = Array.from(document.querySelectorAll('a'));
  const target = links.find(link => link.textContent.trim() === '登录');
  if (target) target.click();
});
```

✅ 优点：灵活，可自定义匹配逻辑  
⚠️ 注意：需确保元素可见且可点击

---

## 方法三：使用 $$eval 批量查找并点击

```javascript
await page.$$eval('a', (links, text) => {
  const target = links.find(link => link.textContent.trim() === text);
  if (target) target.click();
}, '登录');
```

✅ 优点：语法简洁，适合批量处理  
⚠️ 注意：同样需确保元素可交互

---

## 方法四：封装为通用函数（推荐用于生产）

```javascript
async function clickLinkByText(page, text, options = {}) {
  const { exact = true, timeout = 5000 } = options;

  try {
    const xpath = exact
      ? `//a[normalize-space()='${text}']`
      : `//a[contains(normalize-space(), '${text}')]`;

    await page.waitForXPath(xpath, { timeout });
    await page.click(xpath);
    return true;
  } catch (err) {
    console.error(`未找到链接文字: "${text}"`);
    return false;
  }
}

// 使用示例
await clickLinkByText(page, '登录');
await clickLinkByText(page, '下一步', { exact: false });
```

✅ 优点：可复用、支持模糊匹配、带错误处理  
⚠️ 注意：建议加入日志或异常捕获机制

---

## 方法五：处理嵌套元素中的文字

如果链接文字在子元素中（如 `<a><span>登录</span></a>`）：

```javascript
// XPath 匹配所有后代文本节点
await page.click(`//a[contains(., '登录')]`);

// 或使用 evaluate 匹配 innerText
await page.evaluate((text) => {
  const link = Array.from(document.querySelectorAll('a')).find(
    el => el.innerText.includes(text)
  );
  if (link) link.click();
}, '登录');
```

---

## 方法六：等待链接出现后再点击

```javascript
// 等待 XPath 元素出现
await page.waitForXPath(`//a[text()='登录']`, { timeout: 5000 });
await page.click(`//a[text()='登录']`);

// 或等待后使用 evaluate 点击
await page.waitForFunction(
  (text) => Array.from(document.querySelectorAll('a')).some(a => a.textContent.includes(text)),
  {}, '登录'
);
await page.evaluate((text) => {
  document.querySelector(`a[href*="${text}"]`)?.click();
}, '登录');
```

---

## 常见匹配场景示例

```javascript
// 精确匹配
await page.click(`//a[text()='首页']`);

// 包含匹配
await page.click(`//a[contains(text(), '下载')]`);

// 去除空格后匹配
await page.click(`//a[normalize-space()='联系我们']`);

// 匹配 href 中包含关键字
await page.click(`//a[contains(@href, 'login')]`);

// 匹配 title 属性
await page.click(`//a[@title='用户中心']`);
```

---

## 完整示例

```javascript
const puppeteer = require('puppeteer');

(async () => {
  const browser = await puppeteer.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');

  try {
    // 等待链接出现
    await page.waitForXPath(`//a[contains(text(), '登录')]`, { timeout: 5000 });

    // 点击链接
    await page.click(`//a[contains(text(), '登录')]`);

    // 等待跳转
    await page.waitForNavigation();

    console.log('点击成功');
  } catch (err) {
    console.error('点击失败:', err.message);
  }

  await browser.close();
})();
```

---

## 注意事项与调试技巧

- ✅ 使用 `normalize-space()` 去除多余空格
- ✅ 使用 `waitForXPath` 确保元素已加载
- ✅ 检查元素是否可见、可点击（可能被遮挡）
- ✅ 使用 `page.$x()` 调试 XPath 是否匹配成功：
  ```javascript
  const links = await page.$x(`//a[contains(text(), '登录')]`);
  console.log(links.length);  // 查看匹配数量
  ```
- ✅ 处理 iframe 中的链接需先切换上下文：
  ```javascript
  const frame = page.frames().find(f => f.url().includes('xxx'));
  await frame.click(`//a[text()='登录']`);
  ```

---

## 方法对比总结

| 方法            | 优点                     | 缺点                     | 推荐场景           |
|-----------------|--------------------------|--------------------------|--------------------|
| XPath           | 简洁、语义清晰           | 对文字匹配要求严格       | 精确/模糊匹配链接  |
| evaluate        | 灵活、可自定义逻辑       | 代码略复杂               | 特殊匹配需求       |
| $$eval          | 简洁、适合批量处理       | 功能略受限               | 简单查找点击       |
| 封装函数        | 可复用、易维护           | 需额外定义               | 生产环境推荐使用   |

---

✅ 推荐优先使用 XPath 搭配 `waitForXPath`，并结合封装函数提升可维护性。根据具体页面结构选择最适合的方式，确保稳定性与兼容性。


# 如何判断文件上传是否结束？

在 Puppeteer 中，`page.setInputFiles()` 只是将文件选择到输入框中，**并不代表服务器端上传已完成**。判断上传是否结束，需要根据网站的具体实现方式（AJAX、表单提交、WebSocket 等）来监控不同的信号。

以下是几种常用的判断方法及代码示例：

---

## 方法一：监听网络请求响应（最推荐，适用于 AJAX 上传）

大多数现代网站使用 XHR 或 Fetch 上传文件。你可以监听特定的上传接口响应。

```javascript
// 等待特定的上传接口返回成功状态
const [response] = await Promise.all([
  page.waitForResponse(res => 
    res.url().includes('/api/upload') &&  // 匹配上传接口 URL
    res.status() === 200                 // 匹配成功状态码
  ),
  page.setInputFiles('input[type="file"]', './test.pdf') // 触发上传
]);

const result = await response.json();
console.log('上传完成，服务器返回:', result);
```

✅ **优点**：最准确，直接确认服务器接收成功。  
⚠️ **注意**：需要知道上传接口的 URL 特征。

---

## 方法二：监听 DOM 元素变化（适用于有进度条或提示信息的页面）

如果上传成功后页面会出现“上传成功”提示，或进度条消失，可以监听这些元素。

```javascript
// 1. 等待进度条消失
await page.setInputFiles('input[type="file"]', './test.pdf');
await page.waitForSelector('.upload-progress', { hidden: true, timeout: 30000 });

// 2. 或者等待成功提示出现
await page.waitForSelector('.upload-success-message', { timeout: 30000 });
console.log('检测到成功提示，上传结束');
```

✅ **优点**：不依赖网络细节，贴近用户视角。  
⚠️ **注意**：需确保选择器稳定，避免误判。

---

## 方法三：等待页面导航（适用于传统表单提交）

如果是老式表单，上传后可能会刷新页面或跳转。

```javascript
await Promise.all([
  page.waitForNavigation({ waitUntil: 'networkidle2' }), // 等待跳转且网络空闲
  page.click('#submit-upload-btn') // 点击提交按钮
]);
console.log('页面已跳转，上传流程结束');
```

✅ **优点**：简单直接。  
⚠️ **注意**：仅适用于会刷新/跳转的场景。

---

## 方法四：自定义轮询检查（适用于复杂逻辑）

如果上述方法都不可用，可以在页面上下文中轮询检查状态。

```javascript
await page.setInputFiles('input[type="file"]', './test.pdf');

// 轮询检查某个全局变量或 DOM 状态
await page.waitForFunction(() => {
  // 假设网站有一个全局变量 window.uploadStatus
  return window.uploadStatus === 'completed'; 
  // 或者检查 DOM: return !!document.querySelector('.file-list .done');
}, { timeout: 60000, polling: 500 });

console.log('自定义检查通过，上传完成');
```

✅ **优点**：极度灵活，可处理任何前端逻辑。  
⚠️ **注意**：性能稍差，需合理设置 polling 间隔。

---

## 方法五：监听控制台日志（适用于调试或特定框架）

有些框架会在上传完成时打印特定日志。

```javascript
let uploadDone = false;

const listener = msg => {
  if (msg.text().includes('Upload finished')) {
    uploadDone = true;
  }
};

page.on('console', listener);

await page.setInputFiles('input[type="file"]', './test.pdf');

// 等待日志出现
await page.waitForFunction(() => window.uploadDone, { timeout: 30000 });
// 注意：上面这行需要配合 evaluate 设置 window.uploadDone，或者直接用 Promise 包装 listener
```

*更实用的日志监听写法：*
```javascript
const uploadPromise = new Promise(resolve => {
  const handler = msg => {
    if (msg.text().includes('upload success')) {
      page.off('console', handler);
      resolve();
    }
  };
  page.on('console', handler);
});

await page.setInputFiles('input[type="file"]', './test.pdf');
await uploadPromise; // 等待日志触发
```

---

## 综合实战示例（健壮版）

结合网络监听和 DOM 检查，防止单一条件失效。

```javascript
async function uploadFileAndWait(page, selector, filePath) {
  try {
    // 1. 设置文件
    await page.setInputFiles(selector, filePath);

    // 2. 定义上传完成的信号（多重保障）
    const uploadComplete = Promise.race([
      // 信号 A: 网络响应
      page.waitForResponse(res => res.url().includes('upload') && res.ok(), { timeout: 30000 }),
      
      // 信号 B: DOM 变化 (成功提示)
      page.waitForSelector('.toast-success', { timeout: 30000 }).catch(() => null),
      
      // 信号 C: 进度条消失
      page.waitForSelector('.progress-bar', { hidden: true, timeout: 30000 }).catch(() => null)
    ]);

    await uploadComplete;
    console.log('✅ 文件上传确认完成');
    return true;

  } catch (error) {
    console.error('❌ 上传超时或失败:', error.message);
    // 可选：截图调试
    await page.screenshot({ path: 'upload_error.png' });
    return false;
  }
}

// 调用
await uploadFileAndWait(page, 'input[type="file"]', './my-file.pdf');
```

---

## 调试技巧与注意事项

1.  **区分“选择完成”与“上传完成”**：
    *   `setInputFiles` 返回仅代表文件已选入输入框。
    *   真正的上传通常由 JS 自动触发或需点击“提交”按钮。

2.  **大文件超时**：
    *   默认超时是 30 秒，大文件需增加 `timeout` 参数。
    *   ```javascript
        page.waitForResponse(..., { timeout: 120000 }) // 2 分钟
        ```

3.  **多文件上传**：
    *   如果是多文件，可能需要监听多次响应，或等待所有文件名列出。
    *   ```javascript
        await page.setInputFiles('input', ['file1.jpg', 'file2.jpg']);
        // 等待列表中有 2 个文件项
        await page.waitForFunction((count) => {
          return document.querySelectorAll('.file-item').length >= count;
        }, {}, 2);
        ```

4.  **使用 DevTools Protocol (CDP)**：
    *   如果网站使用 WebSocket 上传，常规网络监听可能无效，需监听 CDP 事件。
    *   ```javascript
        const client = await page.target().createCDPSession();
        await client.send('Network.enable');
        client.on('Network.webSocketFrameReceived', (event) => {
           if (event.response.payloadData.includes('upload_complete')) {
             // 触发完成逻辑
           }
        });
        ```

5.  **失败重试机制**：
    *   上传容易受网络波动影响，建议包裹在重试逻辑中。

---

## 总结建议

| 场景 | 推荐判断方式 |
| :--- | :--- |
| **REST API / AJAX 上传** | `page.waitForResponse()` 监听接口 |
| **有进度条/提示 UI** | `page.waitForSelector()` 监听 DOM 变化 |
| **上传后跳转页面** | `page.waitForNavigation()` |
| **未知/复杂逻辑** | `page.waitForFunction()` 轮询状态 |
| **WebSocket 上传** | CDP Session 监听帧数据 |

**最佳实践**：优先使用 **网络响应监听**，因为它最直接反映服务器状态；辅以 **DOM 检查** 作为兜底，确保用户体验层面的完成。
