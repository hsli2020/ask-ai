# 编写一个php函数：callAPI($method, $url, $headers = [], $data = null)

以下是一个健壮、安全且符合现代实践的 PHP API 调用函数，支持自动数据序列化、Header 处理、错误诊断，并附带详细注释：

```php
<?php
/**
 * 通用 HTTP API 调用函数
 * 
 * 功能特点：
 * - 自动处理 JSON/array 数据序列化并设置 Content-Type
 * - 严格区分 GET/HEAD（不发送 Body）与其他方法
 * - 完整错误诊断（cURL 错误 + HTTP 状态码）
 * - 安全 Header 处理（避免重复/冲突）
 * - 启用 SSL 验证（生产环境安全）
 * - 返回结构化结果便于调用方处理
 * 
 * @param string $method HTTP 方法 (GET, POST, PUT, DELETE 等)
 * @param string $url    请求地址（需 urlencode 处理参数）
 * @param array  $headers 自定义请求头（关联数组或字符串数组）
 * @param mixed  $data    请求体（字符串/数组/对象，GET/HEAD 时忽略）
 * @return array          ['success'=>bool, 'http_code'=>int, 'body'=>string, 'error'=>string, 'headers'=>array]
 */
function callAPI($method, $url, $headers = [], $data = null) {
    $ch = curl_init();
    $result = [
        'success'   => false,
        'http_code' => 0,
        'body'      => '',
        'error'     => '',
        'headers'   => []
    ];

    // ===== 1. 基础配置 =====
    curl_setopt_array($ch, [
        CURLOPT_URL            => $url,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_FOLLOWLOCATION => true,
        CURLOPT_TIMEOUT        => 30,
        CURLOPT_SSL_VERIFYPEER => true, // 安全：验证 SSL 证书
        CURLOPT_SSL_VERIFYHOST => 2,
        CURLOPT_HEADER         => true, // 用于后续分离响应头
    ]);

    // ===== 2. 处理请求方法 =====
    $methodUpper = strtoupper(trim($method));
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $methodUpper);

    // ===== 3. 处理请求体与 Content-Type（仅非 GET/HEAD）=====
    $isBodyMethod = !in_array($methodUpper, ['GET', 'HEAD', 'OPTIONS', 'TRACE']);
    $needsJson = false;

    if ($isBodyMethod && $data !== null) {
        // 自动序列化数组/对象为 JSON
        if (is_array($data) || is_object($data)) {
            $jsonData = json_encode($data, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
            if ($jsonData === false) {
                curl_close($ch);
                $result['error'] = 'JSON encode failed: ' . json_last_error_msg();
                return $result;
            }
            $data = $jsonData;
            $needsJson = true; // 标记需检查 Content-Type
        }

        // 仅当 $data 为非空字符串时设置请求体
        if (is_string($data) && $data !== '') {
            curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
        }
    }

    // ===== 4. 构建并设置请求头 =====
    $formattedHeaders = [];
    $hasContentType = false;

    foreach ($headers as $key => $value) {
        if (is_string($key)) {
            // 处理关联数组: ['Content-Type' => '...']
            $headerLine = trim($key) . ': ' . trim($value);
            $formattedHeaders[] = $headerLine;
            if (stripos($key, 'content-type') !== false) {
                $hasContentType = true;
            }
        } else {
            // 处理字符串数组: ['Content-Type: ...']
            $headerLine = trim($value);
            $formattedHeaders[] = $headerLine;
            if (stripos($headerLine, 'content-type:') === 0) {
                $hasContentType = true;
            }
        }
    }

    // 自动补充 Content-Type（仅当需要且未设置时）
    if ($needsJson && !$hasContentType) {
        $formattedHeaders[] = 'Content-Type: application/json; charset=utf-8';
    }

    if (!empty($formattedHeaders)) {
        curl_setopt($ch, CURLOPT_HTTPHEADER, $formattedHeaders);
    }

    // ===== 5. 执行请求 =====
    $rawResponse = curl_exec($ch);
    $curlError = curl_error($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);

    // ===== 6. 错误处理 =====
    if ($rawResponse === false || $curlError) {
        $result['error'] = 'cURL Error (#' . curl_errno($ch) . '): ' . ($curlError ?: 'Unknown error');
        curl_close($ch);
        return $result;
    }

    // ===== 7. 分离响应头与响应体 =====
    $responseHeaders = substr($rawResponse, 0, $headerSize);
    $result['body'] = substr($rawResponse, $headerSize);
    $result['http_code'] = $httpCode;

    // 简易解析响应头（每行一个）
    $result['headers'] = array_filter(array_map('trim', explode("\r\n", $responseHeaders)));

    // ===== 8. 设置成功标志（业务层错误如 4xx/5xx 仍视为"成功执行"）=====
    $result['success'] = true;
    if ($httpCode >= 400) {
        $result['error'] = "HTTP {$httpCode}: Server returned error status";
    }

    curl_close($ch);
    return $result;
}

// ============ 使用示例 ============
/*
// POST JSON
$res = callAPI('POST', 'https://api.example.com/users', [], ['name' => 'John']);
if ($res['success'] && $res['http_code'] == 201) {
    $data = json_decode($res['body'], true);
    // 处理 $data
} else {
    error_log("API Error: {$res['error']} (HTTP {$res['http_code']})");
}

// GET（参数应已拼在 URL 中）
$res = callAPI('GET', 'https://api.example.com/users?id=123');

// 带自定义 Header
$res = callAPI(
    'DELETE',
    'https://api.example.com/resource/1',
    ['Authorization' => 'Bearer xyz', 'X-Custom' => 'value']
);
*/
?>
```

