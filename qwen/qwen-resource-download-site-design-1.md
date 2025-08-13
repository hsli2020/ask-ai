# 我想开发一个资源(如电子书、视频等)下载网站，资源是必须付费才能下载的，用户付费后得到一个有时效性的下载链接或者提取码，来到我的网站下载，请给我讲解如何开发这样的网站，有哪些细节需要考虑，并列出一些开源的系统可供参考。

开发一个付费资源下载网站（通常称为“知识付费”或“资源站”）是一个典型的电商 + 内容交付系统。虽然逻辑看似简单（付钱 -> 给链接），但在**安全性、支付合规、文件保护、高并发**等方面有很多细节需要处理。

以下是一份完整的开发指南，涵盖架构设计、关键细节、法律风险及开源推荐。

---

### 一、核心业务流程设计

在写代码之前，必须理清业务闭环：

1.  **用户端：** 注册/登录 -> 浏览资源 -> 加入购物车/直接购买 -> 支付 -> **获取下载凭证（链接/提取码）** -> 下载资源。
2.  **管理端：** 上传资源 -> 设置价格/库存 -> 查看订单 -> 财务统计 -> 用户管理。
3.  **系统后台：** 监听支付回调 -> 验证支付状态 -> **生成有时效性的下载凭证** -> 记录下载日志。

---

### 二、技术架构与关键模块

建议采用前后端分离的架构，便于扩展和维护。

#### 1. 前端 (Frontend)
*   **技术栈：** Vue.js / React / Next.js (推荐 Next.js，利于 SEO)。
*   **关键点：** 资源详情页需要被搜索引擎收录（SEO），所以 SSR（服务端渲染）很重要。

#### 2. 后端 (Backend)
*   **技术栈：** Node.js (NestJS), Python (Django/FastAPI), Go, 或 PHP (Laravel)。
*   **关键点：** 事务处理（保证支付和发链接的原子性）、接口鉴权。

#### 3. 数据库 (Database)
*   **选型：** MySQL 或 PostgreSQL。
*   **关键表：** 用户表、资源表、订单表、下载记录表、支付流水表。

#### 4. 文件存储 (Storage) —— **最重要的一环**
*   **不要存在 Web 服务器上！** 否则带宽会瞬间打满，且文件容易泄露。
*   **方案：** 使用对象存储 (Object Storage)。
    *   **国内：** 阿里云 OSS、腾讯云 COS、七牛云。
    *   **国外：** AWS S3、Cloudflare R2 (免流量费，强烈推荐)。
*   **加速：** 必须配合 CDN，否则大文件下载体验极差。

#### 5. 支付系统 (Payment)
*   **国内：** 支付宝/微信支付官方接口（需要营业执照）、第四方聚合支付（如虎皮椒、Payjs，适合个人，但有风险）。
*   **国外：** Stripe, PayPal。
*   **逻辑：** 必须处理**异步回调 (Webhook)**，不能只依赖前端跳转。

---

### 三、核心细节与难点解决方案

#### 1. 下载链接的时效性与安全性
这是防止资源被倒卖的核心。
*   **签名 URL (Signed URL)：** 对象存储（如 S3/OSS）都支持生成带签名的临时链接。
    *   *原理：* 链接中包含 `Expires` 时间戳和 `Signature` 签名。
    *   *效果：* 链接在 1 小时后失效，且无法被篡改。
    *   *限制：* 可以限制 IP 或 Referer，防止用户把链接发给别人。
*   **动态提取码：** 如果资源放在百度网盘等第三方，需生成一次性或限时有效的提取码（需对接第三方 API 或人工录入）。
*   **下载次数限制：** 数据库记录该订单已下载次数，超过 N 次链接失效。

#### 2. 文件防盗与版权保护
一旦用户下载了文件，你就失去了控制权。只能增加倒卖成本：
*   **数字水印：** 对 PDF 添加购买者 ID/邮箱的水印；对视频添加动态跑马灯水印。
*   **加密压缩：** 发送加密压缩包，密码通过站内信发送（增加一道门槛）。
*   **在线预览：** 提供部分内容的在线预览，吸引购买，但禁止右键下载。

#### 3. 支付安全与对账
*   **幂等性：** 支付回调可能触发多次，后端必须保证同一订单号只处理一次成功逻辑。
*   **金额校验：** 后端校验回调中的金额是否与订单一致，防止篡改。
*   **掉单处理：** 用户付了钱但没收到链接。需要有一个“查询订单状态”或“手动补发”的功能。

