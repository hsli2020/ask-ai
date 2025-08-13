# 用PHP开发web应用，选择框架时面临两个困扰，有的框架频繁升级，有的框架放弃维护，这都给我们带来风险。如果不采用框架，有哪些最佳实践，提供一些可参考的例子

这是一个非常深刻且实际的问题。在 PHP 生态中，框架的“过度迭代”（如 Laravel 大版本升级带来的破坏性变更）和“停止维护”（如某些老旧框架）确实是长期维护的痛点。

**“不使用框架”并不等于“不使用任何库”或“写原生过程式代码”。** 最佳实践应该是 **“基于组件的组合式架构” (Component-Based Architecture)**。即：自己掌控核心架构，但使用稳定、独立的库来处理具体功能（如数据库、路由、日志）。

以下是不采用重型框架时的 **7 大最佳实践**，以及一个 **可参考的最小化架构示例**。

---

### 一、核心原则：组合优于继承，稳定优于新颖

1.  **拒绝“全家桶”**：不要引入一个解决所有问题的框架，而是引入解决特定问题的库。
2.  **遵循 PSR 标准**：这是 PHP 社区互操作性的基石。遵循 PSR-4 (自动加载)、PSR-7 (HTTP 消息)、PSR-11 (容器) 等，可以让你随时替换底层库而不影响业务代码。
3.  **锁定依赖版本**：这是避免“频繁升级风险”的关键。

---

### 二、7 大最佳实践指南

#### 1. 项目结构设计 (标准化)
即使没有框架，也要有清晰的目录结构。参考 PSR-1 和现代 PHP 项目结构：

```text
project/
├── public/             # 唯一对外目录，存放 index.php, assets
├── src/                # 你的业务代码 (Controllers, Models, Services)
├── config/             # 配置文件 (不硬编码在代码里)
├── templates/          # 视图文件
├── tests/              # 测试代码
├── vendor/             # 第三方库 (Composer 管理)
├── composer.json       # 依赖声明
└── composer.lock       # 依赖锁定 (至关重要！)
```

#### 2. 依赖管理 (Composer 是必须的)
不要手动下载 `.zip` 包放入项目。使用 Composer 管理依赖。
*   **规避升级风险**：在 `composer.json` 中精确锁定版本。
    *   ❌ 风险写法：`"monolog/monolog": "^3.0"` (允许小版本升级，可能变)
    *   ✅ 稳定写法：`"monolog/monolog": "3.5.0"` (锁定具体版本)
    *   **提交 `composer.lock` 到 Git**：确保所有环境和生产环境使用的库版本完全一致。

#### 3. 入口文件与自动加载 (Bootstrap)
在 `public/index.php` 中统一处理请求入口，利用 Composer 的 PSR-4 自动加载。

```php
// public/index.php
<?php
declare(strict_types=1);

// 1. 引入自动加载
require __DIR__ . '/../vendor/autoload.php';

// 2. 全局错误处理 (避免直接暴露错误)
set_error_handler(function($severity, $message, $file, $line) {
    throw new ErrorException($message, 0, $severity, $file, $line);
});
set_exception_handler(function($e) {
    // 记录日志，返回友好提示
    error_log($e->getMessage());
    http_response_code(500);
    echo "Internal Server Error";
});

// 3. 初始化核心服务 (数据库、路由等)
$container = require __DIR__ . '/../config/container.php';

// 4. 分发请求
$dispatcher = $container->get('router');
$response = $dispatcher->dispatch();
$response->send();
```

#### 4. 数据库访问 (安全与抽象)
不要自己写数据库连接类，使用成熟、轻量且长期维护的库，或者严格使用原生 PDO。
*   **推荐**：`doctrine/dbal` (稳定，抽象层) 或 原生 `PDO`。
*   **安全铁律**：
    *   必须使用 **预处理语句 (Prepared Statements)** 防止 SQL 注入。
    *   密码必须使用 `password_hash()` 和 `password_verify()`。
    *   事务管理要在业务层控制。

```php
// src/Database/UserRepository.php
class UserRepository {
    private PDO $pdo;
    public function __construct(PDO $pdo) { $this->pdo = $pdo; }
    
    public function findByEmail(string $email): ?array {
        $stmt = $this->pdo->prepare("SELECT * FROM users WHERE email = :email");
        $stmt->execute(['email' => $email]);
        return $stmt->fetch(PDO::FETCH_ASSOC) ?: null;
    }
}
```

#### 5. 路由控制 (轻量级)
不要自己写复杂的正则路由，容易出安全漏洞。使用一个独立的、无依赖的路由库。
*   **推荐**：`nikic/fast-route` (非常稳定，被许多框架底层使用) 或 `league/route`。

#### 6. 视图渲染 (防止 XSS)
不要在 PHP 文件中直接 `echo $_GET['name']`。
*   **方案 A (推荐)**：使用轻量模板引擎，如 `league/plates` 或 `twig` (即使不用框架，Twig 也可以独立使用)。它们默认会自动转义输出。
*   **方案 B (原生)**：编写一个简单的 `e()` 辅助函数进行 `htmlspecialchars` 转义。

#### 7. 日志与监控
不要使用 `echo` 或 `error_log` 散落在代码中。
*   **推荐**：`monolog/monolog`。它是 PSR-3 日志接口的标准实现，几乎不会被废弃。
*   **实践**：记录所有异常、慢查询、关键业务操作。