### 核心设计说明：
1. **安全优先**  
   - 启用 `CURLOPT_SSL_VERIFYPEER` 和 `CURLOPT_SSL_VERIFYHOST`，避免中间人攻击
   - 不自动关闭 SSL 验证（开发环境需自行配置 CURLOPT_CAINFO）

2. **智能数据处理**  
   - 数组/对象 → 自动 `json_encode` + 补充 `Content-Type: application/json`
   - 保留原始字符串处理能力（支持 XML/form-data 等）
   - GET/HEAD/OPTIONS/TRACE 严格不发送 Body（符合 RFC 标准）

3. **健壮的错误反馈**  
   - 区分 cURL 网络错误 与 HTTP 业务错误
   - 返回完整结构：`success`（执行状态）、`http_code`、`body`、`error`、`headers`
   - 4xx/5xx 状态码不中断流程，由调用方决策处理逻辑

4. **Header 灵活处理**  
   - 兼容关联数组 `['Key' => 'Val']` 和字符串数组 `['Key: Val']`
   - 智能检测并避免重复设置 Content-Type

5. **生产就绪**  
   - 超时控制（30秒）
   - 响应头解析便于调试
   - 详细注释与使用示例

> **提示**：  
> - GET/HEAD 请求的查询参数**必须**由调用方拼接到 `$url` 中  
> - 如需禁用 SSL 验证（仅限测试环境），可在函数内添加 `CURLOPT_SSL_VERIFYPEER => false`（不推荐）  
> - 复杂场景建议使用 Guzzle 等专业 HTTP 客户端库


# 我想把这个函数修改一下，callAPI($method, $url, $data = []); 
header、body、query Params等都通过data传入，甚至可以传入更多参数，如debug，logger等
首先实现这个函数，然后对比两种方案的优劣

我们重新设计函数，将header、body、query等都通过$data参数传入，并支持debug、logger等扩展参数。
 函数签名：callAPI($method, $url, $data = [])
 $data结构示例：
   [
     'headers' => ['Authorization' => 'Bearer xxx'], // 请求头（关联数组或字符串数组）
     'body' => ['key' => 'value'],                  // 请求体（自动根据Content-Type序列化）
     'query' => ['id' => 123],                      // 查询参数（自动合并到URL）
     'debug' => true,                               // 是否开启调试模式
     'logger' => function($msg) { ... },            // 日志回调（可选）
     'timeout' => 30,                               // 超时时间（秒）
     'ssl_verify' => true,                          // 是否验证SSL（默认true）
     // ... 其他可扩展参数
   ]

 实现步骤：
 1. 提取$data中的各个配置项，设置默认值
 2. 处理URL：将query参数合并到$url中（注意保留原URL的查询参数和fragment）
 3. 处理请求头：从$data['headers']获取，并准备后续设置Content-Type
 4. 处理请求体：
      - 判断请求方法是否允许携带body（非GET/HEAD/OPTIONS/TRACE）
      - 根据headers中的Content-Type或自动推断，序列化body
      - 若body为数组/对象且未指定Content-Type，默认使用JSON
 5. 构建cURL选项，应用超时、SSL验证等
 6. 若开启debug，通过logger记录请求详情
 7. 执行请求，处理响应
 8. 若开启debug，记录响应详情
 9. 返回结构化结果（含success, http_code, body, error, headers等）

 注意：需谨慎处理URL合并（避免重复参数）、Content-Type推断、错误处理等。

 下面开始编写代码。
