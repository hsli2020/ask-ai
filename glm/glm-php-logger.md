# 用php编写一个Logger类，实现常用的日志功能

`这是一个功能完善、符合现代PHP开发规范的 `Logger` 类。它参考了 PSR-3 (PHP-FIG 日志接口) 的标准设计，但保持了轻量化，没有引入任何外部依赖。

### 主要功能：
1. **多级别日志**：支持 DEBUG, INFO, WARNING, ERROR, CRITICAL 五个常用级别。
2. **级别过滤**：可以设置最低记录级别（例如在生产环境只记录 ERROR 及以上）。
3. **上下文支持（插值）**：支持 `{key}` 占位符替换，方便记录动态数据。
4. **按日期自动分割**：自动按天生成不同的日志文件（如 `app-2023-10-27.log`），防止单个文件过大。
5. **并发安全**：写入文件时使用 `LOCK_EX` 排他锁，防止多进程同时写入导致日志错乱。
6. **自动创建目录**：如果日志目录不存在，会自动递归创建。

### PHP 代码实现：

```php
<?php

class Logger
{
    // 日志级别常量
    const DEBUG = 100;
    const INFO = 200;
    const WARNING = 300;
    const ERROR = 400;
    const CRITICAL = 500;

    // 级别名称映射
    private static $levelNames = [
        self::DEBUG => 'DEBUG',
        self::INFO => 'INFO',
        self::WARNING => 'WARNING',
        self::ERROR => 'ERROR',
        self::CRITICAL => 'CRITICAL',
    ];

    // 允许通过字符串设置级别
    private static $levelMap = [
        'DEBUG' => self::DEBUG,
        'INFO' => self::INFO,
        'WARNING' => self::WARNING,
        'ERROR' => self::ERROR,
        'CRITICAL' => self::CRITICAL,
    ];

    /**
     * @var string 日志文件路径
     */
    private $logFile;

    /**
     * @var int 最低记录级别
     */
    private $minLevel;

    /**
     * @var string 时间格式
     */
    private $dateFormat;

    /**
     * 构造函数
     *
     * @param string $logFilePath 日志文件路径 (如: /var/log/myapp/app.log)
     * @param mixed $minLevel 最低日志级别 (默认 DEBUG，可传字符串 'INFO' 或数字 200)
     * @param bool $dailyRotation 是否按日期分割日志 (默认 true)
     */
    public function __construct(string $logFilePath, $minLevel = self::DEBUG, bool $dailyRotation = true)
    {
        // 处理最低级别
        if (is_string($minLevel)) {
            $this->minLevel = self::$levelMap[strtoupper($minLevel)] ?? self::DEBUG;
        } else {
            $this->minLevel = (int)$minLevel;
        }

        $this->dateFormat = 'Y-m-d H:i:s';

        // 按日期分割处理文件名
        if ($dailyRotation) {
            $pathInfo = pathinfo($logFilePath);
            $ext = isset($pathInfo['extension']) ? '.' . $pathInfo['extension'] : '';
            $this->logFile = $pathInfo['dirname'] . '/' . $pathInfo['filename'] . '-' . date('Y-m-d') . $ext;
        } else {
            $this->logFile = $logFilePath;
        }

        // 自动创建目录
        $logDir = dirname($this->logFile);
        if (!is_dir($logDir)) {
            mkdir($logDir, 0755, true);
        }
    }

    /**
     * 核心日志记录方法
     */
    public function log($level, string $message, array $context = []): void
    {
        // 转换字符串级别为数字
        if (is_string($level)) {
            $level = self::$levelMap[strtoupper($level)] ?? self::DEBUG;
        }

        // 级别过滤
        if ($level < $this->minLevel) {
            return;
        }

        // 替换上下文中的占位符
        $interpolatedMessage = $this->interpolate($message, $context);

        // 格式化日志内容
        $levelName = self::$levelNames[$level] ?? 'UNKNOWN';
        $time = date($this->dateFormat);
        $logStr = "[$time] [$levelName] $interpolatedMessage";

        // 如果上下文没有被完全替换，将剩余的上下文以 JSON 格式追加到末尾
        if (!empty($context)) {
            $logStr .= " | Context: " . $this->toJson($context);
        }

        $logStr .= PHP_EOL;

        // 写入文件 (FILE_APPEND追加, LOCK_EX排他锁防止并发错乱)
        file_put_contents($this->logFile, $logStr, FILE_APPEND | LOCK_EX);
    }