#### 4. 高并发与性能
*   **热点资源：** 如果某个资源突然爆火，数据库查询压力大，使用 Redis 缓存资源信息。
*   **下载带宽：** 对象存储 + CDN 是必须的，不要用自己的服务器带宽扛下载流量。

---

### 四、法律与合规风险 (非常重要)

在中国大陆运营此类网站，风险极高，请务必注意：

1.  **ICP 备案与许可证：** 服务器在国内必须备案。如果涉及收费，理论上需要《增值电信业务经营许可证》(ICP 证)，个人很难办理，通常以“个人博客捐赠”或“技术服务费”名义规避，但有风险。
2.  **版权 (Copyright)：**
    *   **自有资源：**  safest。
    *   **搬运资源：** 极高风险。如果下载的是受版权保护的书籍、课程、软件，会被起诉甚至承担刑事责任（侵犯著作权罪）。
    *   **建议：** 只做原创，或获得授权的分销（如分销虚拟课程）。
3.  **内容安全：** 严禁色情、政治、赌博等违规内容。需要接入内容安全 API 进行自动审核。
4.  **支付合规：** 个人收款码用于商业经营容易被微信/支付宝风控冻结。建议使用有执照的企业账户或合规的聚合支付。

---

### 五、开源系统推荐

不要重复造轮子，以下系统可以作为基础进行二次开发：

#### 1. WordPress + 插件 (最推荐，成本低，生态好)
WordPress 本身是博客，但配合插件可变成资源站。
*   **核心插件：** WooCommerce (电商功能)。
*   **付费下载主题/插件：**
    *   **Zibll (子比主题)：** 国内非常流行，专为资源站设计，支持付费下载、会员制、微信支付/支付宝对接。
    *   **Erphp-Download：** 老牌的 WordPress 付费下载插件。
    *   **MemberPress：** 强大的会员付费内容限制插件。
*   **优点：** 开发快，插件多，SEO 好。
*   **缺点：** 性能相对较重，需优化。

#### 2. Cloudreve (专注于文件管理)
*   **简介：** 基于 Go 开发的网盘系统，支持多种存储策略（本地、S3、WebDAV 等）。
*   **付费功能：** 原生主要做网盘，但可以通过插件或二次开发实现“付费下载组”。
*   **适用：** 如果你想做一个类似“网盘”体验的资源站。

#### 3. AList (文件列表程序)
*   **简介：** 支持挂载多种网盘（阿里云盘、百度网盘等）并生成下载链接。
*   **付费方案：** AList 本身免费，但你可以配合一个独立的支付系统（如 ZCart），支付成功后调用 AList API 生成临时链接。
*   **适用：** 资源主要托管在第三方网盘，不想自己买存储空间的场景。

#### 4. 基于 Laravel/ThinkPHP 的源码 (适合二次开发)
国内有很多基于 PHP 框架开发的资源站源码（通常称为“资源站系统”）：
*   **BageCMS (八哥 CMS)：** 适合做资源下载。
*   **Likeshop / CRMEB：** 虽然是电商系统，但支持虚拟商品，逻辑通用。
*   **注意：** 很多此类源码在 GitHub 或码云上，但质量参差不齐，需注意后门风险。

#### 5. 自研参考架构 (GitHub 项目)
如果你想自己写，可以参考以下类型的开源项目：
*   **Saleor / Medusa:** 头部开源电商系统（Node.js/Python），适合学习订单和支付逻辑。
*   **Halo:** 优秀的 Java 博客系统，插件机制完善，可参考其插件开发模式。

---

### 六、开发路线图建议

如果你决定开始，建议按以下步骤：

1.  **MVP 阶段 (最小可行性产品)：**
    *   使用 **WordPress + 子比主题**。
    *   配置好阿里云 OSS 存储。
    *   对接好支付接口（先测试沙箱环境）。
    *   上传几个测试资源，跑通“支付 - 下载”流程。
    *   *耗时：1-2 周。*

2.  **优化阶段：**
    *   如果 WordPress 性能不够，考虑使用 **Laravel + Vue** 自研。
    *   增加**水印功能**（PDF/视频）。
    *   增加**会员体系**（月费/年费无限下载）。
    *   增加**推广员系统**（用户邀请他人购买得佣金）。