</think>

以下为重构后的 `callAPI` 函数，采用统一 `$data` 参数承载所有配置，并附带方案对比分析：

```php
<?php
/**
 * 通用 HTTP API 调用函数（统一配置版）
 * 
 * @param string $method HTTP 方法（GET/POST/PUT/DELETE 等）
 * @param string $url    基础请求地址（查询参数将自动合并）
 * @param array  $data   统一配置数组，支持：
 *                       - headers: ['Key' => 'Val'] 或 ['Key: Val']
 *                       - body: 请求体（数组/对象/字符串）
 *                       - query: 查询参数数组（自动合并到 URL）
 *                       - debug: bool（开启调试日志）
 *                       - logger: callable（日志回调函数，接收 string）
 *                       - timeout: int（超时秒数，默认30）
 *                       - ssl_verify: bool（是否验证 SSL，默认 true）
 *                       - curl_opts: array（额外 cURL 选项，谨慎使用）
 * @return array         ['success'=>bool, 'http_code'=>int, 'body'=>string, 'error'=>string, 'headers'=>array, 'debug_info'=>?array]
 */
function callAPI($method, $url, $data = []) {
    // ===== 1. 提取配置项（带默认值）=====
    $headers     = $data['headers'] ?? [];
    $body        = $data['body'] ?? null;
    $queryParams = $data['query'] ?? [];
    $debug       = !empty($data['debug']);
    $logger      = is_callable($data['logger'] ?? null) ? $data['logger'] : null;
    $timeout     = intval($data['timeout'] ?? 30) ?: 30;
    $sslVerify   = !isset($data['ssl_verify']) || !empty($data['ssl_verify']);
    $curlOpts    = $data['curl_opts'] ?? [];

    // ===== 2. 合并查询参数到 URL =====
    if (!empty($queryParams)) {
        $urlParts = parse_url($url);
        if ($urlParts === false) {
            return _buildResult(false, 0, '', 'Invalid base URL', []);
        }
        
        $existingQuery = [];
        if (!empty($urlParts['query'])) {
            parse_str($urlParts['query'], $existingQuery);
        }
        $mergedQuery = array_merge($existingQuery, $queryParams);
        $newQuery = http_build_query($mergedQuery);
        
        $scheme = $urlParts['scheme'] ?? '';
        $host = $urlParts['host'] ?? '';
        $port = isset($urlParts['port']) ? ':' . $urlParts['port'] : '';
        $path = $urlParts['path'] ?? '';
        $fragment = !empty($urlParts['fragment']) ? '#' . $urlParts['fragment'] : '';
        
        $url = ($scheme ? $scheme . '://' : '') . $host . $port . $path 
             . ($newQuery ? '?' . $newQuery : '') . $fragment;
    }

    // ===== 3. 初始化 cURL =====
    $ch = curl_init();
    $result = [
        'success'    => false,
        'http_code'  => 0,
        'body'       => '',
        'error'      => '',
        'headers'    => [],
        'debug_info' => $debug ? [
            'request_url' => $url,
            'request_method' => strtoupper($method),
            'request_headers' => [],
            'request_body' => null,
            'curl_error' => null,
            'response_time' => 0
        ] : null
    ];

    // ===== 4. 处理请求体与 Content-Type =====
    $methodUpper = strtoupper(trim($method));
    $isBodyMethod = !in_array($methodUpper, ['GET', 'HEAD', 'OPTIONS', 'TRACE'], true);
    $autoContentType = null;
    $originalBody = $body;

    if ($isBodyMethod && $body !== null) {
        // 检测已有 Content-Type
        $hasContentType = false;
        foreach ($headers as $k => $v) {
            $key = is_string($k) ? $k : (is_string($v) ? explode(':', $v, 2)[0] : '');
            if (stripos(trim($key), 'content-type') !== false) {
                $hasContentType = true;
                break;
            }
        }

        // 序列化 body
        if (is_array($body) || is_object($body)) {
            if (!$hasContentType) {
                // 无 Content-Type 时默认 JSON
                $body = json_encode($body, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);
                $autoContentType = 'application/json; charset=utf-8';
            } else {
                // 根据已有 Content-Type 序列化
                $ct = '';
                foreach ($headers as $k => $v) {
                    if (is_string($k) && stripos($k, 'content-type') !== false) {
                        $ct = $v; break;
                    } elseif (is_string($v) && stripos($v, 'content-type:') === 0) {
                        $ct = trim(substr($v, strpos($v, ':') + 1)); break;
                    }
                }
                if (stripos($ct, 'application/x-www-form-urlencoded') !== false) {
                    $body = http_build_query($body);
                } elseif (stripos($ct, 'application/json') !== false) {
                    $body = json_encode($body, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES | JSON_THROW_ON_ERROR);
                } // 其他类型由调用者自行处理
            }
        }
        // 字符串 body 直接使用
    } else {
        $body = null; // 确保非 body 方法不发送数据
    }

    // ===== 5. 构建最终请求头 =====
    $finalHeaders = [];
    $hasContentTypeFinal = false;
    foreach ($headers as $k => $v) {
        if (is_string($k)) {
            $line = trim($k) . ': ' . trim($v);
            $finalHeaders[] = $line;
            if (stripos($k, 'content-type') !== false) $hasContentTypeFinal = true;
        } else {
            $line = trim($v);
            $finalHeaders[] = $line;
            if (stripos($line, 'content-type:') === 0) $hasContentTypeFinal = true;
        }
    }
    if ($autoContentType && !$hasContentTypeFinal) {
        $finalHeaders[] = 'Content-Type: ' . $autoContentType;
    }

    // ===== 6. 配置 cURL =====
    $curlConfig = [
        CURLOPT_URL            => $url,
        CURLOPT_CUSTOMREQUEST  => $methodUpper,
        CURLOPT_RETURNTRANSFER => true,
        CURLOPT_HEADER         => true,
        CURLOPT_FOLLOWLOCATION => true,
        CURLOPT_TIMEOUT        => $timeout,
        CURLOPT_SSL_VERIFYPEER => $sslVerify,
        CURLOPT_SSL_VERIFYHOST => ($sslVerify ? 2 : 0),
        CURLOPT_HTTPHEADER     => $finalHeaders,
    ];

    // 应用 body（仅 body 方法）
    if ($isBodyMethod && $body !== null && $body !== '') {
        $curlConfig[CURLOPT_POSTFIELDS] = $body;
    }

    // 应用用户自定义 cURL 选项（谨慎！）
    foreach ($curlOpts as $opt => $val) {
        if (is_int($opt)) $curlConfig[$opt] = $val;
    }

    curl_setopt_array($ch, $curlConfig);

    // ===== 7. 调试日志（请求）=====
    if ($debug) {
        $result['debug_info']['request_headers'] = $finalHeaders;
        $result['debug_info']['request_body'] = is_string($originalBody) ? substr($originalBody, 0, 500) : json_encode($originalBody, JSON_UNESCAPED_UNICODE);
        $logMsg = sprintf(
            "[API REQUEST] %s %s\nHeaders: %s\nBody: %s",
            $methodUpper,
            $url,
            json_encode($finalHeaders),
            $result['debug_info']['request_body']
        );
        _logDebug($logMsg, $logger);
    }

    // ===== 8. 执行请求 =====
    $startTime = microtime(true);
    $rawResponse = curl_exec($ch);
    $execTime = round(microtime(true) - $startTime, 4);
    $curlErrno = curl_errno($ch);
    $curlError = curl_error($ch);
    $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);

    if ($debug) {
        $result['debug_info']['response_time'] = $execTime;
        $result['debug_info']['curl_error'] = $curlError ?: null;
    }

    // ===== 9. 错误处理 =====
    if ($rawResponse === false || $curlErrno) {
        $errMsg = sprintf('cURL Error (#%d): %s', $curlErrno, $curlError ?: 'Unknown error');
        curl_close($ch);
        if ($debug) _logDebug("[API ERROR] $errMsg", $logger);
        return _buildResult(false, $httpCode, '', $errMsg, [], $result['debug_info']);
    }

    // ===== 10. 解析响应 =====
    $responseHeadersRaw = substr($rawResponse, 0, $headerSize);
    $responseBody = substr($rawResponse, $headerSize);
    $responseHeaders = array_filter(array_map('trim', explode("\r\n", $responseHeadersRaw)));

    // ===== 11. 调试日志（响应）=====
    if ($debug) {
        $logMsg = sprintf(
            "[API RESPONSE] HTTP/%d | Time: %.4fs\nHeaders: %s\nBody (500 chars): %s",
            $httpCode,
            $execTime,
            json_encode($responseHeaders),
            substr($responseBody, 0, 500)
        );
        _logDebug($logMsg, $logger);
    }

    // ===== 12. 构建结果 =====
    $success = ($httpCode >= 200 && $httpCode < 300);
    if (!$success && empty($result['error'])) {
        $result['error'] = "HTTP {$httpCode}: Request failed";
    }

    curl_close($ch);
    return _buildResult($success, $httpCode, $responseBody, $result['error'], $responseHeaders, $result['debug_info']);
}

// ===== 辅助函数 =====
function _buildResult($success, $httpCode, $body, $error, $headers, $debugInfo = null) {
    $res = [
        'success'   => $success,
        'http_code' => $httpCode,
        'body'      => $body,
        'error'     => $error,
        'headers'   => $headers
    ];
    if ($debugInfo !== null) $res['debug_info'] = $debugInfo;
    return $res;
}

function _logDebug($message, $logger = null) {
    if ($logger) {
        call_user_func($logger, $message);
    } else {
        error_log($message);
    }
}
?>
```