    /**
     * 快捷方法
     */
    public function debug(string $message, array $context = []): void
    {
        $this->log(self::DEBUG, $message, $context);
    }

    public function info(string $message, array $context = []): void
    {
        $this->log(self::INFO, $message, $context);
    }

    public function warning(string $message, array $context = []): void
    {
        $this->log(self::WARNING, $message, $context);
    }

    public function error(string $message, array $context = []): void
    {
        $this->log(self::ERROR, $message, $context);
    }

    public function critical(string $message, array $context = []): void
    {
        $this->log(self::CRITICAL, $message, $context);
    }

    /**
     * 替换消息中的 {key} 占位符
     */
    private function interpolate(string $message, array $context): string
    {
        if (empty($context)) {
            return $message;
        }

        $replace = [];
        foreach ($context as $key => $val) {
            // 如果值是数组或对象，转为JSON字符串
            if (is_array($val) || (is_object($val) && !method_exists($val, '__toString'))) {
                $val = $this->toJson($val);
            } elseif (is_object($val)) {
                $val = (string)$val;
            }
            $replace['{' . $key . '}'] = $val;
        }

        return strtr($message, $replace);
    }

    /**
     * 安全的 JSON 编码
     */
    private function toJson($data): string
    {
        try {
            return json_encode($data, JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR);
        } catch (\Throwable $e) {
            // 如果包含无法编码的资源(如数据库连接)等，退回到 print_r
            return print_r($data, true);
        }
    }
}
```

### 使用示例：

```php
<?php
// 引入 Logger 类
// require_once 'Logger.php';

// 1. 初始化 (按天分割，只记录 INFO 及以上级别)
$logger = new Logger(__DIR__ . '/logs/app.log', 'INFO');

// 2. 记录不同级别的日志
$logger->debug("这条不会被执行，因为最低级别是 INFO"); 

$logger->info("系统启动成功");
$logger->warning("内存使用率较高", ['usage' => '85%', 'limit' => '1024M']);
$logger->error("数据库连接失败", [
    'host' => '127.0.0.1',
    'error_code' => 2002,
    'error_msg' => 'Connection refused'
]);
$logger->critical("磁盘空间已满，服务即将停止！", ['free_space' => '0MB']);

// 3. 测试占位符替换 (PSR-3 风格)
$userId = 105;
$action = 'login';
$logger->info("用户 {user_id} 执行了 {action} 操作", [
    'user_id' => $userId,
    'action' => $action,
    'ip' => '192.168.1.100' // 多余的上下文会自动追加到日志末尾
]);
```

### 生成的日志文件 (`logs/app-2023-10-27.log`) 内容示例：

```text
[2023-10-27 14:30:15] [INFO] 系统启动成功
[2023-10-27 14:30:16] [WARNING] 内存使用率较高 | Context: {"usage":"85%","limit":"1024M"}
[2023-10-27 14:30:16] [ERROR] 数据库连接失败 | Context: {"host":"127.0.0.1","error_code":2002,"error_msg":"Connection refused"}
[2023-10-27 14:30:16] [CRITICAL] 磁盘空间已满，服务即将停止！ | Context: {"free_space":"0MB"}
[2023-10-27 14:30:17] [INFO] 用户 105 执行了 login 操作 | Context: {"ip":"192.168.1.100"}
```

### 生产环境建议：
对于**中小型项目**或**单机部署**的 PHP 项目，这个类完全足够了。
但如果您的项目是**大型分布式架构**，或者使用了 **Laravel / Symfony 等现代框架**，建议直接使用遵循 PSR-3 标准的第三方库，例如 **[Monolog](https://github.com/Seldaek/monolog)**。Monolog 支持将日志发送到 Redis、Kafka、Elasticsearch、钉钉/企业微信报警等高级功能。


# 我想做一点改进，先不将log信息直接写入文件，而是先存入内存中，当程序结束时才写入文件

这是一个非常实用的改进！这种机制被称为**日志缓冲**。将日志存入内存数组，在程序结束时一次性写入文件，可以大幅减少磁盘 I/O 次数，显著提升高性能接口的响应速度。

在 PHP 中，实现“程序结束时执行”的最佳方式是使用 `register_shutdown_function()` 注册一个关机函数。

以下是改进后的代码：

```php
<?php