3.  **运营与风控：**
    *   部署 WAF (Web 应用防火墙) 防攻击。
    *   设置下载频率限制（防爬虫）。
    *   准备法务条款（用户协议、免责声明）。

### 七、总结建议

1.  **技术不是最难，合规最难。** 务必确认资源的版权来源。
2.  **存储分离。** 永远不要把资源文件放在 Web 服务器的本地磁盘。
3.  **先买后下。** 所有的下载链接生成逻辑必须在**支付成功回调之后**触发。
4.  **推荐起步方案：** 购买一个成熟的 WordPress 资源主题（如子比），配合 Cloudflare R2 (存储) + 虎皮椒/易支付 (支付)，这是成本最低、上线最快的方式。验证商业模式成功后，再考虑自研系统。

如果你需要具体的代码片段（如生成 S3 签名 URL 的代码）或特定语言的架构建议，可以继续提问。


# 我考虑使用 Laravel + Vue 自研的线路，应该开发哪些模块，创建哪些数据库表，每个表的结构。

选择 **Laravel + Vue** 是自研资源站的绝佳组合。Laravel 拥有强大的生态（支付、队列、存储、权限），Vue 则能提供流畅的用户交互体验。

以下是针对“付费下载 + 时效性链接”场景的详细开发规划，包括核心模块划分和数据库设计。

---

### 一、核心功能模块划分

你需要开发以下 6 个核心模块：

#### 1. 用户与认证模块 (Auth & User)
*   **功能**：注册、登录、忘记密码、个人中心。
*   **扩展**：用户等级（普通/ VIP）、余额管理、积分系统。
*   **Laravel 特性**：使用 `Laravel Breeze` 或 `Jetstream` 快速搭建基础 Auth。

#### 2. 资源管理模块 (Resource Management)
*   **功能**：
    *   后台上传文件（自动转存至 OSS/S3）。
    *   设置资源属性：标题、描述、封面图、价格、文件大小、格式。
    *   **关键逻辑**：文件不存本地，只存 OSS 的 `key` 或 `path`。
    *   状态管理：上架、下架、审核中。

#### 3. 订单与支付模块 (Order & Payment)
*   **功能**：
    *   创建订单（锁定库存/资源）。
    *   对接支付网关（支付宝/微信/Stripe）。
    *   **核心逻辑**：监听支付回调 (Webhook)，验证签名，更新订单状态为“已支付”。
    *   处理“掉单”（用户付了钱但没回调）的查询接口。

#### 4. 交付与鉴权模块 (Delivery & Security) —— **最核心**
*   **功能**：
    *   支付成功后，生成**带签名的临时下载链接** (Signed URL)。
    *   记录用户的购买记录。
    *   **限制逻辑**：
        *   链接有效期（如 2 小时）。
        *   下载次数限制（如仅限 3 次）。
        *   IP 绑定（可选，防止分享链接）。
    *   提供“我的订单/下载记录”页面，用户可在此重新获取未过期的链接。

#### 5. 统计与日志模块 (Analytics & Logs)
*   **功能**：
    *   销售统计（日/月营收）。
    *   下载日志（谁、在什么时间、下载了什么、IP 地址）。
    *   异常监控（支付失败、生成链接失败）。

#### 6. 后台管理系统 (Admin Dashboard)
*   **功能**：资源 CRUD、订单管理、用户管理、财务对账、系统配置（OSS 密钥、支付密钥）。
*   **技术建议**：可以使用 `Laravel Nova` (收费但强大) 或 `Filament PHP` (免费且流行，强烈推荐) 快速构建后台。

---

### 二、数据库表结构设计

以下是基于 MySQL 的核心表结构建议（使用 Laravel Migration 风格描述）：

#### 1. 用户表 (`users`)
*Laravel 默认表，需扩展字段*
```php
Schema::create('users', function (Blueprint $table) {
    $table->id();
    $table->string('name');
    $table->string('email')->unique();
    $table->string('password');
    $table->decimal('balance', 10, 2)->default(0); // 账户余额
    $table->string('role')->default('user'); // user, admin, vip
    $table->timestamp('vip_expire_at')->nullable(); // VIP 过期时间
    $table->rememberToken();
    $table->timestamps();
});
```