---

### 三、可参考的最小化架构示例

这是一个基于上述实践的“微架构”示例，展示了如何组装稳定组件。

**1. composer.json (依赖选择注重稳定性)**
```json
{
    "require": {
        "php": ">=8.1",
        "nikic/fast-route": "^1.3", 
        "monolog/monolog": "^3.5",
        "vlucas/phpdotenv": "^5.5"
    },
    "autoload": {
        "psr-4": {
            "App\\": "src/"
        }
    }
}
```
*注：`nikic/fast-route` 已经非常成熟，几乎不再变动，非常适合此场景。*

**2. config/container.php (简单的依赖注入容器)**
不使用重型容器，手写一个简单的数组容器或工厂模式。
```php
<?php
// 创建 PDO 实例
$pdo = new PDO(
    "mysql:host={$_ENV['DB_HOST']};dbname={$_ENV['DB_NAME']}",
    $_ENV['DB_USER'],
    $_ENV['DB_PASS'],
    [PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION]
);

// 简单容器
return [
    'db' => $pdo,
    'logger' => new Monolog\Logger('app'),
    // 业务类
    'user_repo' => new App\Repository\UserRepository($pdo),
];
```

**3. src/Controller/HomeController.php (业务逻辑)**
```php
namespace App\Controller;

class HomeController {
    public function index() {
        // 纯 PHP 渲染或调用模板引擎
        require __DIR__ . '/../../templates/home.php';
    }
}
```

**4. public/index.php (路由分发)**
```php
$dispatcher = \FastRoute\simpleDispatcher(function(\FastRoute\RouteCollector $r) {
    $r->addRoute('GET', '/', [App\Controller\HomeController::class, 'index']);
    $r->addRoute('POST', '/login', [App\Controller\AuthController::class, 'login']);
});

$routeInfo = $dispatcher->dispatch($_SERVER['REQUEST_METHOD'], $_SERVER['REQUEST_URI']);

switch ($routeInfo[0]) {
    case \FastRoute\Dispatcher::NOT_FOUND:
        http_response_code(404);
        echo "404 Not Found";
        break;
    case \FastRoute\Dispatcher::METHOD_NOT_ALLOWED:
        http_response_code(405);
        echo "405 Method Not Allowed";
        break;
    case \FastRoute\Dispatcher::FOUND:
        $handler = $routeInfo[1];
        $vars = $routeInfo[2];
        // 实例化控制器并调用
        $controller = new $handler[0]();
        call_user_func_array([$controller, $handler[1]], $vars);
        break;
}
```

---

### 四、另一种选择：微框架 (Micro-Frameworks)

如果你发现完全“裸奔”维护成本还是太高（比如要自己处理 Cookie、Session、中间件），可以考虑 **微框架**。它们提供了必要的基础设施，但不会强迫你使用特定的 ORM 或目录结构。

**推荐列表（注重长期维护）：**

1.  **Slim Framework (v4/v5)**
    *   **特点**：极度轻量，核心只有路由、中间件、依赖注入。
    *   **稳定性**：存在超过 10 年，社区庞大，升级平滑。
    *   **适用**：API 服务、小型 Web 应用。
2.  **Laminas (原 Zend Framework)**
    *   **特点**：组件化程度最高。你可以只用它的 `Validator` 或 `Http` 组件，而不必用整个框架。
    *   **稳定性**：企业级，由 Linux 基金会维护，几乎不可能倒闭。
3.  **Symfony Components**
    *   **特点**：Laravel 的底层。你可以只安装 `symfony/http-foundation`, `symfony/routing`, `symfony/console`。
    *   **稳定性**：PHP 界最稳定的组件库，长期支持版本 (LTS) 维护期长。

---

### 五、决策矩阵：什么时候该“去框架化”？

| 场景 | 建议 | 理由 |
| :--- | :--- | :--- |
| **企业内部小工具/脚本** | **无框架 / 脚本式** | 开发快，部署简单，无需复杂架构。 |
| **高性能 API 网关** | **微框架 (Slim)** | 需要极致性能，不需要视图层，逻辑简单。 |
| **长期维护的 SaaS 产品** | **重型框架 (Laravel/Symfony)** | 需要权限、队列、邮件等复杂生态，框架能节省 80% 重复工作。 |
| **对依赖极其敏感的项目** | **组件组合 (你的需求)** | 需要完全控制依赖版本，避免框架升级带来的破坏性变更。 |
| **团队新手较多** | **重型框架** | 框架强制规范了代码风格，减少“乱写”的风险。 |

### 六、总结建议

1.  **不要重复造轮子**：安全（加密、Hash）、HTTP 解析、数据库连接池，这些请使用成熟的库。
2.  **锁定版本是核心**：无论是否用框架，`composer.lock` 必须提交到版本控制，生产环境严禁 `composer update`。
3.  **测试是安全网**：没有框架的约束，代码结构容易散乱。**PHPUnit** 测试覆盖率是保证重构和升级不崩盘的唯一手段。
4.  **考虑 Symfony Components**：如果你想要框架的稳定性，但又不想被框架绑定，使用 **Symfony 组件** 自己组装是业界公认最稳健的“无框架”方案。

