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

        echo $logStr;

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