#### 2. 资源表 (`resources`)
*存储资源元数据，不存文件实体*
```php
Schema::create('resources', function (Blueprint $table) {
    $table->id();
    $table->unsignedBigInteger('category_id')->nullable(); // 分类 ID
    $table->string('title'); // 标题
    $table->text('description'); // 描述
    $table->string('cover_image')->nullable(); // 封面图 URL
    $table->string('file_path'); // OSS 上的 Key 或路径 (例如：books/book_01.pdf)
    $table->string('file_name_original'); // 原始文件名 (用于下载时重命名)
    $table->bigInteger('file_size'); // 文件大小 (字节)
    $table->decimal('price', 10, 2); // 价格
    $table->boolean('is_vip_free')->default(false); // VIP 是否免费
    $table->integer('download_limit_count')->default(0); // 0 表示无限，>0 表示限制次数
    $table->integer('link_expire_minutes')->default(120); // 链接默认有效时长 (分钟)
    $table->enum('status', ['pending', 'active', 'offline'])->default('pending'); // 状态
    $table->json('meta_data')->nullable(); // 额外信息 (如页数、时长等)
    $table->timestamps();
    
    $table->index('status');
});
```

#### 3. 订单表 (`orders`)
*核心交易表*
```php
Schema::create('orders', function (Blueprint $table) {
    $table->id();
    $table->string('order_no')->unique(); // 业务订单号 (生成规则：年月日 + 随机数)
    $table->foreignId('user_id')->constrained();
    $table->foreignId('resource_id')->constrained();
    $table->decimal('amount', 10, 2); // 实际支付金额
    $table->enum('status', ['pending', 'paid', 'failed', 'refunded'])->default('pending');
    $table->string('payment_method')->nullable(); // alipay, wechat, stripe
    $table->string('transaction_id')->nullable(); // 第三方支付流水号
    $table->text('callback_data')->nullable(); // 存储支付回调的原始数据，便于排查
    $table->timestamp('paid_at')->nullable(); // 支付成功时间
    $table->timestamps();
    
    $table->index('order_no');
    $table->index(['user_id', 'status']);
});
```

#### 4. 下载凭证/记录表 (`download_records`)
*这是实现“时效性”和“防倒卖”的关键表*
```php
Schema::create('download_records', function (Blueprint $table) {
    $table->id();
    $table->foreignId('user_id')->constrained();
    $table->foreignId('order_id')->constrained(); // 关联哪个订单
    $table->foreignId('resource_id')->constrained();
    
    // 核心安全字段
    $table->string('token')->unique(); // 随机生成的唯一令牌，用于构建下载链接 /download/{token}
    $table->timestamp('expires_at'); // 链接过期时间
    $table->integer('max_downloads'); // 允许最大下载次数
    $table->integer('current_downloads')->default(0); // 当前已下载次数
    $table->string('last_download_ip')->nullable(); // 最后一次下载 IP (用于风控)
    
    $table->boolean('is_used_up')->default(false); // 是否已用完 (次数耗尽或过期)
    $table->timestamps();
    
    $table->index('token');
    $table->index('expires_at');
});
```
*注意：也可以不在数据库存 Signed URL，而是每次用户请求下载时，后端实时根据 `order` 和 `resource` 动态生成 S3/OSS 的 Signed URL。但为了控制“总下载次数”，必须在数据库记录 `download_records`。*

#### 5. 操作日志表 (`operation_logs`) - 可选但推荐
```php
Schema::create('operation_logs', function (Blueprint $table) {
    $table->id();
    $table->foreignId('user_id')->nullable();
    $table->string('action'); // e.g., 'download_attempt', 'payment_success'
    $table->string('ip_address');
    $table->string('user_agent');
    $table->json('context')->nullable(); // 上下文信息，如 resource_id
    $table->timestamps();
});
```

---

### 三、关键业务逻辑实现思路

#### 1. 支付成功后的流程 (The Happy Path)
当收到支付网关的 `Success` 回调时：
1.  **验证签名**：确保请求来自支付宝/微信。
2.  **检查订单状态**：防止重复处理。
3.  **事务处理 (DB Transaction)**：
    *   更新 `orders` 表状态为 `paid`。
    *   插入 `download_records` 表：
        *   生成一个高强度的随机 `token` (如 `Str::random(32)`)。
        *   计算 `expires_at` = `now()` + `resource.link_expire_minutes`。
        *   设置 `max_downloads` (从资源表读取或默认值)。