通过这种方式，你既避免了重型框架的“升级焦虑”，又规避了原生开发的“安全与维护风险”。


# 列出至少10个适合“无框架”开发时使用的PHP库

在“无框架”（或称“组件式”）开发模式下，选择库的核心标准是：**遵循 PSR 标准、社区生态庞大、长期维护记录良好、功能单一且稳定**。这样即使某个库停止更新，你也可以轻松替换它，而不会导致整个应用崩溃。

以下是为你精选的 **10+ 个适合无框架开发的 PHP 库**，按功能分类，并附带了选择理由和稳定性分析。

---

### 1. 日志记录：`monolog/monolog`
*   **功能**：将日志发送到文件、Socket、邮件、数据库等。
*   **为什么选它**：它是 PSR-3 日志接口的标准实现，几乎所有现代 PHP 框架（包括 Laravel、Symfony）底层都在用它。
*   **稳定性**：⭐⭐⭐⭐⭐ (极高)
*   **风险规避**：即使未来不再更新，由于其接口标准化，替换成本极低。
*   **安装**：`composer require monolog/monolog`

### 2. 数据库抽象层：`doctrine/dbal`
*   **功能**：提供比原生 PDO 更强大的数据库抽象层（查询构建器、类型安全、跨数据库兼容）。
*   **为什么选它**：比 ORM 轻量，比原生 PDO 安全。它不强制你使用特定的实体映射，适合自由架构。
*   **稳定性**：⭐⭐⭐⭐⭐ (企业级标准，维护超过 15 年)
*   **风险规避**：它是独立组件，不依赖 Doctrine ORM，升级影响小。
*   **安装**：`composer require doctrine/dbal`

### 3. 模板引擎：`twig/twig`
*   **功能**：安全的模板渲染，分离业务逻辑与视图。
*   **为什么选它**：默认开启自动转义（防 XSS），语法简洁。即使不用 Symfony 框架，Twig 也可以独立运行。
*   **稳定性**：⭐⭐⭐⭐⭐ (行业标准)
*   **风险规避**：语法极其稳定，大版本升级破坏性小。
*   **安装**：`composer require twig/twig`

### 4. 依赖注入容器：`php-di/php-di`
*   **功能**：管理类的实例化和依赖关系，解耦代码。
*   **为什么选它**：配置简单（支持注解、PHP 定义），性能优秀。在无框架项目中，它是组织代码结构的核心。
*   **稳定性**：⭐⭐⭐⭐ (社区活跃，长期维护)
*   **风险规避**：遵循 PSR-11 标准，未来可无缝切换到其他容器。
*   **安装**：`composer require php-di/php-di`

### 5. 数据验证：`respect/validation`
*   **功能**：灵活、强大的数据验证规则链。
*   **为什么选它**：独立于任何框架，规则丰富（如 `v::string()->length(1, 15)->alnum()`）。
*   **稳定性**：⭐⭐⭐⭐ (老牌库，逻辑稳定)
*   **风险规避**：验证逻辑通常是纯函数式的，极少受外部环境影响。
*   **安装**：`composer require respect/validation`

### 6. 环境变量管理：`vlucas/phpdotenv`
*   **功能**：从 `.env` 文件加载环境变量到 `$_ENV` 和 `getenv()`。
*   **为什么选它**：开发环境和生产环境配置分离的标准做法。
*   **稳定性**：⭐⭐⭐⭐⭐ (几乎每个现代 PHP 项目都在用)
*   **风险规避**：功能单一，几乎没有升级破坏的风险。
*   **安装**：`composer require vlucas/phpdotenv`

### 7. HTTP 消息实现 (PSR-7)：`laminas/laminas-diactoros`
*   **功能**：提供标准的 Request 和 Response 对象实现。
*   **为什么选它**：无框架开发需要自己处理 HTTP 请求。遵循 PSR-7 标准，意味着你的代码可以兼容任何支持该标准的中间件。
*   **稳定性**：⭐⭐⭐⭐⭐ (由 Linux 基金会支持的 Laminas 项目维护)
*   **风险规避**：作为基础标准实现，几乎不会发生逻辑变更。
*   **安装**：`composer require laminas/laminas-diactoros`

### 8. HTTP 客户端：`guzzlehttp/guzzle`
*   **功能**：发送 HTTP 请求（调用第三方 API、微服务通信）。
*   **为什么选它**：功能强大，支持异步、中间件、Mock 测试。
*   **稳定性**：⭐⭐⭐⭐⭐ (事实上的行业标准)
*   **风险规避**：生态极大，即使出问题，社区修复速度极快。
*   **安装**：`composer require guzzlehttp/guzzle`

### 9. 命令行工具：`symfony/console`
*   **功能**：构建强大的 CLI 命令（如定时任务、数据迁移脚本）。
*   **为什么选它**：你可以只引入这个组件，而不需要整个 Symfony 框架。它让无框架项目也能拥有像 Laravel Artisan 那样的命令行体验。
*   **稳定性**：⭐⭐⭐⭐⭐ (长期支持版本 LTS 极多)
*   **风险规避**：Symfony 组件以向后兼容性著称。
*   **安装**：`composer require symfony/console`

