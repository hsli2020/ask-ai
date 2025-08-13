# 用PHP能开发Model Context Protocol(MCP)吗？如果能，请用PHP编写一个MCP例子，让我能明白MCP的工作原理

截至 **2026 年 1 月**，**Model Context Protocol（MCP）** 是一种新兴的、用于在 AI 模型（如大语言模型）与外部工具或服务之间进行结构化通信的协议。它由 **OpenRouter** 等平台推动，旨在标准化 LLM 如何请求上下文信息、调用工具、获取实时数据等。

MCP 的核心思想是：  
> **LLM 可以通过标准格式“请求”上下文或执行操作，外部服务（称为 MCP Server）响应这些请求并返回结构化结果。**

MCP 通常基于 **JSON-RPC 2.0** 或类似机制，通过 **stdin/stdout**（进程间通信）或 **HTTP/WebSocket** 进行交互。虽然官方参考实现多为 Python/Node.js，但 **PHP 完全可以实现 MCP Server**，只要它能处理输入/输出流或 HTTP 请求，并遵循 MCP 的消息格式。

---

### ✅ 用 PHP 实现一个简单的 MCP Server（基于 stdin/stdout）

以下是一个 **最小可行的 MCP 示例**，模拟一个“天气查询”工具。LLM 会发送 MCP 请求，PHP 脚本解析并返回模拟天气数据。

#### 步骤说明：
1. LLM 通过标准输入（stdin）发送 MCP 请求（JSON 格式）。
2. PHP 脚本读取 stdin，解析 JSON。
3. 根据请求类型（如 `get_weather`），执行逻辑。
4. 将结果通过 stdout 输出为 MCP 响应。

---

### 📜 PHP MCP 示例代码 (`mcp_weather_server.php`)

```php
#!/usr/bin/env php
<?php

// 设置错误报告（开发时使用）
error_reporting(E_ALL);
ini_set('display_errors', 1);

// 从 stdin 读取一行（MCP 通常按行分隔 JSON 消息）
$input = fgets(STDIN);
if (!$input) {
    fwrite(STDERR, "No input received.\n");
    exit(1);
}

// 解码 JSON
$request = json_decode($input, true);

// 基本验证
if (!isset($request['method']) || !isset($request['id'])) {
    $response = [
        'jsonrpc' => '2.0',
        'id' => $request['id'] ?? null,
        'error' => ['code' => -32600, 'message' => 'Invalid Request']
    ];
    echo json_encode($response) . "\n";
    exit(1);
}

$method = $request['method'];
$params = $request['params'] ?? [];
$id = $request['id'];

// 模拟工具注册表
$tools = [
    'get_weather' => function($params) {
        $location = $params['location'] ?? 'Unknown';
        // 模拟天气数据（实际可调用 API）
        return [
            'location' => $location,
            'temperature' => rand(-10, 35),
            'unit' => 'Celsius',
            'condition' => ['sunny', 'cloudy', 'rainy'][array_rand([0,1,2])]
        ];
    }
];

if (isset($tools[$method])) {
    try {
        $result = $tools[$method]($params);
        $response = [
            'jsonrpc' => '2.0',
            'id' => $id,
            'result' => $result
        ];
    } catch (Exception $e) {
        $response = [
            'jsonrpc' => '2.0',
            'id' => $id,
            'error' => ['code' => -32000, 'message' => $e->getMessage()]
        ];
    }
} else {
    $response = [
        'jsonrpc' => '2.0',
        'id' => $id,
        'error' => ['code' => -32601, 'message' => 'Method not found: ' . $method]
    ];
}

// 输出响应（必须换行，MCP 依赖行分隔）
echo json_encode($response, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES) . "\n";
// 刷新输出缓冲
flush();
```

---

### ▶️ 如何测试这个 MCP Server？

1. **保存代码为 `mcp_weather_server.php`**
2. **赋予执行权限（Linux/macOS）**：
   ```bash
   chmod +x mcp_weather_server.php
   ```