4.  **通知用户**：发送邮件或站内信，告知“购买成功，前往个人中心下载”。

#### 2. 下载链接的生成与验证
用户点击“下载”按钮时：

**方案 A：前端请求后端，后端重定向 (推荐)**
1.  用户访问 `/api/download/{record_token}`。
2.  后端中间件/控制器执行：
    *   查找 `download_records` 表中该 token。
    *   **校验 1**：记录是否存在？
    *   **校验 2**：`expires_at` 是否已过？ -> 若过期，标记 `is_used_up` 并返回错误。
    *   **校验 3**：`current_downloads` >= `max_downloads`？ -> 若超限，返回错误。
    *   **校验 4** (可选)：IP 是否变动过大？
3.  **通过校验后**：
    *   `current_downloads` + 1 (原子操作 `increment`)。
    *   记录 `last_download_ip`。
    *   **调用 OSS/S3 SDK** 生成一个临时的 Signed URL (例如有效期 5 分钟)。
    *   返回 JSON: `{ "url": "https://oss...?signature=..." }` 或直接 `redirect()->to($signedUrl)`。

**为什么这样做？**
*   真正的文件链接 (OSS Signed URL) 只有几分钟有效期，即使被截获也很快失效。
*   主要的权限控制（总次数、总时效）由你的数据库 `download_records` 掌控。
*   用户可以多次进入网站点击“获取下载链接”，只要没超总次数，就能拿到新的临时 OSS 链接。

#### 3. 定时任务 (Scheduler)
在 `app/Console/Kernel.php` 中添加定时任务，清理过期数据或标记过期记录：
```php
// 每天运行一次，标记过期的记录
$schedule->call(function () {
    DB::table('download_records')
      ->where('expires_at', '<', now())
      ->where('is_used_up', false)
      ->update(['is_used_up' => true]);
})->daily();
```

---

### 四、Laravel + Vue 开发特别提示

1.  **API 设计**：
    *   使用 `Laravel Sanctum` 进行 API 认证（Vue 通过 Token 访问）。
    *   资源列表接口需要支持分页、搜索、筛选（分类、价格区间）。

2.  **文件上传**：
    *   **不要**让 Vue 直接传大文件到 Laravel 服务器。
    *   **最佳实践**：Vue 请求后端获取 OSS 的 `Pre-signed Put URL` -> Vue 直接上传文件到 OSS -> 上传完成后，Vue 通知后端保存 `file_path` 到数据库。这样节省服务器带宽和 CPU。

3.  **队列 (Queue)**：
    *   支付回调处理、发送通知邮件、生成水印（如果需要）等操作，务必放入 Laravel Queue (Redis driver)，避免阻塞主线程导致用户等待。

4.  **安全性**：
    *   **CSRF/XSS**：Laravel 默认防护，Vue 需注意 `v-html` 的使用。
    *   **SQL 注入**：使用 Eloquent ORM 即可避免。
    *   **接口限流**：在 `Route::middleware('throttle:60,1')` 中限制下载接口的请求频率，防止暴力破解 token。

### 五、下一步行动清单

1.  **环境搭建**：安装 Laravel 10/11, Vue 3 (Vite), MySQL, Redis。
2.  **迁移数据库**：根据上述结构编写 Migration 文件。
3.  **模型定义**：创建 `Resource`, `Order`, `DownloadRecord` Model 并建立关联关系 (`hasMany`, `belongsTo`)。
4.  **OSS 配置**：安装 `league/flysystem-aws-s3-v3` 或阿里云 OSS SDK，配置 `.env`。
5.  **核心接口开发**：
    *   `POST /api/orders` (创建订单)
    *   `POST /api/payments/notify` (支付回调)
    *   `GET /api/download/{token}` (获取下载链接)
6.  **前端页面**：
    *   资源详情页 (展示 + 购买按钮)。
    *   个人中心 (订单列表 + 下载按钮)。

这个架构既保证了安全性（通过数据库控制权限 + OSS 临时链接），又具有良好的扩展性。如果你需要具体的 Controller 代码示例或 Migration 文件代码，请告诉我！


# 请提供下面几个API的Controller代码示例
POST /api/orders (创建订单)
POST /api/payments/notify (支付回调)
GET /api/download/{token} (获取下载链接)


以下是基于 **Laravel 10/11** 的三个核心 Controller 代码示例。

### 前置准备

