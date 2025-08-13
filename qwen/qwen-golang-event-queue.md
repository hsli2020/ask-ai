# 请用golang + mysql编写一个事件队列系统

以下是一个使用 Go + MySQL 构建的事件队列系统的完整示例，涵盖数据库设计、事件入队与出队、并发处理、重试机制、状态管理、配置与日志等关键功能，适用于生产环境。

---

## 一、数据库设计（MySQL）

```sql
CREATE TABLE IF NOT EXISTS events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    payload JSON NOT NULL,
    status ENUM('pending', 'processing', 'completed', 'failed') DEFAULT 'pending',
    retry_count INT DEFAULT 0,
    max_retry INT DEFAULT 3,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    processed_at TIMESTAMP NULL,
    error_message TEXT,
    INDEX idx_status_created (status, created_at)
);
```

---

## 二、项目结构

```
event-queue/
├── main.go
├── config/
│   └── config.go
├── models/
│   └── event.go
├── queue/
│   ├── producer.go
│   └── consumer.go
├── db/
│   └── db.go
└── go.mod
```

---

## 三、核心代码实现

### 1. 配置管理（config/config.go）

```go
package config

type Config struct {
    DBHost     string
    DBPort     string
    DBUser     string
    DBPassword string
    DBName     string
    WorkerNum  int
    PollIntervalSec int
}

func Load() *Config {
    return &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnv("DB_PORT", "3306"),
        DBUser:     getEnv("DB_USER", "root"),
        DBPassword: getEnv("DB_PASSWORD", ""),
        DBName:     getEnv("DB_NAME", "event_queue"),
        WorkerNum:  5,
        PollIntervalSec: 2,
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

---

### 2. 数据库连接（db/db.go）

```go
package db

import (
    "database/sql"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
)

func Connect(host, port, user, pass, dbname string) (*sql.DB, error) {
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", user, pass, host, port, dbname)
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    db.SetMaxOpenConns(20)
    db.SetMaxIdleConns(5)
    return db, db.Ping()
}
```

---

### 3. 事件模型（models/event.go）

```go
package models

import (
    "database/sql"
    "encoding/json"
    "time"
)

type EventStatus string

const (
    StatusPending   EventStatus = "pending"
    StatusProcessing            = "processing"
    StatusCompleted             = "completed"
    StatusFailed                = "failed"
)

type Event struct {
    ID           int64
    EventType    string
    Payload      map[string]interface{}
    Status       EventStatus
    RetryCount   int
    MaxRetry     int
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ProcessedAt  sql.NullTime
    ErrorMessage sql.NullString
}

func (e *Event) ToJSON() ([]byte, error) {
    return json.Marshal(e.Payload)
}

func (e *Event) FromJSON(data []byte) error {
    return json.Unmarshal(data, &e.Payload)
}
```

---

### 4. 事件生产者（queue/producer.go）

```go
package queue

import (
    "database/sql"
    "encoding/json"
    "fmt"
)

type Producer struct {
    db *sql.DB
}

func NewProducer(db *sql.DB) *Producer {
    return &Producer{db: db}
}

func (p *Producer) Enqueue(eventType string, payload map[string]interface{}) (int64, error) {
    data, _ := json.Marshal(payload)
    res, err := p.db.Exec(`
        INSERT INTO events (event_type, payload, status, retry_count, max_retry)
        VALUES (?, ?, 'pending', 0, 3)
    `, eventType, string(data))
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}
```

---

### 5. 事件消费者（queue/consumer.go）

```go
package queue

import (
    "database/sql"
    "encoding/json"
    "log"
    "time"
    "yourproject/models"
)

type Consumer struct {
    db           *sql.DB
    workerCount  int
    pollInterval time.Duration
    handlers     map[string]func(*models.Event) error
}

func NewConsumer(db *sql.DB, workers int, interval time.Duration) *Consumer {
    return &Consumer{
        db:           db,
        workerCount:  workers,
        pollInterval: interval,
        handlers:     make(map[string]func(*models.Event) error),
    }
}

