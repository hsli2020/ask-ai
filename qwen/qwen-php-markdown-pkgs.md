# 在PHP中，进行markdown处理有哪些比较好的库？

在 PHP 中处理 Markdown 有多个成熟、功能丰富且广泛使用的库。以下是一些比较好的选择，按推荐程度和特性分类：

---

### 1. **League\CommonMark**（最推荐）
- **GitHub**: https://github.com/thephpleague/commonmark
- **特点**：
  - 完全兼容 [CommonMark](https://commonmark.org/) 规范（以及部分 GFM 扩展）。
  - 高度可扩展：支持自定义解析器、渲染器、扩展（如表格、任务列表、脚注等）。
  - 安全性好，默认会自动转义 HTML（可配置）。
  - 性能优秀，社区活跃，文档完善。
- **安装**（通过 Composer）：
  ```bash
  composer require league/commonmark
  ```
- **基本用法**：
  ```php
  use League\CommonMark\MarkdownConverter;

  $converter = new MarkdownConverter();
  echo $converter->convert('# Hello World!'); // 输出 HTML
  ```

> ✅ **适合大多数项目，尤其是需要安全、标准兼容性和可扩展性的场景。**

---

### 2. **Michelf\Markdown / MarkdownExtra**
- **GitHub**: https://github.com/michelf/php-markdown
- **作者**：Michel Fortin（Markdown 的早期 PHP 实现者之一）
- **特点**：
  - 支持标准 Markdown 和 Markdown Extra（支持表格、脚注、定义列表等）。
  - 轻量级，无依赖。
  - 不完全兼容 CommonMark（语法细节略有差异）。
- **安装**：
  ```bash
  composer require michelf/php-markdown
  ```
- **用法**：
  ```php
  use Michelf\Markdown;
  echo Markdown::defaultTransform('# Hello');
  ```

> ⚠️ **适合遗留项目或对 CommonMark 兼容性要求不高的场景。新项目建议优先考虑 CommonMark。**

---

### 3. **Parsedown**
- **GitHub**: https://github.com/erusev/parsedown
- **特点**：
  - 极简、快速，单文件实现。
  - 语法接近 GitHub Flavored Markdown（GFM），但非完全兼容。
  - 不支持扩展（如表格需手动启用）。
  - 已多年未积极维护（截至 2024 年）。
- **安装**：
  ```bash
  composer require erusev/parsedown
  ```
- **用法**：
  ```php
  $parser = new Parsedown();
  echo $parser->text('# Hello');
  ```

> ⚠️ **性能好但功能有限，适合简单场景；但因缺乏维护，不建议用于新项目。**

---

### 4. **CommonMark + GFM 扩展（基于 League\CommonMark）**
如果你需要 GitHub 风格的 Markdown（如任务列表、删除线、表格等），可以使用官方扩展包：

```bash
composer require league/commonmark league/commonmark-gfm-extension
```

```php
use League\CommonMark\Environment\Environment;
use League\CommonMark\Extension\GithubFlavoredMarkdownExtension;
use League\CommonMark\MarkdownConverter;

$environment = Environment::createCommonMarkEnvironment();
$environment->addExtension(new GithubFlavoredMarkdownExtension());

$converter = new MarkdownConverter($environment);
echo $converter->convert('~~strikethrough~~ and - [x] task list');
```

---

### 安全提示
无论使用哪个库，请注意：
- 默认应**禁用原始 HTML**（防止 XSS），除非你信任输入源。
- `League\CommonMark` 默认会转义 HTML，而 `Parsedown` 和 `Michelf` 默认允许 HTML，需手动配置。

---

### 总结推荐

| 场景 | 推荐库 |
|------|--------|
| 新项目、标准兼容、可扩展 | ✅ **League\CommonMark** |
| 需要 GitHub 风格 Markdown | ✅ **League\CommonMark + GFM Extension** |
| 轻量级、简单需求（旧项目） | ⚠️ Parsedown 或 Michelf\Markdown |
| 遗留系统迁移 | Michelf\Markdown（兼容旧语法） |

> 📌 **当前（2026 年）最主流、最安全、最现代的选择是 `league/commonmark`。**

如有特定需求（如自定义语法、性能极致优化等），可进一步探讨扩展方案。

