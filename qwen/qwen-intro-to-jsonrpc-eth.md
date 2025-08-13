# 请详细介绍jsonrpc

当然可以！下面是对 **JSON-RPC** 的详细介绍，涵盖其定义、设计原则、版本、消息结构、工作流程、优缺点、应用场景以及示例等内容。

---

## 🌐 什么是 JSON-RPC？

**JSON-RPC**（JavaScript Object Notation - Remote Procedure Call）是一种轻量级的远程过程调用（RPC）协议，使用 **JSON**（JavaScript Object Notation）作为数据格式，通过 HTTP、WebSocket、TCP 等传输层进行通信。它允许客户端调用服务器端的函数或方法，就像调用本地函数一样，而无需了解底层实现细节。

JSON-RPC 最初由 **Ryan Dahl**（Node.js 的创建者）在 2005 年左右提出，后由 **JSON-RPC Working Group** 标准化，并被广泛用于 Web 服务、区块链（如以太坊）、微服务架构等领域。

---

## 🔧 设计原则

JSON-RPC 遵循以下核心设计原则：

1. **轻量级**：基于 JSON，结构简单，易于解析。
2. **无状态**：每个请求独立，服务器不保存客户端状态（除非应用层实现）。
3. **语言无关**：JSON 是通用数据格式，支持任何编程语言。
4. **可扩展**：支持自定义参数和错误码。
5. **简单易用**：相比 SOAP 或 gRPC，学习和实现成本低。
6. **双向通信支持**：支持通知（notification）和响应（response），可用于事件推送。

---

## 📦 JSON-RPC 版本

目前主流有两个版本：

| 版本 | 发布时间 | 特点 |
|------|----------|------|
| **JSON-RPC 1.0** | 2005 | 初始版本，功能简单，无标准错误码，已逐渐淘汰 |
| **JSON-RPC 2.0** | 2010 | 官方标准化版本，推荐使用，支持命名参数、批处理、标准错误码等 |

> ✅ **强烈建议使用 JSON-RPC 2.0**，它是当前行业标准。

---

## 📄 JSON-RPC 2.0 消息结构

一个 JSON-RPC 请求或响应是一个标准的 JSON 对象，包含以下字段：

### ✅ 请求对象（Request）