func (c *Consumer) RegisterHandler(eventType string, handler func(*models.Event) error) {
    c.handlers[eventType] = handler
}

func (c *Consumer) Start(done <-chan struct{}) {
    for i := 0; i < c.workerCount; i++ {
        go c.workerLoop(done)
    }
}

func (c *Consumer) workerLoop(done <-chan struct{}) {
    for {
        select {
        case <-done:
            return
        default:
            c.processNextEvent()
            time.Sleep(c.pollInterval)
        }
    }
}

func (c *Consumer) processNextEvent() {
    tx, err := c.db.Begin()
    if err != nil {
        return
    }
    defer tx.Rollback()

    var e models.Event
    err = tx.QueryRow(`
        SELECT id, event_type, payload, status, retry_count, max_retry, created_at, updated_at, processed_at, error_message
        FROM events
        WHERE status = 'pending'
        ORDER BY created_at ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    `).Scan(&e.ID, &e.EventType, &e.Payload, &e.Status, &e.RetryCount, &e.MaxRetry, &e.CreatedAt, &e.UpdatedAt, &e.ProcessedAt, &e.ErrorMessage)

    if err == sql.ErrNoRows {
        return
    }
    if err != nil {
        log.Println("Fetch event error:", err)
        return
    }

    // 标记为 processing
    _, err = tx.Exec("UPDATE events SET status = 'processing' WHERE id = ?", e.ID)
    if err != nil {
        return
    }
    tx.Commit()

    // 执行处理
    handler, ok := c.handlers[e.EventType]
    var processErr error
    if !ok {
        processErr = fmt.Errorf("no handler for event type: %s", e.EventType)
    } else {
        processErr = handler(&e)
    }

    // 更新状态
    c.updateEventStatus(e.ID, processErr)
}

func (c *Consumer) updateEventStatus(id int64, err error) {
    if err == nil {
        _, _ = c.db.Exec(`
            UPDATE events SET status = 'completed', processed_at = NOW()
            WHERE id = ?
        `, id)
    } else {
        var e models.Event
        _ = c.db.QueryRow("SELECT retry_count, max_retry FROM events WHERE id = ?", id).Scan(&e.RetryCount, &e.MaxRetry)
        if e.RetryCount >= e.MaxRetry {
            _, _ = c.db.Exec(`
                UPDATE events SET status = 'failed', error_message = ?, processed_at = NOW()
                WHERE id = ?
            `, err.Error(), id)
        } else {
            _, _ = c.db.Exec(`
                UPDATE events SET status = 'pending', retry_count = retry_count + 1, error_message = ?
                WHERE id = ?
            `, err.Error(), id)
        }
    }
}
```

---

### 6. 主程序（main.go）

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    "yourproject/config"
    "yourproject/db"
    "yourproject/queue"
    "yourproject/models"
)

func main() {
    cfg := config.Load()
    database, err := db.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
    if err != nil {
        log.Fatal("DB connect failed:", err)
    }
    defer database.Close()

    consumer := queue.NewConsumer(database, cfg.WorkerNum, time.Duration(cfg.PollIntervalSec)*time.Second)
    consumer.RegisterHandler("user.created", func(e *models.Event) error {
        log.Printf("Processing user.created event: %+v\n", e.Payload)
        return nil
    })

    done := make(chan struct{})
    consumer.Start(done)

    // 示例：生产一个事件
    producer := queue.NewProducer(database)
    _, _ = producer.Enqueue("user.created", map[string]interface{}{
        "user_id":   123,
        "username":  "alice",
        "email":     "alice@example.com",
    })

    // 优雅退出
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    close(done)
    log.Println("Shutting down...")
}
```

---

## 四、功能特性总结

