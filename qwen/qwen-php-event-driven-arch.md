# 我正在用PHP开发一个系统，我想通过“任务队列”或者“消息队列”之类的机制提高程序的灵活性，比如，当某件事情发生时，比如“收到新订单”，只需向队列中发送一个事件“NEW_ORDER_RECEIVED"，后台收到信息后，会触发多种处理过程，比如，检查库存，打印工作单，打印邮寄标签等等，多个处理过程互不干扰，并可以灵活添加新的处理过程。我的问题是，我应该如何实现这样的架构？

您描述的是典型的**事件驱动架构（Event-Driven Architecture）**，结合消息队列来实现。
以下是几种可行的实现方案，从简单到复杂：

## 方案一：基于数据库的简单队列（适合中小型项目）

### 1. 数据库表结构
```sql
-- 事件队列表
CREATE TABLE event_queue (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_name VARCHAR(100) NOT NULL,
    payload JSON NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    retry_count INT DEFAULT 0,
    INDEX idx_status (status),
    INDEX idx_event_name (event_name)
);

-- 事件处理器注册表（可选）
CREATE TABLE event_handlers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    event_name VARCHAR(100) NOT NULL,
    handler_class VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    priority INT DEFAULT 0
);
```

### 2. 核心代码实现

```php
// EventDispatcher.php - 事件分发器
class EventDispatcher {
    private $pdo;
    
    public function __construct(PDO $pdo) {
        $this->pdo = $pdo;
    }
    
    // 发布事件到队列
    public function dispatch(string $eventName, array $payload): void {
        $stmt = $this->pdo->prepare("
            INSERT INTO event_queue (event_name, payload) 
            VALUES (?, ?)
        ");
        $stmt->execute([$eventName, json_encode($payload)]);
    }
    
    // 处理队列中的事件
    public function processQueue(int $limit = 10): void {
        // 获取待处理的事件
        $stmt = $this->pdo->prepare("
            SELECT * FROM event_queue 
            WHERE status = 'pending' 
            ORDER BY created_at ASC 
            LIMIT ?
            FOR UPDATE SKIP LOCKED
        ");
        $stmt->execute([$limit]);
        $events = $stmt->fetchAll(PDO::FETCH_ASSOC);
        
        foreach ($events as $event) {
            $this->processEvent($event);
        }
    }
    
    private function processEvent(array $event): void {
        try {
            // 标记为处理中
            $this->updateEventStatus($event['id'], 'processing');
            
            // 获取所有处理器
            $handlers = $this->getEventHandlers($event['event_name']);
            
            foreach ($handlers as $handler) {
                if (class_exists($handler['handler_class'])) {
                    $handlerInstance = new $handler['handler_class']($this->pdo);
                    $handlerInstance->handle(json_decode($event['payload'], true));
                }
            }
            
            $this->updateEventStatus($event['id'], 'completed');
            
        } catch (Exception $e) {
            $retryCount = $event['retry_count'] + 1;
            if ($retryCount < 3) { // 最多重试3次
                $this->updateEventStatus($event['id'], 'pending', $retryCount);
            } else {
                $this->updateEventStatus($event['id'], 'failed');
                error_log("Event {$event['id']} failed: " . $e->getMessage());
            }
        }
    }
    
    private function updateEventStatus(int $eventId, string $status, int $retryCount = null): void {
        if ($retryCount !== null) {
            $stmt = $this->pdo->prepare("
                UPDATE event_queue 
                SET status = ?, retry_count = ?, processed_at = NOW() 
                WHERE id = ?
            ");
            $stmt->execute([$status, $retryCount, $eventId]);
        } else {
            $stmt = $this->pdo->prepare("
                UPDATE event_queue 
                SET status = ?, processed_at = NOW() 
                WHERE id = ?
            ");
            $stmt->execute([$status, $eventId]);
        }
    }
    
    private function getEventHandlers(string $eventName): array {
        $stmt = $this->pdo->prepare("
            SELECT handler_class FROM event_handlers 
            WHERE event_name = ? AND is_active = 1 
            ORDER BY priority ASC
        ");
        $stmt->execute([$eventName]);
        return $stmt->fetchAll(PDO::FETCH_ASSOC);
    }
}
```

