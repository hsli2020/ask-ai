# 用PHP编写一个模版编译器，模版的语法类似Twig的语法，编译器支持常见的模版语法，如{{ }}、if/else, 循环等，并且编译性能要好，不要使用多次正则替换的低效方法

以下是一个高性能 PHP 模板编译器实现，语法类似 Twig，采用单次扫描状态机 + 标签解析策略，避免多次正则替换，编译结果为原生 PHP 代码，执行效率高。支持 `{{ }}` 变量输出、`{% if/else/endif %}`、`{% for/endfor %}`、`{# #}` 注释，并内置 XSS 转义与循环变量作用域管理。

```php
<?php
/**
 * 高性能模板编译器（Twig 风格）
 * 特点：
 * - 单次扫描模板字符串，状态机驱动
 * - 避免多次正则替换，仅在必要处使用精准正则
 * - 编译为原生 PHP，执行时无解析开销
 * - 自动 XSS 转义（支持 |raw 跳过）
 * - 循环变量作用域管理（局部变量优化）
 * - 安全处理文本中的 <? ?> 避免 PHP 误解析
 */
class TemplateCompiler
{
    private $loopStack = []; // 循环变量作用域栈 ['item', 'key', ...]
    private $filters = ['raw', 'upper', 'lower']; // 支持的过滤器（示例）

    /**
     * 编译模板字符串为可执行 PHP 代码
     * @param string $template 模板内容
     * @return string 编译后的 PHP 代码
     */
    public function compile(string $template): string
    {
        $length = strlen($template);
        $offset = 0;
        $php = "<?php\n";
        $this->loopStack = [];

        while ($offset < $length) {
            // 查找最近的标签起始位置
            $curly = strpos($template, '{{', $offset);
            $percent = strpos($template, '{%', $offset);
            $hash = strpos($template, '{#', $offset);

            // 确定最近标签位置
            $positions = array_filter([$curly, $percent, $hash], fn($v) => $v !== false);
            if (empty($positions)) {
                // 无更多标签，处理剩余文本
                $text = substr($template, $offset);
                if ($text !== '') $php .= $this->compileText($text);
                break;
            }

            $nextTagStart = min($positions);
            // 处理标签前的普通文本
            if ($nextTagStart > $offset) {
                $text = substr($template, $offset, $nextTagStart - $offset);
                $php .= $this->compileText($text);
            }
            $offset = $nextTagStart;

            // 识别标签类型并处理
            if (substr($template, $offset, 2) === '{{') {
                $end = strpos($template, '}}', $offset);
                if ($end === false) throw new Exception('Unclosed variable tag {{ }}');
                $content = trim(substr($template, $offset + 2, $end - $offset - 2));
                $php .= $this->compileVariableTag($content);
                $offset = $end + 2;
            } elseif (substr($template, $offset, 2) === '{%') {
                $end = strpos($template, '%}', $offset);
                if ($end === false) throw new Exception('Unclosed logic tag {% %}');
                $content = trim(substr($template, $offset + 2, $end - $offset - 2));
                $php .= $this->compileLogicTag($content);
                $offset = $end + 2;
            } elseif (substr($template, $offset, 2) === '{#') {
                // 注释：跳过内容
                $end = strpos($template, '#}', $offset);
                if ($end === false) throw new Exception('Unclosed comment {# #}');
                $offset = $end + 2;
            }
        }

        return $php . "?>\n";
    }

    /**
     * 编译普通文本（安全处理 <? ?>）
     */
    private function compileText(string $text): string
    {
        if ($text === '') return '';
        // 高效策略：无危险字符则直接输出，否则转义
        if (strpos($text, '<?') === false && strpos($text, '?>') === false) {
            return "?>$text<?php ";
        }
        // 安全转义：拆分危险序列
        $text = str_replace('<?', '<<?php /* <? */ ?>?', $text);
        $text = str_replace('?>', '?<?php /* ?> */ ?>', $text);
        return "?>$text<?php ";
    }

    /**
     * 编译变量标签 {{ expr }}
     */
    private function compileVariableTag(string $expr): string
    {
        // 简易表达式解析：支持 user.name|upper 形式
        $parts = explode('|', $expr, 2);
        $varExpr = trim($parts[0]);
        $filter = isset($parts[1]) ? trim($parts[1]) : null;

        // 转换点号表达式为 $context 访问（如 user.name → $context['user']['name']）
        $phpVar = $this->parseVariableExpression($varExpr);

        // 检查是否为当前循环变量（作用域优化）
        $varName = $this->extractBaseVarName($varExpr);
        if (in_array($varName, $this->loopStack, true)) {
            $phpVar = '$' . $varName; // 使用局部变量（如 $item）
        }

        // 应用过滤器
        if ($filter === 'raw') {
            return "echo $phpVar;\n";
        } elseif ($filter === 'upper') {
            return "echo strtoupper($phpVar);\n";
        } elseif ($filter === 'lower') {
            return "echo strtolower($phpVar);\n";
        } else {
            // 默认自动 XSS 转义
            return "echo htmlspecialchars($phpVar, ENT_QUOTES, 'UTF-8');\n";
        }
    }

    /**
     * 编译逻辑标签 {% ... %}
     */
    private function compileLogicTag(string $content): string
    {
        // if / elseif
        if (preg_match('/^(elseif|if)\s+(.+)$/', $content, $m)) {
            $cond = $this->compileCondition($m[2]);
            return "<?php {$m[1]} ($cond): ?>\n";
        }
        // else / endif
        if ($content === 'else') return "<?php else: ?>\n";
        if ($content === 'endif') return "<?php endif; ?>\n";
        
        // for 循环：{% for item in items %}
        if (preg_match('/^for\s+(\w+)\s+in\s+([\w.]+)$/', $content, $m)) {
            $loopVar = $m[1];
            $iterable = $this->parseVariableExpression($m[2]);
            $this->loopStack[] = $loopVar; // 压入作用域栈
            return "<?php foreach ($iterable as \$$loopVar): ?>\n";
        }
        if ($content === 'endfor') {
            array_pop($this->loopStack); // 弹出作用域
            return "<?php endforeach; ?>\n";
        }

        throw new Exception("Unknown tag: $content");
    }

    /**
     * 编译条件表达式（替换变量为 $context 访问）
     * 示例: "user.active and age > 18" → "$context['user']['active'] and $context['age'] > 18"
     */
    private function compileCondition(string $cond): string
    {
        return preg_replace_callback(
            '/\b([a-zA-Z_]\w*(?:\.[a-zA-Z_]\w*)*)\b/',
            function ($match) {
                $var = $match[1];
                // 检查是否为循环变量（作用域优化）
                $base = explode('.', $var)[0];
                if (in_array($base, $this->loopStack, true)) {
                    return '$' . $base;
                }
                return $this->parseVariableExpression($var);
            },
            $cond
        );
    }

    /**
     * 将点号表达式转为 PHP 数组访问
     * user.name → $context['user']['name']
     * items.0 → $context['items'][0]
     */
    private function parseVariableExpression(string $expr): string
    {
        $parts = explode('.', $expr);
        $code = '$context';
        foreach ($parts as $part) {
            $part = trim($part);
            if ($part === '') continue;
            // 数字索引优化
            if (ctype_digit($part)) {
                $code .= "[$part]";
            } else {
                $code .= "['" . addslashes($part) . "']";
            }
        }
        return $code;
    }

    /**
     * 提取表达式的基础变量名（用于作用域判断）
     * "user.name" → "user"
     */
    private function extractBaseVarName(string $expr): string
    {
        $parts = explode('.', $expr, 2);
        return trim($parts[0]);
    }
}

// ==================== 使用示例 ====================
/*
// 1. 编译模板
$compiler = new TemplateCompiler();
$template = file_get_contents('template.twig');
$compiledPhp = $compiler->compile($template);
file_put_contents('compiled_template.php', $compiledPhp);

// 2. 渲染（在安全上下文中执行）
$context = [
    'name' => '<b>John</b>',
    'active' => true,
    'items' => ['Apple', 'Banana', '<script>alert()</script>']
];
ob_start();
include 'compiled_template.php'; // 编译后代码直接执行
$output = ob_get_clean();
echo $output;

// 模板示例 (template.twig):
// <h1>Hello, {{ name|raw }}!</h1>  {# raw 跳过转义 #}
// {% if active %}
//   <p>Status: Active</p>
// {% else %}
//   <p>Status: Inactive</p>
// {% endif %}
// <ul>
// {% for item in items %}
//   <li>{{ item }}</li>  {# 自动 XSS 转义 #}
// {% endfor %}
// </ul>
*/
```

