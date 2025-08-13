# 🤵 列举一些在使用mysql时容易踩到的坑，以及如何避免

在使用 MySQL 时，开发者常常因为对数据库特性、配置或 SQL 语法理解不充分而“踩坑”。
以下是一些常见的陷阱以及相应的避免方法：

---

### 1. **字符集和排序规则不一致导致乱码或比较错误**
**问题表现**：
- 插入中文乱码
- 字符串比较时大小写敏感或不敏感行为异常
- 跨表 JOIN 时出现错误

**原因**：
- 表、列、连接、客户端、服务器的字符集设置不一致（如 `utf8` vs `utf8mb4`）

**避免方法**：
```sql
-- 创建数据库时指定字符集
CREATE DATABASE db_name CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建表时指定
CREATE TABLE t (
    name VARCHAR(50)
) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 连接时设置
SET NAMES utf8mb4;
```
✅ 建议统一使用 `utf8mb4`（支持 emoji）和 `utf8mb4_unicode_ci` 排序规则。

---

### 2. **NULL 值处理不当**
**问题表现**：
- `WHERE column = NULL` 永远不成立
- `COUNT(*)` 和 `COUNT(column)` 结果不同
- `IN`、`NOT IN` 与 NULL 一起使用导致意外结果

**示例**：
```sql
SELECT * FROM users WHERE age NOT IN (20, NULL); -- 返回空结果
```

**避免方法**：
- 使用 `IS NULL` 或 `IS NOT NULL`
- 用 `NOT EXISTS` 替代 `NOT IN` 避免 NULL 陷阱
- 查询时使用 `COALESCE()` 或 `IFNULL()` 处理 NULL

```sql
SELECT * FROM users WHERE NOT EXISTS (
    SELECT 1 FROM blacklist b WHERE b.age = users.age
);
```

---

### 3. **索引未生效（隐式类型转换）**
**问题表现**：
- 查询很慢，明明有索引却未使用

**原因**：
- 字段类型与查询值类型不匹配，如：
  ```sql
  -- phone 是 VARCHAR，但用数字查询
  SELECT * FROM users WHERE phone = 13812345678; -- 触发隐式转换，索引失效
  ```

**避免方法**：
- 保证查询值类型与字段一致
- 使用 EXPLAIN 分析执行计划
```sql
EXPLAIN SELECT * FROM users WHERE phone = '13812345678';
```

---

### 4. **事务未正确使用或隔离级别误解**
**问题表现**：
- 并发写入时数据不一致
- 事务中读取不到最新数据（如 REPEATABLE READ 下的幻读）

**避免方法**：
- 显式开启事务：`START TRANSACTION; ... COMMIT;`
- 根据业务选择隔离级别：
  ```sql
  SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED;
  ```
- 注意长事务导致的锁等待和 undo log 膨胀

---

### 5. **自动提交（autocommit）设置不当**
**问题表现**：
- 每条语句自动提交，无法回滚
- 误删数据无法恢复

**避免方法**：
- 在需要事务的场景中关闭自动提交：
  ```sql
  SET autocommit = 0;
  -- 执行多条语句
  COMMIT; -- 或 ROLLBACK;
  SET autocommit = 1;
  ```

---

### 6. **LIMIT 和 OFFSET 分页性能差**
**问题表现**：
- `LIMIT 1000000, 10` 查询极慢

**原因**：
- MySQL 仍需扫描前 1000000 条记录

**避免方法**：
- 使用“游标分页”（基于上一页最后一条记录的 ID）
  ```sql
  SELECT * FROM users WHERE id > last_id ORDER BY id LIMIT 10;
  ```
- 或使用覆盖索引 + 延迟关联

---

### 7. **唯一索引缺失导致重复数据**
**问题表现**：
- 插入重复数据，业务逻辑出错

**避免方法**：
- 对业务上要求唯一的字段建立唯一索引
  ```sql
  ALTER TABLE users ADD UNIQUE INDEX uk_email (email);
  ```
