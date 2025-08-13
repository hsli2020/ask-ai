# 请帮我编写一个PHP类，实现常用的http方法调用，代码示例如下：
$response = Http::request("GET http://www.example.com')
->header('content-type', 'application/json')
->send();

要求：
1、链式调用
2、支持所有http方法，get/post/put等等
3、错误处理，timeout


以下是一个完整的PHP HTTP客户端类，支持链式调用和所有HTTP方法：

```php
<?php

class Http
{
    private $method;
    private $url;
    private $headers = [];
    private $data = null;
    private $timeout = 30;
    private $options = [];

    /**
     * 构造函数
     * @param string $method HTTP方法
     * @param string $url 请求URL
     */
    public function __construct($method = 'GET', $url = '')
    {
        $this->method = strtoupper($method);
        $this->url = $url;
    }

    /**
     * 静态工厂方法，创建HTTP请求实例
     * @param string $request 请求字符串，格式如 "GET http://example.com"
     * @return Http
     */
    public static function request($request)
    {
        $parts = explode(' ', $request, 2);
        $method = $parts[0];
        $url = isset($parts[1]) ? $parts[1] : '';
        
        return new self($method, $url);
    }

    /**
     * 设置GET请求
     * @param string $url
     * @return Http
     */
    public static function get($url)
    {
        return new self('GET', $url);
    }

    /**
     * 设置POST请求
     * @param string $url
     * @return Http
     */
    public static function post($url)
    {
        return new self('POST', $url);
    }

    /**
     * 设置PUT请求
     * @param string $url
     * @return Http
     */
    public static function put($url)
    {
        return new self('PUT', $url);
    }

    /**
     * 设置DELETE请求
     * @param string $url
     * @return Http
     */
    public static function delete($url)
    {
        return new self('DELETE', $url);
    }

    /**
     * 设置PATCH请求
     * @param string $url
     * @return Http
     */
    public static function patch($url)
    {
        return new self('PATCH', $url);
    }

    /**
     * 设置HEAD请求
     * @param string $url
     * @return Http
     */
    public static function head($url)
    {
        return new self('HEAD', $url);
    }

    /**
     * 添加请求头
     * @param string $key 头部键名
     * @param string $value 头部值
     * @return Http
     */
    public function header($key, $value)
    {
        $this->headers[$key] = $value;
        return $this;
    }

    /**
     * 批量添加请求头
     * @param array $headers 头部数组
     * @return Http
     */
    public function headers($headers)
    {
        $this->headers = array_merge($this->headers, $headers);
        return $this;
    }

    /**
     * 设置请求数据
     * @param mixed $data 请求数据
     * @return Http
     */
    public function data($data)
    {
        $this->data = $data;
        return $this;
    }

    /**
     * 设置请求体（JSON数据）
     * @param mixed $data JSON数据
     * @return Http
     */
    public function json($data)
    {
        $this->data = is_string($data) ? $data : json_encode($data);
        $this->header('Content-Type', 'application/json');
        return $this;
    }

    /**
     * 设置超时时间
     * @param int $seconds 超时秒数
     * @return Http
     */
    public function timeout($seconds)
    {
        $this->timeout = $seconds;
        return $this;
    }

    /**
     * 设置其他cURL选项
     * @param array $options cURL选项数组
     * @return Http
     */
    public function options($options)
    {
        $this->options = array_merge($this->options, $options);
        return $this;
    }

    /**
     * 发送HTTP请求
     * @return HttpResponse
     * @throws HttpException
     */
    public function send()
    {
        if (empty($this->url)) {
            throw new HttpException('URL is required');
        }

        // 初始化cURL
        $ch = curl_init();

        // 设置基本选项
        $curlOptions = [
            CURLOPT_URL => $this->url,
            CURLOPT_RETURNTRANSFER => true,
            CURLOPT_TIMEOUT => $this->timeout,
            CURLOPT_CONNECTTIMEOUT => $this->timeout,
            CURLOPT_CUSTOMREQUEST => $this->method,
            CURLOPT_HEADER => true,
            CURLOPT_NOBODY => ($this->method === 'HEAD'),
        ];

        // 设置请求头
        if (!empty($this->headers)) {
            $headers = [];
            foreach ($this->headers as $key => $value) {
                $headers[] = $key . ': ' . $value;
            }
            $curlOptions[CURLOPT_HTTPHEADER] = $headers;
        }

        // 设置请求数据
        if ($this->data !== null && in_array($this->method, ['POST', 'PUT', 'PATCH', 'DELETE'])) {
            if (is_array($this->data) || is_object($this->data)) {
                $curlOptions[CURLOPT_POSTFIELDS] = is_array($this->data) ? http_build_query($this->data) : json_encode($this->data);
            } else {
                $curlOptions[CURLOPT_POSTFIELDS] = $this->data;
            }
        }

        // 合并自定义选项
        $curlOptions = array_merge($curlOptions, $this->options);

        // 应用cURL选项
        curl_setopt_array($ch, $curlOptions);

        // 执行请求
        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        $error = curl_error($ch);
        $errno = curl_errno($ch);

        // 关闭cURL
        curl_close($ch);

        // 处理错误
        if ($errno !== 0) {
            throw new HttpException("cURL Error ({$errno}): {$error}", $errno);
        }

        if ($response === false) {
            throw new HttpException('Failed to execute request');
        }

        // 解析响应
        return $this->parseResponse($response, $httpCode);
    }

    /**
     * 解析HTTP响应
     * @param string $response 原始响应
     * @param int $httpCode HTTP状态码
     * @return HttpResponse
     */
    private function parseResponse($response, $httpCode)
    {
        $headerSize = strpos($response, "\r\n\r\n");
        if ($headerSize === false) {
            $headerSize = 0;
        } else {
            $headerSize += 4;
        }

        $headerString = substr($response, 0, $headerSize);
        $body = substr($response, $headerSize);

        // 解析响应头
        $headers = [];
        $headerLines = explode("\r\n", trim($headerString));
        foreach ($headerLines as $line) {
            if (strpos($line, ':') !== false) {
                list($key, $value) = explode(':', $line, 2);
                $headers[trim($key)] = trim($value);
            }
        }

        return new HttpResponse($httpCode, $body, $headers);
    }
}

/**
 * HTTP响应类
 */
class HttpResponse
{
    private $statusCode;
    private $body;
    private $headers;

    public function __construct($statusCode, $body, $headers)
    {
        $this->statusCode = $statusCode;
        $this->body = $body;
        $this->headers = $headers;
    }

    /**
     * 获取状态码
     * @return int
     */
    public function getStatusCode()
    {
        return $this->statusCode;
    }

    /**
     * 获取响应体
     * @return string
     */
    public function getBody()
    {
        return $this->body;
    }

    /**
     * 获取响应头
     * @param string|null $key 头部键名，如果为null则返回所有头部
     * @return string|array|null
     */
    public function getHeader($key = null)
    {
        if ($key === null) {
            return $this->headers;
        }
        return isset($this->headers[$key]) ? $this->headers[$key] : null;
    }

    /**
     * 获取JSON格式的响应体
     * @return array|null
     */
    public function getJson()
    {
        $json = json_decode($this->body, true);
        return json_last_error() === JSON_ERROR_NONE ? $json : null;
    }

    /**
     * 检查请求是否成功
     * @return bool
     */
    public function isSuccess()
    {
        return $this->statusCode >= 200 && $this->statusCode < 300;
    }

    /**
     * 检查是否为重定向
     * @return bool
     */
    public function isRedirect()
    {
        return $this->statusCode >= 300 && $this->statusCode < 400;
    }

    /**
     * 检查是否有客户端错误
     * @return bool
     */
    public function isClientError()
    {
        return $this->statusCode >= 400 && $this->statusCode < 500;
    }

    /**
     * 检查是否有服务器错误
     * @return bool
     */
    public function isServerError()
    {
        return $this->statusCode >= 500 && $this->statusCode < 600;
    }

    /**
     * 转换为字符串
     * @return string
     */
    public function __toString()
    {
        return $this->body;
    }
}

/**
 * HTTP异常类
 */
class HttpException extends Exception
{
    public function __construct($message = "", $code = 0, Exception $previous = null)
    {
        parent::__construct($message, $code, $previous);
    }
}

// 使用示例
try {
    // 示例1: 基本GET请求
    $response = Http::request("GET http://httpbin.org/get")
        ->header('Content-Type', 'application/json')
        ->timeout(10)
        ->send();
    
    echo "Status: " . $response->getStatusCode() . "\n";
    echo "Body: " . $response->getBody() . "\n";
    echo "Success: " . ($response->isSuccess() ? 'Yes' : 'No') . "\n";

    // 示例2: POST请求
    $response = Http::post("http://httpbin.org/post")
        ->json(['name' => 'John', 'age' => 30])
        ->timeout(15)
        ->send();
    
    $json = $response->getJson();
    if ($json) {
        echo "JSON Response: " . json_encode($json, JSON_PRETTY_PRINT) . "\n";
    }

    // 示例3: PUT请求
    $response = Http::put("http://httpbin.org/put")
        ->data(['id' => 123, 'status' => 'active'])
        ->header('Authorization', 'Bearer token123')
        ->send();

    echo "PUT Status: " . $response->getStatusCode() . "\n";

} catch (HttpException $e) {
    echo "HTTP Error: " . $e->getMessage() . "\n";
} catch (Exception $e) {
    echo "General Error: " . $e->getMessage() . "\n";
}
```

这个HTTP客户端类具有以下特性：

## 主要功能：

1. **链式调用** - 所有方法都返回`$this`，支持流畅的链式调用
2. **支持所有HTTP方法** - GET, POST, PUT, DELETE, PATCH, HEAD
3. **灵活的请求构建** - 可以通过静态方法或`request()`方法创建请求
4. **完整的错误处理** - 包括超时、网络错误等
5. **丰富的响应处理** - 提供状态码、响应体、响应头的访问方法

## 主要方法：

- `Http::request()` - 通过字符串创建请求
- `Http::get()`, `Http::post()`, 等 - 直接创建对应方法的请求
- `header()` - 添加单个请求头
- `headers()` - 批量添加请求头
- `data()` - 设置请求数据
- `json()` - 设置JSON数据并自动设置Content-Type
- `timeout()` - 设置超时时间
- `send()` - 发送请求并返回HttpResponse对象

## 响应处理：

- `getStatusCode()` - 获取HTTP状态码
- `getBody()` - 获取原始响应体
- `getHeader()` - 获取响应头
- `getJson()` - 获取解析后的JSON数据
- `isSuccess()`, `isRedirect()` 等 - 状态检查方法

## 错误处理：

- `HttpException` - 专门的HTTP异常类
- 自动处理超时、网络错误、cURL错误等

使用这个类可以轻松地进行各种HTTP请求操作，代码简洁且功能完整。