```json
{
  "jsonrpc": "2.0",
  "method": "string",
  "params": array or object,
  "id": number or string or null
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `jsonrpc` | string | ✅ 是 | 固定为 `"2.0"`，标识协议版本 |
| `method` | string | ✅ 是 | 要调用的远程方法名（如 `"getUser"`） |
| `params` | array / object | ❌ 否 | 方法参数，可以是位置参数（数组）或命名参数（对象） |
| `id` | number / string / null | ❌ 否 | 请求唯一标识符，用于匹配响应；若为 `null`，表示是“通知”（无响应） |

#### 示例 1：带位置参数的请求

```json
{
  "jsonrpc": "2.0",
  "method": "add",
  "params": [5, 3],
  "id": 1
}
```

#### 示例 2：带命名参数的请求

```json
{
  "jsonrpc": "2.0",
  "method": "createUser",
  "params": {
    "name": "Alice",
    "email": "alice@example.com"
  },
  "id": 2
}
```

#### 示例 3：通知（Notification）—— 无响应

```json
{
  "jsonrpc": "2.0",
  "method": "logEvent",
  "params": {
    "event": "user_login",
    "userId": 123
  }
  // 注意：没有 id 字段 → 表示是通知，服务器不应回复
}
```

---

### ✅ 响应对象（Response）

```json
{
  "jsonrpc": "2.0",
  "result": any,
  "id": number or string
}
```

或出错时：

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": integer,
    "message": string,
    "data": any
  },
  "id": number or string
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `jsonrpc` | string | ✅ 是 | 固定为 `"2.0"` |
| `result` | any | ❌ 否 | 成功时返回的结果（任意类型：数字、字符串、对象、数组等） |
| `error` | object | ❌ 否 | 错误信息对象（仅当出错时存在） |
| `id` | number/string | ✅ 是 | 必须与请求中的 `id` 一致，用于匹配请求-响应对 |

#### ✅ 成功响应示例：

```json
{
  "jsonrpc": "2.0",
  "result": 8,
  "id": 1
}
```

#### ❌ 错误响应示例：

```json
{
  "jsonrpc": "2.0",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": "The method 'deleteUser' does not exist"
  },
  "id": 2
}
```

---

## 🚫 JSON-RPC 2.0 标准错误码

| 错误码 | 描述 | 说明 |
|--------|------|------|
| `-32700` | Parse error | JSON 解析失败（如非法字符） |
| `-32600` | Invalid Request | 请求格式无效（缺少必要字段） |
| `-32601` | Method not found | 服务器找不到指定的方法 |
| `-32602` | Invalid params | 参数类型或数量错误 |
| `-32603` | Internal error | 服务器内部错误（未定义） |
| `-32000` ~ `-32099` | Server error | 服务器应用层错误（预留范围） |

> ⚠️ 服务器可以定义自己的错误码（如 `-32001`），但应避免与标准冲突。

---

## 🔄 工作流程（Client ↔ Server）

1. **客户端** 构造一个 JSON-RPC 请求对象（含 `method`, `params`, `id`）。
2. 通过 HTTP POST（最常见）或其他协议（如 WebSocket）发送到服务器。
3. **服务器** 接收并解析 JSON，查找对应方法，执行逻辑。
4. 服务器生成响应（成功结果或错误），返回给客户端。
5. **客户端** 根据 `id` 匹配响应，处理结果或错误。

> ✅ 如果 `id` 为 `null`，则为“通知”，服务器不应回复。

---

## 📦 批量请求（Batch Requests）

JSON-RPC 2.0 支持一次发送多个请求（批量处理），提高效率。

### 示例：批量请求

```json
[
  {
    "jsonrpc": "2.0",
    "method": "getBalance",
    "params": ["user1"],
    "id": 1
  },
  {
    "jsonrpc": "2.0",
    "method": "getTransactions",
    "params": ["user1"],
    "id": 2
  },
  {
    "jsonrpc": "2.0",
    "method": "logActivity",
    "params": ["login"],
    "id": null  // 通知，无响应
  }
]
```

### 响应（数组形式）

```json
[
  {
    "jsonrpc": "2.0",
    "result": 1500,
    "id": 1
  },
  {
    "jsonrpc": "2.0",
    "result": [...transactions...],
    "id": 2
  }
]
```

> ⚠️ 批量请求中不能混用通知和普通请求？其实可以，但响应中只返回有 `id` 的请求结果。

---

## ✅ 优点

| 优点 | 说明 |
|------|------|
| **简单易用** | JSON 结构清晰，开发门槛低 |
| **跨平台/语言** | 几乎所有语言都有 JSON 库，便于集成 |
| **轻量高效** | 相比 XML/SOAP，体积小，解析快 |
| **支持异步和通知** | 适合事件驱动架构、实时推送 |
| **支持批处理** | 减少网络往返，提升性能 |
| **兼容性好** | 可运行于 HTTP、WebSocket、TCP 等多种传输层 |

---

## ❌ 缺点

| 缺点 | 说明 |
|------|------|
| **缺乏强类型** | JSON 是动态类型，无法像 gRPC 那样做编译时检查 |
| **无内置安全机制** | 依赖 HTTPS、认证、授权等外部机制 |
| **无服务发现** | 不像 gRPC 或 RESTful API 有 OpenAPI/Swagger 这类文档标准 |
| **调试困难** | 错误信息较模糊，需依赖日志和工具辅助 |
| **不适合复杂契约** | 对于大型企业系统，可能不如 Protobuf/gRPC 严谨 |

---

## 🌍 应用场景

| 场景 | 说明 |
|------|------|
| **区块链（以太坊）** | Ethereum JSON-RPC API 是节点交互的标准接口（如 `eth_getBlockByNumber`） |
| **Web 前端与后端通信** | 替代 REST，尤其在需要 RPC 风格调用时 |
| **微服务间通信** | 简单服务之间使用 HTTP + JSON-RPC 实现轻量通信 |
| **物联网（IoT）设备控制** | 轻量协议适合资源受限设备 |
| **插件系统 / 扩展框架** | 如 VS Code 的 Language Server Protocol (LSP) 使用 JSON-RPC 通信 |
| **实时通知系统** | 使用 WebSocket + JSON-RPC 实现服务器推送（如股票行情、聊天） |

> 💡 **VS Code 的 LSP** 就是基于 JSON-RPC 构建的，实现了编辑器与语言服务器之间的高效通信。

---

## 🔧 实现示例（Python + Flask）

### 服务端（Flask）

```python
from flask import Flask, request, jsonify
import json

