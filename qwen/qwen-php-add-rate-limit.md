# 如何为PHP开发的Web API应用引入Rate Limit机制

在 PHP 开发的 Web API 应用中引入 **Rate Limit（限流）机制**，可以有效防止滥用、DDoS 攻击或资源耗尽。以下是几种常见的实现方式，从简单到复杂，适用于不同规模和需求的项目。

---

## 一、基本原理

Rate Limit 通常基于以下维度：

- **IP 地址**
- **API Key / Token**
- **用户 ID（认证后）**

限制策略包括：

- 固定窗口（Fixed Window）
- 滑动窗口（Sliding Window）
- 令牌桶（Token Bucket）
- 漏桶（Leaky Bucket）

---

## 二、使用 Redis 实现滑动窗口限流（推荐）

Redis 性能高、支持原子操作，是实现限流的理想选择。

### 示例：每分钟最多 100 次请求（按 IP 限流）

```php
<?php
// rate_limit.php

function isRateLimited($ip, $maxRequests = 100, $windowSeconds = 60) {
    $redis = new Redis();
    $redis->connect('127.0.0.1', 6379);

    $key = "rate_limit:$ip";
    $now = time();

    // 移除窗口外的旧记录（滑动窗口）
    $redis->zremrangebyscore($key, 0, $now - $windowSeconds);

    // 获取当前窗口内请求数
    $requestCount = $redis->zcard($key);

    if ($requestCount >= $maxRequests) {
        return true; // 被限流
    }

    // 添加当前请求时间戳
    $redis->zadd($key, $now, uniqid());
    $redis->expire($key, $windowSeconds); // 自动过期

    return false;
}

// 使用示例
$ip = $_SERVER['REMOTE_ADDR'];
if (isRateLimited($ip, 100, 60)) {
    http_response_code(429);
    echo json_encode(['error' => 'Too Many Requests']);
    exit;
}
```

> ✅ 优点：精确、支持高并发  
> ⚠️ 注意：`uniqid()` 用于避免重复 score 导致覆盖，也可用微秒时间戳。

---

## 三、使用中间件（Laravel / Symfony 等框架）

### Laravel 示例（使用内置限流）

Laravel 自带 `throttle` 中间件：

```php
// routes/api.php
Route::middleware('throttle:100,1')->group(function () {
    Route::get('/data', [ApiController::class, 'getData']);
});
```

- `100,1` 表示每 1 分钟最多 100 次请求。
- 默认按 IP + 用户 ID（如果已认证）限流。

你也可以自定义限流键（如按 API Key）：

```php
// app/Http/Middleware/CustomThrottle.php
use Illuminate\Cache\RateLimiting\Limit;
use Illuminate\Support\Facades\RateLimiter;

RateLimiter::for('api-key', function ($request) {
    return Limit::perMinute(100)->by($request->header('X-API-Key'));
});

// 路由中使用
Route::middleware('throttle:api-key')->get('/data', ...);
```

---

## 四、轻量级纯 PHP 实现（无 Redis，仅用于开发/低并发）

> ⚠️ 不推荐用于生产环境（无法跨进程、重启清空）

```php
session_start();

function simpleRateLimit($max = 10, $window = 60) {
    $now = time();
    $key = 'rate_limit_' . $_SERVER['REMOTE_ADDR'];

    if (!isset($_SESSION[$key])) {
        $_SESSION[$key] = [];
    }

    // 清理过期记录
    $_SESSION[$key] = array_filter($_SESSION[$key], fn($t) => $t > $now - $window);

    if (count($_SESSION[$key]) >= $max) {
        return true;
    }

    $_SESSION[$key][] = $now;
    return false;
}

if (simpleRateLimit(10, 60)) {
    http_response_code(429);
    exit('Too Many Requests');
}
```

---

## 五、返回标准限流响应头（RFC 6585）

增强 API 友好性，返回如下 HTTP 头：

