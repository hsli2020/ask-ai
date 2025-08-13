# 我有一个用php开发的监控网站，需要将从各处上传的csv文件导入到数据库中，
问题是，这些csv文件的格式非常混乱，字段名、字段数差别很大，各不相同，
并且随着监控目标的增加，更多格式不同的csv文件会上传上来。
我要把csv文件导入到数据库中，csv文件中有些字段是我需要的，另外一些是我不需要的，
我应该如何做才能应对这种不一致的文件格式，请给出代码示例


面对这种“字段名不一、字段数量不定、格式混乱”的CSV导入需求，核心解决思路是**“配置化映射”**。你不需要为每种CSV写一段代码，而是建立一套规则，将CSV的“原始字段”映射到你数据库的“标准字段”。

以下是具体的解决方案和代码示例。

### 1. 核心设计思路

#### A. 数据库设计
你需要两张核心表：
1.  **数据主表 (`monitor_data`)**：存储最终清洗后的标准化数据。
2.  **映射规则表 (`import_mappings`)**：用于定义哪种CSV格式对应哪些字段。

**表结构示例：**

```sql
-- 最终存储数据的标准表
CREATE TABLE monitor_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    server_ip VARCHAR(50),      -- 标准字段：IP
    cpu_usage FLOAT,            -- 标准字段：CPU使用率
    memory_usage FLOAT,         -- 标准字段：内存使用率
    log_time DATETIME,          -- 标准字段：时间
    raw_content TEXT,           -- 可选：存储原始行数据以备查
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 映射规则配置表
CREATE TABLE import_mappings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    source_name VARCHAR(100),   -- 来源标识，如 "阿里云监控", "Zabbix服务器A"
    csv_header_name VARCHAR(100), -- CSV文件中的原始表头名（如 "IP地址", "CPU%"）
    db_field_name VARCHAR(50),    -- 对应数据库字段名（如 "server_ip", "cpu_usage"）
    is_required TINYINT DEFAULT 0 -- 是否必须字段
);

-- 插入一些示例规则
-- 假设 "阿里云" 的CSV里叫 "IP地址"，对应 db 的 "server_ip"
INSERT INTO import_mappings (source_name, csv_header_name, db_field_name) VALUES 
('aliyun', 'IP地址', 'server_ip'),
('aliyun', 'CPU利用率(%)', 'cpu_usage'),
('zabbix', 'Host IP', 'server_ip'),
('zabbix', 'Memory Used', 'memory_usage');
```

#### B. 处理流程
1.  **识别来源**：上传时，让用户选择这是哪个监控源的文件（或者根据文件名/表头特征自动识别）。
2.  **读取表头**：读取CSV第一行，获取所有字段名。
3.  **查找映射**：从数据库查询该来源的映射规则。
4.  **数据清洗**：遍历CSV数据行，根据映射规则提取需要的字段，忽略不需要的字段。
5.  **入库**：组装SQL语句插入数据库。

---

### 2. PHP 代码实现示例

这是一个简化版的 PHP 类，演示如何实现动态映射导入。

