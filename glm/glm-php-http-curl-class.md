## 编写一个PHP类，封装常用的curl客户端函数调用，提供一个清晰易用的接口，用来调用各
种API，类似这种写法：
$response = Http::request('GET https://www.example.com/')->send();
$value = $response->header('a-key');
$body = $reponse->body;



我来为您创建一个功能完整且易用的HTTP客户端类，支持链式调用和常见的HTTP操作需求。

```php
<?php

/**
 * HTTP客户端类 - 封装cURL操作
 * 支持链式调用，提供简洁的API接口
 */
class Http
{
    /**
     * 创建一个新的HTTP请求
     * 
     * @param string $url 请求URL
     * @return static
     */
    public static function request(string $url = ''): self
    {
        return new static($url);
    }

    /**
     * GET请求快捷方法
     * 
     * @param string $url 请求URL
     * @return static
     */
    public static function get(string $url): self
    {
        return (new static($url))->method('GET');
    }

    /**
     * POST请求快捷方法
     * 
     * @param string $url 请求URL
     * @param array $data 请求数据
     * @return static
     */
    public static function post(string $url, array $data = []): self
    {
        return (new static($url))->method('POST')->data($data);
    }

    /**
     * PUT请求快捷方法
     * 
     * @param string $url 请求URL
     * @param array $data 请求数据
     * @return static
     */
    public static function put(string $url, array $data = []): self
    {
        return (new static($url))->method('PUT')->data($data);
    }

    /**
     * DELETE请求快捷方法
     * 
     * @param string $url 请求URL
     * @return static
     */
    public static function delete(string $url): self
    {
        return (new static($url))->method('DELETE');
    }

    /**
     * PATCH请求快捷方法
     * 
     * @param string $url 请求URL
     * @param array $data 请求数据
     * @return static
     */
    public static function patch(string $url, array $data = []): self
    {
        return (new static($url))->method('PATCH')->data($data);
    }

    // ===== 实例属性和方法 =====

    private string $url = '';
    private string $method = 'GET';
    private array $headers = [];
    private array $query = [];
    private $body = null;
    private array $options = [];
    private array $files = [];
    private int $timeout = 30;
    private bool $verifySSL = true;
    private ?string $proxy = null;

    public function __construct(string $url = '')
    {
        $this->url = $url;
        $this->setDefaultHeaders();
    }

    /**
     * 设置请求方法
     * 
     * @param string $method HTTP方法
     * @return self
     */
    public function method(string $method): self
    {
        $this->method = strtoupper($method);
        return $this;
    }

    /**
     * 设置请求头
     * 
     * @param string|array $key 头名称或头数组
     * @param string|null $value 头值
     * @return self
     */
    public function header($key, ?string $value = null): self
    {
        if (is_array($key)) {
            $this->headers = array_merge($this->headers, $key);
        } else {
            $this->headers[$key] = $value;
        }
        return $this;
    }

    /**
     * 批量设置请求头
     * 
     * @param array $headers 头数组
     * @return self
     */
    public function headers(array $headers): self
    {
        $this->headers = array_merge($this->headers, $headers);
        return $this;
    }

    /**
     * 设置查询参数
     * 
     * @param string|array $key 参数名或参数数组
     * @param mixed $value 参数值
     * @return self
     */
    public function query($key, $value = null): self
    {
        if (is_array($key)) {
            $this->query = array_merge($this->query, $key);
        } else {
            $this->query[$key] = $value;
        }
        return $this;
    }

    /**
     * 设置请求体数据
     * 
     * @param mixed $data 数据
     * @return self
     */
    public function data($data): self
    {
        $this->body = $data;
        return $this;
    }

    /**
     * 设置JSON数据
     * 
     * @param mixed $data 数据
     * @return self
     */
    public function json($data): self
    {
        $this->body = json_encode($data);
        $this->header('Content-Type', 'application/json');
        return $this;
    }

    /**
     * 添加文件上传
     * 
     * @param string $name 字段名
     * @param string $path 文件路径
     * @param string|null $filename 文件名
     * @param string|null $mimeType MIME类型
     * @return self
     */
    public function file(string $name, string $path, ?string $filename = null, ?string $mimeType = null): self
    {
        $this->files[$name] = [
            'path' => $path,
            'filename' => $filename ?: basename($path),
            'mimeType' => $mimeType
        ];
        return $this;
    }

    /**
     * 设置超时时间（秒）
     * 
     * @param int $seconds 超时时间
     * @return self
     */
    public function timeout(int $seconds): self
    {
        $this->timeout = $seconds;
        return $this;
    }

    /**
     * 设置SSL验证
     * 
     * @param bool $verify 是否验证
     * @return self
     */
    public function verify(bool $verify = true): self
    {
        $this->verifySSL = $verify;
        return $this;
    }

    /**
     * 设置代理
     * 
     * @param string $proxy 代理地址，如：http://proxy:port
     * @return self
     */
    public function proxy(string $proxy): self
    {
        $this->proxy = $proxy;
        return $this;
    }

    /**
     * 设置Bearer Token认证
     * 
     * @param string $token Token值
     * @return self
     */
    public function bearer(string $token): self
    {
        $this->header('Authorization', 'Bearer ' . $token);
        return $this;
    }

    /**
     * 设置Basic认证
     * 
     * @param string $username 用户名
     * @param string $password 密码
     * @return self
     */
    public function basic(string $username, string $password): self
    {
        $this->header('Authorization', 'Basic ' . base64_encode("$username:$password"));
        return $this;
    }

    /**
     * 设置自定义cURL选项
     * 
     * @param int|array $option 选项名或选项数组
     * @param mixed $value 选项值
     * @return self
     */
    public function option($option, $value = null): self
    {
        if (is_array($option)) {
            $this->options = array_merge($this->options, $option);
        } else {
            $this->options[$option] = $value;
        }
        return $this;
    }

    /**
     * 发送请求
     * 
     * @return Response
     * @throws HttpException
     */
    public function send(): Response
    {
        // 构建完整URL
        $url = $this->buildUrl();
        
        // 初始化cURL
        $ch = curl_init($url);
        
        // 设置基本选项
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HEADER, true);
        curl_setopt($ch, CURLOPT_TIMEOUT, $this->timeout);
        curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $this->verifySSL);
        curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $this->verifySSL ? 2 : 0);
        
        // 设置请求方法
        $this->setMethod($ch);
        
        // 设置请求头
        if (!empty($this->headers)) {
            curl_setopt($ch, CURLOPT_HTTPHEADER, $this->formatHeaders());
        }
        
        // 设置请求体
        $this->setBody($ch);
        
        // 设置代理
        if ($this->proxy) {
            curl_setopt($ch, CURLOPT_PROXY, $this->proxy);
        }
        
        // 设置自定义选项
        foreach ($this->options as $option => $value) {
            curl_setopt($ch, $option, $value);
        }
        
        // 执行请求
        $response = curl_exec($ch);
        
        // 检查错误
        if ($response === false) {
            $error = curl_error($ch);
            $errno = curl_errno($ch);
            curl_close($ch);
            throw new HttpException("cURL Error [$errno]: $error");
        }
        
        // 获取响应信息
        $info = curl_getinfo($ch);
        curl_close($ch);
        
        // 解析响应
        return $this->parseResponse($response, $info);
    }

    /**
     * 异步发送请求（需要安装curl_multi扩展）
     * 
     * @param array $requests 请求数组
     * @return array 响应数组
     */
    public static function sendAsync(array $requests): array
    {
        $multiHandle = curl_multi_init();
        $curlHandles = [];
        
        // 初始化所有cURL句柄
        foreach ($requests as $index => $request) {
            if (!$request instanceof self) {
                throw new HttpException("All requests must be Http instances");
            }
            
            $url = $request->buildUrl();
            $ch = curl_init($url);
            
            // 设置选项
            curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
            curl_setopt($ch, CURLOPT_HEADER, true);
            curl_setopt($ch, CURLOPT_TIMEOUT, $request->timeout);
            curl_setopt($ch, CURLOPT_SSL_VERIFYPEER, $request->verifySSL);
            curl_setopt($ch, CURLOPT_SSL_VERIFYHOST, $request->verifySSL ? 2 : 0);
            
            $request->setMethod($ch);
            
            if (!empty($request->headers)) {
                curl_setopt($ch, CURLOPT_HTTPHEADER, $request->formatHeaders());
            }
            
            $request->setBody($ch);
            
            if ($request->proxy) {
                curl_setopt($ch, CURLOPT_PROXY, $request->proxy);
            }
            
            foreach ($request->options as $option => $value) {
                curl_setopt($ch, $option, $value);
            }
            
            curl_multi_add_handle($multiHandle, $ch);
            $curlHandles[$index] = $ch;
        }
        
        // 执行所有请求
        $running = null;
        do {
            curl_multi_exec($multiHandle, $running);
            curl_multi_select($multiHandle);
        } while ($running > 0);
        
        // 收集响应
        $responses = [];
        foreach ($curlHandles as $index => $ch) {
            $response = curl_multi_getcontent($ch);
            $info = curl_getinfo($ch);
            $responses[$index] = $requests[$index]->parseResponse($response, $info);
            curl_multi_remove_handle($multiHandle, $ch);
            curl_close($ch);
        }
        
        curl_multi_close($multiHandle);
        return $responses;
    }

    // ===== 私有方法 =====

    /**
     * 设置默认请求头
     */
    private function setDefaultHeaders(): void
    {
        $this->headers = [
            'User-Agent' => 'PHP-Http-Client/1.0',
            'Accept' => '*/*',
            'Connection' => 'close',
        ];
    }

    /**
     * 构建完整URL
     * 
     * @return string
     */
    private function buildUrl(): string
    {
        $url = $this->url;
        
        if (!empty($this->query)) {
            $queryString = http_build_query($this->query);
            $url .= (strpos($url, '?') === false ? '?' : '&') . $queryString;
        }
        
        return $url;
    }

    /**
     * 设置请求方法
     * 
     * @param resource $ch cURL句柄
     */
    private function setMethod($ch): void
    {
        switch ($this->method) {
            case 'POST':
                curl_setopt($ch, CURLOPT_POST, true);
                break;
            case 'PUT':
            case 'DELETE':
            case 'PATCH':
                curl_setopt($ch, CURLOPT_CUSTOMREQUEST, $this->method);
                break;
        }
    }

    /**
     * 设置请求体
     * 
     * @param resource $ch cURL句柄
     */
    private function setBody($ch): void
    {
        if ($this->body !== null || !empty($this->files)) {
            if (!empty($this->files)) {
                // 文件上传
                $data = $this->body ? (is_array($this->body) ? $this->body : []) : [];
                
                foreach ($this->files as $name => $file) {
                    if (function_exists('curl_file_create')) {
                        $data[$name] = curl_file_create(
                            $file['path'],
                            $file['mimeType'] ?? null,
                            $file['filename']
                        );
                    } else {
                        $data[$name] = '@' . $file['path'];
                    }
                }
                
                curl_setopt($ch, CURLOPT_POSTFIELDS, $data);
            } else {
                curl_setopt($ch, CURLOPT_POSTFIELDS, $this->body);
            }
        }
    }

    /**
     * 格式化请求头
     * 
     * @return array
     */
    private function formatHeaders(): array
    {
        $formatted = [];
        foreach ($this->headers as $key => $value) {
            $formatted[] = "$key: $value";
        }
        return $formatted;
    }

    /**
     * 解析响应
     * 
     * @param string $response 原始响应
     * @param array $info cURL信息
     * @return Response
     */
    private function parseResponse(string $response, array $info): Response
    {
        $headerSize = $info['header_size'];
        $headerText = substr($response, 0, $headerSize);
        $body = substr($response, $headerSize);
        
        // 解析响应头
        $headers = [];
        $headerLines = explode("\r\n", $headerText);
        array_shift($headerLines); // 移除状态行
        
        foreach ($headerLines as $line) {
            if (trim($line) === '') continue;
            
            $parts = explode(':', $line, 2);
            if (count($parts) === 2) {
                $headers[trim($parts[0])] = trim($parts[1]);
            }
        }
        
        return new Response(
            $info['http_code'],
            $headers,
            $body,
            $info
        );
    }
}

/**
 * HTTP响应类
 */
class Response
{
    private int $statusCode;
    private array $headers;
    private string $body;
    private array $info;
    private ?array $jsonData = null;

    public function __construct(int $statusCode, array $headers, string $body, array $info)
    {
        $this->statusCode = $statusCode;
        $this->headers = $headers;
        $this->body = $body;
        $this->info = $info;
    }

    /**
     * 获取状态码
     * 
     * @return int
     */
    public function status(): int
    {
        return $this->statusCode;
    }

    /**
     * 获取响应头
     * 
     * @param string|null $key 头名称，为null时返回所有头
     * @return mixed
     */
    public function header(?string $key = null)
    {
        if ($key === null) {
            return $this->headers;
        }
        
        return $this->headers[$key] ?? null;
    }

    /**
     * 获取响应体
     * 
     * @return string
     */
    public function body(): string
    {
        return $this->body;
    }

    /**
     * 获取JSON数据
     * 
     * @param bool $assoc 是否返回关联数组
     * @return mixed
     */
    public function json(bool $assoc = true)
    {
        if ($this->jsonData === null) {
            $this->jsonData = json_decode($this->body, $assoc);
        }
        return $this->jsonData;
    }

    /**
     * 获取响应信息
     * 
     * @param string|null $key 信息键名
     * @return mixed
     */
    public function info(?string $key = null)
    {
        if ($key === null) {
            return $this->info;
        }
        
        return $this->info[$key] ?? null;
    }

    /**
     * 判断请求是否成功（2xx状态码）
     * 
     * @return bool
     */
    public function successful(): bool
    {
        return $this->statusCode >= 200 && $this->statusCode < 300;
    }

    /**
     * 判断是否为客户端错误（4xx状态码）
     * 
     * @return bool
     */
    public function clientError(): bool
    {
        return $this->statusCode >= 400 && $this->statusCode < 500;
    }

    /**
     * 判断是否为服务器错误（5xx状态码）
     * 
     * @return bool
     */
    public function serverError(): bool
    {
        return $this->statusCode >= 500;
    }

    /**
     * 判断是否失败（4xx或5xx状态码）
     * 
     * @return bool
     */
    public function failed(): bool
    {
        return $this->clientError() || $this->serverError();
    }

    /**
     * 获取响应体（魔术方法）
     * 
     * @return string
     */
    public function __toString(): string
    {
        return $this->body;
    }

    /**
     * 获取响应体（魔术属性）
     * 
     * @return string
     */
    public function __get(string $name)
    {
        if ($name === 'body') {
            return $this->body;
        }
        
        if ($name === 'status') {
            return $this->statusCode;
        }
        
        return null;
    }
}

/**
 * HTTP异常类
 */
class HttpException extends Exception
{
    // 可以添加更多HTTP特定的异常处理
}

// ===== 使用示例 =====

/*
// 基本GET请求
$response = Http::get('https://api.example.com/users')
    ->query(['page' => 1, 'limit' => 10])
    ->header('Accept', 'application/json')
    ->send();

if ($response->successful()) {
    $users = $response->json();
    echo "Total users: " . count($users);
}

// POST请求
$response = Http::post('https://api.example.com/users', [
    'name' => 'John Doe',
    'email' => 'john@example.com'
])
    ->header('Authorization', 'Bearer your-token')
    ->send();

// 文件上传
$response = Http::post('https://api.example.com/upload')
    ->file('avatar', '/path/to/avatar.jpg')
    ->data(['user_id' => 123])
    ->send();

// JSON请求
$response = Http::put('https://api.example.com/users/123')
    ->json([
        'name' => 'Jane Doe',
        'email' => 'jane@example.com'
    ])
    ->bearer('your-jwt-token')
    ->send();

// 异步请求
$requests = [
    Http::get('https://api.example.com/users'),
    Http::get('https://api.example.com/posts'),
    Http::get('https://api.example.com/comments')
];

$responses = Http::sendAsync($requests);
foreach ($responses as $response) {
    echo $response->status() . "\n";
}

// 代理和SSL设置
$response = Http::get('https://api.example.com/data')
    ->proxy('http://proxy.example.com:8080')
    ->verify(false)
    ->timeout(60)
    ->send();

// 错误处理
try {
    $response = Http::get('https://api.example.com/data')->send();
    if ($response->failed()) {
        echo "Request failed with status: " . $response->status();
    }
} catch (HttpException $e) {
    echo "HTTP Error: " . $e->getMessage();
}
*/
```

这个HTTP客户端类提供了以下特性：

## 主要功能

1. **链式调用** - 支持流畅的API设计
2. **多种HTTP方法** - GET、POST、PUT、DELETE、PATCH
3. **灵活的数据发送** - 支持表单数据、JSON、文件上传
4. **完整的响应处理** - 状态码、头部、响应体
5. **认证支持** - Bearer Token、Basic Auth
6. **高级功能** - 代理、SSL验证、超时设置
7. **异步请求** - 支持并发请求

## 使用示例

```php
// 简单GET请求
$response = Http::get('https://api.example.com/users')
    ->query(['page' => 1])
    ->send();

// POST JSON数据
$response = Http::post('https://api.example.com/users')
    ->json(['name' => 'John', 'email' => 'john@example.com'])
    ->bearer('token')
    ->send();

// 获取响应数据
if ($response->successful()) {
    $data = $response->json();
    $contentType = $response->header('Content-Type');
    echo $response->body;
}
```

这个实现提供了专业级的HTTP客户端功能，可以满足大多数API调用需求。