```php
$remaining = $maxRequests - $requestCount;
$resetTime = $now + $windowSeconds - ($now % $windowSeconds);

header('X-RateLimit-Limit: ' . $maxRequests);
header('X-RateLimit-Remaining: ' . max(0, $remaining));
header('X-RateLimit-Reset: ' . $resetTime);
```

客户端可根据这些头进行重试控制。

---

## 六、安全建议

1. **不要仅依赖 IP**：NAT 或代理下多个用户共享同一 IP。
2. **优先使用 API Key 或 Token**：更精准识别调用者。
3. **日志记录被限流的请求**：用于分析攻击行为。
4. **结合 WAF 或 CDN**：如 Cloudflare、阿里云 WAF 提供边缘限流。

---

## 七、扩展方案

- 使用 **APCu** 替代 Redis（单机场景）
- 集成 **Guzzle** 的限流中间件（客户端限流）
- 使用 **Swoole** + 内存表实现高性能限流（常驻内存）

---

## 总结

| 方案 | 适用场景 | 优点 | 缺点 |
|------|--------|------|------|
| Redis + 滑动窗口 | 生产环境、高并发 | 精确、可扩展 | 需要 Redis |
| Laravel Throttle | Laravel 项目 | 开箱即用 | 框架绑定 |
| Session 限流 | 开发/测试 | 无需额外依赖 | 不可靠、不跨进程 |

> 推荐：**生产环境务必使用 Redis 实现滑动窗口限流**，并配合标准响应头。

如需完整代码示例或集成到特定框架（如 Slim、CodeIgniter），可进一步说明。

# 用纯PHP实现 固定窗口，令牌桶，漏桶 算法

以下是使用 **纯 PHP（不依赖 Redis、数据库等外部服务）** 实现的三种经典限流算法：**固定窗口（Fixed Window）**、**令牌桶（Token Bucket）** 和 **漏桶（Leaky Bucket）**。这些实现适用于单机、低并发或开发测试场景，**不推荐用于高并发生产环境**（因无法跨进程共享状态，且重启会丢失数据），但有助于理解算法原理。

---

## 📌 公共基础：模拟“全局”存储

由于纯 PHP 无持久内存，我们使用 **APCu**（推荐）或 **文件/静态变量** 模拟状态存储。  
> ✅ 若服务器支持 APCu（PHP 扩展），优先使用；否则用 `static` 变量（仅限单次请求内有效，仅用于演示）。

为通用性，以下示例使用 **APCu**（若未安装，可替换为文件缓存或说明限制）。

```php
// 检查 APCu 是否可用
if (!extension_loaded('apcu')) {
    // 降级到静态变量（⚠️ 仅限单次请求内有效，实际无法限流！仅用于演示逻辑）
    class FallbackStorage {
        private static $data = [];
        public static function get($key) { return self::$data[$key] ?? null; }
        public static function set($key, $value, $ttl = 0) { self::$data[$key] = $value; }
        public static function delete($key) { unset(self::$data[$key]); }
    }
    function apcu_get($key) { return FallbackStorage::get($key); }
    function apcu_set($key, $value, $ttl = 0) { return FallbackStorage::set($key, $value, $ttl); }
    function apcu_delete($key) { return FallbackStorage::delete($key); }
}
```

> ⚠️ 注意：**静态变量方案在真实 Web 请求中无效**（每个请求独立）。生产环境必须用 APCu、Redis 或共享内存。

---

## 1️⃣ 固定窗口（Fixed Window）

- **原理**：将时间划分为固定长度窗口（如 60 秒），在窗口内计数，超限则拒绝。
- **缺点**：窗口边界处可能突发双倍流量（如 59 秒发 100 次，0 秒又发 100 次）。