class BufferLogger
{
    // 日志级别常量
    const DEBUG = 100;
    const INFO = 200;
    const WARNING = 300;
    const ERROR = 400;
    const CRITICAL = 500;

    private static $levelNames = [
        self::DEBUG => 'DEBUG',
        self::INFO => 'INFO',
        self::WARNING => 'WARNING',
        self::ERROR => 'ERROR',
        self::CRITICAL => 'CRITICAL',
    ];

    private static $levelMap = [
        'DEBUG' => self::DEBUG,
        'INFO' => self::INFO,
        'WARNING' => self::WARNING,
        'ERROR' => self::ERROR,
        'CRITICAL' => self::CRITICAL,
    ];

    /** @var string 日志文件路径 */
    private $logFile;

    /** @var int 最低记录级别 */
    private $minLevel;

    /** @var string 时间格式 */
    private $dateFormat;

    /** @var array 内存缓冲区 */
    private $buffer = [];

    /**
     * 构造函数
     */
    public function __construct(string $logFilePath, $minLevel = self::DEBUG, bool $dailyRotation = true)
    {
        if (is_string($minLevel)) {
            $this->minLevel = self::$levelMap[strtoupper($minLevel)] ?? self::DEBUG;
        } else {
            $this->minLevel = (int)$minLevel;
        }

        $this->dateFormat = 'Y-m-d H:i:s';

        if ($dailyRotation) {
            $pathInfo = pathinfo($logFilePath);
            $ext = isset($pathInfo['extension']) ? '.' . $pathInfo['extension'] : '';
            $this->logFile = $pathInfo['dirname'] . '/' . $pathInfo['filename'] . '-' . date('Y-m-d') . $ext;
        } else {
            $this->logFile = $logFilePath;
        }

        $logDir = dirname($this->logFile);
        if (!is_dir($logDir)) {
            mkdir($logDir, 0755, true);
        }

        // 【核心改进】注册 PHP 关机函数，程序结束时自动将缓冲区写入文件
        register_shutdown_function([$this, 'flush']);
    }

    /**
     * 核心日志记录方法（仅存入内存）
     */
    public function log($level, string $message, array $context = []): void
    {
        if (is_string($level)) {
            $level = self::$levelMap[strtoupper($level)] ?? self::DEBUG;
        }

        if ($level < $this->minLevel) {
            return;
        }

        $interpolatedMessage = $this->interpolate($message, $context);
        $levelName = self::$levelNames[$level] ?? 'UNKNOWN';
        $time = date($this->dateFormat);
        
        $logStr = "[$time] [$levelName] $interpolatedMessage";
        
        if (!empty($context)) {
            $logStr .= " | Context: " . $this->toJson($context);
        }

        $logStr .= PHP_EOL;

        // 【核心改进】存入内存数组，而不是直接写文件
        $this->buffer[] = $logStr;
    }