3. **手动测试（通过命令行模拟 LLM 输入）**：
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"get_weather","params":{"location":"Beijing"}}' | php mcp_weather_server.php
   ```

4. **预期输出（示例）**：
   ```json
   {"jsonrpc":"2.0","id":1,"result":{"location":"Beijing","temperature":22,"unit":"Celsius","condition":"sunny"}}
   ```

---

### 🔧 MCP 关键点说明

| 组件 | 说明 |
|------|------|
| **通信方式** | 通常为 `stdin/stdout`（子进程）或 HTTP/WebSocket |
| **消息格式** | 基于 JSON-RPC 2.0，每条消息独立一行 |
| **方法（method）** | 对应工具名称，如 `get_weather`、`search_web` |
| **参数（params）** | 工具所需参数，由 LLM 根据上下文生成 |
| **响应（result/error）** | 成功返回 `result`，失败返回 `error` |

---

### 💡 扩展建议

- 支持 **多个工具**：扩展 `$tools` 数组。
- 添加 **认证/限流**：适用于生产环境。
- 使用 **ReactPHP 或 Swoole** 实现高性能异步 MCP Server（支持 WebSocket/HTTP）。
- 与 **LLM 框架集成**：如 LangChain、LlamaIndex（需适配其 MCP 客户端）。

---

### ✅ 总结

> **是的，PHP 完全可以开发 MCP Server！**  
> 虽然 PHP 不是 AI 领域的主流语言，但因其强大的 I/O 和 JSON 处理能力，实现 MCP 协议毫无障碍。关键在于遵循 **JSON-RPC 2.0 格式** 和 **行分隔消息** 的约定。

如果你有具体的 MCP 使用场景（如与 Claude、OpenRouter 集成），我可以提供更针对性的示例！


# 如果我开发了一个MCP，比如上面的get_weather，如何在Claude中使用它呢？

要在 **Claude（由 Anthropic 开发的大语言模型）** 中使用你开发的 **Model Context Protocol (MCP) 工具**（如 `get_weather`），你需要通过 **支持 MCP 的平台或客户端** 将你的 PHP MCP Server 与 Claude 连接起来。

截至 **2026 年 1 月**，**Claude 官方本身并不直接运行 MCP Server**，但 **Anthropic 支持通过“工具调用”（Tool Use）机制** 与外部工具集成。而 **MCP 是一种标准化的工具通信协议**，通常由 **中间层（如 OpenRouter、LM Studio、Jan.ai、或自建代理）** 来桥接 LLM 与你的 MCP Server。

---

### ✅ 正确路径：通过 **支持 MCP 的代理/平台** 连接 Claude + 你的 PHP MCP

目前最主流的方式是使用 **OpenRouter** 或 **本地 AI 框架（如 LM Studio、Continue.dev、Jan）**，它们支持：
- 调用 Claude（通过 Anthropic API）
- 同时加载 MCP 工具
- 自动在 Claude 需要上下文时触发你的 MCP Server

---

## 🛠️ 方案一：使用 OpenRouter（推荐，云端）

> OpenRouter 是一个支持 MCP 的聚合 LLM 平台，允许你将自定义 MCP 工具绑定到 Claude 等模型。

### 步骤：

#### 1. **部署你的 PHP MCP Server 到公网**
由于 OpenRouter 需要能访问你的 MCP 服务，你必须让 PHP 脚本可通过 HTTP 访问（MCP 也支持 HTTP 模式）。

> ⚠️ 原始 stdin/stdout 版本只能用于本地进程，**需升级为 HTTP 服务**。

##### 修改 PHP 脚本为 HTTP 服务（使用内置服务器或框架）：

```php
<?php
// mcp_http_server.php

header('Content-Type: application/json');

if ($_SERVER['REQUEST_METHOD'] !== 'POST') {
    http_response_code(405);
    echo json_encode(['error' => 'Method not allowed']);
    exit;
}

$input = file_get_contents('php://input');
$request = json_decode($input, true);

if (!$request || !isset($request['method'])) {
    http_response_code(400);
    echo json_encode(['error' => 'Invalid JSON']);
    exit;
}

$method = $request['method'];
$params = $request['params'] ?? [];
$id = $request['id'] ?? null;

// 工具定义
$tools = [
    'get_weather' => function($params) {
        $location = $params['location'] ?? 'Unknown';
        return [
            'location' => $location,
            'temperature' => rand(-10, 35),
            'unit' => 'Celsius',
            'condition' => ['sunny', 'cloudy', 'rainy'][array_rand([0,1,2])]
        ];
    }
];