### 10. 测试框架：`phpunit/phpunit`
*   **功能**：单元测试、集成测试。
*   **为什么选它**：无框架项目缺乏框架自带的约束，**测试是防止代码腐化的唯一防线**。
*   **稳定性**：⭐⭐⭐⭐⭐ (PHP 测试的唯一标准)
*   **风险规避**：版本升级通常有详细的迁移指南，且测试代码本身能帮你发现升级带来的问题。
*   **安装**：`composer require --dev phpunit/phpunit`

---

### Bonus: 额外推荐 (针对特定场景)

#### 11. 路由：`league/route` 或 `nikic/fast-route`
*   **说明**：
    *   `nikic/fast-route`：极度稳定，几乎不再更新（功能已完备），适合追求“零变动”的项目。
    *   `league/route`：基于 PSR-7/15 标准，更现代，属于 The PHP League 组织，维护活跃。
*   **建议**：追求极致稳定选 `nikic`，追求标准兼容选 `league`。

#### 12. 加密安全：`defuse/php-encryption`
*   **说明**：简单易用的安全加密库。不要自己写加密算法。
*   **建议**：用于加密 Cookie、敏感数据存储。

---

### 如何管理这些库以避免“升级风险”？

选择了稳定的库只是第一步，**管理策略**才是关键：

1.  **锁定版本 (Lock Versions)**：
    在 `composer.json` 中，尽量锁定具体版本号，而不是使用通配符。
    *   ❌ `"monolog/monolog": "^3.0"` (允许 3.0 到 3.99，可能引入变更)
    *   ✅ `"monolog/monolog": "3.5.0"` (锁定死版本)
    *   **关键**：务必将 `composer.lock` 文件提交到 Git 仓库。

2.  **遵循 PSR 标准**：
    上述库大多遵循 PSR 标准（如 PSR-3 日志、PSR-7 HTTP、PSR-11 容器）。
    *   **好处**：如果你的日志库 `monolog` 真的停止维护了，你只需要换一个实现了 PSR-3 的库，你的业务代码**一行都不用改**。

3.  **定期审计依赖**：
    使用 `composer audit` 检查安全漏洞。无框架开发意味着你要自己对安全负责。

4.  **编写集成测试**：
    在升级任何库之前，运行你的测试套件。如果测试通过，升级风险就降低了 90%。

### 总结建议

这套组合拳（**Monolog + Doctrine DBAL + Twig + PHP-DI + PHPUnit**）构成了一个**“隐形框架”**。
*   它没有框架的“强制约束”，你拥有完全的控制权。
*   它拥有框架的“稳定性”，因为这些组件都是经过成千上万个项目验证的。
*   它避免了“框架废弃风险”，因为即使其中一个组件挂了，你可以单独替换它，而不需要重构整个应用。


# 使用这套库，开发一个简单的示例程序，实现homepage，login/logout，和welcome page

这是一个基于**组件式架构**的完整示例。我们将使用之前推荐的稳定库（Twig, Doctrine DBAL, PHP-DI, Monolog, Dotenv, FastRoute）构建一个具备 **首页、登录、登出、受保护欢迎页** 功能的最小化 Web 应用。

### 1. 项目目录结构

```text
my-app/
├── config/
│   ├── container.php       # 依赖注入配置
│   └── routes.php          # 路由定义
├── public/
│   ├── index.php           # 唯一入口文件
│   └── .htaccess           # Apache 重写规则 (可选)
├── src/
│   ├── Controller/         # 控制器
│   ├── Repository/         # 数据访问层
│   └── Service/            # 业务逻辑层 (认证)
├── templates/              # Twig 视图
├── tests/                  # 测试
├── .env                    # 环境变量 (不要提交到 Git)
├── .env.example            # 环境变量模板
├── composer.json           # 依赖管理
└── vendor/                 # 第三方库
```

### 2. 依赖安装 (`composer.json`)

```json
{
    "name": "demo/no-framework-app",
    "type": "project",
    "require": {
        "php": ">=8.1",
        "vlucas/phpdotenv": "^5.5",
        "php-di/php-di": "^7.0",
        "doctrine/dbal": "^3.6",
        "twig/twig": "^3.0",
        "monolog/monolog": "^3.5",
        "nikic/fast-route": "^1.3"
    },
    "autoload": {
        "psr-4": {
            "App\\": "src/"
        }
    }
}
```
**执行安装：** `composer install`

### 3. 数据库准备 (SQL)

在 MySQL 中创建一个数据库，并运行以下 SQL 创建用户表（预置一个密码为 `123456` 的用户）：

```sql
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 密码是 '123456' 的 bcrypt 哈希
INSERT INTO users (email, password) VALUES 
('admin@example.com', '$2y$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi');
```

### 4. 环境配置 (`.env`)

```ini
DB_HOST=localhost
DB_NAME=demo_db
DB_USER=root
DB_PASS=secret
APP_DEBUG=true
```

### 5. 核心代码实现

#### A. 入口文件 (`public/index.php`)
负责引导应用、启动会话、处理路由和异常。