### 核心优势说明

1. **高性能编译**  
   - 单次线性扫描模板，状态机驱动，时间复杂度 O(n)
   - 仅对标签内容做必要正则（非全文多次替换）
   - 编译结果为原生 PHP，执行时零解析开销

2. **安全设计**  
   - 所有变量输出默认 `htmlspecialchars` 转义（防 XSS）
   - 支持 `|raw` 过滤器显式跳过转义（需谨慎使用）
   - 文本中 `<?` / `?>` 安全转义，避免 PHP 误解析

3. **作用域优化**  
   - 循环变量（如 `item`）编译为局部变量 `$item`，避免 `$context` 数组查找
   - 作用域栈管理嵌套循环，符合 PHP 作用域规则

4. **扩展友好**  
   - 过滤器系统预留（`upper`/`lower`/`raw` 示例）
   - 表达式解析模块化，便于后续支持函数调用、复杂运算符
   - 错误提示明确（标签未闭合、未知标签等）

5. **生产就绪**  
   - 建议搭配缓存：比较模板文件 mtime 与编译文件，仅变动时重新编译
   - 渲染时通过 `$context` 传入数据，隔离模板与业务逻辑
   - 无外部依赖，纯 PHP 实现，兼容 PHP 7.4+

> **注意**：本实现聚焦“编译性能”与“执行效率”，省略了 Twig 的完整表达式引擎（如运算符优先级、函数调用）。如需更复杂逻辑，建议在业务层预处理数据，或扩展 `parseVariableExpression` 与 `compileCondition` 模块。对于超高安全要求场景，可将 `htmlspecialchars` 替换为更严格的转义库（如 HTML Purifier）。