- ✅ 支持事件入队与出队
- ✅ 多消费者并发处理
- ✅ 事件状态管理（pending/processing/completed/failed）
- ✅ 自动重试机制（可配置最大重试次数）
- ✅ 使用 `FOR UPDATE SKIP LOCKED` 避免重复消费
- ✅ 支持优雅关闭
- ✅ 可扩展事件处理器注册机制

---

## 五、扩展建议

- 引入消息中间件（如 Kafka、RabbitMQ）提升吞吐
- 添加事件调度（延迟队列）
- 增加监控指标（如 Prometheus）
- 使用连接池、事务优化性能
- 添加事件归档/清理机制

---

该系统可作为事件驱动架构的基础组件，适用于异步任务处理、通知系统、日志收集等场景。


# 添加事件调度（延迟队列）
添加事件归档/清理机制

以下是针对原有系统的增强版本，主要添加了 **延迟队列（事件调度）** 和 **数据归档/清理机制**。

### 一、数据库架构升级

我们需要修改 `events` 表以支持调度时间，并创建 `events_archive` 表用于归档旧数据。

```sql
-- 1. 修改 events 表，增加 scheduled_at 字段
ALTER TABLE events 
ADD COLUMN scheduled_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
ADD INDEX idx_status_scheduled (status, scheduled_at);

-- 2. 创建归档表 (结构与 events 一致)
CREATE TABLE IF NOT EXISTS events_archive LIKE events;

-- 3. 为归档表增加归档时间索引
ALTER TABLE events_archive ADD INDEX idx_archived_at (updated_at);
```

> **注意**：`SKIP LOCKED` 需要 MySQL 8.0+ 版本。

---

### 二、模型层更新 (`models/event.go`)

在 Event 结构体中增加 `ScheduledAt` 字段。

```go
package models

import (
    "database/sql"
    "time"
)

// ... (保留原有的 EventStatus 常量)

type Event struct {
    ID           int64
    EventType    string
    Payload      map[string]interface{}
    Status       EventStatus
    RetryCount   int
    MaxRetry     int
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ScheduledAt  time.Time      // 新增：计划执行时间
    ProcessedAt  sql.NullTime
    ErrorMessage sql.NullString
}
```

---

### 三、生产者更新 (`queue/producer.go`)

支持在入队时指定延迟时间。

```go
package queue

import (
    "database/sql"
    "encoding/json"
    "time"
)

type Producer struct {
    db *sql.DB
}

func NewProducer(db *sql.DB) *Producer {
    return &Producer{db: db}
}

// Enqueue 立即执行
func (p *Producer) Enqueue(eventType string, payload map[string]interface{}) (int64, error) {
    return p.EnqueueWithDelay(eventType, payload, 0)
}

// EnqueueWithDelay 支持延迟执行
func (p *Producer) EnqueueWithDelay(eventType string, payload map[string]interface{}, delay time.Duration) (int64, error) {
    data, err := json.Marshal(payload)
    if err != nil {
        return 0, err
    }

    // 计算计划执行时间
    scheduledAt := time.Now().Add(delay)

    res, err := p.db.Exec(`
        INSERT INTO events (event_type, payload, status, retry_count, max_retry, scheduled_at)
        VALUES (?, ?, 'pending', 0, 3, ?)
    `, eventType, string(data), scheduledAt)
    
    if err != nil {
        return 0, err
    }
    return res.LastInsertId()
}
```

---

### 四、消费者更新 (`queue/consumer.go`)

修改查询逻辑，只获取 `scheduled_at <= NOW()` 的事件。

