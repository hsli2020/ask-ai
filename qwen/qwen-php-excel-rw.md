# 用php编写一个程序，打开给出的execel文件，在当前sheet中，从第二行开始，将第三列和第四列填上内容

要使用 PHP 操作 Excel 文件（读取并写入），推荐使用 **PhpSpreadsheet** 库，这是目前最流行、功能强大的 PHP Excel 处理库（是 PHPExcel 的继任者）。

下面是一个完整的示例程序，实现以下功能：

- 打开一个已存在的 Excel 文件（.xlsx 格式）；
- 获取当前活动的工作表（active sheet）；
- 从第 2 行开始（即跳过表头），向第 3 列（C 列）和第 4 列（D 列）写入指定内容；
- 保存修改后的文件。

---

### ✅ 步骤 1：安装 PhpSpreadsheet

使用 Composer 安装：

```bash
composer require phpoffice/phpspreadsheet
```

---

### ✅ 步骤 2：PHP 代码示例

```php
<?php

require 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\IOFactory;

// 配置
$inputFileName = 'example.xlsx';      // 输入的 Excel 文件路径
$outputFileName = 'output.xlsx';      // 输出文件（可与输入相同，但建议先备份）

// 要填充的数据（示例：从第2行开始，每行填入 C 和 D 列的内容）
// 你可以根据实际需求动态生成这些数据
$dataToFill = [
    ['张三', '工程师'],
    ['李四', '设计师'],
    ['王五', '产品经理'],
    // 可以继续添加更多行
];

try {
    // 1. 加载 Excel 文件
    $spreadsheet = IOFactory::load($inputFileName);
    
    // 2. 获取当前活动的工作表
    $worksheet = $spreadsheet->getActiveSheet();

    // 3. 从第2行开始填充（行号从1开始）
    $startRow = 2;
    foreach ($dataToFill as $index => $row) {
        $currentRow = $startRow + $index;
        // 第3列是 C，第4列是 D
        $worksheet->setCellValue("C{$currentRow}", $row[0]); // 第三列
        $worksheet->setCellValue("D{$currentRow}", $row[1]); // 第四列
    }

    // 4. 保存文件
    $writer = IOFactory::createWriter($spreadsheet, 'Xlsx');
    $writer->save($outputFileName);

    echo "Excel 文件已成功更新并保存为 {$outputFileName}\n";

} catch (Exception $e) {
    die('错误: ' . $e->getMessage());
}
```

---

### 📝 说明：

- `example.xlsx` 是你要打开的原始 Excel 文件，请确保它存在且格式为 `.xlsx`。
- `$dataToFill` 是你要写入的数据数组，每个子数组包含两个元素，分别对应 C 列和 D 列。
- 如果你想覆盖原文件，可以将 `$outputFileName` 设为和 `$inputFileName` 相同，但**强烈建议先备份原文件**。
- 该脚本假设你从第 2 行开始写入，并且数据行数与 `$dataToFill` 一致。你可以根据实际行数动态判断（例如读取最大行号）。

---

### 🔁 如果你想动态填充（比如填满已有数据的行）：

你可以先读取已有行数，再循环填充：

```php
$highestRow = $worksheet->getHighestRow(); // 获取最大行号
for ($row = 2; $row <= $highestRow; $row++) {
    $worksheet->setCellValue("C{$row}", "默认值C");
    $worksheet->setCellValue("D{$row}", "默认值D");
}
```

---

如有更多需求（如处理 .xls 文件、多 sheet、格式设置等），也可以进一步扩展。

需要我帮你根据具体场景调整代码吗？

# 如何取得指定单元格的内容

在 PhpSpreadsheet 中，获取指定单元格的内容非常简单。以下是详细的方法和示例：

## 🔹 基本语法

```php
// 获取单元格的值
$value = $worksheet->getCell('A1')->getValue();

// 或者使用行列坐标（行和列从1开始）
$value = $worksheet->getCellByColumnAndRow(1, 1)->getValue(); // A1 单元格
```

## 🔸 完整示例代码

