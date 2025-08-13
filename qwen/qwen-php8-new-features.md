# 总结一下 php 自从 8.0 之后添加的新特性

PHP 8.0 是 PHP 发展史上的一个里程碑，随后的 8.1、8.2、8.3 以及即将到来的 8.4 持续引入了大量现代化特性，显著提升了代码的**类型安全性**、**表达力**和**性能**。

以下是 PHP 8.0 及之后版本主要新特性的详细总结：

---

### 🚀 PHP 8.0 (2020 年 11 月) - 奠基之作
8.0 带来了语法层面的重大革新，让 PHP 看起来更像现代语言。

1.  **JIT (Just In Time) 编译器**
    *   **作用**：将 PHP 代码编译为机器码执行，显著提升 CPU 密集型任务的性能（如数学计算、图像处理）。
    *   **注意**：对于典型的 Web I/O 密集型应用，性能提升可能不明显。
2.  **联合类型 (Union Types)**
    *   允许参数或返回值是多种类型之一。
    *   ```php
        function foo(int|float $num): int|float { ... }
        ```
3.  **命名参数 (Named Arguments)**
    *   调用函数时可以跳过默认参数，通过参数名传递。
    *   ```php
        // 以前
        array_fill(0, 100, 50);
        // 现在
        array_fill(start_index: 0, count: 100, value: 50);
        ```
4.  **属性 (Attributes)**
    *   原生支持元数据（注解），替代了 PHPDoc 注释。
    *   ```php
        #[Route("/home", methods: ["GET"])]
        class HomeController { ... }
        ```
5.  **构造函数属性提升 (Constructor Property Promotion)**
    *   在构造函数中直接定义并初始化类属性，减少样板代码。
    *   ```php
        class User {
            public function __construct(
                public string $name,
                private int $id
            ) {}
        }
        ```
6.  **Match 表达式**
    *   `switch` 的增强版，支持返回值、严格比较、更简洁。
    *   ```php
        $status = match($code) {
            200, 300 => 'OK',
            404 => 'Not Found',
            default => 'Unknown',
        };
        ```
7.  **Nullsafe 操作符 (`?->`)**
    *   链式调用中自动处理 null，避免层层 `if` 判断。
    *   ```php
        // 以前
        $country = $session?->user?->address?->country;
        // 现在
        $country = $session?->user?->address?->country; // 语法更简洁
        ```
8.  **混合类型 (Mixed Type)**
    *   `mixed` 成为显式类型，等同于 `array|bool|float|int|...` 等所有类型的联合。

---

### 🔥 PHP 8.1 (2021 年 11 月) - 类型与结构增强
8.1 重点增强了类型系统和面向对象的能力。

1.  **枚举 (Enums)**
    *   原生支持枚举类型，支持纯枚举和 backed enums（关联标量值）。
    *   ```php
        enum Status: int {
            case PENDING = 0;
            case APPROVED = 1;
        }
        ```
2.  **只读属性 (Readonly Properties)**
    *   属性初始化后不可修改（构造函数外）。
    *   ```php
        class User {
            public readonly string $id;
        }
        ```
3.  **Fibers (纤程)**
    *   轻量级的并发单元，为异步编程框架（如 ReactPHP, Amp）提供了底层支持。
    *   ```php
        $fiber = new Fiber(function(): void {
            // 可暂停/恢复的任务
        });
        ```
4.  **第一类可调用语法 (First-class Callable Syntax)**
    *   获取可调用对象更简洁。
    *   ```php
        // 以前
        $cb = [$obj, 'method'];
        // 现在
        $cb = $obj->method(...);
        ```
5.  **交叉类型 (Intersection Types)**
    *   要求值同时满足多个类型（主要用于对象）。
    *   ```php
        function foo(HasId & HasName $obj) { ... }
        ```
6.  **Never 类型**
    *   表示函数不返回（抛出异常或终止脚本）。
    *   ```php
        function redirect(): never {
            header('Location: ...');
            exit;
        }
        ```

---

### 🛡️ PHP 8.2 (2022 年 12 月) - 安全与一致性
8.2 继续完善类型系统，并增强了数据不可变性。

1.  **只读类 (Readonly Classes)**
    *   类中所有属性默认为 `readonly`，无需逐个声明。
    *   ```php
        readonly class User {
            public string $name; // 自动变为 readonly
        }
        ```