```php
<?php
declare(strict_types=1);

require __DIR__ . '/../vendor/autoload.php';

use Dotenv\Dotenv;
use DI\ContainerBuilder;
use Monolog\Logger;
use Monolog\Handler\StreamHandler;

// 1. 加载环境变量
$dotenv = Dotenv::createImmutable(__DIR__ . '/..');
$dotenv->load();

// 2. 启动会话 (必须在输出前)
// 安全配置：防止 JS 访问 Cookie，仅 HTTPS 传输 (生产环境)
session_start([
    'cookie_httponly' => true,
    'cookie_secure' => false, // 生产环境设为 true
    'use_strict_mode' => true
]);

// 3. 构建依赖容器
$builder = new ContainerBuilder();
$builder->addDefinitions(__DIR__ . '/../config/container.php');
$container = $builder->build();

// 4. 初始化日志
$logger = $container->get(Logger::class);
$logger->info("Request started: " . $_SERVER['REQUEST_URI']);

// 5. 路由分发
$dispatcher = $container->get(\FastRoute\Dispatcher::class);
$routeInfo = $dispatcher->dispatch($_SERVER['REQUEST_METHOD'], $_SERVER['REQUEST_URI']);

// 6. 处理路由结果
switch ($routeInfo[0]) {
    case \FastRoute\Dispatcher::NOT_FOUND:
        http_response_code(404);
        echo "404 Not Found";
        break;
    case \FastRoute\Dispatcher::METHOD_NOT_ALLOWED:
        http_response_code(405);
        echo "405 Method Not Allowed";
        break;
    case \FastRoute\Dispatcher::FOUND:
        $handler = $routeInfo[1];
        $vars = $routeInfo[2];
        
        // 实例化控制器并调用方法
        // 注意：这里简单处理，生产环境建议通过容器解析控制器
        $controller = new $handler[0]($container);
        try {
            echo call_user_func_array([$controller, $handler[1]], $vars);
        } catch (\Exception $e) {
            $logger->error($e->getMessage());
            http_response_code(500);
            echo "Internal Server Error";
        }
        break;
}
```

#### B. 依赖配置 (`config/container.php`)
组装所有组件，这是“无框架”项目的核心骨架。

```php
<?php
use DI\Container;
use Doctrine\DBAL\DriverManager;
use Monolog\Logger;
use Monolog\Handler\StreamHandler;
use Twig\Environment;
use Twig\Loader\FilesystemLoader;
use FastRoute\simpleDispatcher;

return [
    // 数据库连接
    \Doctrine\DBAL\Connection::class => function() {
        return DriverManager::getConnection([
            'driver' => 'pdo_mysql',
            'host' => $_ENV['DB_HOST'],
            'dbname' => $_ENV['DB_NAME'],
            'user' => $_ENV['DB_USER'],
            'password' => $_ENV['DB_PASS'],
            'charset' => 'utf8mb4',
        ]);
    },

    // 日志
    Logger::class => function() {
        $log = new Logger('app');
        $log->pushHandler(new StreamHandler(__DIR__ . '/../var/app.log', Logger::DEBUG));
        return $log;
    },

    // Twig 模板引擎
    Environment::class => function() {
        $loader = new FilesystemLoader(__DIR__ . '/../templates');
        $twig = new Environment($loader, [
            'cache' => __DIR__ . '/../var/cache/twig',
            'debug' => $_ENV['APP_DEBUG'] === 'true',
            'auto_reload' => true
        ]);
        // 添加全局函数或变量
        $twig->addGlobal('session', $_SESSION);
        return $twig;
    },

    // 路由配置
    \FastRoute\Dispatcher::class => function() {
        return simpleDispatcher(function(\FastRoute\RouteCollector $r) {
            $r->addRoute('GET', '/', [App\Controller\HomeController::class, 'index']);
            $r->addRoute('GET', '/login', [App\Controller\AuthController::class, 'showLogin']);
            $r->addRoute('POST', '/login', [App\Controller\AuthController::class, 'login']);
            $r->addRoute('GET', '/logout', [App\Controller\AuthController::class, 'logout']);
            $r->addRoute('GET', '/welcome', [App\Controller\HomeController::class, 'welcome']);
        });
    },
    
    // 认证服务
    App\Service\AuthService::class => function(Container $c) {
        return new App\Service\AuthService(
            $c->get(\Doctrine\DBAL\Connection::class),
            $c->get(Logger::class)
        );
    },
];
```

#### C. 认证服务 (`src/Service/AuthService.php`)
处理核心业务逻辑，保持控制器干净。

```php
<?php
namespace App\Service;

use Doctrine\DBAL\Connection;
use Monolog\Logger;

class AuthService {
    private Connection $db;
    private Logger $logger;

    public function __construct(Connection $db, Logger $logger) {
        $this->db = $db;
        $this->logger = $logger;
    }

    public function attemptLogin(string $email, string $password): bool {
        // 使用 DBAL 防止 SQL 注入
        $user = $this->db->fetchAssociative(
            "SELECT * FROM users WHERE email = ?",
            [$email]
        );

        if ($user && password_verify($password, $user['password'])) {
            // 登录成功，写入 Session
            $_SESSION['user_id'] = $user['id'];
            $_SESSION['user_email'] = $user['email'];
            
            // 防止会话固定攻击
            session_regenerate_id(true);
            
            $this->logger->info("User logged in: " . $email);
            return true;
        }

        $this->logger->warning("Login failed for: " . $email);
        return false;
    }

    public function logout(): void {
        $_SESSION = [];
        session_destroy();
        $this->logger->info("User logged out");
    }

    public function check(): bool {
        return isset($_SESSION['user_id']);
    }
    
    public function getUserEmail(): ?string {
        return $_SESSION['user_email'] ?? null;
    }
}
```