- 使用 `INSERT IGNORE` 或 `ON DUPLICATE KEY UPDATE`

---

### 8. **大字段（TEXT/BLOB）影响性能**
**问题表现**：
- 查询变慢，即使只查主键
- 内存临时表变成磁盘临时表

**避免方法**：
- 避免在 SELECT * 中包含大字段
- 将大字段拆到单独的表中（垂直分表）
- 使用 `pt-online-schema-change` 工具修改大表结构

---

### 9. **未使用连接池或连接泄漏**
**问题表现**：
- `Too many connections` 错误
- 应用响应变慢

**避免方法**：
- 使用连接池（如 HikariCP、Druid）
- 及时关闭数据库连接（使用 try-with-resources 或 finally）
- 监控连接数，调整 `max_connections`

---

### 10. **误用 ORDER BY RAND()**
**问题表现**：
- `SELECT * FROM users ORDER BY RAND() LIMIT 1;` 随着数据量增长极慢

**原因**：
- 每行都计算随机值，需全表扫描 + 文件排序

**避免方法**：
- 使用随机主键或预生成随机值字段 + 索引
- 用应用层生成随机偏移（注意分页跳跃问题）

---

### 11. **未设置合适的主键或使用 UUID 作为主键**
**问题表现**：
- 插入性能下降
- 索引碎片严重

**原因**：
- UUID 是无序的，导致 B+ 树频繁分裂

**避免方法**：
- 使用自增主键（InnoDB 推荐）
- 若必须用 UUID，可考虑 `UUID_TO_BIN()` + 重排版本（如去掉 `-` 并调整时间部分位置）

---

### 12. **忽视慢查询日志和性能分析**
**问题表现**：
- 数据库负载高，但不知道瓶颈在哪

**避免方法**：
- 开启慢查询日志：
  ```sql
  SET GLOBAL slow_query_log = 'ON';
  SET GLOBAL long_query_time = 1;
  ```
- 使用 `pt-query-digest` 分析慢日志
- 定期使用 `EXPLAIN` 分析关键 SQL

---

### 13. **ALTER TABLE 阻塞写入（尤其大表）**
**问题表现**：
- 加字段时表锁，服务不可用

**避免方法**：
- 使用在线 DDL 工具（如 `pt-online-schema-change` 或 `gh-ost`）
- MySQL 5.6+ 支持部分 `ALTER` 操作不锁表（如加索引）
- 避免在高峰期执行 DDL

---

### 14. **误删数据且无备份**
**问题表现**：
- `DELETE FROM users;` 未加 WHERE，数据全丢

**避免方法**：
- 开启 `sql_safe_updates` 模式
  ```sql
  SET sql_safe_updates = 1;
  ```
- 操作前备份：`mysqldump` 或使用 binlog 恢复
- 使用 `DELETE` 前先 `SELECT` 验证条件

---

### 15. **忽略 binlog 和主从复制配置**
**问题表现**：
- 主从延迟严重
- 故障恢复时无法找回数据

**避免方法**：
- 正确配置 `binlog_format=ROW`（推荐）
- 定期检查主从状态：`SHOW SLAVE STATUS`
- 启用 GTID 简化复制管理

---

### 总结：最佳实践建议

| 类别 | 建议 |
|------|------|
| 字符集 | 统一使用 `utf8mb4` |
| 索引 | 合理设计，避免隐式转换 |
| 事务 | 显式控制，避免长事务 |
| NULL | 使用 `IS NULL`，避免 `NOT IN` |
| 分页 | 避免大偏移，使用游标 |
| 安全 | 开启 `sql_safe_updates`，备份 |
| 性能 | 开启慢查询日志，定期优化 |

---

通过了解这些常见“坑”，并结合良好的开发规范和监控机制，可以显著提升 MySQL 的稳定性与性能。
#-------------------------------------------------------------------------------

