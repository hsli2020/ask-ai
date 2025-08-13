# 介绍一下Doctrine ORM，总结它的常见用法

Doctrine ORM 是 PHP 生态中最流行、功能最强大的**对象关系映射（Object-Relational Mapper, ORM）**工具之一。它允许开发者使用面向对象的方式与数据库交互，而不是直接编写 SQL 语句。Doctrine 也是 Symfony 框架默认的 ORM 组件。

以下是对 Doctrine ORM 的详细介绍及其常见用法的总结。

---

### 1. 核心概念

在使用之前，需要理解以下几个核心概念：

*   **Entity (实体)**: 对应数据库中的一张表。一个 PHP 类就是一个实体，类的属性对应表的字段。
*   **EntityManager (实体管理器)**: ORM 的核心入口。负责管理实体的生命周期（创建、读取、更新、删除），维护内存中对象与数据库记录的一致性（单位工作模式 Unit of Work）。
*   **Repository (仓库)**: 专门用于查询特定实体的类。通常每个 Entity 都有一个对应的 Repository。
*   **Metadata (元数据)**: 描述 Entity 如何映射到数据库的信息（可以通过 PHP 8 Attributes、Annotations、XML 或 YAML 配置）。
*   **DQL (Doctrine Query Language)**: 类似 SQL 的查询语言，但操作的是对象和属性，而不是表和字段。

---

### 2. 常见用法总结

#### 2.1 安装与配置
通常通过 Composer 安装。在 Symfony 中已预装，独立使用需安装 `doctrine/orm` 和 `doctrine/dbal`。

```bash
composer require doctrine/orm
```

配置核心在于数据库连接信息（DSN, user, password）和映射驱动（如 Attributes）。

#### 2.2 定义实体 (Entity)
使用 PHP 8 的 **Attributes (特性)** 是现代推荐的方式（旧版使用 Annotations 注释）。

```php
<?php
// src/Entity/User.php
namespace App\Entity;

use Doctrine\ORM\Mapping as ORM;
use Doctrine\Common\Collections\ArrayCollection;
use Doctrine\Common\Collections\Collection;

#[ORM\Entity]
#[ORM\Table(name: 'users')]
class User
{
    #[ORM\Id]
    #[ORM\GeneratedValue]
    #[ORM\Column(type: 'integer')]
    private int $id;

    #[ORM\Column(type: 'string', length: 255)]
    private string $username;

    #[ORM\Column(type: 'string', length: 255)]
    private string $email;

    // 一对多关系示例：一个用户有多篇文章
    #[ORM\OneToMany(mappedBy: 'author', targetEntity: Post::class, cascade: ['persist', 'remove'])]
    private Collection $posts;

    public function __construct()
    {
        $this->posts = new ArrayCollection();
    }

    // Getter 和 Setter ...
}
```

#### 2.3 基础 CRUD 操作
通过 `EntityManager` 进行数据的增删改查。

```php
// 获取 EntityManager (在 Symfony 中通常注入 EntityManagerInterface)
$em = $this->getDoctrine()->getManager(); 

// 1. Create (创建)
$user = new User();
$user->setUsername('john_doe');
$user->setEmail('john@example.com');

$em->persist($user); // 标记为待持久化
$em->flush();        // 执行 SQL 插入

// 2. Read (查询)
$repo = $em->getRepository(User::class);

// 根据主键查找
$user = $repo->find(1); 

// 根据条件查找
$users = $repo->findBy(['email' => 'john@example.com']);

// 查找单个
$user = $repo->findOneBy(['username' => 'john_doe']);

// 3. Update (更新)
// 直接修改对象属性即可，Doctrine 会检测变化
$user->setEmail('new_email@example.com');
$em->flush(); // 执行 SQL UPDATE

// 4. Delete (删除)
$em->remove($user);
$em->flush(); // 执行 SQL DELETE
```

#### 2.4 高级查询 (DQL & QueryBuilder)
当 `findBy` 无法满足复杂需求时，使用 DQL 或 QueryBuilder。

**使用 QueryBuilder (推荐，防注入):**
```php
$qb = $em->createQueryBuilder();
$query = $qb->select('u')
            ->from(User::class, 'u')
            ->where('u.email LIKE :email')
            ->setParameter('email', '%@example.com')
            ->orderBy('u.id', 'DESC')
            ->setMaxResults(10)
            ->getQuery();

$users = $query->getResult();
```