    /**
     * 【新增】将缓冲区内容刷入磁盘
     * 注意：此方法被设为 public，以便在长驻进程或特定场景下手动触发
     */
    public function flush(): void
    {
        // 如果缓冲区为空，直接返回
        if (empty($this->buffer)) {
            return;
        }

        // 一次性将数组拼接成字符串
        $content = implode('', $this->buffer);

        // 一次性写入文件
        file_put_contents($this->logFile, $content, FILE_APPEND | LOCK_EX);

        // 清空缓冲区，防止重复写入
        $this->buffer = [];
    }

    // --- 以下是快捷方法（保持不变） ---
    
    public function debug(string $message, array $context = []): void { $this->log(self::DEBUG, $message, $context); }
    public function info(string $message, array $context = []): void { $this->log(self::INFO, $message, $context); }
    public function warning(string $message, array $context = []): void { $this->log(self::WARNING, $message, $context); }
    public function error(string $message, array $context = []): void { $this->log(self::ERROR, $message, $context); }
    public function critical(string $message, array $context = []): void { $this->log(self::CRITICAL, $message, $context); }

    // --- 以下是私有辅助方法（保持不变） ---

    private function interpolate(string $message, array $context): string
    {
        if (empty($context)) return $message;

        $replace = [];
        foreach ($context as $key => $val) {
            if (is_array($val) || (is_object($val) && !method_exists($val, '__toString'))) {
                $val = $this->toJson($val);
            } elseif (is_object($val)) {
                $val = (string)$val;
            }
            $replace['{' . $key . '}'] = $val;
        }
        return strtr($message, $replace);
    }

    private function toJson($data): string
    {
        try {
            return json_encode($data, JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR);
        } catch (\Throwable $e) {
            return print_r($data, true);
        }
    }
    
    /**
     * 析构函数（双重保险）
     * 正常情况下 register_shutdown_function 已经会触发 flush
     * 但如果对象被 unset 或作用域销毁，这里可以保证刷出
     */
    public function __destruct()
    {
        $this->flush();
    }
}
```

### 改进点解析：

1. **内存缓冲区 (`$buffer`)**：原来每次调用 `log()` 都会触发一次磁盘 I/O。现在只是执行 `$this->buffer[] = $logStr`，这只是一次内存操作，速度极快。
2. **自动关机刷出 (`register_shutdown_function`)**：无论你的 PHP 代码是正常执行完毕，还是因为 Fatal Error / Exception 崩溃退出，PHP 引擎在关闭前都会执行注册进去的 `flush()` 方法，保证日志不丢失。
3. **高效 I/O (`implode`)**：在 `flush()` 时，使用 `implode('', $this->buffer)` 将所有日志拼成一个巨大的字符串，然后**只调用一次** `file_put_contents`。相比循环写入，这是最高效的写法。
4. **析构函数 `__destruct` (双重保险)**：增加了一层保险，如果在脚本运行中途，Logger 对象被销毁了，也会自动把已收集的日志写入文件。

### ⚠️ 重要注意事项（使用场景限制）：

这种模式非常适合 **传统的 PHP-FPM / Apache 环境**（每次请求启动，执行完毕销毁）。

**但不适合以下场景：**
* **常驻内存框架（如 Swoole, Workerman, RoadRunner）**：因为程序永远不会结束，`shutdown_function` 永远不会触发，日志会一直堆积在内存中，最终导致内存溢出（OOM）。
* **超长耗时的 CLI 脚本**：如果你跑一个需要处理几小时的定时任务脚本，期间产生的日志全在内存里，不仅占内存，而且如果中途宕机，还没 flush 的日志就全丢了。

**如果是长驻进程/长脚本怎么办？**
你可以定期手动调用 `flush()`，例如：
```php
// 在 Swoole 或长脚本中
$logger = new BufferLogger('/tmp/test.log');

for ($i = 0; $i < 100000; $i++) {
    $logger->info("处理第 {$i} 个任务");
    
    // 每处理 1000 条，手动刷入一次磁盘
    if ($i % 1000 === 0) {
        $logger->flush(); 
    }
}
// 循环结束后再刷一次，确保尾数日志写入
$logger->flush(); 
```