```php
class FixedWindowRateLimiter
{
    private string $key;
    private int $maxRequests;
    private int $windowSeconds;

    public function __construct(string $key, int $maxRequests, int $windowSeconds)
    {
        $this->key = $key;
        $this->maxRequests = $maxRequests;
        $this->windowSeconds = $windowSeconds;
    }

    public function allow(): bool
    {
        $now = time();
        $windowStart = floor($now / $this->windowSeconds) * $this->windowSeconds;
        $storageKey = "fixed_window:{$this->key}:{$windowStart}";

        $count = apcu_fetch($storageKey);
        if ($count === false) {
            // 新窗口，初始化计数
            apcu_add($storageKey, 1, $this->windowSeconds);
            return true;
        }

        if ($count >= $this->maxRequests) {
            return false; // 超限
        }

        // 原子递增（APCu 的 increment 是原子的）
        apcu_inc($storageKey);
        return true;
    }
}

// 使用示例
$limiter = new FixedWindowRateLimiter('ip_192.168.1.1', 5, 60); // 每分钟最多 5 次
if (!$limiter->allow()) {
    http_response_code(429);
    exit('Too Many Requests (Fixed Window)');
}
echo "Request allowed!";
```

---

## 2️⃣ 令牌桶（Token Bucket）

- **原理**：以固定速率向桶中添加令牌，请求需消耗令牌。桶有容量上限，令牌满则丢弃。
- **优点**：允许突发流量（只要桶中有令牌）。

```php
class TokenBucketRateLimiter
{
    private string $key;
    private float $capacity;      // 桶容量
    private float $refillRate;    // 每秒生成令牌数
    private float $refillInterval; // 计算精度（秒），越小越精确

    public function __construct(string $key, float $capacity, float $refillRate, float $refillInterval = 1.0)
    {
        $this->key = $key;
        $this->capacity = $capacity;
        $this->refillRate = $refillRate;
        $this->refillInterval = $refillInterval;
    }

    public function allow(int $tokens = 1): bool
    {
        $storageKey = "token_bucket:{$this->key}";
        $now = microtime(true);

        // 获取当前状态 [lastRefillTime, currentTokens]
        $state = apcu_fetch($storageKey);
        if ($state === false) {
            // 初始化：满桶 + 当前时间
            $state = [$now, $this->capacity];
            apcu_store($storageKey, $state, 3600); // TTL 1小时
        }

        [$lastRefill, $currentTokens] = $state;

        // 计算自上次 refill 后新增的令牌数
        $elapsed = $now - $lastRefill;
        $newTokens = $elapsed * $this->refillRate;
        $currentTokens = min($this->capacity, $currentTokens + $newTokens);

        if ($currentTokens < $tokens) {
            // 令牌不足，更新状态（防止负数）
            apcu_store($storageKey, [$now, $currentTokens], 3600);
            return false;
        }

        // 消耗令牌
        $currentTokens -= $tokens;
        apcu_store($storageKey, [$now, $currentTokens], 3600);
        return true;
    }
}

// 使用示例：每秒生成 2 个令牌，桶容量 10（允许突发 10 次）
$limiter = new TokenBucketRateLimiter('api_key_abc', 10, 2);
if (!$limiter->allow()) {
    http_response_code(429);
    exit('Too Many Requests (Token Bucket)');
}
echo "Request allowed!";
```

---

## 3️⃣ 漏桶（Leaky Bucket）

- **原理**：请求进入“桶”，以固定速率“漏出”（处理）。桶满则拒绝新请求。
- **效果**：强制平滑输出，不允许突发。

```php
class LeakyBucketRateLimiter
{
    private string $key;
    private int $capacity;     // 桶容量（队列长度）
    private float $leakRate;   // 每秒漏出请求数（即处理速率）

    public function __construct(string $key, int $capacity, float $leakRate)
    {
        $this->key = $key;
        $this->capacity = $capacity;
        $this->leakRate = $leakRate;
    }

    public function allow(): bool
    {
        $storageKey = "leaky_bucket:{$this->key}";
        $now = microtime(true);

        // 获取队列 [lastLeakTime, queueSize]
        $state = apcu_fetch($storageKey);
        if ($state === false) {
            $state = [$now, 0]; // 初始空桶
            apcu_store($storageKey, $state, 3600);
        }

        [$lastLeak, $queueSize] = $state;

        // 模拟“漏水”：计算自上次 leak 后应处理的请求数
        $elapsed = $now - $lastLeak;
        $leaked = $elapsed * $this->leakRate;
        $queueSize = max(0, $queueSize - $leaked);

        // 检查桶是否已满
        if ($queueSize >= $this->capacity) {
            apcu_store($storageKey, [$now, $queueSize], 3600);
            return false;
        }

        // 新请求入队
        $queueSize += 1;
        apcu_store($storageKey, [$now, $queueSize], 3600);
        return true;
    }
}

// 使用示例：桶容量 5，每秒处理 1 个请求
$limiter = new LeakyBucketRateLimiter('user_123', 5, 1);
if (!$limiter->allow()) {
    http_response_code(429);
    exit('Too Many Requests (Leaky Bucket)');
}
echo "Request allowed!";
```

