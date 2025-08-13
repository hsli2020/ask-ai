# 编写一个PHP函数，callAPI($method, $url, $data)，$data 是一个数组，可以包含queryParams, headers, body等参数，
# 甚至可以包含debug, logger等方便调试的参数，函数的返回值要包含充分的信息

这是一个设计完善、功能丰富的 PHP cURL 封装函数。它支持灵活的参数配置、详细的返回信息、以及方便的调试和日志功能。

### 函数代码

```php
<?php

/**
 * 通用 API 调用函数
 *
 * @param string $method HTTP 方法 (GET, POST, PUT, DELETE, etc.)
 * @param string $url 请求的 URL
 * @param array $data 参数数组，支持以下键：
 *      - 'queryParams' => array, URL 查询参数 (?key=value)
 *      - 'headers'     => array, HTTP 头信息
 *      - 'body'        => array|string, 请求体内容
 *      - 'timeout'     => int, 超时时间（秒），默认 30
 *      - 'debug'       => bool, 是否开启调试模式，默认 false
 *      - 'logger'      => callable, 日志记录回调函数
 *
 * @return array 包含详细信息的响应数组
 */
function callAPI($method, $url, $data = []) {
    // 1. 参数初始化与默认值设置
    $queryParams = $data['queryParams'] ?? [];
    $headers     = $data['headers'] ?? [];
    $body        = $data['body'] ?? null;
    $timeout     = $data['timeout'] ?? 30;
    $debug       = $data['debug'] ?? false;
    $logger      = $data['logger'] ?? null;

    // 2. 处理 URL 和查询参数
    if (!empty($queryParams)) {
        $queryString = http_build_query($queryParams);
        // 检查 URL 是否已包含问号
        $url .= (strpos($url, '?') === false ? '?' : '&') . $queryString;
    }

    // 3. 初始化 cURL
    $ch = curl_init();
    
    // 基础 cURL 设置
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true); // 返回字符串而非直接输出
    curl_setopt($ch, CURLOPT_TIMEOUT, $timeout);
    curl_setopt($ch, CURLOPT_FOLLOWLOCATION, true); // 跟随重定向
    curl_setopt($ch, CURLOPT_HEADER, true); // 在输出中包含 Header
    curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $method);

    // 4. 处理请求头
    // 如果没有显式设置 Content-Type，且 body 是数组，默认当作 application/x-www-form-urlencoded 处理
    // 如果希望默认 JSON，可取消下方的注释或手动传入 header
    $formattedHeaders = [];
    foreach ($headers as $key => $val) {
        $formattedHeaders[] = "$key: $val";
    }
    curl_setopt($ch, CURLOPT_HTTPHEADER, $formattedHeaders);

    // 5. 处理请求体
    if (in_array($method, ['POST', 'PUT', 'PATCH']) && !empty($body)) {
        // 如果 body 是数组，根据 Content-Type 决定如何编码
        if (is_array($body)) {
            // 检查 headers 中是否包含 application/json
            $isJson = false;
            foreach ($headers as $h) {
                if (stripos($h, 'application/json') !== false) {
                    $isJson = true;
                    break;
                }
            }
            // 默认行为：如果未指定 json，则自动转为 json (现代 API 常用做法)
            // 你也可以根据需要改为 http_build_query
            $body = $isJson ? $body : json_encode($body);
            if (!$isJson) {
                curl_setopt($ch, CURLOPT_HTTPHEADER, array_merge($formattedHeaders, ['Content-Type: application/json']));
            }
        }
        curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
    }

    // 6. 调试模式设置
    if ($debug) {
        curl_setopt($ch, CURLOPT_VERBOSE, true);
        // 如果需要将 verbose 输出到临时文件而不是 STDERR
        // $verbose = fopen('php://temp', 'w+');
        // curl_setopt($ch, CURLOPT_STDERR, $verbose);
    }

    // 7. 执行请求
    $startTime = microtime(true);
    $responseRaw = curl_exec($ch);
    $endTime = microtime(true);
    $executionTime = round(($endTime - $startTime) * 1000, 2); // 毫秒

    // 8. 处理结果
    $result = [
        'success'      => false,
        'status_code'  => 0,
        'headers'      => [],
        'body'         => null,
        'raw_response' => $responseRaw,
        'error'        => null,
        'request_info' => [
            'url'            => $url,
            'method'         => $method,
            'payload'        => $body,
            'execution_time' => $executionTime . ' ms'
        ]
    ];

    if ($responseRaw === false) {
        $result['error'] = curl_error($ch);
        $result['error_code'] = curl_errno($ch);
    } else {
        // 分离 Header 和 Body
        $headerSize = curl_getinfo($ch, CURLINFO_HEADER_SIZE);
        $headerContent = substr($responseRaw, 0, $headerSize);
        $bodyContent = substr($responseRaw, $headerSize);

        $result['status_code'] = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $result['headers'] = parseHeaders($headerContent);
        $result['body'] = $bodyContent;
        
        // 自动尝试解析 JSON
        if (is_string($bodyContent) && !empty($bodyContent)) {
             // 检查 Content-Type 是否为 JSON
             $contentType = $result['headers']['Content-Type'] ?? '';
             if (stripos($contentType, 'application/json') !== false) {
                 $decoded = json_decode($bodyContent, true);
                 if (json_last_error() === JSON_ERROR_NONE) {
                     $result['body'] = $decoded;
                 }
             }
        }

        $result['success'] = ($result['status_code'] >= 200 && $result['status_code'] < 300);
    }

    // 9. 获取详细的请求调试信息
    if ($debug) {
        $result['debug'] = curl_getinfo($ch);
        // 如果使用了临时文件存储 verbose，这里可以读取
    }

    curl_close($ch);

    // 10. 调用日志记录器
    if (is_callable($logger)) {
        call_user_func($logger, $result);
    }

    return $result;
}

/**
 * 辅助函数：解析响应头字符串为数组
 */
function parseHeaders($headerContent) {
    $headers = [];
    $lines = explode("\r\n", trim($headerContent));
    foreach ($lines as $line) {
        if (strpos($line, ':') !== false) {
            list($key, $value) = explode(':', $line, 2);
            $headers[trim($key)] = trim($value);
        }
    }
    return $headers;
}
```

