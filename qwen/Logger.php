<?php

declare(strict_types=1);

class Logger
{
    // 日志级别常量 (数值越小，级别越低)
    const DEBUG     = 100;
    const INFO      = 200;
    const NOTICE    = 250;
    const WARNING   = 300;
    const ERROR     = 400;
    const CRITICAL  = 500;
    const ALERT     = 550;
    const EMERGENCY = 600;

    // 级别名称映射
    const LEVEL_NAMES = [
        self::DEBUG     => 'DEBUG',
        self::INFO      => 'INFO',
        self::NOTICE    => 'NOTICE',
        self::WARNING   => 'WARNING',
        self::ERROR     => 'ERROR',
        self::CRITICAL  => 'CRITICAL',
        self::ALERT     => 'ALERT',
        self::EMERGENCY => 'EMERGENCY',
    ];

    private int $minLevel;
    private string $logDir;
    private string $dateFormat;
    private array $defaultContext;

    /**
     * @param string $logDir       日志存储目录
     * @param int    $minLevel     最低记录级别（低于此级别的日志将被忽略）
     * @param string $dateFormat   时间格式
     * @param array  $defaultContext 默认附加上下文（如 request_id, user_id 等）
     */
    public function __construct(
        string $logDir = '/var/log/app',
        int $minLevel = self::INFO,
        string $dateFormat = 'Y-m-d H:i:s',
        array $defaultContext = []
    ) {
        $this->logDir = rtrim($logDir, DIRECTORY_SEPARATOR);
        $this->minLevel = $minLevel;
        $this->dateFormat = $dateFormat;
        $this->defaultContext = $defaultContext;

        $this->ensureDirectoryExists($this->logDir);
    }

    // ==================== PSR-3 风格便捷方法 ====================

    public function emergency(string|\Stringable $message, array $context = []): void { $this->log(self::EMERGENCY, $message, $context); }
    public function alert(string|\Stringable $message, array $context = []): void     { $this->log(self::ALERT, $message, $context); }
    public function critical(string|\Stringable $message, array $context = []): void  { $this->log(self::CRITICAL, $message, $context); }
    public function error(string|\Stringable $message, array $context = []): void     { $this->log(self::ERROR, $message, $context); }
    public function warning(string|\Stringable $message, array $context = []): void   { $this->log(self::WARNING, $message, $context); }
    public function notice(string|\Stringable $message, array $context = []): void    { $this->log(self::NOTICE, $message, $context); }
    public function info(string|\Stringable $message, array $context = []): void      { $this->log(self::INFO, $message, $context); }
    public function debug(string|\Stringable $message, array $context = []): void     { $this->log(self::DEBUG, $message, $context); }

    /**
     * 核心日志记录方法
     */
    public function log(int $level, string|\Stringable $message, array $context = []): void
    {
        if ($level < $this->minLevel) {
            return;
        }

        $context = array_merge($this->defaultContext, $context);
        $formatted = $this->formatMessage($level, $message, $context);
        $this->writeToLog($formatted, $level);
    }

    /**
     * 动态调整最低日志级别
     */
    public function setMinLevel(int $level): void
    {
        $this->minLevel = $level;
    }

    // ==================== 内部实现 ====================

    /**
     * 格式化日志内容（支持 {key} 占位符替换）
     */
    private function formatMessage(int $level, string|\Stringable $message, array $context): string
    {
        $message = (string) $message;

        // 1. 替换占位符
        foreach ($context as $key => $value) {
            if (is_string($key) && str_contains($message, '{' . $key . '}')) {
                $message = str_replace('{' . $key . '}', $this->valueToString($value), $message);
            }
        }

        // 2. 提取未使用的上下文（附加到末尾）
        $unused = [];
        foreach ($context as $key => $value) {
            if (is_string($key) && !str_contains($message, '{' . $key . '}')) {
                $unused[$key] = $value;
            }
        }

        $timestamp = date($this->dateFormat);
        $levelName = self::LEVEL_NAMES[$level] ?? 'UNKNOWN';
        $logLine = "$timestamp $levelName $message";

        if (!empty($unused)) {
            $logLine .= ' | ' . json_encode($unused, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES);
        }

        return $logLine;
    }

    /**
     * 安全地将任意类型转为字符串
     */
    private function valueToString(mixed $value): string
    {
        return match (true) {
            is_string($value), is_numeric($value) => (string) $value,
            is_bool($value) => $value ? 'true' : 'false',
            is_null($value) => 'null',
            is_array($value), is_object($value) => json_encode($value, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
            is_resource($value) => '(resource)',
            default => '(unknown)'
        };
    }

    /**
     * 写入文件（按天分割 + 文件锁防并发）
     */
    private function writeToLog(string $message, int $level): void
    {
        $logFile = $this->logDir . DIRECTORY_SEPARATOR . 'app-' . date('Y-m-d') . '.log';

        try {
            $fp = fopen($logFile, 'a');
            if ($fp === false) {
                throw new RuntimeException("无法打开日志文件: $logFile");
            }

            echo $message, PHP_EOL;

            // 独占锁（阻塞模式）
            if (flock($fp, LOCK_EX)) {
                fwrite($fp, $message . PHP_EOL);
                flock($fp, LOCK_UN);
            } else {
                throw new RuntimeException("无法获取文件锁: $logFile");
            }
            fclose($fp);
        } catch (Throwable $e) {
            // 降级写入 PHP 错误日志，避免日志系统自身导致应用崩溃
            error_log($message . ' | LoggerError: ' . $e->getMessage());
        }
    }

    /**
     * 自动创建日志目录
     */
    private function ensureDirectoryExists(string $dir): void
    {
        if (!is_dir($dir)) {
            if (!mkdir($dir, 0755, true) && !is_dir($dir)) {
                throw new RuntimeException("无法创建日志目录: $dir");
            }
        }
    }
}

/*
// 1. 基础初始化
$logger = new Logger(
    logDir: './logs',
    minLevel: Logger::DEBUG, // 开发环境记录所有级别
    defaultContext: ['app' => 'user-service', 'env' => 'dev']
);

// 2. 记录不同级别日志
$logger->info("用户登录成功", ['user_id' => 1001]);
$logger->warning('API 响应较慢', ['endpoint' => '/api/orders', 'time_ms' => 2300]);

// 3. 占位符替换（PSR-3 规范）
$logger->error('支付失败: {reason}', [
    'reason' => '余额不足',
    'order_id' => 'ORD-9981',
    'user_id' => 1001,
    'trace_id' => 'abc-123' // 未出现在消息中的键会自动转为 JSON 附加到末尾
]);

// 4. 动态调整级别（如线上仅记录 WARNING 及以上）
$logger->setMinLevel(Logger::WARNING);
$logger->debug('这条不会被写入'); // 忽略
*/