---

## 🔍 对比总结

| 算法 | 突发流量支持 | 平滑性 | 实现复杂度 | 适用场景 |
|------|-------------|--------|-----------|---------|
| **固定窗口** | ❌ 边界处可能双倍 | 差 | ⭐ | 简单场景，容忍边界问题 |
| **令牌桶** | ✅（桶内令牌允许突发） | 中 | ⭐⭐ | 需要突发能力的 API |
| **漏桶** | ❌（强制平滑） | 优 | ⭐⭐ | 严格匀速处理（如消息队列） |

---

## ⚠️ 重要提醒

1. **纯 PHP 限流的局限性**：
   - 多进程/多服务器无法共享状态（需 Redis/Memcached）。
   - PHP-FPM 每次请求结束后状态重置（除非用 APCu/共享内存）。
   - **静态变量方案仅用于算法演示，实际无效！**

2. **生产环境建议**：
   - 使用 **Redis + Lua 脚本** 实现原子操作。
   - 框架集成（如 Laravel Throttle、Symfony RateLimiter）。
   - CDN/WAF 层限流（Cloudflare、阿里云）。

3. **APCu 安装**（如需）：
   ```bash
   # Ubuntu/Debian
   sudo apt install php-apcu
   # CentOS/RHEL
   sudo yum install php-pecl-apcu
   ```

---

如需 **Redis 版本** 或 **与特定框架集成** 的实现，请告知！


# 用PHP+Redis实现 固定窗口，令牌桶，漏桶 算法

以下是使用 **PHP + Redis** 实现的三种经典限流算法：**固定窗口（Fixed Window）**、**令牌桶（Token Bucket）** 和 **漏桶（Leaky Bucket）**。这些实现具备**高并发安全性**（通过 Redis 原子操作）、**跨进程/服务器共享状态**，适用于生产环境。

> ✅ **前提**：已安装 Redis 并可通过 PHP 访问（推荐使用 `phpredis` 扩展）。

---

## 🔧 公共 Redis 连接封装

```php
// RedisClient.php
class RedisClient
{
    private static ?Redis $instance = null;

    public static function getInstance(): Redis
    {
        if (self::$instance === null) {
            self::$instance = new Redis();
            self::$instance->connect('127.0.0.1', 6379);
            // 如需密码：self::$instance->auth('password');
        }
        return self::$instance;
    }
}
```

---

## 1️⃣ 固定窗口（Fixed Window）

- **原理**：以固定时间窗口（如 60 秒）计数，超限拒绝。
- **Redis 结构**：`String`（键名包含窗口起始时间）
- **原子操作**：`INCR` + `EXPIRE`