```go
// ... (引用部分保持不变)

func (c *Consumer) processNextEvent() {
    tx, err := c.db.Begin()
    if err != nil {
        return
    }
    defer tx.Rollback()

    var e models.Event
    // 关键修改：增加 scheduled_at <= NOW() 条件
    err = tx.QueryRow(`
        SELECT id, event_type, payload, status, retry_count, max_retry, created_at, updated_at, scheduled_at, processed_at, error_message
        FROM events
        WHERE status = 'pending' 
          AND scheduled_at <= NOW()
        ORDER BY scheduled_at ASC, created_at ASC
        LIMIT 1
        FOR UPDATE SKIP LOCKED
    `).Scan(
        &e.ID, &e.EventType, &e.Payload, &e.Status, &e.RetryCount, &e.MaxRetry, 
        &e.CreatedAt, &e.UpdatedAt, &e.ScheduledAt, &e.ProcessedAt, &e.ErrorMessage,
    )

    if err == sql.ErrNoRows {
        return
    }
    if err != nil {
        // 记录错误日志
        return
    }

    // 标记为 processing
    _, err = tx.Exec("UPDATE events SET status = 'processing' WHERE id = ?", e.ID)
    if err != nil {
        return
    }
    
    // 提交事务以释放行锁，允许其他 worker 处理其他事件
    if err := tx.Commit(); err != nil {
        return
    }

    // --- 业务处理 (在事务外执行，避免长事务锁表) ---
    handler, ok := c.handlers[e.EventType]
    var processErr error
    if !ok {
        processErr = fmt.Errorf("no handler for event type: %s", e.EventType)
    } else {
        // 需要重新解析 Payload，因为 Scan 到 map 可能需要特殊处理，这里简化处理
        // 实际生产中建议在 Scan 时直接处理 JSON 或统一封装
        processErr = handler(&e)
    }

    // 更新状态
    c.updateEventStatus(e.ID, processErr)
}

// ... (updateEventStatus 方法保持不变)
```

---

### 五、归档与清理组件 (`queue/cleaner.go`)

新建文件，负责定期将已完成/失败的历史数据移动到归档表，防止主表膨胀。

```go
package queue

import (
    "database/sql"
    "log"
    "time"
)

type Cleaner struct {
    db            *sql.DB
    retentionDays int
    batchSize     int
    interval      time.Duration
}

func NewCleaner(db *sql.DB, retentionDays int, batchSize int, interval time.Duration) *Cleaner {
    return &Cleaner{
        db:            db,
        retentionDays: retentionDays,
        batchSize:     batchSize,
        interval:      interval,
    }
}

func (c *Cleaner) Start(done <-chan struct{}) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-done:
            return
        case <-ticker.C:
            c.runCleanup()
        }
        }
}

func (c *Cleaner) runCleanup() {
    // 1. 归档旧数据
    // 将超过保留时间的 completed/failed 数据插入到 archive 表
    archiveSQL := `
        INSERT INTO events_archive 
        SELECT * FROM events 
        WHERE status IN ('completed', 'failed') 
          AND updated_at < NOW() - INTERVAL ? DAY
        LIMIT ?
    `
    res, err := c.db.Exec(archiveSQL, c.retentionDays, c.batchSize)
    if err != nil {
        log.Printf("Archive error: %v\n", err)
        return
    }
    rowsAffected, _ := res.RowsAffected()
    if rowsAffected == 0 {
        return // 没有需要清理的数据
    }

    // 2. 删除主表中的旧数据
    // 注意：这里使用相同的条件删除，生产环境建议根据 ID 删除更精确，但为了示例简洁使用条件删除
    deleteSQL := `
        DELETE FROM events 
        WHERE status IN ('completed', 'failed') 
          AND updated_at < NOW() - INTERVAL ? DAY
        LIMIT ?
    `
    _, err = c.db.Exec(deleteSQL, c.retentionDays, c.batchSize)
    if err != nil {
        log.Printf("Delete error: %v\n", err)
        // 如果删除失败，归档数据可能会重复，但不会丢失数据（幂等性）
        return
    }

    log.Printf("Cleaner: Archived and deleted %d events\n", rowsAffected)
}
```

---

### 六、主程序集成 (`main.go`)

将清理协程整合进主程序，并演示延迟队列的使用。