```php
<?php

require 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\IOFactory;

try {
    // 加载 Excel 文件
    $inputFileName = 'example.xlsx';
    $spreadsheet = IOFactory::load($inputFileName);
    $worksheet = $spreadsheet->getActiveSheet();

    // 方法1：使用单元格地址（如 A1, B2, C3 等）
    $cellA1 = $worksheet->getCell('A1')->getValue();
    $cellB2 = $worksheet->getCell('B2')->getValue();
    $cellC3 = $worksheet->getCell('C3')->getValue();

    echo "A1 单元格内容: " . ($cellA1 ?? '空') . "\n";
    echo "B2 单元格内容: " . ($cellB2 ?? '空') . "\n";
    echo "C3 单元格内容: " . ($cellC3 ?? '空') . "\n";

    // 方法2：使用列号和行号（列A=1, B=2, C=3...）
    // 获取第3列第5行（即 C5）的内容
    $cellC5 = $worksheet->getCellByColumnAndRow(3, 5)->getValue();
    echo "C5 单元格内容: " . ($cellC5 ?? '空') . "\n";

    // 方法3：循环读取多个单元格
    echo "\n读取 A1 到 C3 范围的内容:\n";
    for ($row = 1; $row <= 3; $row++) {
        for ($col = 1; $col <= 3; $col++) {
            $value = $worksheet->getCellByColumnAndRow($col, $row)->getValue();
            echo "[$col,$row]: " . ($value ?? '空') . "\t";
        }
        echo "\n";
    }

} catch (Exception $e) {
    die('错误: ' . $e->getMessage());
}
```

## 🔹 获取不同类型的值

### 1. 原始值（getValue）
```php
$rawValue = $worksheet->getCell('A1')->getValue();
// 返回原始存储的值，可能是公式、数字、字符串等
```

### 2. 计算后的值（getCalculatedValue）
```php
// 如果单元格包含公式，这个方法会返回计算后的结果
$calculatedValue = $worksheet->getCell('A1')->getCalculatedValue();
```

### 3. 格式化后的值（getFormattedValue）
```php
// 返回应用了单元格格式后的值（如日期格式、货币格式等）
$formattedValue = $worksheet->getCell('A1')->getFormattedValue();
```

## 🔸 实用函数封装

你可以创建一个辅助函数来简化操作：

```php
function getCellValue($worksheet, $cellAddress) {
    try {
        return $worksheet->getCell($cellAddress)->getValue();
    } catch (Exception $e) {
        return null; // 单元格不存在或出错时返回 null
    }
}

function getCellValueByCoord($worksheet, $column, $row) {
    try {
        return $worksheet->getCellByColumnAndRow($column, $row)->getValue();
    } catch (Exception $e) {
        return null;
    }
}

// 使用示例
$worksheet = $spreadsheet->getActiveSheet();
$value1 = getCellValue($worksheet, 'D5');
$value2 = getCellValueByCoord($worksheet, 4, 5); // D5
```

## 🔹 处理空单元格

```php
// 检查单元格是否为空
$cell = $worksheet->getCell('A1');
if ($cell->getValue() === null || $cell->getValue() === '') {
    echo "单元格为空\n";
} else {
    echo "单元格内容: " . $cell->getValue() . "\n";
}

// 或者使用更简洁的方式
$value = $worksheet->getCell('A1')->getValue() ?? '默认值';
```

## 🔸 批量读取示例

如果你需要读取整列或整行：

```php
// 读取 C 列（第3列）从第2行到第10行
for ($row = 2; $row <= 10; $row++) {
    $value = $worksheet->getCellByColumnAndRow(3, $row)->getValue();
    echo "C{$row}: " . ($value ?? '空') . "\n";
}

// 读取第2行从A到E列
for ($col = 1; $col <= 5; $col++) {
    $value = $worksheet->getCellByColumnAndRow($col, 2)->getValue();
    echo "第2行第{$col}列: " . ($value ?? '空') . "\n";
}
```

这样你就可以灵活地获取 Excel 文件中任意单元格的内容了！