```php
// FixedWindowLimiter.php
class FixedWindowLimiter
{
    private string $keyPrefix;
    private int $maxRequests;
    private int $windowSeconds;

    public function __construct(string $keyPrefix, int $maxRequests, int $windowSeconds)
    {
        $this->keyPrefix = $keyPrefix;
        $this->maxRequests = $maxRequests;
        $this->windowSeconds = $windowSeconds;
    }

    public function allow(string $identifier): bool
    {
        $redis = RedisClient::getInstance();
        $now = time();
        $windowStart = floor($now / $this->windowSeconds) * $this->windowSeconds;
        $key = "fixed_window:{$this->keyPrefix}:{$identifier}:{$windowStart}";

        // 原子递增并设置过期时间
        $count = $redis->incr($key);
        if ($count === 1) {
            $redis->expire($key, $this->windowSeconds);
        }

        return $count <= $this->maxRequests;
    }
}

// 使用示例
$limiter = new FixedWindowLimiter('api', 100, 60); // 每分钟最多 100 次
if (!$limiter->allow($_SERVER['REMOTE_ADDR'])) {
    http_response_code(429);
    exit('Too Many Requests (Fixed Window)');
}
echo "Request allowed!";
```

> ✅ **优点**：简单高效  
> ⚠️ **缺点**：窗口边界可能突发双倍流量（如 59 秒发 100 次，0 秒又发 100 次）

---

## 2️⃣ 令牌桶（Token Bucket）

- **原理**：以固定速率生成令牌，请求消耗令牌。桶有容量上限。
- **Redis 结构**：`Hash` 存储 `{ last_refill_time, tokens }`
- **原子操作**：Lua 脚本保证一致性

```php
// TokenBucketLimiter.php
class TokenBucketLimiter
{
    private string $keyPrefix;
    private float $capacity;
    private float $refillRate; // 每秒生成令牌数

    public function __construct(string $keyPrefix, float $capacity, float $refillRate)
    {
        $this->keyPrefix = $keyPrefix;
        $this->capacity = $capacity;
        $this->refillRate = $refillRate;
    }

    public function allow(string $identifier, int $tokens = 1): bool
    {
        $redis = RedisClient::getInstance();
        $key = "token_bucket:{$this->keyPrefix}:{$identifier}";
        $now = microtime(true);

        // Lua 脚本：原子计算令牌
        $script = '
            local key = KEYS[1]
            local now = tonumber(ARGV[1])
            local capacity = tonumber(ARGV[2])
            local refill_rate = tonumber(ARGV[3])
            local requested_tokens = tonumber(ARGV[4])

            -- 获取当前状态
            local last_refill = redis.call("HGET", key, "last_refill")
            local current_tokens = redis.call("HGET", key, "tokens")

            if not last_refill then
                -- 初始化：满桶
                last_refill = now
                current_tokens = capacity
            else
                last_refill = tonumber(last_refill)
                current_tokens = tonumber(current_tokens)
                -- 计算新增令牌
                local elapsed = now - last_refill
                local new_tokens = elapsed * refill_rate
                current_tokens = math.min(capacity, current_tokens + new_tokens)
            end

            if current_tokens < requested_tokens then
                -- 令牌不足，更新状态（防止负数）
                redis.call("HMSET", key, "last_refill", now, "tokens", current_tokens)
                redis.call("EXPIRE", key, 3600) -- 1小时TTL
                return 0
            end

            -- 消耗令牌
            current_tokens = current_tokens - requested_tokens
            redis.call("HMSET", key, "last_refill", now, "tokens", current_tokens)
            redis.call("EXPIRE", key, 3600)
            return 1
        ';

        $result = $redis->eval(
            $script,
            [$key, $now, $this->capacity, $this->refillRate, $tokens],
            1
        );

        return (bool)$result;
    }
}

// 使用示例：每秒生成 2 个令牌，桶容量 10
$limiter = new TokenBucketLimiter('api', 10, 2);
if (!$limiter->allow($_SERVER['REMOTE_ADDR'])) {
    http_response_code(429);
    exit('Too Many Requests (Token Bucket)');
}
echo "Request allowed!";
```

> ✅ **优点**：支持突发流量（只要桶中有令牌）  
> ✅ **精确性**：Lua 脚本保证原子性

---

## 3️⃣ 漏桶（Leaky Bucket）

- **原理**：请求进入队列，以固定速率“漏出”（处理）。队列满则拒绝。
- **Redis 结构**：`Hash` 存储 `{ last_leak_time, queue_size }`
- **原子操作**：Lua 脚本

