# 🤵 我想编写一个sql语句，删除表中的记录，只保留最新的1000行

Qwen3-235B-A22B-2507


要删除表中的记录，只保留最新的1000行（按某个时间字段或自增ID排序），可以使用以下 SQL 语句。假设你的表名为 `your_table`，且有一个可以表示“最新”的字段，比如 `id`（自增主键）或 `created_at`（时间戳）。

### 方法一：基于自增ID（假设 `id` 是主键且递增）

```sql
DELETE FROM your_table
WHERE id NOT IN (
    SELECT id FROM (
        SELECT id
        FROM your_table
        ORDER BY id DESC
        LIMIT 1000
    ) AS latest
);
```

### 方法二：基于时间戳字段（例如 `created_at`）

```sql
DELETE FROM your_table
WHERE created_at NOT IN (
    SELECT created_at FROM (
        SELECT created_at
        FROM your_table
        ORDER BY created_at DESC
        LIMIT 1000
    ) AS latest
);
```

⚠️ **注意**：
- 如果 `created_at` 字段不唯一（多个记录时间相同），可能会误删或保留错误数据。
- 为避免此问题，建议结合 `id` 和时间字段使用。

### 方法三（推荐）：使用窗口函数（适用于 MySQL 8.0+、PostgreSQL、SQL Server 等支持窗口函数的数据库）

```sql
DELETE FROM your_table
WHERE id IN (
    SELECT id FROM (
        SELECT id,
               ROW_NUMBER() OVER (ORDER BY id DESC) AS rn
        FROM your_table
    ) AS ranked
    WHERE rn > 1000
);
```

或者使用 `created_at`：

```sql
DELETE FROM your_table
WHERE id IN (
    SELECT id FROM (
        SELECT id,
               ROW_NUMBER() OVER (ORDER BY created_at DESC) AS rn
        FROM your_table
    ) AS ranked
    WHERE rn > 1000
);
```

### 注意事项：
1. **备份数据**：执行删除前务必备份。
2. **测试**：先在小数据集或测试环境运行。
3. **性能**：大表删除可能较慢，考虑分批删除。
4. **唯一排序字段**：确保排序字段能唯一确定“最新”的记录，避免歧义。

---

✅ **推荐使用方法三（窗口函数）**，逻辑清晰、性能好、不易出错（前提是数据库支持）。

如果你告诉我你使用的数据库类型（MySQL、PostgreSQL、SQL Server等）和表结构，我可以给出更精确的语句。