假设你已经安装了以下依赖：
*   `laravel/sanctum` (用于 API 认证)
*   AWS S3 SDK 或 阿里云 OSS SDK (此处以 Laravel 原生支持的 S3 协议为例，阿里云/腾讯云配置类似)
*   数据库迁移已完成 (参考上一轮的表结构)

为了代码清晰，我将逻辑集中在 Controller 中，实际项目中建议将“创建订单逻辑”、“支付验证逻辑”提取到 Service 层。

---

### 1. `OrderController.php` - 创建订单

**功能**：验证用户、检查资源状态、创建待支付订单、返回支付参数。

```php
<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Order;
use App\Models\Resource;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Str;
use Symfony\Component\HttpKernel\Exception\HttpException;

class OrderController extends Controller
{
    /**
     * 创建订单
     * POST /api/orders
     */
    public function store(Request $request)
    {
        $request->validate([
            'resource_id' => 'required|integer|exists:resources,id',
            'payment_method' => 'required|in:alipay,wechat,stripe', // 根据实际支持的渠道调整
        ]);

        $user = Auth::user();
        $resource = Resource::findOrFail($request->resource_id);

        // 1. 基础校验
        if ($resource->status !== 'active') {
            throw new HttpException(400, '该资源已下架或不可购买');
        }

        // 2. 检查是否已购买过 (可选策略：允许重复购买或提示已拥有)
        $existingOrder = Order::where('user_id', $user->id)
            ->where('resource_id', $resource->id)
            ->where('status', 'paid')
            ->first();

        if ($existingOrder) {
            // 如果已购买，直接返回已有的下载记录或提示
            return response()->json([
                'message' => '您已购买过该资源',
                'order_no' => $existingOrder->order_no,
                'redirect_to' => '/user/downloads' 
            ], 200);
        }

        // 3. 生成订单号 (规则：年月日 + 随机字符串)
        $orderNo = date('Ymd') . strtoupper(Str::random(8));

        // 4. 事务创建订单
        $order = DB::transaction(function () use ($user, $resource, $orderNo, $request) {
            return Order::create([
                'order_no' => $orderNo,
                'user_id' => $user->id,
                'resource_id' => $resource->id,
                'amount' => $resource->price, // 这里可以加入优惠券逻辑修改最终金额
                'status' => 'pending',
                'payment_method' => $request->payment_method,
            ]);
        });

        // 5. 调用支付服务获取支付参数 (伪代码，需对接具体支付 SDK)
        // $paymentService = app('PaymentService');
        // $payParams = $paymentService->createOrder($order, $request->payment_method);
        
        // 模拟返回支付参数
        $payParams = [
            'order_no' => $orderNo,
            'pay_url' => 'https://mock-payment-gateway.com/pay/' . $orderNo, // 实际应为支付宝/微信返回的串或二维码链接
            'expire_time' => now()->addMinutes(15)->toDateTimeString(),
        ];

        return response()->json([
            'success' => true,
            'data' => [
                'order' => $order,
                'payment' => $payParams
            ]
        ]);
    }
}
```

---

### 2. `PaymentCallbackController.php` - 支付回调

**功能**：接收第三方支付通知、验签、更新订单状态、**生成下载凭证**。
*注意：此接口通常不需要 Sanctum 认证，但需要严格的签名验证。*

