# 用php编写一个程序，功能如下： 生成一个execel文件，加入几个sheets, 向每个sheet中加入若干行和列，设置单元格的内容，设置单元格的背景色，设置单元格的宽度

我来为您编写一个PHP程序来生成Excel文件。我们将使用 **PhpSpreadsheet** 库，这是目前最流行的PHP Excel处理库。

## 1. 安装 PhpSpreadsheet

首先需要安装 PhpSpreadsheet 库：

```bash
composer require phpoffice/phpspreadsheet
```

## 2. PHP 代码实现

```php
<?php

require_once 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;
use PhpOffice\PhpSpreadsheet\Style\Color;
use PhpOffice\PhpSpreadsheet\Style\Fill;

// 创建新的 Spreadsheet 对象
$spreadsheet = new Spreadsheet();

// 删除默认的工作表
$spreadsheet->removeSheetByIndex(0);

// 创建第一个工作表
$sheet1 = $spreadsheet->createSheet();
$sheet1->setTitle('员工信息');

// 向第一个工作表添加数据
$sheet1->setCellValue('A1', '员工ID');
$sheet1->setCellValue('B1', '姓名');
$sheet1->setCellValue('C1', '部门');
$sheet1->setCellValue('D1', '薪资');

$sheet1->setCellValue('A2', '001');
$sheet1->setCellValue('B2', '张三');
$sheet1->setCellValue('C2', '技术部');
$sheet1->setCellValue('D2', '8000');

$sheet1->setCellValue('A3', '002');
$sheet1->setCellValue('B3', '李四');
$sheet1->setCellValue('C3', '销售部');
$sheet1->setCellValue('D3', '7500');

$sheet1->setCellValue('A4', '003');
$sheet1->setCellValue('B4', '王五');
$sheet1->setCellValue('C4', '人事部');
$sheet1->setCellValue('D4', '6500');

// 设置标题行背景色
$sheet1->getStyle('A1:D1')->getFill()
    ->setFillType(Fill::FILL_SOLID)
    ->getStartColor()->setARGB('FF4CAF50'); // 绿色

// 设置单元格宽度
$sheet1->getColumnDimension('A')->setWidth(15);
$sheet1->getColumnDimension('B')->setWidth(20);
$sheet1->getColumnDimension('C')->setWidth(25);
$sheet1->getColumnDimension('D')->setWidth(18);

// 创建第二个工作表
$sheet2 = $spreadsheet->createSheet();
$sheet2->setTitle('产品信息');

// 向第二个工作表添加数据
$sheet2->setCellValue('A1', '产品ID');
$sheet2->setCellValue('B1', '产品名称');
$sheet2->setCellValue('C1', '价格');
$sheet2->setCellValue('D1', '库存');

$sheet2->setCellValue('A2', 'P001');
$sheet2->setCellValue('B2', '笔记本电脑');
$sheet2->setCellValue('C2', '5999');
$sheet2->setCellValue('D2', '50');

$sheet2->setCellValue('A3', 'P002');
$sheet2->setCellValue('B3', '智能手机');
$sheet2->setCellValue('C3', '3999');
$sheet2->setCellValue('D3', '100');

$sheet2->setCellValue('A4', 'P003', '平板电脑');
$sheet2->setCellValue('B4', '平板电脑');
$sheet2->setCellValue('C4', '2999');
$sheet2->setCellValue('D4', '75');

// 设置标题行背景色（蓝色）
$sheet2->getStyle('A1:D1')->getFill()
    ->setFillType(Fill::FILL_SOLID)
    ->getStartColor()->setARGB('FF2196F3'); // 蓝色

// 设置单元格宽度
$sheet2->getColumnDimension('A')->setWidth(12);
$sheet2->getColumnDimension('B')->setWidth(25);
$sheet2->getColumnDimension('C')->setWidth(15);
$sheet2->getColumnDimension('D')->setWidth(15);

// 创建第三个工作表
$sheet3 = $spreadsheet->createSheet();
$sheet3->setTitle('销售统计');

// 向第三个工作表添加数据
$sheet3->setCellValue('A1', '月份');
$sheet3->setCellValue('B1', '销售额');
$sheet3->setCellValue('C1', '利润');
$sheet3->setCellValue('D1', '增长率');

$sheet3->setCellValue('A2', '1月');
$sheet3->setCellValue('B2', '150000');
$sheet3->setCellValue('C2', '45000');
$sheet3->setCellValue('D2', '12%');

$sheet3->setCellValue('A3', '2月');
$sheet3->setCellValue('B3', '180000');
$sheet3->setCellValue('C3', '54000');
$sheet3->setCellValue('D3', '20%');

$sheet3->setCellValue('A4', '3月');
$sheet3->setCellValue('B4', '200000');
$sheet3->setCellValue('C4', '60000');
$sheet3->setCellValue('D4', '11%');

// 设置标题行背景色（橙色）
$sheet3->getStyle('A1:D1')->getFill()
    ->setFillType(Fill::FILL_SOLID)
    ->getStartColor()->setARGB('FFFF9800'); // 橙色

// 设置单元格宽度
$sheet3->getColumnDimension('A')->setWidth(10);
$sheet3->getColumnDimension('B')->setWidth(20);
$sheet3->getColumnDimension('C')->setWidth(20);
$sheet3->getColumnDimension('D')->setWidth(15);

// 设置第一个工作表为活动工作表
$spreadsheet->setActiveSheetIndex(0);

// 保存为 Excel 文件
$writer = new Xlsx($spreadsheet);
$filename = '示例文件_' . date('Y-m-d_H-i-s') . '.xlsx';
$writer->save($filename);

echo "Excel 文件已成功生成：{$filename}\n";

// 如果要在浏览器中直接下载，可以使用以下代码（取消注释）
/*
header('Content-Type: application/vnd.openxmlformats-officedocument.spreadsheetml.sheet');
header('Content-Disposition: attachment;filename="' . $filename . '"');
header('Cache-Control: max-age=0');

$writer->save('php://output');
exit;
*/
?>
```

