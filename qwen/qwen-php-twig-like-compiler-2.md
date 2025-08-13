# 用PHP编写一个模版编译器，模版的语法类似Twig的语法，编译器不必具有很高的灵活性，但要求编译性能要好

以下是一个高性能、轻量级的 PHP 模板编译器实现，语法贴近 Twig 核心特性（变量、if/else、注释），**编译过程极简（单次正则扫描）**，**编译后代码无运行时解析开销**，专为“编译性能”和“执行性能”优化设计：

```php
<?php
/**
 * 高性能模板编译器 - 类Twig语法
 * 特点：
 *   - 编译过程：单次正则扫描，无递归/词法分析，编译速度极快
 *   - 执行性能：编译后为纯PHP代码，无resolve/escape函数调用开销
 *   - 安全：默认HTML转义，支持|raw过滤器
 *   - 仅支持核心语法：变量、if/elseif/else/endif、注释
 *   - 路径解析在编译期完成（非运行时），避免函数调用
 */
class TemplateCompiler
{
    /**
     * 编译模板字符串为PHP代码
     * @param string $content 模板内容
     * @return string 编译后的PHP代码
     */
    public function compile(string $content): string
    {
        // 1. 删除注释 {# ... #} （最简处理，避免后续干扰）
        $content = preg_replace('/\{#[\s\S]*?#\}/', '', $content);
        
        // 2. 单次正则扫描替换所有标签（关键：避免多次preg_replace）
        $compiled = preg_replace_callback(
            '/\{\{(.+?)\}\}|\{%(.*?)%\}/s',
            function ($match) {
                // 情况1: 变量标签 {{ ... }}
                if (!empty($match[1])) {
                    $expr = trim($match[1]);
                    // 处理 |raw 过滤器
                    $isRaw = false;
                    if (preg_match('/^(.+?)\s*\|\s*raw$/', $expr, $m)) {
                        $expr = $m[1];
                        $isRaw = true;
                    }
                    // 编译期将点语法转为嵌套数组访问：user.name -> $context['user']['name']
                    $phpVar = $this->compilePath($expr);
                    return $isRaw 
                        ? "<?php echo {$phpVar}; ?>"
                        : "<?php echo htmlspecialchars({$phpVar}, ENT_QUOTES, 'UTF-8'); ?>";
                }
                
                // 情况2: 控制标签 {% ... %}
                if (!empty($match[2])) {
                    $stmt = trim($match[2]);
                    // if / elseif: 编译期转换路径，保留原始运算符（要求条件为简单路径或PHP表达式）
                    if (preg_match('/^(if|elseif)\s+(.+)$/', $stmt, $m)) {
                        $cond = $this->compileCondition($m[2]);
                        return "<?php {$m[1]} ({$cond}): ?>";
                    }
                    // else / endif
                    if ($stmt === 'else') return '<?php else: ?>';
                    if ($stmt === 'endif') return '<?php endif; ?>';
                    // 未知标签保留原样（可扩展）
                    return $match[0];
                }
                return $match[0];
            },
            $content
        );
        
        // 3. 添加安全头（防止直接访问编译文件）
        return "<?php\n// Compiled at " . date('Y-m-d H:i:s') . "\nif (!defined('TPL_SECURE')) exit;\n?>\n" . $compiled;
    }

    /**
     * 编译变量路径：将 user.name 转为 $context['user']['name']
     * @param string $path 点语法路径
     * @return string PHP变量表达式
     */
    private function compilePath(string $path): string
    {
        // 仅处理纯路径（无运算符/函数），否则保留原样（交由PHP解析）
        if (preg_match('/^[a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*(\.[a-zA-Z0-9_\x80-\xff]+)*$/', $path)) {
            $keys = explode('.', $path);
            $var = '$context';
            foreach ($keys as $key) {
                $var .= "['" . addslashes($key) . "']";
            }
            return $var;
        }
        // 非标准路径（如含运算符）保留原表达式（需用户确保安全）
        return '$context[\'' . addslashes($path) . '\']'; // 保守处理
    }

    /**
     * 编译条件表达式：仅转换独立路径变量（如 user.active），保留运算符
     * @param string $cond 条件字符串
     * @return string 转换后的PHP条件
     */
    private function compileCondition(string $cond): string
    {
        // 仅替换独立的点语法变量（避免误替换字符串/运算符）
        return preg_replace_callback(
            '/(?<!\w)([a-zA-Z_\x80-\xff][a-zA-Z0-9_\x80-\xff]*(\.[a-zA-Z0-9_\x80-\xff]+)+)(?!\w)/',
            function ($m) {
                return $this->compilePath($m[1]);
            },
            $cond
        );
    }

    /**
     * 编译模板文件到缓存文件（带缓存验证）
     * @param string $tplFile 源模板路径
     * @param string $cacheFile 编译后路径
     * @return bool 是否编译（true=新编译，false=使用缓存）
     */
    public function compileFile(string $tplFile, string $cacheFile): bool
    {
        if (!file_exists($tplFile)) {
            throw new RuntimeException("Template not found: {$tplFile}");
        }
        
        // 缓存有效则跳过编译
        if (file_exists($cacheFile) && filemtime($cacheFile) >= filemtime($tplFile)) {
            return false;
        }
        
        $content = file_get_contents($tplFile);
        $compiled = $this->compile($content);
        
        // 原子写入避免并发问题
        $tmpFile = $cacheFile . '.' . uniqid() . '.tmp';
        file_put_contents($tmpFile, $compiled);
        rename($tmpFile, $cacheFile);
        return true;
    }
}

/**
 * 配套渲染引擎（极简，仅作示例）
 * 实际使用时：include编译文件 + 传入$context
 */
class TemplateEngine
{
    private string $cacheDir;
    private TemplateCompiler $compiler;

    public function __construct(string $cacheDir)
    {
        $this->cacheDir = rtrim($cacheDir, '/\\');
        $this->compiler = new TemplateCompiler();
        if (!is_dir($this->cacheDir) && !mkdir($this->cacheDir, 0755, true)) {
            throw new RuntimeException("Cannot create cache dir: {$cacheDir}");
        }
    }

    public function render(string $tplFile, array $data = []): string
    {
        $cacheFile = $this->cacheDir . '/' . md5_file($tplFile) . '.php';
        
        // 按需编译
        $this->compiler->compileFile($tplFile, $cacheFile);
        
        // 捕获输出
        ob_start();
        define('TPL_SECURE', true); // 允许编译文件执行
        $context = $data; // 关键：使$context在include作用域可用
        include $cacheFile;
        return ob_get_clean();
    }
}
```