#### D. 控制器 (`src/Controller/`)

**HomeController.php**
```php
<?php
namespace App\Controller;

use DI\Container;
use Twig\Environment;
use App\Service\AuthService;

class HomeController {
    private Container $container;
    public function __construct(Container $container) {
        $this->container = $container;
    }

    public function index(): string {
        /** @var Environment $twig */
        $twig = $this->container->get(Environment::class);
        return $twig->render('home.html.twig', ['title' => '首页']);
    }

    public function welcome(): string {
        /** @var AuthService $auth */
        $auth = $this->container->get(AuthService::class);
        
        // 简单的权限拦截
        if (!$auth->check()) {
            header('Location: /login');
            exit;
        }

        /** @var Environment $twig */
        $twig = $this->container->get(Environment::class);
        return $twig->render('welcome.html.twig', [
            'title' => '欢迎',
            'email' => $auth->getUserEmail()
        ]);
    }
}
```

**AuthController.php**
```php
<?php
namespace App\Controller;

use DI\Container;
use Twig\Environment;
use App\Service\AuthService;

class AuthController {
    private Container $container;
    public function __construct(Container $container) {
        $this->container = $container;
    }

    public function showLogin(): string {
        /** @var Environment $twig */
        $twig = $this->container->get(Environment::class);
        return $twig->render('login.html.twig', ['title' => '登录', 'error' => null]);
    }

    public function login(): string {
        /** @var AuthService $auth */
        $auth = $this->container->get(AuthService::class);
        /** @var Environment $twig */
        $twig = $this->container->get(Environment::class);

        $email = $_POST['email'] ?? '';
        $password = $_POST['password'] ?? '';

        if ($auth->attemptLogin($email, $password)) {
            header('Location: /welcome');
            exit;
        }

        // 登录失败，重新渲染表单
        return $twig->render('login.html.twig', [
            'title' => '登录',
            'error' => '邮箱或密码错误'
        ]);
    }

    public function logout(): void {
        /** @var AuthService $auth */
        $auth = $this->container->get(AuthService::class);
        $auth->logout();
        header('Location: /');
        exit;
    }
}
```

#### E. 视图模板 (`templates/`)

**base.html.twig** (布局文件)
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{ title }}</title>
</head>
<body>
    <nav>
        <a href="/">首页</a>
        {% if session.user_id %}
            | <a href="/welcome">欢迎页</a>
            | <a href="/logout">退出 ({{ session.user_email }})</a>
        {% else %}
            | <a href="/login">登录</a>
        {% endif %}
    </nav>
    <hr>
    {% block content %}{% endblock %}
</body>
</html>
```

**home.html.twig**
```html
{% extends "base.html.twig" %}
{% block content %}
    <h1>欢迎来到无框架应用</h1>
    <p>这是一个基于组件构建的安全示例。</p>
{% endblock %}
```

**login.html.twig**
```html
{% extends "base.html.twig" %}
{% block content %}
    <h1>登录</h1>
    {% if error %}
        <p style="color:red">{{ error }}</p>
    {% endif %}
    <form method="POST" action="/login">
        <label>邮箱：<input type="email" name="email" required></label><br><br>
        <label>密码：<input type="password" name="password" required></label><br><br>
        <button type="submit">登录</button>
    </form>
{% endblock %}
```

**welcome.html.twig**
```html
{% extends "base.html.twig" %}
{% block content %}
    <h1>受保护的欢迎页</h1>
    <p>你好，{{ email }}！你已成功登录。</p>
    <p>如果你未登录，是无法看到此页面的。</p>
{% endblock %}
```

---

### 6. 运行与测试

1.  **启动服务器**：
    在项目根目录运行 PHP 内置服务器：
    ```bash
    php -S localhost:8000 -t public
    ```
2.  **访问**：
    *   打开 `http://localhost:8000` 查看首页。
    *   点击登录，使用 `admin@example.com` / `123456` 登录。
    *   登录后会自动跳转到 `/welcome`。
    *   尝试直接在浏览器地址栏输入 `/welcome` (未登录状态)，应被重定向回登录页。
    *   点击退出，会话清除。

### 7. 这个示例体现了哪些最佳实践？

1.  **安全性 (Security)**：
    *   **SQL 注入**：使用 `Doctrine DBAL` 的参数化查询 (`?` 占位符)。
    *   **XSS 攻击**：`Twig` 默认自动转义所有输出变量。
    *   **密码安全**：使用 `password_verify` 而不是明文比对。
    *   **会话固定**：登录成功后调用 `session_regenerate_id(true)`。
    *   **入口收敛**：所有请求通过 `public/index.php`，无法直接访问 `src` 或 `config` 代码。

2.  **可维护性 (Maintainability)**：
    *   **依赖注入**：控制器不直接 `new` 数据库连接，而是通过容器获取。如果想换数据库库，只需修改 `container.php`。
    *   **职责分离**：认证逻辑在 `AuthService`，路由在 `routes.php`，视图在 `templates`。
    *   **配置分离**：数据库密码在 `.env` 中，不硬编码在代码里。

3.  **稳定性 (Stability)**：
    *   所有使用的库都是各自领域的“事实标准”，即使某天你想换掉 Twig，因为代码中通过接口交互，迁移成本很低。
    *   没有框架的“魔法”，代码流向清晰，调试容易。