## 3. 高级功能示例（可选）

如果您需要更多高级功能，这里是一个包含更多样式的示例：

```php
<?php
// ... (前面的代码相同)

// 高级样式设置示例
function applyAdvancedStyles($sheet) {
    // 设置标题行样式
    $sheet->getStyle('A1:D1')->applyFromArray([
        'font' => [
            'bold' => true,
            'color' => ['rgb' => 'FFFFFF'],
            'size' => 12
        ],
        'alignment' => [
            'horizontal' => \PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_CENTER,
            'vertical' => \PhpOffice\PhpSpreadsheet\Style\Alignment::VERTICAL_CENTER
        ],
        'borders' => [
            'allBorders' => [
                'borderStyle' => \PhpOffice\PhpSpreadsheet\Style\Border::BORDER_THIN,
                'color' => ['rgb' => '000000']
            ]
        ]
    ]);
    
    // 设置数据行样式
    $highestRow = $sheet->getHighestRow();
    $sheet->getStyle("A2:D{$highestRow}")->applyFromArray([
        'alignment' => [
            'horizontal' => \PhpOffice\PhpSpreadsheet\Style\Alignment::HORIZONTAL_CENTER
        ],
        'borders' => [
            'allBorders' => [
                'borderStyle' => \PhpOffice\PhpSpreadsheet\Style\Border::BORDER_THIN,
                'color' => ['rgb' => 'CCCCCC']
            ]
        ]
    ]);
}

// 应用高级样式
applyAdvancedStyles($sheet1);
applyAdvancedStyles($sheet2);
applyAdvancedStyles($sheet3);

// ... (保存文件的代码)
?>
```

## 4. 常用颜色代码参考

- 红色: `FFFF0000`
- 绿色: `FF00FF00` 或 `FF4CAF50`
- 蓝色: `FF0000FF` 或 `FF2196F3`
- 黄色: `FFFFFF00`
- 橙色: `FFFF9800`
- 紫色: `FF9C27B0`
- 灰色: `FF9E9E9E`

## 5. 运行说明

1. 确保已安装 Composer
2. 在项目目录中运行 `composer require phpoffice/phpspreadsheet`
3. 将上述代码保存为 `generate_excel.php`
4. 在命令行中运行 `php generate_excel.php`
5. 程序会生成一个包含三个工作表的Excel文件

这个程序实现了您要求的所有功能：
- ✅ 生成Excel文件
- ✅ 创建多个sheets
- ✅ 向每个sheet添加多行多列数据
- ✅ 设置单元格内容
- ✅ 设置单元格背景色
- ✅ 设置单元格列宽

