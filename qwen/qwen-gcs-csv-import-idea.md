# 我有一个用php开发的监控网站，需要将从各处上传的csv文件导入到数据库中，
问题是，这些csv文件的格式非常混乱，字段名、字段数差别很大，各不相同，
并且随着监控目标的增加，更多格式不同的csv文件会上传上来。
我要把csv文件导入到数据库中，csv文件中有些字段是我需要的，另外一些是我不需要的，
我应该如何做才能应对这种不一致的文件格式，请给出代码示例


# 灵活CSV导入系统解决方案

针对格式不统一的CSV文件导入问题，我推荐采用**配置驱动的映射系统**。以下是完整解决方案：

## 一、系统架构设计

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   CSV文件   │───▶│  格式识别器   │───▶│  字段映射器  │
└─────────────┘    └──────────────┘    └─────────────┘
                                              │
                                              ▼
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   数据库    │◀───│  数据验证器   │◀───│  数据清洗器  │
└─────────────┘    └──────────────┘    └─────────────┘
```

## 二、完整代码实现

### 1. 数据库表结构设计

```sql
-- 监控数据主表（标准格式）
CREATE TABLE monitoring_data (
    id INT AUTO_INCREMENT PRIMARY KEY,
    device_id VARCHAR(50) NOT NULL,
    timestamp DATETIME NOT NULL,
    temperature DECIMAL(5,2),
    humidity DECIMAL(5,2),
    pressure DECIMAL(8,2),
    status VARCHAR(20),
    source_file VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_device_time (device_id, timestamp),
    INDEX idx_timestamp (timestamp)
);