### 8. 下一步建议

如果要将其投入生产环境，你还需要补充：
1.  **CSRF 保护**：在表单中添加 CSRF Token 验证（可以写一个简单的 Middleware 或 Service）。
2.  **错误页面**：自定义 404 和 500 的 Twig 页面，而不是直接输出文本。
3.  **缓存**：引入 `symfony/cache` 组件缓存数据库查询或模板。
4.  **部署**：配置 Nginx/Apache 将 `public` 设为根目录，并禁止访问其他目录。

这个示例展示了如何在没有重型框架束缚的情况下，依然构建出结构清晰、安全且易于维护的 PHP 应用。

# 如何添加 CSRF 保护

在无框架开发中，添加 CSRF（跨站请求伪造）保护是**必须**的安全措施。最稳健的最佳实践是使用 **`symfony/security-csrf`** 组件。它是一个独立的库，不依赖整个 Symfony 框架，但提供了企业级的安全性。

为了适配我们之前使用的 **原生 `$_SESSION`**（而不是 Symfony 的 Session 组件），我们需要写一个简单的适配器。

以下是完整的实施步骤：

### 1. 安装依赖

```bash
composer require symfony/security-csrf
```

### 2. 创建 Session 存储适配器

Symfony 的 CSRF 组件默认期望使用 Symfony 的 Session 对象。为了兼容我们之前的原生 `$_SESSION` 实现，我们需要实现 `TokenStorageInterface`。

**文件：** `src/Security/CsrfSessionStorage.php`

```php
<?php
namespace App\Security;

use Symfony\Component\Security\Csrf\TokenStorage\TokenStorageInterface;

class CsrfSessionStorage implements TokenStorageInterface
{
    private string $sessionKey = '_csrf_tokens';

    public function getToken(string $tokenId): ?string
    {
        if (!isset($_SESSION[$this->sessionKey][$tokenId])) {
            return null;
        }
        return $_SESSION[$this->sessionKey][$tokenId];
    }

    public function setToken(string $tokenId, string $token): void
    {
        // 确保 session 数组已初始化
        if (!isset($_SESSION[$this->sessionKey])) {
            $_SESSION[$this->sessionKey] = [];
        }
        $_SESSION[$this->sessionKey][$tokenId] = $token;
    }

    public function removeToken(string $tokenId): ?string
    {
        $token = $this->getToken($tokenId);
        if ($token !== null) {
            unset($_SESSION[$this->sessionKey][$tokenId]);
        }
        return $token;
    }
}
```

### 3. 创建 CSRF 服务封装

为了方便在控制器中调用，我们封装一下 `CsrfTokenManager`。

**文件：** `src/Security/CsrfService.php`

```php
<?php
namespace App\Security;

use Symfony\Component\Security\Csrf\CsrfToken;
use Symfony\Component\Security\Csrf\CsrfTokenManager;

class CsrfService
{
    private CsrfTokenManager $manager;

    public function __construct(CsrfTokenManager $manager)
    {
        $this->manager = $manager;
    }

    // 生成 Token (用于视图)
    public function generateToken(string $tokenId): string
    {
        return $this->manager->getToken($tokenId)->getValue();
    }

    // 验证 Token (用于控制器)
    public function validateToken(string $tokenId, string $submittedToken): bool
    {
        $token = new CsrfToken($tokenId, $submittedToken);
        return $this->manager->isTokenValid($token);
    }
}
```

### 4. 创建 Twig 扩展 (可选但推荐)

为了在模板中方便地生成 Token，我们添加一个 Twig 函数 `csrf_token('login')`。

**文件：** `src/Twig/CsrfExtension.php`

```php
<?php
namespace App\Twig;

use App\Security\CsrfService;
use Twig\Extension\AbstractExtension;
use Twig\TwigFunction;

class CsrfExtension extends AbstractExtension
{
    private CsrfService $csrfService;

    public function __construct(CsrfService $csrfService)
    {
        $this->csrfService = $csrfService;
    }

    public function getFunctions(): array
    {
        return [
            new TwigFunction('csrf_token', [$this, 'generateToken']),
        ];
    }

    public function generateToken(string $tokenId): string
    {
        return $this->csrfService->generateToken($tokenId);
    }
}
```

### 5. 更新依赖容器 (`config/container.php`)

将新组件注册到容器中。

```php
// 在 return 数组中添加以下定义

use Symfony\Component\Security\Csrf\CsrfTokenManager;
use App\Security\CsrfSessionStorage;
use App\Security\CsrfService;
use App\Twig\CsrfExtension;

return [
    // ... 之前的配置 ...

    // CSRF 存储 (适配原生 Session)
    \Symfony\Component\Security\Csrf\TokenStorage\TokenStorageInterface::class => function() {
        return new CsrfSessionStorage();
    },

    // CSRF 管理器
    CsrfTokenManager::class => function(Container $c) {
        return new CsrfTokenManager(
            null, // 默认使用哈希生成器
            $c->get(\Symfony\Component\Security\Csrf\TokenStorage\TokenStorageInterface::class)
        );
    },

    // CSRF 业务服务
    CsrfService::class => function(Container $c) {
        return new CsrfService($c->get(CsrfTokenManager::class));
    },

    // Twig 扩展
    CsrfExtension::class => function(Container $c) {
        return new CsrfExtension($c->get(CsrfService::class));
    },
    
    // 将扩展注册到 Twig 环境中 (需要修改之前的 Twig 定义)
    Environment::class => function(Container $c) {
        $loader = new FilesystemLoader(__DIR__ . '/../templates');
        $twig = new Environment($loader, [
            'cache' => __DIR__ . '/../var/cache/twig',
            'debug' => $_ENV['APP_DEBUG'] === 'true',
        ]);
        
        // 添加全局变量
        $twig->addGlobal('session', $_SESSION);
        
        // 注册 CSRF 扩展
        $twig->addExtension($c->get(CsrfExtension::class));
        
        return $twig;
    },
];
```