**使用 DQL (字符串形式):**
```php
$dql = "SELECT u FROM App\Entity\User u WHERE u.username = :name";
$query = $em->createQuery($dql)->setParameter('name', 'john_doe');
$user = $query->getOneOrNullResult();
```

#### 2.5 实体关系映射 (Relationships)
Doctrine 强大之处在于处理表关联。

*   **ManyToOne (多对一)**: 多篇文章属于一个作者。
*   **OneToMany (一对多)**: 一个作者拥有多篇文章。
*   **OneToOne (一对一)**: 用户对应一个用户详情。
*   **ManyToMany (多对多)**: 学生对应课程（通常需要中间表）。

**示例：双向关联 (User <-> Post)**
```php
// User 端 (OneToMany)
#[ORM\OneToMany(mappedBy: 'author', targetEntity: Post::class)]
private Collection $posts;

// Post 端 (ManyToOne)
#[ORM\ManyToOne(targetEntity: User::class, inversedBy: 'posts')]
#[ORM\JoinColumn(nullable: false)]
private User $author;
```
*注意：维护双向关联时，需要在 Setter 方法中同步两边关系，否则可能导致数据不一致。*

#### 2.6 数据库迁移 (Migrations)
Doctrine Migrations 用于版本化管理数据库结构变更。

```bash
# 生成迁移文件 (对比 Entity 变化和当前数据库)
php bin/console doctrine:migrations:diff

# 执行迁移 (应用变更到数据库)
php bin/console doctrine:migrations:migrate
```

---

### 3. 最佳实践与常见坑

#### 3.1 N+1 查询问题 (性能杀手)
**问题**: 循环遍历实体集合时，访问关联对象会触发额外的 SQL 查询。
```php
// 糟糕的代码
$posts = $repo->findAll(); // 1 次查询
foreach ($posts as $post) {
    echo $post->getAuthor()->getName(); // 每次循环触发 1 次查询 (N 次)
}
```
**解决**: 使用 `JOIN` 预加载关联数据。
```php
// 优化代码
$query = $em->createQueryBuilder()
    ->select('p', 'a') // 同时选择 Post 和 Author
    ->from(Post::class, 'p')
    ->leftJoin('p.author', 'a') // 预加载
    ->getQuery();
$posts = $query->getResult();
```

#### 3.2 事务管理 (Transactions)
涉及多个写操作时，务必使用事务保证数据一致性。
```php
$em->getConnection()->beginTransaction();
try {
    // ... 多个 persist/remove 操作
    $em->flush();
    $em->getConnection()->commit();
} catch (\Exception $e) {
    $em->getConnection()->rollBack();
    throw $e;
}
```

#### 3.3 延迟加载 (Lazy Loading) vs 即时加载 (Eager Loading)
*   **Lazy (默认)**: 访问关联对象时才查询数据库。适合大多数场景，但需警惕 N+1。
*   **Eager**: 查询主实体时立即加载关联对象。配置：`#[ORM\ManyToOne(fetch: 'EAGER')]`。适合关联数据量小且必用的场景。

#### 3.4 不要在新建实体时设置 ID
除非特殊需求，否则让数据库自增 ID。Doctrine 会在 `flush` 后将生成的 ID 回填到对象中。

#### 3.5 清理 EntityManager
在长运行脚本（如 CLI 命令、队列消费者）中，处理大量数据后应调用 `$em->clear()` 释放内存，防止内存泄漏。

---

### 4. 总结：优缺点

**优点:**
*   **开发效率高**: 面向对象操作，无需手写大量 SQL。
*   **数据库无关性**: 轻松切换 MySQL, PostgreSQL, SQLite 等。
*   **功能丰富**: 支持复杂关系、缓存、事件监听、生命周期回调。
*   **生态成熟**: 与 Symfony 深度集成，社区活跃。

**缺点:**
*   **学习曲线**: 概念较多（实体状态、关联维护等）。
*   **性能开销**: 相比原生 SQL，ORM 层有性能损耗（可通过优化查询缓解）。
*   **复杂查询**: 极度复杂的报表查询有时用原生 SQL 或 DQL 更直观。

**一句话总结**:
Doctrine ORM 是 PHP 中处理复杂业务逻辑和领域模型的首选工具，掌握其**实体定义、EntityManager 操作、关联映射**以及**避免 N+1 问题**是高效使用的关键。