```go
package main

import (
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"
    "yourproject/config"
    "yourproject/db"
    "yourproject/queue"
    "yourproject/models"
)

func main() {
    cfg := config.Load()
    database, err := db.Connect(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
    if err != nil {
        log.Fatal("DB connect failed:", err)
    }
    defer database.Close()

    // 1. 初始化消费者
    consumer := queue.NewConsumer(database, cfg.WorkerNum, time.Duration(cfg.PollIntervalSec)*time.Second)
    consumer.RegisterHandler("user.created", func(e *models.Event) error {
        log.Printf("Processing user.created event: %+v\n", e.Payload)
        // 模拟耗时操作
        time.Sleep(1 * time.Second) 
        return nil
    })

    // 2. 初始化清理器 (保留 7 天数据，每次处理 1000 条，每 1 小时运行一次)
    cleaner := queue.NewCleaner(database, 7, 1000, 1*time.Hour)

    // 3. 启动协程
    done := make(chan struct{})
    consumer.Start(done)
    cleaner.Start(done)

    // 4. 示例：生产事件
    producer := queue.NewProducer(database)
    
    // 立即执行
    _, _ = producer.Enqueue("user.created", map[string]interface{}{
        "user_id": 101,
        "type":    "immediate",
    })

    // 延迟 10 秒执行
    _, _ = producer.EnqueueWithDelay("user.created", map[string]interface{}{
        "user_id": 102,
        "type":    "delayed_10s",
    }, 10*time.Second)

    // 5. 优雅退出
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    
    log.Println("Shutting down gracefully...")
    close(done)
    
    // 给予少量时间处理完当前任务
    time.Sleep(2 * time.Second)
}
```

---

### 七、关键实现细节说明

#### 1. 延迟队列原理
*   **实现方式**：利用 `scheduled_at` 字段 + 轮询。
*   **查询条件**：`WHERE status = 'pending' AND scheduled_at <= NOW()`。
*   **优点**：不需要引入 Redis 或 RabbitMQ 等额外中间件，架构简单。
*   **缺点**：依赖消费者轮询，实时性取决于 `PollInterval`。如果延迟精度要求极高（毫秒级），建议使用 Redis ZSet 实现延迟队列。

#### 2. 归档清理策略
*   **批量处理**：使用 `LIMIT` 防止单次操作锁表时间过长，影响线上业务。
*   **先归档后删除**：保证数据不丢失。即使删除失败，下次清理时会尝试再次归档（`events_archive` 建议设置 `id` 唯一索引或忽略重复插入）。
*   **独立协程**：清理逻辑与业务消费逻辑分离，避免清理任务阻塞事件处理。

#### 3. 性能优化建议
*   **索引**：确保 `(status, scheduled_at)` 和 `(status, updated_at)` 有联合索引，否则 `SELECT` 和 `DELETE` 会全表扫描。
*   **事务粒度**：在 Consumer 中，获取任务后立即 `COMMIT` 事务，将业务逻辑放在事务外执行。这能极大提高并发吞吐量，避免长事务锁住大量行。
*   **分区表**：如果数据量达到千万级，建议对 `events` 表按 `created_at` 进行 MySQL 分区（Partitioning），以便快速 `DROP PARTITION` 来清理旧数据，代替 `DELETE`。

### 八、补充：归档表去重（可选优化）

为了防止清理程序重复归档，可以在 `events_archive` 上对原主键 `id` 做处理，或者在归档时忽略错误。

```sql
-- 在 events_archive 上创建唯一索引，防止重复归档同一 ID
-- 注意：如果 events 表 id 是自增且复用（极少见），则不能这样做。
-- 通常事件表 id 只增不减，可以直接归档。
ALTER TABLE events_archive ADD UNIQUE KEY uk_event_id (id);
```

并在 Cleaner 的 `INSERT` 语句中使用 `INSERT IGNORE` 或 `ON DUPLICATE KEY UPDATE`：

```go
archiveSQL := `
    INSERT IGNORE INTO events_archive 
    SELECT * FROM events 
    WHERE ...
`
```

这样即使清理程序重启或并发运行，也不会导致归档表数据重复。