您可以根据实际需求修改数据内容、样式和文件名。

# 如何设置某个单元格的内容为公式？

在 PhpSpreadsheet 中设置单元格公式非常简单，只需要在 `setCellValue()` 方法中传入以等号 `=` 开头的公式字符串即可。

## 1. 基本公式设置

```php
<?php

require_once 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

$spreadsheet = new Spreadsheet();
$sheet = $spreadsheet->getActiveSheet();
$sheet->setTitle('公式示例');

// 设置基础数据
$sheet->setCellValue('A1', '数值1');
$sheet->setCellValue('B1', '数值2');
$sheet->setCellValue('C1', '求和');
$sheet->setCellValue('D1', '平均值');

$sheet->setCellValue('A2', 10);
$sheet->setCellValue('B2', 20);
$sheet->setCellValue('A3', 30);
$sheet->setCellValue('B3', 40);
$sheet->setCellValue('A4', 50);
$sheet->setCellValue('B4', 60);

// 设置公式
$sheet->setCellValue('C2', '=A2+B2');           // 求和
$sheet->setCellValue('D2', '=AVERAGE(A2:B2)');  // 平均值

$sheet->setCellValue('C3', '=SUM(A3:B3)');      // 使用SUM函数求和
$sheet->setCellValue('D3', '=AVERAGE(A3:B3)');  // 平均值

// 跨行求和
$sheet->setCellValue('A5', '总计');
$sheet->setCellValue('C5', '=SUM(C2:C4)');      // C2到C4的总和
$sheet->setCellValue('D5', '=AVERAGE(D2:D4)');  // D2到D4的平均值

// 保存文件
$writer = new Xlsx($spreadsheet);
$writer->save('公式示例.xlsx');

echo "包含公式的Excel文件已生成！\n";
?>
```

## 2. 常用公式示例

```php
<?php

require_once 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

$spreadsheet = new Spreadsheet();
$sheet = $spreadsheet->getActiveSheet();
$sheet->setTitle('常用公式');

// 设置数据
$sheet->setCellValue('A1', '产品');
$sheet->setCellValue('B1', '单价');
$sheet->setCellValue('C1', '数量');
$sheet->setCellValue('D1', '总价');
$sheet->setCellValue('E1', '折扣');
$sheet->setCellValue('F1', '最终价格');

// 产品数据
$products = [
    ['笔记本电脑', 5999, 2],
    ['智能手机', 3999, 3],
    ['平板电脑', 2999, 1],
    ['耳机', 299, 5]
];

$row = 2;
foreach ($products as $product) {
    $sheet->setCellValue('A' . $row, $product[0]);
    $sheet->setCellValue('B' . $row, $product[1]);
    $sheet->setCellValue('C' . $row, $product[2]);
    
    // 总价 = 单价 * 数量
    $sheet->setCellValue('D' . $row, '=B' . $row . '*C' . $row);
    
    // 折扣 (假设10%)
    $sheet->setCellValue('E' . $row, 0.1);
    
    // 最终价格 = 总价 * (1 - 折扣)
    $sheet->setCellValue('F' . $row, '=D' . $row . '*(1-E' . $row . ')');
    
    $row++;
}

// 统计行
$sheet->setCellValue('A' . $row, '总计');
// 总数量
$sheet->setCellValue('C' . $row, '=SUM(C2:C' . ($row-1) . ')');
// 总金额（不打折）
$sheet->setCellValue('D' . $row, '=SUM(D2:D' . ($row-1) . ')');
// 平均折扣
$sheet->setCellValue('E' . $row, '=AVERAGE(E2:E' . ($row-1) . ')');
// 总最终价格
$sheet->setCellValue('F' . $row, '=SUM(F2:F' . ($row-1) . ')');

// 设置列宽
$sheet->getColumnDimension('A')->setWidth(15);
$sheet->getColumnDimension('B')->setWidth(10);
$sheet->getColumnDimension('C')->setWidth(10);
$sheet->getColumnDimension('D')->setWidth(12);
$sheet->getColumnDimension('E')->setWidth(10);
$sheet->getColumnDimension('F')->setWidth(12);

// 保存文件
$writer = new Xlsx($spreadsheet);
$writer->save('销售计算公式.xlsx');

echo "销售计算公式Excel文件已生成！\n";
?>
```