2.  **析取范式类型 (DNF Types)**
    *   支持更复杂的类型组合：`(A & B) | C`。
    *   ```php
        public function foo((A&B)|C $param) { ... }
        ```
3.  **Traits 中的常量 (Constants in Traits)**
    *   允许在 Trait 中定义常量。
4.  **独立的 Null/False/True 类型**
    *   `null`, `false`, `true` 可以作为独立的类型声明（不再仅仅是联合类型的一部分）。
    *   ```php
        function foo(): true { return true; }
        ```
5.  **敏感参数属性 (SensitiveParameter)**
    *   标记参数为敏感，在堆栈跟踪中隐藏其值（增强安全）。
    *   ```php
        function login(#[SensitiveParameter] string $password) { ... }
        ```
6.  **随机扩展重构 (Random Extension)**
    *   新的 `Random\Randomizer` 类，提供更安全、更面向对象的随机数生成。

---

### ⚡ PHP 8.3 (2023 年 11 月) - 细节完善
8.3 主要是一些长期请求的特性落地，提升了开发体验。

1.  **类型化的类常量 (Typed Class Constants)**
    *   常量现在可以声明类型。
    *   ```php
        class Config {
            public const int MAX_RETRY = 5;
        }
        ```
2.  **动态获取类常量 (Dynamic Class Constant Fetch)**
    *   支持通过变量获取常量值。
    *   ```php
        $name = 'MAX_RETRY';
        echo Config::{$name};
        ```
3.  **JSON 验证函数 (`json_validate`)**
    *   无需解码即可验证 JSON 字符串是否合法，节省内存。
    *   ```php
        if (json_validate($jsonString)) { ... }
        ```
4.  **Override 属性 (`#[Override]`)**
    *   标记方法意在重写父类方法，如果父类方法不存在或签名不匹配则报错。
    *   ```php
        class Child extends Parent {
            #[Override]
            public function run() { ... }
        }
        ```
5.  **深克隆 (Deep Cloning)**
    *   在 `__clone` 方法中可以更精细地控制对象克隆过程。

---

### 🔮 PHP 8.4 (预计 2024 年 11 月发布) - 前瞻特性
*注：以下特性目前处于开发/RC 阶段，可能会微调。*

1.  **属性钩子 (Property Hooks)**
    *   无需编写 `get/set` 方法即可拦截属性的读写，类似 C# 或 Kotlin。
    *   ```php
        class User {
            public string $name {
                get => strtoupper($this->name);
                set {
                    if (empty($value)) throw new Exception();
                    $this->name = $value;
                }
            }
        }
        ```
2.  **非对称可见性 (Asymmetric Visibility)**
    *   属性的读取和设置可以有不同的可见性（例如：公开读取，私有设置）。
    *   ```php
        class User {
            public private(set) string $name;
        }
        ```
3.  **匿名类无需括号**
    *   `new class { }` 代替 `new class() { }`。
4.  **请求 - 响应流 (Request-Response Flow)**
    *   旨在标准化 Web 框架处理请求和响应的方式（RFC 阶段）。

---

### 📊 总结：PHP 8.x 的核心趋势

| 特性领域 | 关键改进 | 受益点 |
| :--- | :--- | :--- |
| **类型系统** | 联合、交叉、DNF、类型化常量、独立 Null/True/False | **代码更健壮**，IDE 提示更准确，减少运行时错误。 |
| **语法糖** | 构造函数提升、Match、Nullsafe、命名参数 | **代码更简洁**，可读性更强，减少样板代码。 |
| **面向对象** | 枚举、只读类/属性、Traits 常量、Override 属性 | **设计更规范**，数据不可变性更容易实现。 |
| **性能与安全** | JIT、json_validate、SensitiveParameter、随机数重构 | **运行更快**，敏感数据处理更安全。 |
| **开发体验** | 属性 (Attributes)、Fibers、属性钩子 (8.4) | **框架开发更强大**，元数据管理更原生。 |

**建议：**
如果你的项目还在 PHP 7.x，**强烈建议升级到 PHP 8.2 或 8.3**。这不仅能获得性能提升，更重要的是类型系统的完善能极大降低维护成本。对于新项目，直接以 **PHP 8.3+** 为基准进行开发是最佳选择。