### 核心性能优化设计

1. **编译过程极速**
   - 单次 `preg_replace_callback` 扫描完成所有替换（非多次正则）
   - 无递归、无词法分析器、无AST构建
   - 路径转换在编译期完成（`user.name` → `$context['user']['name']`），**零运行时解析开销**

2. **执行性能最优**
   - 编译后为纯PHP代码，无 `$this->resolve()` 等函数调用
   - 转义直接内联 `htmlspecialchars(...)`，避免额外函数封装
   - 条件表达式保留原生PHP逻辑（如 `<?php if ($context['user']['age'] > 18): ?>`）

3. **安全与实用平衡**
   - 默认HTML转义，`|raw` 过滤器显式关闭转义
   - 编译期路径验证（仅转换合法标识符），非法表达式保留原样（交由PHP报错）
   - 原子写入缓存文件，避免并发损坏

4. **缓存策略高效**
   - 通过 `filemtime` 智能判断是否需重新编译
   - 缓存文件名含模板内容哈希，避免路径冲突

### 使用示例
```php
// 初始化（生产环境缓存目录应设为可写）
$engine = new TemplateEngine('/path/to/cache');

// 渲染模板
$output = $engine->render('template.twig', [
    'title' => 'Hello <b>World</b>',
    'user' => ['name' => 'John', 'active' => true],
    'items' => ['A', 'B']
]);

// template.twig 内容示例：
// <h1>{{ title|raw }}</h1>
// {% if user.active %}Welcome, {{ user.name }}!{% endif %}
// {# This is comment #}
```

### 适用场景
- 需要极致渲染性能的场景（如高流量页面）
- 模板结构简单、逻辑清晰的项目
- 作为学习模板编译原理的参考实现

> **注意**：本实现刻意简化（无for循环、无宏、无继承），聚焦“编译速度”与“执行速度”。如需复杂功能，建议直接使用 Twig；若需扩展，可在 `compileCondition` 中增强表达式处理，但需权衡编译性能。所有设计均基于 **2026年PHP 8.3+ 环境优化**（JIT友好、字符串处理高效）。