```php
<?php

class CsvImporter {
    private $pdo;

    public function __construct($pdo) {
        $this->pdo = $pdo;
    }

    /**
     * 导入主函数
     * @param string $filePath CSV文件路径
     * @param string $sourceName 来源标识 (对应数据库 import_mappings.source_name)
     */
    public function import($filePath, $sourceName) {
        // 1. 检查文件
        if (!file_exists($filePath)) {
            throw new Exception("文件不存在");
        }

        // 2. 获取映射规则
        // 返回格式：['CSV原始字段名' => '数据库字段名']
        $mappingRules = $this->getMappingRules($sourceName);
        if (empty($mappingRules)) {
            throw new Exception("未找到该来源 [{$sourceName}] 的映射配置，请先在后台配置。");
        }

        // 3. 打开文件
        $handle = fopen($filePath, 'r');
        if ($handle === false) {
            throw new Exception("无法打开文件");
        }

        // 4. 读取表头 (假设第一行是表头)
        $csvHeaders = fgetcsv($handle);
        if ($csvHeaders === false) {
            throw new Exception("无法读取CSV表头");
        }

        // 处理BOM头（很多UTF-8 CSV文件有BOM头，会导致第一个字段名识别失败）
        $csvHeaders[0] = preg_replace('/^\x{FEFF}/u', '', $csvHeaders[0]);

        // 5. 建立索引映射
        // 目的：找出CSV列的索引位置，与数据库字段的对应关系
        // 结构：[ 列索引 => 数据库字段名 ]
        $columnMap = [];
        foreach ($csvHeaders as $index => $headerName) {
            // 去除首尾空格，转小写进行匹配（增强容错）
            $cleanHeader = trim($headerName);
            
            // 查找该表头是否在我们的配置规则中
            if (isset($mappingRules[$cleanHeader])) {
                $dbField = $mappingRules[$cleanHeader];
                $columnMap[$index] = $dbField;
            }
            // 如果没有匹配到，说明这个字段不需要，直接忽略
        }

        if (empty($columnMap)) {
            throw new Exception("CSV表头与配置规则无法匹配，请检查表头名称是否正确。");
        }

        // 6. 开始读取数据行并导入
        $this->pdo->beginTransaction();
        try {
            $stmt = $this->pdo->prepare("INSERT INTO monitor_data (server_ip, cpu_usage, memory_usage, log_time, raw_content) VALUES (:server_ip, :cpu_usage, :memory_usage, :log_time, :raw_content)");
            
            $rowCount = 0;
            while (($row = fgetcsv($handle)) !== false) {
                // 跳过空行
                if (array_filter($row) === []) continue;

                $data = [];
                $rawLine = implode(',', $row); // 保存原始数据

                // 根据映射提取数据
                foreach ($columnMap as $csvIndex => $dbField) {
                    // 如果CSV该列存在，取值，否则设为null
                    $value = isset($row[$csvIndex]) ? trim($row[$csvIndex]) : null;
                    
                    // 这里可以加简单的清洗逻辑
                    // 例如：去掉百分号，转浮点数
                    if (in_array($dbField, ['cpu_usage', 'memory_usage'])) {
                        $value = floatval(str_replace('%', '', $value));
                    }
                    
                    $data[$dbField] = $value;
                }

                // 填充未映射的字段为null
                $finalData = [
                    ':server_ip' => $data['server_ip'] ?? null,
                    ':cpu_usage' => $data['cpu_usage'] ?? null,
                    ':memory_usage' => $data['memory_usage'] ?? null,
                    ':log_time' => $data['log_time'] ?? date('Y-m-d H:i:s'), // 默认当前时间
                    ':raw_content' => $rawLine
                ];

                $stmt->execute($finalData);
                $rowCount++;
            }

            $this->pdo->commit();
            fclose($handle);
            return "导入成功，共处理 {$rowCount} 条数据。";

        } catch (Exception $e) {
            $this->pdo->rollBack();
            fclose($handle);
            throw $e;
        }
    }

    /**
     * 获取映射规则
     */
    private function getMappingRules($sourceName) {
        $stmt = $this->pdo->prepare("SELECT csv_header_name, db_field_name FROM import_mappings WHERE source_name = ?");
        $stmt->execute([$sourceName]);
        $rules = $stmt->fetchAll(PDO::FETCH_KEY_PAIR);
        
        // 返回 ['CSV原始名' => 'DB字段名']
        return $rules;
    }
}

// ================= 使用示例 =================

// 数据库连接
$pdo = new PDO('mysql:host=localhost;dbname=test', 'root', 'password');

try {
    $importer = new CsvImporter($pdo);
    
    // 假设用户上传了一个文件，并指定来源为 'aliyun'
    // 对应数据库配置：'IP地址' -> 'server_ip', 'CPU利用率(%)' -> 'cpu_usage'
    $message = $importer->import('/tmp/uploaded_file.csv', 'aliyun');
    echo $message;

} catch (Exception $e) {
    echo "导入失败: " . $e->getMessage();
}
```

### 3. 关键点解析

#### 1. 字段名不一致怎么解决？
通过 `import_mappings` 表解决。如果新来一个监控源“华为云”，它的CSV里IP叫“PrivateIp”，你只需要在数据库里加一条记录：
```sql
INSERT INTO import_mappings (source_name, csv_header_name, db_field_name) VALUES ('huawei', 'PrivateIp', 'server_ip');
```
代码完全不用改。

#### 2. 字段数量差别很大怎么办？
代码中的 `$columnMap` 逻辑解决了这个问题。
*   CSV有20列，但你只需要2列：代码只会匹配这2列的索引，其他18列在循环中被自动忽略。
*   CSV只有5列，但你以前有10个字段：代码只会匹配存在的字段，不存在的字段会以 `null` 插入数据库。

#### 3. 数据格式清洗（如“80%” vs “0.8”）
在 `while` 循环中，针对特定字段（如 `cpu_usage`）做了一层处理：
```php
if (in_array($dbField, ['cpu_usage', 'memory_usage'])) {
    $value = floatval(str_replace('%', '', $value));
}
```
这样无论上传的是 "50%" 还是 "50"，都能正确存入浮点数字段。如果格式极度混乱，建议将数据先存入一个 `text` 字段，后台再通过脚本二次清洗。

### 4. 扩展建议

对于实际生产环境，建议增加以下功能：

1.  **JSON 存储法（备选方案）**：
    如果不想频繁修改数据库表结构，可以将监控数据表设计为 `id, source, data_json, created_at`。
    `data_json` 字段直接存 JSON 格式。这样无论CSV有多少列，直接转成JSON存进去。查询时使用 MySQL 的 JSON 函数提取数据。这适合字段极其多且不固定的情况。

2.  **自动识别来源**：
    上传时不必让用户选来源。你可以取CSV的前几行特征（比如包含“CPU%”这个表头），去匹配规则表，自动判断这是哪种监控源的文件。

3.  **表头容错**：
    用户的CSV表头可能带有空格、换行符或BOM头。在匹配前务必对 `$csvHeaders` 做 `trim()` 和编码清洗，否则经常匹配不上。