app = Flask(__name__)

def add(a, b):
    return a + b

@app.route('/jsonrpc', methods=['POST'])
def jsonrpc_handler():
    data = request.get_json()
    
    if isinstance(data, list):  # 批量请求
        responses = []
        for req in data:
            res = handle_request(req)
            if res is not None:  # 通知不返回
                responses.append(res)
        return jsonify(responses)
    else:  # 单个请求
        response = handle_request(data)
        return jsonify(response)

def handle_request(req):
    try:
        if req['jsonrpc'] != '2.0':
            return error_response(-32600, "Invalid JSON-RPC version", req.get('id'))

        method = req.get('method')
        params = req.get('params', [])
        req_id = req.get('id')

        if method == 'add':
            result = add(*params) if isinstance(params, list) else add(**params)
            return success_response(result, req_id)
        else:
            return error_response(-32601, "Method not found", req_id)

    except Exception as e:
        return error_response(-32603, f"Internal error: {str(e)}", req.get('id'))

def success_response(result, req_id):
    return {
        "jsonrpc": "2.0",
        "result": result,
        "id": req_id
    }

def error_response(code, message, req_id):
    return {
        "jsonrpc": "2.0",
        "error": {
            "code": code,
            "message": message
        },
        "id": req_id
    }

if __name__ == '__main__':
    app.run(debug=True)
```

### 客户端调用（Python requests）

```python
import requests

url = "http://localhost:5000/jsonrpc"

payload = {
    "jsonrpc": "2.0",
    "method": "add",
    "params": [10, 20],
    "id": 1
}

response = requests.post(url, json=payload)
print(response.json())  # 输出: {"jsonrpc": "2.0", "result": 30, "id": 1}
```

---

## 🆚 JSON-RPC vs REST vs gRPC

| 对比项 | JSON-RPC | REST | gRPC |
|--------|----------|------|------|
| 数据格式 | JSON | JSON/XML | Protobuf（二进制） |
| 通信方式 | RPC（方法调用） | 资源操作（GET/POST/PUT/DELETE） | RPC（强类型接口） |
| 性能 | 中等 | 中等 | ⭐⭐⭐ 高（二进制+HTTP/2） |
| 类型安全 | ❌ 弱 | ❌ 弱 | ✅ 强（Protobuf） |
| 文档支持 | 较弱 | ✅ Swagger/OpenAPI | ✅ Protocol Buffers + gRPC Gateway |
| 适用场景 | 简单服务、区块链、插件系统 | Web API、公开服务 | 高性能微服务、内部服务 |
| 学习成本 | 低 | 低 | 中高 |

---

## ✅ 最佳实践建议

1. **始终使用 JSON-RPC 2.0**，避免 1.0。
2. **使用 HTTPS** 保证通信安全。
3. **添加身份验证**（如 JWT、API Key）。
4. **记录请求/响应日志**，便于调试。
5. **使用标准错误码**，避免自定义混乱。
6. **对通知（id=null）保持幂等性**，因为无响应。
7. **批处理时注意顺序和超时**。
8. **前端使用封装库**（如 `json-rpc-client`）简化开发。

---

## 📚 推荐资源

- [官方规范](https://www.jsonrpc.org/specification)（推荐阅读）
- [Ethereum JSON-RPC 文档](https://ethereum.org/en/developers/docs/apis/json-rpc/)
- Python 库：[`jsonrpcclient`](https://github.com/pviotti/jsonrpcclient)
- Node.js 库：[`json-rpc-2.0`](https://www.npmjs.com/package/json-rpc-2.0)
- Java 库：[`jsonrpc4j`](https://github.com/briandilley/jsonrpc4j)

---

## ✅ 总结

> **JSON-RPC 是一种简洁、灵活、跨平台的远程调用协议，特别适合轻量级服务、区块链系统和插件架构。虽然不如 gRPC 强大，但在许多场景下“够用、易用、高效”**。

如果你正在构建一个不需要复杂契约、追求快速迭代的系统，或者对接以太坊等区块链节点，**JSON-RPC 是一个极佳的选择**。

---

如有兴趣，我还可以提供：
- JSON-RPC 在 Ethereum 中的具体方法列表
- TypeScript + JSON-RPC 客户端完整示例
- WebSocket + JSON-RPC 实现实时通信方案

欢迎继续提问！

# JSON-RPC 在 Ethereum 中的具体方法列表

当然可以！以下是 **JSON-RPC 在以太坊（Ethereum）中的完整方法列表**，基于当前主流版本（如 **Geth、OpenEthereum、Nethermind、Infura、Alchemy** 等节点实现），遵循 **Ethereum JSON-RPC API 标准**（最新为 **2024 年更新版**）。

---

## 🌐 Ethereum JSON-RPC 方法总览

以太坊的 JSON-RPC 接口是节点与外部应用交互的核心协议。所有客户端（钱包、DApp、链上分析工具）都通过它查询区块链状态、发送交易、监听事件等。

> ✅ 协议版本：**JSON-RPC 2.0**  
> ✅ 传输方式：通常通过 **HTTP/HTTPS** 或 **WebSocket**  
> ✅ 终端路径示例：`http://localhost:8545`（本地 Geth）或 `https://mainnet.infura.io/v3/YOUR_PROJECT_ID`

