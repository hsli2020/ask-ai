# 我想给我的php+mysql开发的网站加上affiliate系统，要考虑哪些问题，请帮我设计实现

给现有的 PHP+MySQL 网站添加 Affiliate（联盟营销/分销）系统是一个能显著增长流量的功能，但同时也涉及复杂的逻辑、资金安全和反作弊。

以下我将从**核心考量点**、**数据库设计**、**核心逻辑实现**、**安全与反作弊**、以及**运营建议**五个方面为你提供完整的设计方案。

---

### 一、核心考量点 (Before You Code)

在写代码之前，必须明确业务规则，否则后期重构成本极高：

1.  **归因模式 (Attribution Model)**
    *   **Last Click (最后点击):** 用户购买前最后一次点击的推广链接有效（最常用）。
    *   **First Click (首次点击):** 用户第一次点击的链接有效。
    *   **Cookie 有效期:** 点击后多长时间内购买算有效？（通常 30-90 天）。
2.  **佣金计算 (Commission)**
    *   **固定金额:** 每单奖励 $10。
    *   **百分比:** 订单金额的 10%。
    *   **层级:** 一级分销（直接推荐）还是二级分销（推荐人的推荐人）？*建议初期只做一级，避免涉嫌传销风险。*
    *   **结算周期:** 用户购买后立即结算，还是过了退款期（如 15 天）后结算？*强烈建议设置“待结算”状态。*
3.  **身份界定**
    *   普通用户是否可以申请成为 Affiliate？
    *   管理员是否可以成为 Affiliate？（通常禁止，防止作弊）。
    *   用户能否推荐自己？（必须禁止）。
4.  **提现方式**
    *   最低提现金额（Threshold）。
    *   支付方式（PayPal, 银行转账，支付宝，站内余额抵扣）。

---

### 二、数据库设计 (Database Schema)

假设你已有一个 `users` 表。我们需要扩展几张表来支撑联盟系统。

```sql
-- 1. 扩展用户表 (记录谁推荐了谁)
ALTER TABLE users ADD COLUMN referrer_id INT DEFAULT NULL; 
ALTER TABLE users ADD COLUMN affiliate_code VARCHAR(20) UNIQUE; -- 用户的专属推广码

-- 2. 联盟账户表 (存储佣金余额、提现信息)
CREATE TABLE affiliate_accounts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    user_id INT NOT NULL,
    balance DECIMAL(10, 2) DEFAULT 0.00, -- 可用余额
    pending_balance DECIMAL(10, 2) DEFAULT 0.00, -- 待结算余额
    payout_method VARCHAR(50), -- 提现方式 (paypal, bank)
    payout_account VARCHAR(100), -- 提现账号
    status TINYINT DEFAULT 1, -- 1 正常，0 禁用
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

-- 3. 点击追踪表 (用于统计流量，反作弊分析)
CREATE TABLE affiliate_clicks (
    id INT PRIMARY KEY AUTO_INCREMENT,
    affiliate_id INT NOT NULL, -- 推广者 ID
    visitor_ip VARCHAR(45),
    user_agent TEXT,
    landing_page VARCHAR(255),
    clicked_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX (affiliate_id),
    INDEX (clicked_at)
);

-- 4. 佣金记录表 (核心财务表)
CREATE TABLE affiliate_commissions (
    id INT PRIMARY KEY AUTO_INCREMENT,
    affiliate_id INT NOT NULL,
    order_id INT NOT NULL, -- 关联订单 ID
    referred_user_id INT NOT NULL, -- 被推荐购买的用户 ID
    amount DECIMAL(10, 2) NOT NULL, -- 订单金额
    commission DECIMAL(10, 2) NOT NULL, -- 佣金金额
    status TINYINT DEFAULT 0, -- 0 待结算 (Pending), 1 已批准 (Approved), 2 已支付 (Paid), 3 已取消 (Cancelled/Refund)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,
    FOREIGN KEY (affiliate_id) REFERENCES affiliate_accounts(user_id)
);

-- 5. 提现记录表
CREATE TABLE affiliate_payouts (
    id INT PRIMARY KEY AUTO_INCREMENT,
    affiliate_id INT NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    status TINYINT DEFAULT 0, -- 0 申请中，1 已打款，2 拒绝
    request_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_date TIMESTAMP NULL,
    admin_note TEXT,
    FOREIGN KEY (affiliate_id) REFERENCES affiliate_accounts(user_id)
);
```

---

### 三、核心逻辑实现 (PHP Implementation)

#### 1. 推广链接生成与追踪 (Tracking)

**生成链接：**
用户的推广链接格式通常为：`https://yoursite.com/register?ref=USER_CODE` 或 `https://yoursite.com/?aff=USER_CODE`。