```php
<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\Order;
use App\Models\DownloadRecord;
use App\Models\Resource;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class PaymentCallbackController extends Controller
{
    /**
     * 支付回调通知
     * POST /api/payments/notify
     * 注意：此路由通常在 web.php 中定义为未授权路由，或在 api.php 中关闭 auth 中间件
     */
    public function notify(Request $request)
    {
        // 1. 获取回调数据 (不同支付平台字段不同，此处以通用逻辑演示)
        $data = $request->all();
        $paymentMethod = $request->input('channel'); // alipay, wechat, etc.
        $outTradeNo = $request->input('out_trade_no'); // 我们的订单号
        $tradeStatus = $request->input('trade_status'); // SUCCESS, TRADE_SUCCESS 等
        $transactionId = $request->input('transaction_id'); // 第三方流水号
        $sign = $request->input('sign');

        // 2. 【至关重要】验证签名
        // 必须使用支付平台提供的公钥/密钥验证 $sign，防止伪造请求
        // if (!$this->verifySignature($data, $sign, $paymentMethod)) {
        //     Log::warning('Invalid payment signature', ['data' => $data]);
        //     return response('FAIL', 403);
        // }

        // 模拟验证通过
        if ($tradeStatus !== 'SUCCESS' && $tradeStatus !== 'TRADE_SUCCESS') {
            return response('FAIL', 400);
        }

        // 3. 查找订单
        $order = Order::where('order_no', $outTradeNo)->first();
        if (!$order) {
            Log::error('Order not found', ['order_no' => $outTradeNo]);
            return response('FAIL', 404);
        }

        // 4. 幂等性检查：如果订单已经是 paid 状态，直接返回成功，避免重复处理
        if ($order->status === 'paid') {
            return response('SUCCESS', 200);
        }

        // 5. 金额校验 (防止篡改金额)
        // if (bccomp($order->amount, $request->input('total_amount'), 2) !== 0) { ... }

        // 6. 事务处理：更新订单 + 生成下载凭证
        try {
            DB::transaction(function () use ($order, $data, $transactionId) {
                // A. 更新订单状态
                $order->update([
                    'status' => 'paid',
                    'transaction_id' => $transactionId,
                    'paid_at' => now(),
                    'callback_data' => json_encode($data), // 留存原始日志
                ]);

                // B. 获取资源信息以配置下载限制
                $resource = Resource::find($order->resource_id);
                
                // C. 生成下载记录 (Download Record)
                DownloadRecord::create([
                    'user_id' => $order->user_id,
                    'order_id' => $order->id,
                    'resource_id' => $resource->id,
                    'token' => bin2hex(random_bytes(16)), // 生成 32 位随机 Token
                    'expires_at' => now()->addMinutes($resource->link_expire_minutes),
                    'max_downloads' => $resource->download_limit_count > 0 ? $resource->download_limit_count : 999999,
                    'current_downloads' => 0,
                    'is_used_up' => false,
                ]);
            });

            Log::info("Payment success and record created", ['order_no' => $outTradeNo]);
            
            // 支付平台通常期望收到纯文本 "SUCCESS" 或 "OK"
            return response('SUCCESS', 200);

        } catch (\Exception $e) {
            Log::error("Payment callback processing failed", ['error' => $e->getMessage()]);
            // 返回失败会让支付平台重试，确保数据一致性
            return response('FAIL', 500);
        }
    }
    
    // 伪代码：签名验证逻辑
    private function verifySignature(array $data, string $sign, string $method): bool {
        // 此处需接入支付宝/微信官方 SDK 的 verify 方法
        return true; 
    }
}
```

---

### 3. `DownloadController.php` - 获取下载链接

**功能**：验证 Token、检查次数/有效期、原子增加下载次数、生成 OSS 临时签名 URL。

```php
<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Models\DownloadRecord;
use App\Models\Resource;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Facades\DB;
use Symfony\Component\HttpKernel\Exception\HttpException;

class DownloadController extends Controller
{
    /**
     * 获取真实下载链接
     * GET /api/download/{token}
     * 注意：此接口建议加上速率限制 middleware('throttle:10,1')
     */
    public function show(Request $request, string $token)
    {
        // 1. 查找记录
        $record = DownloadRecord::where('token', $token)->firstOrFail();

        // 2. 校验是否已用完或被手动禁用
        if ($record->is_used_up) {
            throw new HttpException(410, '下载链接已失效或次数用尽');
        }

        // 3. 校验过期时间
        if ($record->expires_at < now()) {
            // 标记为用完
            $record->update(['is_used_up' => true]);
            throw new HttpException(410, '下载链接已过期');
        }

        // 4. 校验剩余次数 (使用数据库原子操作防止并发超卖)
        // 先检查当前值
        if ($record->current_downloads >= $record->max_downloads) {
            $record->update(['is_used_up' => true]);
            throw new HttpException(429, '下载次数已达上限');
        }

        // 5. 原子增加下载次数并获取最新状态
        // increment 返回 true 表示执行成功，但在高并发下最好结合 where 条件更新
        $updated = DB::table('download_records')
            ->where('id', $record->id)
            ->where('current_downloads', '<', $record->max_downloads)
            ->increment('current_downloads');

        if (!$updated) {
            // 说明并发冲突导致次数超限，或者记录被其他请求修改了
            $record->refresh(); // 刷新模型数据
            if ($record->current_downloads >= $record->max_downloads) {
                $record->update(['is_used_up' => true]);
                throw new HttpException(429, '下载次数已达上限 (并发限制)');
            }
        }
        
        // 更新本地模型的内存值
        $record->current_downloads++; 
        
        // 可选：记录最后下载 IP
        $record->update([
            'last_download_ip' => $request->ip()
        ]);

        // 6. 获取资源信息
        $resource = Resource::find($record->resource_id);
        if (!$resource || $resource->status !== 'active') {
            throw new HttpException(404, '资源不存在或已下架');
        }

        // 7. 生成对象存储 (OSS/S3) 的临时签名 URL
        // 假设配置文件 filesystems.php 中 disk 名为 'oss' 或 's3'
        $disk = Storage::disk('oss'); // 或 's3'
        
        // 生成一个仅有效 5 分钟的临时链接，即使 token 没过期，这个直连链接也会很快过期，增加安全性
        $temporaryUrl = $disk->temporaryUrl(
            $resource->file_path,
            now()->addMinutes(5),
            [
                'ResponseContentDisposition' => 'attachment; filename="' . $resource->file_name_original . '"',
                // 可选：限制 IP (如果存储商支持)
                // 'IpAddress' => $request->ip() 
            ]
        );

        return response()->json([
            'success' => true,
            'data' => [
                'download_url' => $temporaryUrl,
                'filename' => $resource->file_name_original,
                'remaining_downloads' => $record->max_downloads - $record->current_downloads,
                'expires_at' => $record->expires_at->toIso8601String()
            ]
        ]);
    }
}
```