---

## 🔢 Ethereum JSON-RPC 方法分类

官方将方法分为以下几类：

| 类别 | 描述 |
|------|------|
| **区块相关** | 查询区块信息、哈希、叔块等 |
| **交易相关** | 查询交易、发送交易、获取交易收据 |
| **账户与状态** | 查询余额、代码、存储、nonce |
| **智能合约与调用** | 执行静态调用（call）、估算 Gas |
| **日志与事件** | 监听合约事件（Filter / Subscribe） |
| **网络信息** | 获取网络 ID、同步状态、节点信息 |
| **挖矿相关** | （仅适用于 PoW 节点）控制挖矿、设置奖励 |
| **调试与开发** | 开发者工具，用于调试交易和状态 |
| **Web3 兼容** | 为兼容 Web3.js 而保留的旧方法 |
| **订阅（Subscription）** | WebSocket 专用，实时推送事件 |

> ⚠️ **注意**：部分方法（如挖矿、调试）**仅在本地开发节点（如 Ganache、Hardhat、Geth --dev）中可用**，公链节点（Infura、Alchemy）会拒绝或返回错误。

---

## 📜 完整方法列表（按类别）

### ✅ 1. 区块相关（Block）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `eth_getBlockByHash` | `blockHash`, `fullTransactions` (bool) | `Block object` 或 `null` | 按哈希获取区块，可选包含完整交易 |
| `eth_getBlockByNumber` | `blockNumber`（hex 或 "latest"/"earliest"/"pending"）, `fullTransactions` | `Block object` 或 `null` | 按区块编号获取区块 |
| `eth_getBlockTransactionCountByHash` | `blockHash` | `hex string`（交易数量） | 获取区块中交易数量 |
| `eth_getBlockTransactionCountByNumber` | `blockNumber` | `hex string` | 同上，按编号 |
| `eth_getUncleByBlockHashAndIndex` | `blockHash`, `index`（hex） | `Block object` 或 `null` | 获取叔块（uncle） |
| `eth_getUncleByBlockNumberAndIndex` | `blockNumber`, `index` | `Block object` 或 `null` | 同上，按编号 |
| `eth_getUncleCountByBlockHash` | `blockHash` | `hex string` | 叔块数量 |
| `eth_getUncleCountByBlockNumber` | `blockNumber` | `hex string` | 同上 |

> 💡 `blockNumber` 支持字符串：`"latest"`、`"earliest"`、`"pending"`、`"safe"`、`"finalized"`（EIP-1898）

---

