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