---

### 四、路由配置 (`routes/api.php`)

```php
use App\Http\Controllers\Api\OrderController;
use App\Http\Controllers\Api\PaymentCallbackController;
use App\Http\Controllers\Api\DownloadController;
use Illuminate\Support\Facades\Route;

// 需要登录的接口
Route::middleware('auth:sanctum')->group(function () {
    Route::post('/orders', [OrderController::class, 'store']);
    // 下载接口也可以放在这里，或者像上面代码那样通过 Token 匿名访问但内部校验权限
    // 这里推荐通过 Token 匿名访问，方便分享和直接点击下载，但 Token 本身具备权限属性
});

// 支付回调接口 (通常不需要 auth 中间件，由支付平台调用)
// 注意：生产环境务必在此处做 IP 白名单限制或严格的签名验证
Route::post('/payments/notify', [PaymentCallbackController::class, 'notify']);

// 下载接口 (通过 Token 鉴权，无需用户登录态，方便用户复制链接到下载器)
// 建议加上 throttle 限流
Route::get('/download/{token}', [DownloadController::class, 'show'])->middleware('throttle:10,1');
```

### 五、关键细节补充

1.  **Storage 配置**：
    确保 `.env` 中配置了正确的 OSS/S3 信息。
    ```env
    FILESYSTEM_DISK=oss
    OSS_KEY=your_key
    OSS_SECRET=your_secret
    OSS_BUCKET=your_bucket
    OSS_URL=https://your-bucket.oss-region.aliyuncs.com
    ```
    并在 `config/filesystems.php` 中配置好 `oss` disk。

2.  **并发安全**：
    在 `DownloadController` 中，使用了 `DB::table(...)->where(...)->increment()`。这是防止“超卖”（即下载次数超过限制）的关键。如果不加 `where` 条件直接 `$record->increment()`，在高并发下可能会读出旧值导致计数错误。

3.  **安全性增强**：
    *   **IP 绑定**：可以在生成 `DownloadRecord` 时记录用户 IP，并在下载时比对。如果变化过大（如跨省），可拒绝下载。
    *   **User-Agent 指纹**：类似 IP，可记录首次请求的 UA。
    *   **水印**：如果是 PDF/图片，建议在生成临时链接前，后端先流式处理添加动态水印（会消耗 CPU，建议异步队列处理或前端 Canvas 水印）。

4.  **调试技巧**：
    *   支付回调在本地很难测试（因为支付宝/微信无法访问 localhost）。建议使用 **Ngrok** 或 **LocalTunnel** 将本地端口映射到公网 URL 进行测试。
    *   务必查看 `storage/logs/laravel.log` 监控回调失败的原因。

这套代码提供了一个安全、可扩展的基础框架，你可以在此基础上集成具体的支付 SDK（如 `yansongda/pay` 包）和完善前端交互。