### ✅ 2. 交易相关（Transaction）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `eth_getTransactionByHash` | `txHash` | `Transaction object` 或 `null` | 按交易哈希获取交易详情 |
| `eth_getTransactionByBlockHashAndIndex` | `blockHash`, `index` | `Transaction object` 或 `null` | 获取区块内第 index 笔交易 |
| `eth_getTransactionByBlockNumberAndIndex` | `blockNumber`, `index` | `Transaction object` 或 `null` | 同上，按编号 |
| `eth_getTransactionCount` | `address`, `blockNumber` | `hex string`（nonce） | 获取地址的交易计数（nonce） |
| `eth_getTransactionReceipt` | `txHash` | `Receipt object` 或 `null` | 获取交易收据（含 status、gasUsed、logs） |
| `eth_sendRawTransaction` | `signedTxHex` | `txHash`（hex） | 发送已签名的原始交易（最常用） |
| `eth_estimateGas` | `callObject` | `hex string` | 估算执行交易所需 gas |
| `eth_gasPrice` | — | `hex string` | 获取当前建议的 gas price（单位：wei） |

> 📌 `callObject` 示例：
```json
{
  "from": "0x...",
  "to": "0x...",
  "value": "0x123",
  "data": "0x..."
}
```

---

### ✅ 3. 账户与状态（Account & State）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `eth_getBalance` | `address`, `blockNumber` | `hex string`（wei） | 获取账户 ETH 余额 |
| `eth_getStorageAt` | `address`, `position`, `blockNumber` | `hex string` | 获取合约存储槽内容（位置为 hex） |
| `eth_getCode` | `address`, `blockNumber` | `hex string` | 获取合约字节码（部署后） |
| `eth_getProof` | `address`, `storageKeys`, `blockNumber` | `AccountProof object` | 获取 Merkle 证明（用于轻客户端验证） —— EIP-1186 |

> 🔍 `eth_getStorageAt` 示例：
```bash
# 获取合约在 slot 0 的值
curl -X POST --data '{"jsonrpc":"2.0","method":"eth_getStorageAt","params":["0x...","0x00","latest"],"id":1}' http://localhost:8545
```

---

### ✅ 4. 智能合约调用（Call & Estimate）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `eth_call` | `callObject`, `blockNumber` | `hex string`（返回数据） | 执行合约函数（不广播，只读） |
| `eth_estimateGas` | `callObject` | `hex string` | 如前所述，估算 gas 成本 |

> ✅ `eth_call` 是 DApp 查询合约数据的核心方法（如 `balanceOf()`、`name()`）  
> ❌ 不会消耗 gas，不会修改状态！

#### 示例：调用 ERC20 的 `balanceOf`
```json
{
  "jsonrpc": "2.0",
  "method": "eth_call",
  "params": [
    {
      "to": "0xdAC17F958D2ee523a2206206994597C13D831ec7", // USDT 地址
      "data": "0x70a08231000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" // balanceOf(address)
    },
    "latest"
  ],
  "id": 1
}
```

> 🔧 `data` 是 ABI 编码后的函数调用数据，可通过 Web3.js / ethers.js 自动生成。

---

### ✅ 5. 日志与事件（Logs & Filters）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `eth_getLogs` | `filterOptions` | `Log[] array` | 获取符合过滤条件的日志（非实时） |
| `eth_newFilter` | `filterOptions` | `filterId`（hex） | 创建日志过滤器（用于轮询） |
| `eth_newBlockFilter` | — | `filterId` | 创建新区块过滤器 |
| `eth_newPendingTransactionFilter` | — | `filterId` | 创建待处理交易过滤器 |
| `eth_uninstallFilter` | `filterId` | `boolean` | 删除过滤器 |
| `eth_getFilterChanges` | `filterId` | `Log[]` 或 `hash[]` | 获取自上次查询以来的新日志/区块/交易 |

#### `filterOptions` 结构：

```json
{
  "fromBlock": "0x1",
  "toBlock": "latest",
  "address": "0x...",           // 合约地址（可选）
  "topics": [                   // 事件主题（topic0 = event signature hash）
    "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef", // Transfer event
    null,                       // from address（可选）
    "0x000000000000000000000000a0b86991c6218b36c1d19d4a2e9eb0ce3606eb48" // to address
  ]
}
```

> 💡 `topics` 是 Keccak256 哈希，例如 `Transfer(address,address,uint256)` 的 topic0 是：
```js
web3.utils.sha3("Transfer(address,address,uint256)") // => "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
```

---

