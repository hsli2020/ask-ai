# 我有一个mysql数据库，随着数据增多，占用了越来越多的磁盘空间，我删除了很多过期的数据，但磁盘空间并没得到释放，我该如何做才能回收磁盘空间？

在 MySQL 中，即使你使用 `DELETE` 语句删除了大量数据，磁盘空间**不一定立即释放**。这是因为 MySQL 的存储引擎（尤其是 InnoDB）在删除数据时并不会自动收缩数据文件（`.ibd` 文件或共享表空间 `ibdata1`）。以下是几种有效的方法来**真正回收磁盘空间**：

---

### ✅ 1. 使用 `OPTIMIZE TABLE`（推荐用于 InnoDB 表）
`OPTIMIZE TABLE` 会重建表并释放未使用的空间。

```sql
OPTIMIZE TABLE your_table_name;
```

**效果：**
- 重建表并整理碎片。
- 释放未使用的空间回操作系统（对独立表空间有效）。
- 对 InnoDB 引擎有效（前提是 `innodb_file_per_table = ON`）。

> ⚠️ 注意：`OPTIMIZE TABLE` 会锁表，建议在低峰期执行。

---

### ✅ 2. 确认 `innodb_file_per_table` 是否开启（关键！）

检查是否启用了独立表空间：

```sql
SHOW VARIABLES LIKE 'innodb_file_per_table';
```

- 如果值为 `ON`：每个表的数据存储在独立的 `.ibd` 文件中，**删除数据后可通过 `OPTIMIZE TABLE` 释放空间给操作系统**。
- 如果值为 `OFF`：所有表共享 `ibdata1` 文件，**即使删除数据也无法释放空间**。

#### 🔧 如何开启（如果尚未开启）：
在 `my.cnf` 或 `my.ini` 配置文件中添加：

```ini
[mysqld]
innodb_file_per_table = ON
```

> ⚠️ 修改后只对**新创建的表**生效，已有表仍使用共享表空间。

---

### ✅ 3. 重建表（替代 OPTIMIZE TABLE）
如果你不能使用 `OPTIMIZE TABLE`（如锁表时间太长），可以用：

```sql
ALTER TABLE your_table_name ENGINE=InnoDB;
```

效果与 `OPTIMIZE TABLE` 类似，会重建表并释放空间。

---

### ✅ 4. 对于大表：使用 `pt-online-schema-change`（推荐生产环境）

使用 Percona Toolkit 的 `pt-online-schema-change` 工具可以在不锁表的情况下重建表，避免服务中断。

示例：
```bash
pt-online-schema-change --alter "ENGINE=InnoDB" D=your_db,t=your_table --execute
```

---

### ✅ 5. 删除整个表后重建（极端情况）
如果你删除了几乎全部数据，可以考虑：

```sql
RENAME TABLE old_table TO old_table_backup;
-- 重新创建表结构
CREATE TABLE old_table (...);
-- 导入剩余数据（如有）
INSERT INTO old_table SELECT * FROM old_table_backup WHERE ...;
-- 删除备份表
DROP TABLE old_table_backup;
```

这能彻底释放空间。

---

### ❌ 为什么 `DELETE` 不释放磁盘空间？
- `DELETE` 只是标记数据为“可覆盖”，空间仍被 InnoDB 保留用于后续插入。
- 数据页中的碎片不会自动整理。
- 共享表空间（`ibdata1`）永远不会缩小。

---

### ✅ 6. 清理二进制日志（binlog）——额外空间优化
如果你启用了 binlog，旧日志可能占用大量空间。

查看 binlog 文件：

```sql
SHOW BINARY LOGS;
```

清理过期日志（例如保留最近7天）：

```sql
PURGE BINARY LOGS BEFORE DATE_SUB(NOW(), INTERVAL 7 DAY);
```

或在配置中设置自动清理：

```ini
expire_logs_days = 7
```

> 注意：MySQL 8.0+ 使用 `binlog_expire_logs_seconds`。

---

### ✅ 总结：回收空间的推荐步骤

1. 确认 `innodb_file_per_table = ON`
2. 对大表执行 `OPTIMIZE TABLE` 或 `ALTER TABLE ... ENGINE=InnoDB`
3. 清理不必要的 binlog
4. 监控磁盘使用情况：`du -h /var/lib/mysql/`

---

### 🔍 验证空间是否释放
查看表的磁盘使用情况：

```sql
SELECT 
    table_name,
    round(((data_length + index_length) / 1024 / 1024), 2) `Size (MB)`
FROM information_schema.tables
WHERE table_schema = 'your_database_name'
ORDER BY (data_length + index_length) DESC;
```

同时用系统命令查看文件大小变化：

```bash
du -h /var/lib/mysql/your_database/
```

---

如果你提供具体的表结构、引擎类型和 MySQL 版本，我可以给出更精确的建议。