-- CSV格式配置表
CREATE TABLE csv_format_configs (
    id INT AUTO_INCREMENT PRIMARY KEY,
    format_name VARCHAR(100) NOT NULL,
    format_pattern VARCHAR(255),  -- 用于识别格式的特征
    field_mapping JSON NOT NULL,  -- 字段映射配置
    is_active TINYINT(1) DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 2. 核心导入类

```php
<?php
// src/Importer/CsvImporter.php

namespace App\Importer;

class CsvImporter
{
    private $pdo;
    private $configRepository;
    
    public function __construct($pdo, ConfigRepository $configRepo)
    {
        $this->pdo = $pdo;
        $this->configRepository = $configRepo;
    }
    
    /**
     * 导入CSV文件
     */
    public function import(string $filePath, string $sourceName = ''): array
    {
        $result = [
            'success' => false,
            'imported' => 0,
            'failed' => 0,
            'errors' => []
        ];
        
        try {
            // 1. 检测CSV格式
            $format = $this->detectFormat($filePath);
            
            if (!$format) {
                throw new \Exception('无法识别CSV文件格式');
            }
            
            // 2. 获取字段映射配置
            $mapping = $format['field_mapping'];
            
            // 3. 解析CSV
            $rows = $this->parseCsv($filePath);
            
            if (empty($rows)) {
                throw new \Exception('CSV文件为空');
            }
            
            // 4. 处理数据
            $this->pdo->beginTransaction();
            
            foreach ($rows as $index => $row) {
                try {
                    // 跳过标题行
                    if ($index === 0) continue;
                    
                    // 映射字段
                    $mappedData = $this->mapFields($row, $mapping);
                    
                    // 验证数据
                    $validatedData = $this->validateData($mappedData);
                    
                    // 清洗数据
                    $cleanedData = $this->cleanData($validatedData);
                    
                    // 插入数据库
                    $this->insertData($cleanedData, $sourceName ?: basename($filePath));
                    
                    $result['imported']++;
                } catch (\Exception $e) {
                    $result['failed']++;
                    $result['errors'][] = [
                        'row' => $index + 1,
                        'error' => $e->getMessage()
                    ];
                }
            }
            
            $this->pdo->commit();
            $result['success'] = true;
            
        } catch (\Exception $e) {
            if ($this->pdo->inTransaction()) {
                $this->pdo->rollBack();
            }
            $result['errors'][] = ['error' => $e->getMessage()];
        }
        
        return $result;
    }
    
    /**
     * 检测CSV格式
     */
    private function detectFormat(string $filePath): ?array
    {
        $headers = $this->getCsvHeaders($filePath);
        
        // 从配置库中匹配格式
        $formats = $this->configRepository->getActiveFormats();
        
        foreach ($formats as $format) {
            if ($this->matchFormat($headers, $format)) {
                return $format;
            }
        }
        
        // 如果没有匹配，返回默认格式
        return $this->configRepository->getDefaultFormat();
    }
    
    /**
     * 匹配格式
     */
    private function matchFormat(array $headers, array $format): bool
    {
        $mapping = json_decode($format['field_mapping'], true);
        $requiredFields = $mapping['required'] ?? [];
        
        // 检查必需字段是否存在
        foreach ($requiredFields as $required) {
            if (!in_array($required['source'], $headers)) {
                return false;
            }
        }
        
        return true;
    }
    
    /**
     * 获取CSV标题行
     */
    private function getCsvHeaders(string $filePath): array
    {
        $handle = fopen($filePath, 'r');
        $headers = fgetcsv($handle);
        fclose($handle);
        
        return array_map('trim', $headers);
    }
    
    /**
     * 解析CSV文件
     */
    private function parseCsv(string $filePath): array
    {
        $rows = [];
        $handle = fopen($filePath, 'r');
        
        while (($row = fgetcsv($handle)) !== false) {
            $rows[] = $row;
        }
        
        fclose($handle);
        return $rows;
    }
    
    /**
     * 字段映射
     */
    private function mapFields(array $row, array $mapping): array
    {
        $headers = $this->getCsvHeadersFromRow($row, $mapping);
        $mapped = [];
        
        foreach ($mapping['fields'] as $target => $config) {
            $source = $config['source'];
            $value = $row[array_search($source, $headers)] ?? null;
            
            // 应用转换规则
            if (isset($config['transform'])) {
                $value = $this->applyTransform($value, $config['transform']);
            }
            
            $mapped[$target] = $value;
        }
        
        return $mapped;
    }
    
    /**
     * 应用转换规则
     */
    private function applyTransform($value, string $transform)
    {
        switch ($transform) {
            case 'datetime':
                return date('Y-m-d H:i:s', strtotime($value));
            case 'float':
                return floatval($value);
            case 'int':
                return intval($value);
            case 'uppercase':
                return strtoupper($value);
            case 'lowercase':
                return strtolower($value);
            default:
                return $value;
        }
    }
    
    /**
     * 数据验证
     */
    private function validateData(array $data): array
    {
        $rules = [
            'device_id' => 'required|string',
            'timestamp' => 'required|datetime',
            'temperature' => 'numeric|min:-50|max:100',
            'humidity' => 'numeric|min:0|max:100',
        ];
        
        foreach ($rules as $field => $rule) {
            if (strpos($rule, 'required') !== false && empty($data[$field])) {
                throw new \Exception("字段 {$field} 不能为空");
            }
        }
        
        return $data;
    }
    
    /**
     * 数据清洗
     */
    private function cleanData(array $data): array
    {
        // 去除空值
        $data = array_filter($data, function($v) {
            return $v !== null && $v !== '';
        });
        
        // 清理字符串
        foreach ($data as $key => $value) {
            if (is_string($value)) {
                $data[$key] = trim($value);
            }
        }
        
        return $data;
    }
    
    /**
     * 插入数据库
     */
    private function insertData(array $data, string $sourceFile): void
    {
        $fields = array_keys($data);
        $fields[] = 'source_file';
        $data['source_file'] = $sourceFile;
        
        $placeholders = array_map(fn($f) => ":$f", $fields);
        
        $sql = "INSERT INTO monitoring_data (" . implode(',', $fields) . ") 
                VALUES (" . implode(',', $placeholders) . ")";
        
        $stmt = $this->pdo->prepare($sql);
        $stmt->execute($data);
    }
    
    private function getCsvHeadersFromRow(array $row, array $mapping): array
    {
        // 简化处理，实际应从文件读取
        return array_keys($mapping['fields']);
    }
}
```

### 3. 配置仓库类

```php
<?php
// src/Importer/ConfigRepository.php

namespace App\Importer;

class ConfigRepository
{
    private $pdo;
    
    public function __construct($pdo)
    {
        $this->pdo = $pdo;
    }
    
    /**
     * 获取所有活跃格式配置
     */
    public function getActiveFormats(): array
    {
        $stmt = $this->pdo->query(
            "SELECT * FROM csv_format_configs WHERE is_active = 1"
        );
        return $stmt->fetchAll(\PDO::FETCH_ASSOC);
    }
    
    /**
     * 获取默认格式
     */
    public function getDefaultFormat(): ?array
    {
        $stmt = $this->pdo->query(
            "SELECT * FROM csv_format_configs WHERE format_name = 'default' LIMIT 1"
        );
        $format = $stmt->fetch(\PDO::FETCH_ASSOC);
        
        if (!$format) {
            return $this->createDefaultFormat();
        }
        
        return $format;
    }
    
    /**
     * 注册新格式
     */
    public function registerFormat(string $name, array $mapping, string $pattern = ''): int
    {
        $stmt = $this->pdo->prepare(
            "INSERT INTO csv_format_configs (format_name, format_pattern, field_mapping) 
             VALUES (:name, :pattern, :mapping)"
        );
        
        $stmt->execute([
            'name' => $name,
            'pattern' => $pattern,
            'mapping' => json_encode($mapping)
        ]);
        
        return $this->pdo->lastInsertId();
    }
    
    private function createDefaultFormat(): array
    {
        $defaultMapping = [
            'required' => [
                ['source' => 'device_id', 'target' => 'device_id']
            ],
            'fields' => [
                'device_id' => ['source' => 'device_id', 'transform' => 'string'],
                'timestamp' => ['source' => 'time', 'transform' => 'datetime'],
                'temperature' => ['source' => 'temp', 'transform' => 'float'],
                'humidity' => ['source' => 'hum', 'transform' => 'float'],
                'pressure' => ['source' => 'press', 'transform' => 'float'],
                'status' => ['source' => 'status', 'transform' => 'uppercase']
            ]
        ];
        
        return [
            'format_name' => 'default',
            'field_mapping' => json_encode($defaultMapping)
        ];
    }
}
```

### 4. 格式配置示例

```php
<?php
// 注册不同格式的CSV配置

$configRepo = new ConfigRepository($pdo);

// 格式1: 传感器A
$configRepo->registerFormat('sensor_type_a', [
    'required' => [
        ['source' => 'DeviceID']
    ],
    'fields' => [
        'device_id' => ['source' => 'DeviceID', 'transform' => 'string'],
        'timestamp' => ['source' => 'Timestamp', 'transform' => 'datetime'],
        'temperature' => ['source' => 'Temp_C', 'transform' => 'float'],
        'humidity' => ['source' => 'Humidity_%', 'transform' => 'float'],
        'pressure' => ['source' => 'Pressure_hPa', 'transform' => 'float'],
        'status' => ['source' => 'Status', 'transform' => 'uppercase']
    ]
], 'DeviceID,Timestamp');

// 格式2: 传感器B
$configRepo->registerFormat('sensor_type_b', [
    'required' => [
        ['source' => '设备编号']
    ],
    'fields' => [
        'device_id' => ['source' => '设备编号', 'transform' => 'string'],
        'timestamp' => ['source' => '采集时间', 'transform' => 'datetime'],
        'temperature' => ['source' => '温度', 'transform' => 'float'],
        'humidity' => ['source' => '湿度', 'transform' => 'float'],
        'pressure' => ['source' => '气压', 'transform' => 'float'],
        'status' => ['source' => '状态', 'transform' => 'string']
    ]
], '设备编号，采集时间');

// 格式3: 传感器C（英文不同命名）
$configRepo->registerFormat('sensor_type_c', [
    'required' => [
        ['source' => 'sensor_id']
    ],
    'fields' => [
        'device_id' => ['source' => 'sensor_id', 'transform' => 'string'],
        'timestamp' => ['source' => 'recorded_at', 'transform' => 'datetime'],
        'temperature' => ['source' => 'temp_value', 'transform' => 'float'],
        'humidity' => ['source' => 'humidity_value', 'transform' => 'float'],
        'pressure' => ['source' => 'pressure_value', 'transform' => 'float'],
        'status' => ['source' => 'device_status', 'transform' => 'string']
    ]
], 'sensor_id,recorded_at');
```

### 5. 使用示例

```php
<?php
// upload.php - 文件上传处理

require_once 'vendor/autoload.php';

use App\Importer\CsvImporter;
use App\Importer\ConfigRepository;

// 初始化
$pdo = new PDO('mysql:host=localhost;dbname=monitoring', 'user', 'pass');
$configRepo = new ConfigRepository($pdo);
$importer = new CsvImporter($pdo, $configRepo);

// 处理上传
if ($_FILES['csv_file']['error'] === UPLOAD_ERR_OK) {
    $tmpFile = $_FILES['csv_file']['tmp_name'];
    $sourceName = $_FILES['csv_file']['name'];
    
    // 执行导入
    $result = $importer->import($tmpFile, $sourceName);
    
    // 返回结果
    header('Content-Type: application/json');
    echo json_encode($result, JSON_PRETTY_PRINT);
    
    // 清理临时文件
    unlink($tmpFile);
}
```

### 6. 前端上传界面

```html
<!-- upload.html -->
<!DOCTYPE html>
<html>
<head>
    <title>CSV文件导入</title>
</head>
<body>
    <h1>上传监控数据CSV</h1>
    
    <form id="uploadForm" enctype="multipart/form-data">
        <input type="file" name="csv_file" accept=".csv" required>
        <button type="submit">导入</button>
    </form>
    
    <div id="result"></div>
    
    <script>
    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        
        const formData = new FormData(e.target);
        
        const response = await fetch('upload.php', {
            method: 'POST',
            body: formData
        });
        
        const result = await response.json();
        
        document.getElementById('result').innerHTML = `
            <div>成功: ${result.imported}</div>
            <div>失败: ${result.failed}</div>
            <div>错误: ${JSON.stringify(result.errors)}</div>
        `;
    });
    </script>
</body>
</html>
```

## 三、扩展功能

### 1. 自动格式学习

```php
<?php
// 智能学习新格式
class FormatLearner
{
    public function learnFromSample(array $headers, array $sampleData): array
    {
        $mapping = [
            'required' => [],
            'fields' => []
        ];
        
        // 智能匹配字段
        $fieldPatterns = [
            'device_id' => ['/device/i', '/id/i', '/编号/i'],
            'timestamp' => ['/time/i', '/date/i', '/时间/i'],
            'temperature' => ['/temp/i', '/温度/i'],
            'humidity' => ['/humid/i', '/湿度/i'],
            'pressure' => ['/press/i', '/气压/i'],
            'status' => ['/status/i', '/状态/i']
        ];
        
        foreach ($headers as $header) {
            foreach ($fieldPatterns as $target => $patterns) {
                foreach ($patterns as $pattern) {
                    if (preg_match($pattern, $header)) {
                        $mapping['fields'][$target] = [
                            'source' => $header,
                            'transform' => $this->guessTransform($sampleData[$header])
                        ];
                        break;
                    }
                }
            }
        }
        
        return $mapping;
    }
    
    private function guessTransform($value): string
    {
        if (is_numeric($value)) {
            return strpos($value, '.') !== false ? 'float' : 'int';
        }
        if (strtotime($value)) {
            return 'datetime';
        }
        return 'string';
    }
}
```

### 2. 批量导入处理

```php
<?php
// 支持批量导入
class BatchImporter
{
    public function importDirectory(string $directory): array
    {
        $files = glob($directory . '/*.csv');
        $results = [];
        
        foreach ($files as $file) {
            $results[basename($file)] = $this->importer->import($file);
        }
        
        return $results;
    }
}
```

## 四、最佳实践建议

1. **配置管理**：将格式配置存储在数据库或配置文件中，便于管理
2. **日志记录**：记录每次导入的详细信息，便于排查问题
3. **数据验证**：在插入前进行严格的数据验证
4. **错误处理**：部分失败不影响其他数据导入
5. **性能优化**：大批量数据使用批量插入
6. **格式版本**：支持格式配置的版本管理

这个方案的优势：
- ✅ 灵活应对不同格式
- ✅ 易于扩展新格式
- ✅ 配置与代码分离
- ✅ 支持自动格式识别
- ✅ 完善的错误处理

希望这个解决方案能帮助你构建一个健壮的CSV导入系统！ 😊