### ✅ 6. 网络与节点信息（Net & Web3）

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `net_version` | — | `string`（网络 ID） | 返回网络 ID（如 `"1"`=主网，`"5"`=Goerli） |
| `net_listening` | — | `boolean` | 节点是否在监听连接 |
| `net_peerCount` | — | `hex string` | 当前连接的对等节点数 |
| `web3_clientVersion` | — | `string` | 客户端名称和版本（如 `"Geth/v1.10.24-stable..."`） |
| `web3_sha3` | `dataHex` | `hex string` | 对数据做 SHA3（Keccak256）哈希（兼容性方法） |

> 📌 `net_version` 是判断连接哪个网络的关键方法，常用于钱包自动切换网络。

---

### ✅ 7. 挖矿相关（Mining）——仅限私有链/开发环境

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `miner_start` | `threads`（int） | `boolean` | 开始挖矿 |
| `miner_stop` | — | `boolean` | 停止挖矿 |
| `miner_setEtherbase` | `address` | `boolean` | 设置挖矿奖励地址 |
| `eth_mining` | — | `boolean` | 是否正在挖矿 |
| `eth_hashrate` | — | `hex string` | 当前算力（哈希/秒） |
| `eth_getWork` | — | `[header, seed, target]` | 获取工作量证明任务（PoW） |
| `eth_submitWork` | `[nonce, header, mixDigest]` | `boolean` | 提交工作结果 |

> ⚠️ 这些方法在 **以太坊合并后（Merge, 2022）已失效**，因为主网已从 PoW 转为 PoS（权益证明）。  
> 仍可用于 **Ganache、Hardhat、Local Dev Chain**。

---

### ✅ 8. 调试与开发（Debug）——仅限本地节点

| 方法名 | 参数 | 返回值 | 说明 |
|--------|------|--------|------|
| `debug_traceTransaction` | `txHash`, `traceOptions` | `TraceResult` | 跟踪交易执行过程（详细每一步） |
| `debug_traceCall` | `callObject`, `blockNum`, `options` | `TraceResult` | 跟踪模拟调用 |
| `debug_getRawHeader` | `blockHash/number` | `hex string` | 获取原始区块头（RLP 编码） |
| `debug_getRawTransaction` | `txHash` | `hex string` | 获取原始交易（RLP） |
| `debug_storageRangeAt` | `blockHash`, `txIndex`, `addr`, `startKey`, `maxResult` | `StorageRangeResult` | 获取合约存储范围 |

> 💡 `debug_traceTransaction` 是 DApp 调试失败交易的“瑞士军刀”，常用于：
> - 分析 Revert 原因
> - 查看 Gas 消耗细节
> - 复现合约逻辑错误

#### 示例：跟踪交易
```json
{
  "jsonrpc": "2.0",
  "method": "debug_traceTransaction",
  "params": [
    "0x5c504ed432cb51138bcf09aa5e8a410dd4a9e23768929573274248076146344a",
    {
      "tracer": "callTracer",
      "timeout": "5s"
    }
  ],
  "id": 1
}
```

> 输出结构复杂，包含每个调用层级（call stack）、输入输出、gas 使用等。

---

### ✅ 9. Web3 兼容方法（Legacy）

这些是为兼容早期 Web3.js 库而保留的方法，**推荐使用标准替代方法**：

| 旧方法 | 替代方法 | 说明 |
|--------|----------|------|
| `web3_sha3` | `eth_getProof` / `keccak256` | 已被 `eth_getProof` 或通用哈希库取代 |
| `eth_coinbase` | `eth_accounts` | 已废弃，改用 `eth_accounts`（仅本地） |
| `eth_accounts` | — | 返回本地账户列表（仅开发节点） |
| `eth_sign` | `personal_sign` | 签名消息（非交易） |
| `eth_sendTransaction` | `eth_sendRawTransaction` | 需要节点管理私钥（不安全，不推荐） |

> ⚠️ `eth_signTransaction` 和 `eth_sendTransaction` 要求节点持有私钥 → **极不安全**，生产环境禁止使用！

---

### ✅ 10. 订阅（Subscriptions）——WebSocket 专用

> ⚠️ **只能通过 WebSocket 连接使用**（不能用 HTTP）

| 方法 | 参数 | 返回值 | 说明 |
|------|------|--------|------|
| `eth_subscribe` | `"newHeads"`, `"newPendingTransactions"`, `"logs"`, `"syncing"` | `subscriptionId`（string） | 订阅事件 |
| `eth_unsubscribe` | `subscriptionId` | `boolean` | 取消订阅 |