### 3. 事件处理器接口和实现

```php
// EventHandlerInterface.php
interface EventHandlerInterface {
    public function handle(array $payload): void;
}

// CheckInventoryHandler.php
class CheckInventoryHandler implements EventHandlerInterface {
    private $pdo;
    
    public function __construct(PDO $pdo) {
        $this->pdo = $pdo;
    }
    
    public function handle(array $payload): void {
        // 检查库存逻辑
        $orderId = $payload['order_id'];
        // ... 库存检查逻辑
        
        echo "库存检查完成: 订单 {$orderId}\n";
    }
}

// PrintWorkOrderHandler.php
class PrintWorkOrderHandler implements EventHandlerInterface {
    public function handle(array $payload): void {
        // 打印工作单逻辑
        $orderId = $payload['order_id'];
        // ... 打印逻辑
        
        echo "工作单打印完成: 订单 {$orderId}\n";
    }
}

// PrintShippingLabelHandler.php
class PrintShippingLabelHandler implements EventHandlerInterface {
    public function handle(array $payload): void {
        // 打印邮寄标签逻辑
        $orderId = $payload['order_id'];
        // ... 打印逻辑
        
        echo "邮寄标签打印完成: 订单 {$orderId}\n";
    }
}
```

### 4. 使用示例

```php
// 注册处理器（可以在系统初始化时执行）
function registerEventHandlers(PDO $pdo) {
    $handlers = [
        ['NEW_ORDER_RECEIVED', CheckInventoryHandler::class, 1],
        ['NEW_ORDER_RECEIVED', PrintWorkOrderHandler::class, 2],
        ['NEW_ORDER_RECEIVED', PrintShippingLabelHandler::class, 3],
    ];
    
    $stmt = $pdo->prepare("
        INSERT INTO event_handlers (event_name, handler_class, priority) 
        VALUES (?, ?, ?)
    ");
    
    foreach ($handlers as $handler) {
        $stmt->execute($handler);
    }
}

// 发布事件
$dispatcher = new EventDispatcher($pdo);
$dispatcher->dispatch('NEW_ORDER_RECEIVED', [
    'order_id' => 12345,
    'customer_id' => 67890,
    'items' => [['product_id' => 1, 'quantity' => 2]]
]);

// 处理队列（通过定时任务或后台进程调用）
$dispatcher->processQueue();
```

## 方案二：使用 Redis 队列（推荐用于生产环境）

### 1. 安装依赖
```bash
composer require predis/predis
```

### 2. Redis 队列实现

```php
// RedisEventDispatcher.php
class RedisEventDispatcher {
    private $redis;
    
    public function __construct(Redis $redis) {
        $this->redis = $redis;
    }
    
    public function dispatch(string $eventName, array $payload): void {
        $message = json_encode([
            'event_name' => $eventName,
            'payload' => $payload,
            'created_at' => time()
        ]);
        
        // 发布到 Redis List
        $this->redis->lpush('event_queue', $message);
        
        // 同时发布到 Pub/Sub（可选，用于实时通知）
        $this->redis->publish('events', $message);
    }
    
    public function processQueue(): void {
        while (true) {
            // 阻塞式获取消息
            $message = $this->redis->brpop(['event_queue'], 30);
            
            if ($message) {
                $data = json_decode($message[1], true);
                $this->processEvent($data['event_name'], $data['payload']);
            }
        }
    }
    
    private function processEvent(string $eventName, array $payload): void {
        try {
            $handlers = $this->getEventHandlers($eventName);
            foreach ($handlers as $handlerClass) {
                $handler = new $handlerClass();
                $handler->handle($payload);
            }
        } catch (Exception $e) {
            // 记录错误日志
            error_log("Event processing failed: " . $e->getMessage());
            // 可以重新入队或发送到死信队列
        }
    }
    
    private function getEventHandlers(string $eventName): array {
        // 可以从配置文件、数据库或 Redis Hash 中获取
        $config = [
            'NEW_ORDER_RECEIVED' => [
                CheckInventoryHandler::class,
                PrintWorkOrderHandler::class,
                PrintShippingLabelHandler::class
            ]
        ];
        
        return $config[$eventName] ?? [];
    }
}
```