---

## 📊 两种方案深度对比

| **维度**         | **方案A：分离参数**<br>`callAPI($m, $u, $h=[], $d=null)` | **方案B：统一$data**<br>`callAPI($m, $u, $data=[])` |
|------------------|--------------------------------------------------------|-----------------------------------------------------|
| **调用简洁性**   | ✅ 简单场景直观（如纯GET）                             | ✅ 复杂场景更清晰（所有配置集中）                  |
| **参数扩展性**   | ❌ 新增参数需改函数签名（破坏兼容）                    | ✅ 无限扩展（debug/logger/timeout等无缝加入）      |
| **可读性**       | ⚠️ 多参数时易混淆（尤其$null占位）                    | ✅ 语义明确（`'query' => [...]` 一目了然）         |
| **类型安全**     | ✅ 参数类型明确（IDE友好）                             | ❌ $data为关联数组，需文档/注释说明结构            |
| **框架集成**     | ⚠️ 需适配框架参数规范                                  | ✅ 天然契合现代框架风格（Laravel/Guzzle 风格）     |
| **调试能力**     | ❌ 需额外封装                                          | ✅ 内置debug/logger，开箱即用                      |
| **学习成本**     | ✅ 低（传统风格）                                      | ⚠️ 中（需理解$data结构）                           |
| **适用场景**     | 简单脚本、内部工具、参数固定场景                       | 中大型项目、需日志/监控/配置管理的生产环境         |

---

## 💡 选型建议

- **选择方案B（统一$data）当**：
  - 项目需长期维护、频繁迭代
  - 需要调试日志、监控、配置管理
  - 与现代PHP框架（Laravel/Symfony）集成
  - 团队协作需清晰语义（`'query'` 比位置参数更直观）

- **保留方案A（分离参数）当**：
  - 极简脚本（如单次curl调用）
  - 遗留系统兼容
  - 团队严格遵循“显式参数优于魔法数组”原则

> **最佳实践**：在新项目中优先采用方案B。其设计哲学与 Guzzle、Symfony HttpClient 等现代库一致，通过结构化配置提升可维护性，同时通过辅助函数（`_logDebug`）保持代码整洁。对于简单调用，仍可写作：  
> `callAPI('GET', $url, ['query' => ['id'=>1]]);`  
> 语义清晰且无冗余。