#### 示例：订阅新区块头（WebSocket）
```json
{
  "id": 1,
  "method": "eth_subscribe",
  "params": ["newHeads"]
}
```

响应：
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": "0x123abc..."  // subscription id
}
```

后续推送（服务器主动发送）：
```json
{
  "jsonrpc": "2.0",
  "method": "eth_subscription",
  "params": {
    "subscription": "0x123abc...",
    "result": { ... block header object ... }
  }
}
```

#### 常见订阅类型：

| 类型 | 描述 |
|------|------|
| `newHeads` | 新区块头（包括 number, hash, parentHash, timestamp 等） |
| `newPendingTransactions` | 新的待处理交易（未被打包） |
| `logs` | 满足 filter 的合约事件日志 |
| `syncing` | 同步进度（`false` 或 `{startingBlock, currentBlock, highestBlock}`） |

> ✅ 推荐用于 DApp 实时更新：当新交易/区块到达时刷新 UI。

---

## 📌 最佳实践建议（Ethereum JSON-RPC）

| 建议 | 说明 |
|------|------|
| ✅ **使用 `eth_getTransactionReceipt` 判断交易成功** | `status == "0x1"` 表示成功，`"0x0"` 表示 reverted |
| ✅ **避免使用 `eth_sendTransaction`** | 它要求节点托管私钥，危险；改用 `eth_sendRawTransaction` + 自己签名 |
| ✅ **使用 `eth_call` 查询合约状态** | 无需 gas，安全高效 |
| ✅ **用 `eth_getLogs` + `eth_newFilter` 轮询事件** | WebSocket 更优，但 HTTP 环境下轮询是唯一选择 |
| ✅ **区分 `latest`、`pending`、`safe`、`finalized`** | 用于不同一致性要求场景（如交易所提现需 finalized） |
| ✅ **使用 Infura/Alchemy 替代自建节点** | 减少运维成本，提高可用性 |
| ✅ **缓存区块/交易数据** | 避免频繁请求，节省带宽和配额 |

---

## 🛠️ 工具推荐

| 工具 | 用途 |
|------|------|
| **[Remix IDE](https://remix.ethereum.org/)** | 内置控制台，可直接调用 JSON-RPC 方法 |
| **[Etherscan API](https://docs.etherscan.io/)** | 提供 RESTful 封装，适合前端调用 |
| **[Alchemy Notify](https://dashboard.alchemy.com/notify)** | 事件推送服务（替代手动订阅） |
| **[web3.py](https://web3py.readthedocs.io/)** / **[ethers.js](https://docs.ethers.org/)** | JavaScript/Python 封装库，隐藏底层 JSON-RPC |
| **[Postman](https://www.postman.com/)** | 测试 JSON-RPC 请求（设置 Content-Type: application/json） |

---

## 📚 官方文档参考

| 来源 | 链接 |
|------|------|
| Ethereum JSON-RPC 规范 | https://ethereum.org/en/developers/docs/apis/json-rpc/ |
| EIP-1898（块标识符） | https://eips.ethereum.org/EIPS/eip-1898 |
| EIP-1186（Merkle 证明） | https://eips.ethereum.org/EIPS/eip-1186 |
| OpenEthereum RPC 文档 | https://openethereum.github.io/JSONRPC |
| Geth JSON-RPC 文档 | https://geth.ethereum.org/docs/rpc/server |

---

## ✅ 总结：Ethereum JSON-RPC 是什么？

> **Ethereum JSON-RPC 是开发者与以太坊区块链交互的“语言”**，它让你能够：
>
> - 查询余额、合约数据
> - 发送交易
> - 监听事件
> - 调试智能合约
> - 构建钱包、DApp、链上分析工具

无论你使用 **MetaMask、Uniswap、OpenSea、Chainlink、Aave**，其背后都是通过 JSON-RPC 与节点通信。

掌握这些方法，你就掌握了 **以太坊开发的底层钥匙**！

---

如果你需要：
- 某个方法的 **完整请求/响应示例**
- **如何用 ethers.js 调用 eth_call**
- **如何搭建本地节点并测试这些方法**
- **如何用 Python + web3.py 实现订阅**

欢迎继续提问，我可以为你生成完整可运行代码模板！