## 3. 条件公式和逻辑公式

```php
<?php

require_once 'vendor/autoload.php';

use PhpOffice\PhpSpreadsheet\Spreadsheet;
use PhpOffice\PhpSpreadsheet\Writer\Xlsx;

$spreadsheet = new Spreadsheet();
$sheet = $spreadsheet->getActiveSheet();
$sheet->setTitle('条件公式');

// 设置数据
$sheet->setCellValue('A1', '姓名');
$sheet->setCellValue('B1', '成绩');
$sheet->setCellValue('C1', '等级');
$sheet->setCellValue('D1', '是否及格');

$data = [
    ['张三', 85],
    ['李四', 72],
    ['王五', 93],
    ['赵六', 58],
    ['钱七', 67]
];

$row = 2;
foreach ($data as $student) {
    $sheet->setCellValue('A' . $row, $student[0]);
    $sheet->setCellValue('B' . $row, $student[1]);
    
    // 使用IF函数判断等级
    $sheet->setCellValue('C' . $row, '=IF(B' . $row . '>=90,"优秀",IF(B' . $row . '>=80,"良好",IF(B' . $row . '>=70,"中等",IF(B' . $row . '>=60,"及格","不及格"))))');
    
    // 判断是否及格
    $sheet->setCellValue('D' . $row, '=IF(B' . $row . '>=60,"是","否")');
    
    $row++;
}

// 统计信息
$sheet->setCellValue('A' . $row, '统计');
$sheet->setCellValue('B' . $row, '平均分');
$sheet->setCellValue('C' . $row, '最高分');
$sheet->setCellValue('D' . $row, '最低分');

$lastDataRow = $row - 1;
$sheet->setCellValue('B' . $row, '=AVERAGE(B2:B' . $lastDataRow . ')');
$sheet->setCellValue('C' . $row, '=MAX(B2:B' . $lastDataRow . ')');
$sheet->setCellValue('D' . $row, '=MIN(B2:B' . $lastDataRow . ')');

// 保存文件
$writer = new Xlsx($spreadsheet);
$writer->save('条件公式示例.xlsx');

echo "条件公式Excel文件已生成！\n";
?>
```

## 4. 获取公式计算结果

有时候你可能需要获取公式的计算结果而不是公式本身：

```php
<?php

// ... (前面的代码)

// 设置公式
$sheet->setCellValue('C2', '=A2+B2');

// 获取公式本身
$formula = $sheet->getCell('C2')->getValue(); // 返回 "=A2+B2"

// 获取计算后的结果
$calculatedValue = $sheet->getCell('C2')->getCalculatedValue(); // 返回实际计算结果

echo "公式: " . $formula . "\n";
echo "计算结果: " . $calculatedValue . "\n";

// 如果你希望单元格直接显示计算结果而不是公式
// 可以先计算再设置值
$result = $sheet->getCell('C2')->getCalculatedValue();
$sheet->setCellValue('D2', $result); // D2 将显示计算结果，不是公式
?>
```

## 5. 重要注意事项

1. **公式语法**：使用 Excel 的公式语法，不是 PHP 函数
2. **区域设置**：公式中的分隔符（逗号或分号）取决于 Excel 的区域设置
3. **引用格式**：支持 A1 引用样式（如 A1, B2）和 R1C1 引用样式
4. **函数支持**：PhpSpreadsheet 支持大部分 Excel 内置函数
5. **错误处理**：如果公式有错误，Excel 会显示错误值（如 #VALUE!, #REF! 等）

## 6. 常用 Excel 函数参考

- `SUM(range)` - 求和
- `AVERAGE(range)` - 平均值
- `MAX(range)` - 最大值
- `MIN(range)` - 最小值
- `COUNT(range)` - 计数
- `IF(condition, value_if_true, value_if_false)` - 条件判断
- `VLOOKUP(lookup_value, table_array, col_index, range_lookup)` - 垂直查找
- `CONCATENATE(text1, text2, ...)` - 连接文本
- `ROUND(number, num_digits)` - 四舍五入

这样就可以在 PhpSpreadsheet 中轻松设置各种公式了！