### 6. 更新登录视图 (`templates/login.html.twig`)

在表单中添加隐藏字段，使用刚才定义的 `csrf_token` 函数。**Token ID 必须唯一标识这个表单动作**（例如 `'login'`）。

```html
{% extends "base.html.twig" %}
{% block content %}
    <h1>登录</h1>
    {% if error %}
        <p style="color:red">{{ error }}</p>
    {% endif %}
    <form method="POST" action="/login">
        <!-- CSRF Token -->
        <input type="hidden" name="_csrf_token" value="{{ csrf_token('login') }}">
        
        <label>邮箱：<input type="email" name="email" required></label><br><br>
        <label>密码：<input type="password" name="password" required></label><br><br>
        <button type="submit">登录</button>
    </form>
{% endblock %}
```

### 7. 更新认证控制器 (`src/Controller/AuthController.php`)

在 POST 请求处理中验证 Token。

```php
<?php
namespace App\Controller;

use DI\Container;
use Twig\Environment;
use App\Service\AuthService;
use App\Security\CsrfService; // 引入 CSRF 服务

class AuthController {
    private Container $container;
    public function __construct(Container $container) {
        $this->container = $container;
    }

    // ... showLogin 方法不变 ...

    public function login(): string {
        /** @var CsrfService $csrf */
        $csrf = $this->container->get(CsrfService::class);
        
        /** @var AuthService $auth */
        $auth = $this->container->get(AuthService::class);
        
        /** @var Environment $twig */
        $twig = $this->container->get(Environment::class);

        // 1. 获取并验证 CSRF Token
        $submittedToken = $_POST['_csrf_token'] ?? '';
        if (!$csrf->validateToken('login', $submittedToken)) {
            // Token 无效，可能是 CSRF 攻击或 Session 过期
            return $twig->render('login.html.twig', [
                'title' => '登录',
                'error' => '安全验证失败，请刷新页面重试'
            ]);
        }

        // 2. 验证通过后，继续处理业务逻辑
        $email = $_POST['email'] ?? '';
        $password = $_POST['password'] ?? '';

        if ($auth->attemptLogin($email, $password)) {
            header('Location: /welcome');
            exit;
        }

        return $twig->render('login.html.twig', [
            'title' => '登录',
            'error' => '邮箱或密码错误'
        ]);
    }
    
    // ... logout 方法不变 ...
}
```

### 8. 额外防御层：SameSite Cookie

除了 Token 验证，现代浏览器还支持通过 Cookie 属性来防御 CSRF。在 `public/index.php` 的 `session_start` 配置中加强设置：

```php
// public/index.php

session_start([
    'cookie_httponly' => true,
    'cookie_secure' => true,      // 生产环境必须为 true (HTTPS)
    'cookie_samesite' => 'Strict', // 关键：阻止跨站发送 Cookie
    'use_strict_mode' => true
]);
```
*   **`Strict`**: 任何跨站请求（包括链接跳转）都不发送 Cookie。最安全，但用户体验稍差（从外部链接跳回网站可能需要重新登录）。
*   **`Lax`**: 允许顶级导航（如点击链接）发送 Cookie，但阻止 AJAX/表单 POST。这是大多数场景的平衡选择。

### 9. 关键注意事项

1.  **Token ID 唯一性**：
    *   登录表单用 `'login'`。
    *   注册表单用 `'register'`。
    *   修改密码用 `'change_password'`。
    *   **不要**在所有表单中复用同一个 Token ID，否则一个页面的 Token 泄露可能影响其他页面。

2.  **Token 生命周期**：
    *   上面的实现中，Token 存储在 Session 中，Session 过期 Token 即失效。
    *   对于高安全需求，可以在 `CsrfSessionStorage` 中增加时间戳检查，实现 Token 自动过期（例如 30 分钟）。

3.  **AJAX 请求**：
    *   如果使用 AJAX 提交表单，需要将 Token 从 Meta 标签或 JSON 配置中获取，并添加到 HTTP 请求头（如 `X-CSRF-Token`）中，而不是表单字段。

### 总结

通过引入 `symfony/security-csrf` 组件并配合原生 Session 适配器，我们实现了：
1.  **安全性**：使用业界标准的加密算法生成和验证 Token。
2.  **兼容性**：不需要重构现有的 Session 管理机制。
3.  **易用性**：通过 Twig 扩展，在视图中调用 `{{ csrf_token('id') }}` 非常简单。
4.  **解耦**：CSRF 逻辑独立于业务逻辑，符合无框架开发的组件化原则。

这是在不使用重型框架的前提下，能达到的最高安全标准之一。

