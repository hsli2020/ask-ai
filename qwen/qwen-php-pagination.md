# 🤵 请用PHP语言编写一个Pagination类

Qwen3-235B-A22B-2507 3:31 pm
下面是一个使用 PHP 编写的 Pagination 类，用于生成分页链接和相关信息，适用于 Web 分页显示场景。 

<?php

class Pagination
{
    // 当前页码
    private $currentPage;

    // 每页显示的记录数
    private $perPage;

    // 总记录数
    private $totalRecords;

    // 总页数
    private $totalPages;

    // URL 模板，用于生成分页链接（如：/page/%d）
    private $urlTemplate;

    // 显示的页码链接数量（左右各显示几个）
    private $range;

    // 构造函数
    public function __construct($currentPage, $perPage, $totalRecords, $urlTemplate = '/page/%d', $range = 5)
    {
        $this->currentPage = (int)$currentPage;
        $this->perPage = (int)$perPage;
        $this->totalRecords = (int)$totalRecords;
        $this->urlTemplate = $urlTemplate;
        $this->range = (int)$range;

        // 计算总页数
        $this->totalPages = max(1, (int)ceil($this->totalRecords / $this->perPage));

        // 确保当前页在有效范围内
        $this->currentPage = max(1, min($this->currentPage, $this->totalPages));
    }

    // 获取当前页码
    public function getCurrentPage()
    {
        return $this->currentPage;
    }

    // 获取每页记录数
    public function getPerPage()
    {
        return $this->perPage;
    }

    // 获取总记录数
    public function getTotalRecords()
    {
        return $this->totalRecords;
    }

    // 获取总页数
    public function getTotalPages()
    {
        return $this->totalPages;
    }

    // 获取起始记录索引（用于数据库 LIMIT）
    public function getOffset()
    {
        return ($this->currentPage - 1) * $this->perPage;
    }

    // 是否有上一页
    public function hasPrevious()
    {
        return $this->currentPage > 1;
    }

    // 是否有下一页
    public function hasNext()
    {
        return $this->currentPage < $this->totalPages;
    }

    // 获取上一页页码
    public function getPreviousPage()
    {
        return $this->hasPrevious() ? $this->currentPage - 1 : null;
    }

    // 获取下一页页码
    public function getNextPage()
    {
        return $this->hasNext() ? $this->currentPage + 1 : null;
    }

    // 生成分页链接 HTML
    public function render($wrapperTag = 'div', $class = 'pagination')
    {
        if ($this->totalPages <= 1) {
            return ''; // 只有一页时不需要分页
        }

        $html = "<{$wrapperTag} class=\"{$class}\">\n";
        $html .= $this->renderPreviousLink();
        $html .= $this->renderPageLinks();
        $html .= $this->renderNextLink();
        $html .= "</{$wrapperTag}>\n";

        return $html;
    }

    // 生成“上一页”链接
    private function renderPreviousLink()
    {
        if ($this->hasPrevious()) {
            $url = sprintf($this->urlTemplate, $this->getPreviousPage());
            return "<a href=\"{$url}\" class=\"prev\">&laquo; 上一页</a> ";
        }
        return '<span class="prev disabled">&laquo; 上一页</span> ';
    }

    // 生成“下一页”链接
    private function renderNextLink()
    {
        if ($this->hasNext()) {
            $url = sprintf($this->urlTemplate, $this->getNextPage());
            return " <a href=\"{$url}\" class=\"next\">下一页 &raquo;</a>";
        }
        return ' <span class="next disabled">下一页 &raquo;</span>';
    }

    // 生成中间页码链接
    private function renderPageLinks()
    {
        $start = max(1, $this->currentPage - $this->range);
        $end = min($this->totalPages, $this->currentPage + $this->range);

        $links = '';

        // 如果前面有省略号
        if ($start > 1) {
            $url = sprintf($this->urlTemplate, 1);
            $links .= "<a href=\"{$url}\">1</a> ";
            if ($start > 2) {
                $links .= '<span class="ellipsis">...</span> ';
            }
        }

        // 输出中间页码
        for ($i = $start; $i <= $end; $i++) {
            if ($i === $this->currentPage) {
                $links .= "<strong>{$i}</strong> ";
            } else {
                $url = sprintf($this->urlTemplate, $i);
                $links .= "<a href=\"{$url}\">{$i}</a> ";
            }
        }

        // 如果后面有省略号
        if ($end < $this->totalPages) {
            if ($end < $this->totalPages - 1) {
                $links .= '<span class="ellipsis">...</span> ';
            }
            $url = sprintf($this->urlTemplate, $this->totalPages);
            $links .= "<a href=\"{$url}\">{$this->totalPages}</a> ";
        }

        return $links;
    }

    // 获取分页信息数组（可用于 JSON 或模板引擎）
    public function toArray()
    {
        return [
            'current_page' => $this->currentPage,
            'per_page' => $this->perPage,
            'total_records' => $this->totalRecords,
            'total_pages' => $this->totalPages,
            'offset' => $this->getOffset(),
            'has_previous' => $this->hasPrevious(),
            'has_next' => $this->hasNext(),
            'previous_page' => $this->getPreviousPage(),
            'next_page' => $this->getNextPage(),
        ];
    }
}

// 使用示例
/*
$currentPage = $_GET['page'] ?? 1;
$perPage = 10;
$totalRecords = 123;
$urlTemplate = '/articles/page/%d';

$pagination = new Pagination($currentPage, $perPage, $totalRecords, $urlTemplate, 3);

echo $pagination->render();

// 或获取数据
$data = $pagination->toArray();
print_r($data);
*/

📌 功能说明： 

    ✅ 支持自定义 URL 模板（使用 %d 占位符）
    ✅ 自动计算总页数、偏移量
    ✅ 支持上一页/下一页按钮
    ✅ 显示当前页附近的页码（可配置范围）
    ✅ 支持省略号（...）优化大量页码显示
    ✅ 提供 render() 方法输出 HTML 分页
    ✅ 提供 toArray() 方法用于 API 或模板引擎集成

🛠️ 使用方法：

$pagination = new Pagination(
    $page,           // 当前页
    10,              // 每页10条
    250,             // 总共250条数据
    '/list/%d',      // 分页链接模板
    4                // 显示当前页前后各4个页码
);

echo $pagination->render('nav', 'my-pagination');

这个类可以轻松集成到任何 PHP 项目中，无论是原生项目还是自定义 MVC 框架。