## 方案三：使用专业消息队列（RabbitMQ/Beanstalkd）

### 使用 Beanstalkd 示例

```bash
# 安装 Beanstalkd
sudo apt-get install beanstalkd

# PHP 客户端
composer require pda/pheanstalk
```

```php
// BeanstalkEventDispatcher.php
use Pheanstalk\Pheanstalk;

class BeanstalkEventDispatcher {
    private $pheanstalk;
    
    public function __construct(Pheanstalk $pheanstalk) {
        $this->pheanstalk = $pheanstalk;
    }
    
    public function dispatch(string $eventName, array $payload): void {
        $jobData = json_encode([
            'event_name' => $eventName,
            'payload' => $payload
        ]);
        
        $this->pheanstalk->useTube('events')->put($jobData);
    }
    
    public function processQueue(): void {
        $tube = $this->pheanstalk->watch('events');
        
        while ($job = $tube->reserve()) {
            try {
                $data = json_decode($job->getData(), true);
                $this->processEvent($data['event_name'], $data['payload']);
                $this->pheanstalk->delete($job); // 删除已处理的job
            } catch (Exception $e) {
                $this->pheanstalk->bury($job); // 埋葬失败的job
                error_log("Job processing failed: " . $e->getMessage());
            }
        }
    }
}
```

## 运行后台处理器

### 方法1：使用 Cron 定时任务（适用于数据库队列）
```bash
# 每分钟执行一次
* * * * * /usr/bin/php /path/to/your/project/process_queue.php
```

### 方法2：常驻进程（适用于 Redis/Beanstalkd）
```php
// worker.php
require_once 'vendor/autoload.php';

$redis = new Redis();
$redis->connect('127.0.0.1', 6379);

$dispatcher = new RedisEventDispatcher($redis);
$dispatcher->processQueue(); // 持续监听
```

启动进程：
```bash
nohup php worker.php > worker.log 2>&1 &
```

### 方法3：使用 Supervisor 管理进程
```ini
; /etc/supervisor/conf.d/event-worker.conf
[program:event-worker]
command=php /path/to/worker.php
autostart=true
autorestart=true
stderr_logfile=/var/log/event-worker.err.log
stdout_logfile=/var/log/event-worker.out.log
```

## 配置管理建议

创建配置文件来管理事件和处理器的映射关系：

```php
// config/events.php
return [
    'NEW_ORDER_RECEIVED' => [
        \App\Handlers\CheckInventoryHandler::class,
        \App\Handlers\PrintWorkOrderHandler::class,
        \App\Handlers\PrintShippingLabelHandler::class,
        \App\Handlers\SendNotificationHandler::class, // 新增处理器只需在这里添加
    ],
    'ORDER_SHIPPED' => [
        \App\Handlers\UpdateTrackingHandler::class,
        \App\Handlers\SendShippingEmailHandler::class,
    ]
];
```

## 总结建议

1. **小型项目**：使用方案一（数据库队列），简单易维护
2. **中型项目**：使用方案二（Redis队列），性能更好，支持更多特性
3. **大型项目**：使用方案三（专业消息队列），提供更好的可靠性、监控和管理功能

4. **关键考虑点**：
   - 错误处理和重试机制
   - 死信队列处理
   - 监控和日志记录
   - 处理器的幂等性设计
   - 事务一致性（如果需要）

5. **扩展性**：通过配置文件管理事件-处理器映射，新增功能时只需添加新的处理器类并在配置中注册即可。

这样的架构让您的系统具有很好的灵活性和可扩展性，符合开闭原则（对扩展开放，对修改关闭）。