**追踪脚本 (放在所有公共页面的顶部 `init.php` 或 Middleware 中):**

```php
// track_affiliate.php
session_start();
require 'db_connection.php'; // PDO 连接

if (isset($_GET['ref']) && !isset($_COOKIE['affiliate_ref'])) {
    $refCode = trim($_GET['ref']);
    
    // 验证推广码是否存在且有效
    $stmt = $pdo->prepare("SELECT user_id FROM users WHERE affiliate_code = ? AND id NOT IN (SELECT id FROM admin_table)");
    $stmt->execute([$refCode]);
    $affiliate = $stmt->fetch();

    if ($affiliate) {
        // 设置 Cookie，有效期 30 天
        setcookie('affiliate_ref', $refCode, time() + (30 * 24 * 60 * 60), '/');
        $_SESSION['affiliate_ref'] = $refCode; // 同时存 Session 以防 Cookie 被禁
        
        // 记录点击日志 (异步或简单插入)
        $ip = $_SERVER['REMOTE_ADDR'];
        $ua = $_SERVER['HTTP_USER_AGENT'];
        $page = $_SERVER['REQUEST_URI'];
        
        $logStmt = $pdo->prepare("INSERT INTO affiliate_clicks (affiliate_id, visitor_ip, user_agent, landing_page) SELECT user_id, ?, ?, ? FROM users WHERE affiliate_code = ?");
        $logStmt->execute([$ip, $ua, $page, $refCode]);
    }
}
```

#### 2. 订单转化与佣金计算 (Conversion)

**在订单支付成功后的回调函数中 (Checkout Success Hook):**

```php
// process_order_success.php
session_start();
require 'db_connection.php';

$orderId = $_POST['order_id'];
$buyerId = $_SESSION['user_id']; // 当前购买的用户 ID
$orderAmount = $_POST['total_amount'];

// 1. 检查是否有推广来源 (优先取 Session，其次 Cookie)
$refCode = $_SESSION['affiliate_ref'] ?? $_COOKIE['affiliate_ref'] ?? null;

if ($refCode) {
    // 2. 获取推广者 ID
    $stmt = $pdo->prepare("SELECT user_id FROM users WHERE affiliate_code = ?");
    $stmt->execute([$refCode]);
    $affiliate = $stmt->fetch();
    
    if ($affiliate) {
        $affiliateId = $affiliate['user_id'];

        // 3. 防作弊：不能自己推荐自己
        if ($affiliateId != $buyerId) {
            
            // 4. 计算佣金 (例如 10%)
            $commissionRate = 0.10; 
            $commissionAmount = $orderAmount * $commissionRate;

            // 5. 写入佣金记录 (状态为待结算)
            $insertStmt = $pdo->prepare("
                INSERT INTO affiliate_commissions 
                (affiliate_id, order_id, referred_user_id, amount, commission, status) 
                VALUES (?, ?, ?, ?, ?, 0)
            ");
            $insertStmt->execute([$affiliateId, $orderId, $buyerId, $orderAmount, $commissionAmount]);

            // 6. 更新推广者待结算余额
            $updateStmt = $pdo->prepare("
                UPDATE affiliate_accounts 
                SET pending_balance = pending_balance + ? 
                WHERE user_id = ?
            ");
            $updateStmt->execute([$commissionAmount, $affiliateId]);
            
            // 7. (可选) 清除追踪标记，避免重复计算（如果是首次购买才奖励）
            // setcookie('affiliate_ref', '', time() - 3600, '/'); 
        }
    }
}
```

#### 3. 佣金结算任务 (Cron Job)

不要用户一购买就立刻让佣金变为“可提现”，需要等待退款期。创建一个每日运行的 Cron 脚本。

```php
// cron_settle_commissions.php
// 逻辑：将 15 天前 状态为 0 (Pending) 的佣金 改为 1 (Approved)
$daysToHold = 15;
$cutoffDate = date('Y-m-d H:i:s', strtotime("-{$daysToHold} days"));

$stmt = $pdo->prepare("
    UPDATE affiliate_commissions 
    SET status = 1, approved_at = NOW() 
    WHERE status = 0 AND created_at < ?
");
$stmt->execute([$cutoffDate]);

// 同时需要更新 affiliate_accounts 表的余额 (将 pending 转入 balance)
// 这里逻辑稍复杂，建议通过触发器或重新计算总和来同步余额
```

#### 4. 提现申请 (Payout Request)