### 使用示例

#### 1. 基础 POST 请求 (JSON API)
```php
$response = callAPI('POST', 'https://api.example.com/users', [
    'headers' => [
        'Authorization' => 'Bearer YOUR_TOKEN',
        'Content-Type'  => 'application/json'
    ],
    'body' => [
        'name'  => 'John Doe',
        'email' => 'john@example.com'
    ]
]);

if ($response['success']) {
    print_r($response['body']); // 已经自动解析为 PHP 数组
} else {
    echo "Error: " . $response['error'];
}
```

#### 2. GET 请求带查询参数
```php
$response = callAPI('GET', 'https://api.example.com/search', [
    'queryParams' => [
        'q' => 'php curl',
        'page' => 1
    ],
    'headers' => [
        'Accept' => 'application/json'
    ]
]);
// 实际请求 URL 会变成: https://api.example.com/search?q=php+curl&page=1
```

#### 3. 开启调试与日志
```php
$myLogger = function($result) {
    // 将请求信息写入文件或发送到监控系统
    $logMsg = "[" . date('Y-m-d H:i:s') . "] " . 
              $result['request_info']['method'] . " " . 
              $result['request_info']['url'] . 
              " - Status: " . $result['status_code'] . 
              " - Time: " . $result['request_info']['execution_time'] . "\n";
    file_put_contents('api.log', $logMsg, FILE_APPEND);
};

$response = callAPI('GET', 'https://httpbin.org/get', [
    'debug' => true,
    'logger' => $myLogger,
    'timeout' => 10
]);

// 调试模式下，返回数组中会包含详细的 curl_getinfo 信息
print_r($response['debug']);
```

### 函数设计亮点

1.  **返回值结构清晰**：返回数组包含 `success` (布尔值)、`status_code` (状态码)、`body` (自动处理过的内容)、`error` (错误信息) 以及 `request_info` (请求元数据)。
2.  **自动处理 JSON**：如果响应头包含 `application/json`，函数会自动将 `body` 解析为数组；请求时如果传入数组 body，默认也会转为 JSON 格式发送（符合现代 API 习惯）。
3.  **调试友好**：
    *   `debug` 参数：开启后会返回 cURL 的详细连接信息（DNS 解析时间、SSL 握手时间等）。
    *   `request_info`：默认会记录请求耗时，方便性能分析。
4.  **扩展性强**：`logger` 参数允许传入闭包，可以在不修改函数内部逻辑的情况下，将日志写入文件、数据库或发送到 Sentry 等监控系统。
5.  **健壮性**：处理了 URL 拼接（自动判断是否已有 `?`）、Header 分割、超时设置等常见边缘情况。

Http::sendRequest($method, $url, $data);