if (isset($tools[$method])) {
    try {
        $result = $tools[$method]($params);
        $response = [
            'jsonrpc' => '2.0',
            'id' => $id,
            'result' => $result
        ];
    } catch (Exception $e) {
        $response = [
            'jsonrpc' => '2.0',
            'id' => $id,
            'error' => ['code' => -32000, 'message' => $e->getMessage()]
        ];
    }
} else {
    http_response_code(404);
    $response = [
        'jsonrpc' => '2.0',
        'id' => $id,
        'error' => ['code' => -32601, 'message' => 'Method not found']
    ];
}

echo json_encode($response, JSON_UNESCAPED_UNICODE);
```

#### 2. **启动 HTTP 服务（开发测试）**
```bash
php -S 0.0.0.0:8080 mcp_http_server.php
```

#### 3. **使用内网穿透（如 ngrok）暴露公网地址**
```bash
ngrok http 8080
```
你会得到类似：`https://abc123.ngrok.io`

#### 4. **在 OpenRouter 中注册 MCP 工具**
1. 访问 [https://openrouter.ai](https://openrouter.ai)
2. 登录后进入 **Tools** 或 **MCP Integrations**（具体界面可能更新）
3. 添加新工具：
   - **Name**: Weather Tool
   - **MCP Endpoint**: `https://abc123.ngrok.io`
   - **Method**: `get_weather`
   - **Parameters**: `{ "location": "string" }`

> 🔔 注意：OpenRouter 可能要求你提供 **MCP Manifest**（一个描述工具的 JSON 文件），例如：

```json
{
  "name": "weather",
  "description": "Get current weather for a location",
  "methods": {
    "get_weather": {
      "description": "Fetch weather data",
      "parameters": {
        "type": "object",
        "properties": {
          "location": { "type": "string" }
        },
        "required": ["location"]
      }
    }
  }
}
```

#### 5. **在聊天中使用 Claude + 你的工具**
在 OpenRouter 聊天界面选择 **Claude 3.5 Sonnet**，并启用你的天气工具。

然后提问：
> “What’s the weather in Tokyo today?”

Claude 会自动调用你的 PHP MCP Server，获取结果并整合回答。

---

## 🖥️ 方案二：本地使用（如 LM Studio、Continue.dev）

如果你在本地运行 AI 工具链：

1. 安装 [LM Studio](https://lmstudio.ai/) 或 [Continue](https://continue.dev/)
2. 在设置中添加 **MCP Server** 地址（如 `http://localhost:8080`）
3. 选择 Claude（通过 Anthropic API key）
4. 启用 MCP 工具
5. 聊天时自动触发

---

## 📌 关键前提

| 条件 | 说明 |
|------|------|
| ✅ **Claude 必须通过支持 MCP 的中介调用** | Anthropic API 本身不直接连 MCP，需中间层 |
| ✅ **MCP Server 必须可被访问** | 本地测试用 `localhost`，远程需公网 IP 或 ngrok |
| ✅ **遵循 MCP/JSON-RPC 格式** | 请求/响应结构必须正确 |
| ✅ **提供工具描述（Manifest）** | 帮助 LLM 理解何时调用你的工具 |

---

## ❌ 常见误区

- **误区**： “我可以直接在 Anthropic Console 里绑定我的 PHP 脚本”  
  **事实**：不能。Anthropic 不提供 MCP 集成入口，必须通过第三方平台。

- **误区**： “MCP 只能用 stdin/stdout”  
  **事实**：MCP 支持多种传输方式，HTTP 是最通用的。

---

## ✅ 总结

要让 **Claude 使用你的 PHP 编写的 `get_weather` MCP 工具**，你需要：

1. 将 PHP 脚本改为 **HTTP 服务**
2. 通过 **ngrok 等工具暴露公网地址**
3. 在 **OpenRouter / LM Studio / Continue 等支持 MCP 的平台** 中注册该工具
4. 在这些平台中选择 **Claude 模型 + 你的工具**
5. 正常聊天，LLM 会在需要时自动调用你的 PHP 服务！

> 💡 提示：未来如果 Anthropic 官方支持 MCP，流程会更简单。但目前（2026年初），**OpenRouter 是最成熟的桥梁**。

如果你希望我提供完整的 **Docker 化 PHP MCP 服务 + OpenRouter 配置示例**，也可以告诉我！