```php
// request_payout.php
$userId = $_SESSION['user_id'];
$amount = $_POST['amount'];

// 检查余额
$stmt = $pdo->prepare("SELECT balance FROM affiliate_accounts WHERE user_id = ?");
$stmt->execute([$userId]);
$account = $stmt->fetch();

if ($account['balance'] >= $amount && $amount >= 100) { // 假设最低提现 100
    $pdo->beginTransaction();
    try {
        // 1. 扣减余额
        $pdo->prepare("UPDATE affiliate_accounts SET balance = balance - ? WHERE user_id = ?")
            ->execute([$amount, $userId]);
        
        // 2. 创建提现记录
        $pdo->prepare("INSERT INTO affiliate_payouts (affiliate_id, amount, status) VALUES (?, ?, 0)")
            ->execute([$userId, $amount]);
            
        $pdo->commit();
        echo "提现申请成功";
    } catch (Exception $e) {
        $pdo->rollBack();
        echo "系统错误";
    }
}
```

---

### 四、安全与反作弊 (Security & Anti-Fraud)

这是 Affiliate 系统最容易被攻击的地方，务必重视：

1.  **自我推荐检测 (Self-Referral):**
    *   注册/购买时，检查 `referrer_id` 是否等于 `user_id`。
    *   检查 **IP 地址** 和 **设备指纹**。如果推广者和购买者 IP 相同，标记为可疑，人工审核。
2.  **Cookie 填充 (Cookie Stuffing):**
    *   防止恶意网站通过 iframe 隐藏你的链接强制给用户设 Cookie。
    *   **对策:** 设置 Cookie 的 `SameSite` 属性，检查 `HTTP_REFERER`（虽然可伪造，但有参考价值）。
3.  **数据篡改:**
    *   所有金额计算必须在**服务器端**进行，不要信任前端传来的佣金数值。
    *   使用 PDO 预处理语句防止 SQL 注入。
4.  **关联订单与退款:**
    *   如果用户发生退款 (Refund)，必须触发逻辑，将对应的佣金记录状态改为 `3 (Cancelled)`，并扣减推广者余额（如果余额不足，记为负数，从下次佣金扣）。
5.  **审核机制:**
    *   对于大额佣金或异常高频的转化，后台标记为“需人工审核”，暂停自动结算。

---

### 五、前端与用户体验 (UX)

1.  **Affiliate  Dashboard (推广者后台):**
    *   **专属链接:** 提供一键复制功能。
    *   **数据统计:** 图表展示点击量 (Clicks)、转化率 (Conversion Rate)、收入 (Earnings)。
    *   **素材库:** 提供 Banner 图片、文案供推广者使用。
    *   **提现页面:** 显示当前余额、历史提现记录。
2.  **管理后台 (Admin Panel):**
    *   **审核提现:** 列表显示申请，支持“打款”或“拒绝”。
    *   **佣金调整:** 允许手动调整佣金（处理纠纷）。
    *   **封禁功能:** 发现作弊账号，一键封禁并冻结余额。

---

### 六、实施建议与路线图

**阶段 1: MVP (最小可行性产品)**
*   只实现一级分销。
*   只支持站内余额提现（用于抵扣消费），暂不接外部支付打款。
*   固定佣金比例。
*   手动审核提现。

**阶段 2: 自动化**
*   接入 PayPal Payouts API 或 支付宝转账接口 实现自动打款。
*   增加退款自动扣回佣金逻辑。
*   增加更详细的数据报表。

**阶段 3: 扩展**
*   支持二级分销（需谨慎合规）。
*   支持不同产品不同佣金比例。
*   提供 API 供大型推广者对接。

### 七、替代方案 (Build vs Buy)

如果你的核心业务不是联盟系统，而是电商或 SaaS，**自己开发可能会踩很多坑**。你可以考虑以下替代方案：

1.  **开源脚本:** 搜索 "PHP Affiliate Script" (如 AffiliateWP 的 standalone 版本，或 WHMCS 的插件)。
2.  **SaaS 服务:** 使用第三方的联盟追踪系统 (如 Post Affiliate Pro, Refersion, FirstPromoter)。
    *   **优点:** 它们提供 JS 代码，你只需嵌入网站，它们处理追踪、Cookie、反作弊和支付。
    *   **缺点:** 按月付费，数据在第三方。
3.  **CMS 插件:** 如果你是用 WordPress + WooCommerce，直接用 **AffiliateWP** 或 **YITH WooCommerce Affiliate Manager**，不要自己写。

### 总结

自己开发 Affiliate 系统的核心难点不在于代码，而在于**财务逻辑的闭环**（退款怎么处理、余额怎么扣）和**反作弊**。

建议先按照上面的数据库设计搭建框架，**务必加上“佣金审核期”**，并在后台预留人工干预的入口。初期规则越简单越好，避免复杂的层级计算导致财务对账困难。