```php
// LeakyBucketLimiter.php
class LeakyBucketLimiter
{
    private string $keyPrefix;
    private int $capacity;     // 队列最大长度
    private float $leakRate;   // 每秒漏出请求数（处理速率）

    public function __construct(string $keyPrefix, int $capacity, float $leakRate)
    {
        $this->keyPrefix = $keyPrefix;
        $this->capacity = $capacity;
        $this->leakRate = $leakRate;
    }

    public function allow(string $identifier): bool
    {
        $redis = RedisClient::getInstance();
        $key = "leaky_bucket:{$this->keyPrefix}:{$identifier}";
        $now = microtime(true);

        $script = '
            local key = KEYS[1]
            local now = tonumber(ARGV[1])
            local capacity = tonumber(ARGV[2])
            local leak_rate = tonumber(ARGV[3])

            -- 获取当前状态
            local last_leak = redis.call("HGET", key, "last_leak")
            local queue_size = redis.call("HGET", key, "queue_size")

            if not last_leak then
                last_leak = now
                queue_size = 0
            else
                last_leak = tonumber(last_leak)
                queue_size = tonumber(queue_size)
                -- 模拟漏水：计算应处理的请求数
                local elapsed = now - last_leak
                local leaked = elapsed * leak_rate
                queue_size = math.max(0, queue_size - leaked)
            end

            if queue_size >= capacity then
                -- 队列已满
                redis.call("HMSET", key, "last_leak", now, "queue_size", queue_size)
                redis.call("EXPIRE", key, 3600)
                return 0
            end

            -- 新请求入队
            queue_size = queue_size + 1
            redis.call("HMSET", key, "last_leak", now, "queue_size", queue_size)
            redis.call("EXPIRE", key, 3600)
            return 1
        ';

        $result = $redis->eval(
            $script,
            [$key, $now, $this->capacity, $this->leakRate],
            1
        );

        return (bool)$result;
    }
}

// 使用示例：队列容量 5，每秒处理 1 个请求
$limiter = new LeakyBucketLimiter('api', 5, 1);
if (!$limiter->allow($_SERVER['REMOTE_ADDR'])) {
    http_response_code(429);
    exit('Too Many Requests (Leaky Bucket)');
}
echo "Request allowed!";
```

> ✅ **优点**：强制平滑输出，严格匀速处理  
> ⚠️ **注意**：不支持突发流量

---

## 📊 算法对比总结

| 特性 | 固定窗口 | 令牌桶 | 漏桶 |
|------|---------|--------|------|
| **突发流量支持** | ❌（边界问题） | ✅ | ❌ |
| **输出平滑性** | 差 | 中 | 优 |
| **实现复杂度** | ⭐ | ⭐⭐⭐ | ⭐⭐⭐ |
| **适用场景** | 简单限频 | 需要突发能力的 API | 严格匀速处理（如支付队列） |
| **Redis 操作** | `INCR` | Lua 脚本 | Lua 脚本 |

---

## 🛡️ 生产环境最佳实践

1. **使用 Lua 脚本**：确保多命令原子性（避免竞态条件）。
2. **设置 TTL**：防止 Redis 内存泄漏（如 `EXPIRE 3600`）。
3. **按维度限流**：
   - IP：`$_SERVER['REMOTE_ADDR']`
   - API Key：`$request->header('X-API-Key')`
   - 用户 ID：认证后获取
4. **返回标准限流头**：
   ```php
   header('X-RateLimit-Limit: 100');
   header('X-RateLimit-Remaining: ' . max(0, $remaining));
   header('X-RateLimit-Reset: ' . $resetTime);
   ```
5. **监控与告警**：记录被限流的请求，分析异常模式。

---

## 🚀 扩展建议

- **集成到中间件**：在 Laravel/Symfony 中封装为中间件。
- **动态配置**：从数据库读取 `$maxRequests`、`$refillRate` 等参数。
- **多级限流**：先 CDN 限流，再应用层限流。

> 如需 **完整可运行示例** 或 **框架集成代码**（如 Laravel），请告知！
