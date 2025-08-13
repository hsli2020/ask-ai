# 在我用PHP编写的Web应用中，如何加上支付宝、微信支付的收款功能？

在 PHP Web 应用中集成支付宝和微信支付，需要分别接入它们的官方 SDK 或使用 API 接口。以下是实现步骤概览和关键代码示例（截至 2026 年，基于主流做法）。

---

## 一、准备工作

### 1. 注册开发者账号
- **支付宝**：登录 [支付宝开放平台](https://open.alipay.com/)
- **微信支付**：登录 [微信支付商户平台](https://pay.weixin.qq.com/)

### 2. 获取必要参数
#### 支付宝：
- AppID
- 私钥（应用私钥）
- 公钥（支付宝公钥）
- 网关地址（如 `https://openapi.alipay.com/gateway.do`）

#### 微信支付（以 JSAPI 支付为例）：
- 商户号（mch_id）
- AppID（公众号或小程序的）
- APIv3 密钥（用于解密和签名）
- 证书（包括 `apiclient_cert.pem` 和 `apiclient_key.pem`）
- 回调通知 URL

> ⚠️ 注意：微信支付从 2020 年起强制使用 **APIv3** 和 **HTTPS + 证书双向认证**。

---

## 二、推荐使用官方 SDK

### 支付宝：使用 [alipay-sdk-php](https://github.com/alipay/alipay-easysdk-php)
```bash
composer require alipaysdk/easysdk
```

### 微信支付：使用 [wechatpay-php](https://github.com/wechatpay-apiv3/wechatpay-php)
```bash
composer require wechatpay/wechatpay
```

---

## 三、代码示例

### 1. 支付宝网页支付（电脑网站支付）

```php
// config/alipay.php
$config = [
    'protocol' => 'https',
    'gatewayHost' => 'openapi.alipay.com',
    'signType' => 'RSA2',
    'appId' => '你的AppID',
    'merchantPrivateKey' => '你的应用私钥（字符串）',
    'alipayCertPath' => '/path/to/alipayCertPublicKey_RSA2.crt',
    'alipayRootCertPath' => '/path/to/alipayRootCert.crt',
    'merchantCertPath' => '/path/to/appCertPublicKey_*.crt',
];

// 创建支付订单
use Alipay\EasySDK\Kernel\Factory;
use Alipay\EasySDK\Kernel\Config;

Factory::setOptions(new Config($config));

$orderId = 'ORDER_' . time();
$amount = '0.01'; // 单位：元

$result = Factory::payment()
    ->page()
    ->pay(
        '测试商品',
        $orderId,
        $amount,
        'http://yourdomain.com/return_url.php', // 同步回调
        'http://yourdomain.com/notify_url_alipay.php' // 异步通知
    );

// 跳转到支付宝支付页面
echo $result->body;
```

### 2. 微信支付（JSAPI 支付，适用于公众号内）

```php
// config/wechatpay.php
use WeChatPay\Builder;
use WeChatPay\Crypto\Rsa;
use WeChatPay\Util\PemUtil;

$merchantId = '你的商户号';
$merchantSerialNumber = '证书序列号';
$merchantPrivateKey = PemUtil::loadPrivateKey('/path/to/apiclient_key.pem');
$wechatpayCertificate = PemUtil::loadCertificate('/path/to/wechatpay_cert.pem');

$instance = Builder::factory([
    'mchid'      => $merchantId,
    'serial'     => $merchantSerialNumber,
    'privateKey' => $merchantPrivateKey,
    'certs'      => [$wechatpayCertificate],
]);

// 创建预支付订单
$outTradeNo = 'ORDER_' . time();
$body = '测试商品';
$total = 1; // 单位：分

$response = $instance->chain('v3/pay/transactions/jsapi')
    ->post([
        'json' => [
            'appid' => '你的公众号AppID',
            'mchid' => $merchantId,
            'description' => $body,
            'out_trade_no' => $outTradeNo,
            'amount' => ['total' => $total],
            'payer' => ['openid' => '用户的openid'],
            'notify_url' => 'https://yourdomain.com/notify_url_wechat.php'
        ]
    ]);

if ($response->getStatusCode() === 200) {
    $prepayId = json_decode($response->getBody(), true)['prepay_id'];
    
    // 生成前端调用所需参数
    $timestamp = strval(time());
    $nonceStr = uniqid();
    $package = 'prepay_id=' . $prepayId;
    
    $signMessage = "your_appid\n{$timestamp}\n{$nonceStr}\n{$package}\n";
    $signature = Rsa::sign($signMessage, $merchantPrivateKey);
    
    $jsConfig = [
        'appId'     => 'your_appid',
        'timeStamp' => $timestamp,
        'nonceStr'  => $nonceStr,
        'package'   => $package,
        'signType'  => 'RSA',
        'paySign'   => $signature
    ];
    
    // 将 $jsConfig 传给前端，调用微信 JSAPI
    echo json_encode($jsConfig);
}
```

前端 JavaScript 调用：
```html
<script src="https://res.wx.qq.com/open/js/jweixin-1.6.0.js"></script>
<script>
wx.chooseWXPay({
  timestamp: <?= $jsConfig['timeStamp'] ?>,
  nonceStr: '<?= $jsConfig['nonceStr'] ?>',
  package: '<?= $jsConfig['package'] ?>',
  signType: '<?= $jsConfig['signType'] ?>',
  paySign: '<?= $jsConfig['paySign'] ?>',
  success: function(res) { alert('支付成功'); },
  fail: function(res) { alert('支付失败'); }
});
</script>
```

---

## 四、处理异步通知（非常重要！）

支付结果必须通过**服务器异步通知**来确认，不能依赖前端跳转！

### 支付宝通知处理（`notify_url_alipay.php`）：
```php
use Alipay\EasySDK\Kernel\Factory;

// 验证签名
if (Factory::payment()->common()->verify($_POST)) {
    if ($_POST['trade_status'] === 'TRADE_SUCCESS') {
        // 处理业务逻辑：更新订单状态等
        error_log("支付宝支付成功: " . $_POST['out_trade_no']);
        echo 'success'; // 必须返回 success
    }
} else {
    echo 'failure';
}
```

### 微信支付通知处理（`notify_url_wechat.php`）：
```php
use WeChatPay\Notify\Handler;
use WeChatPay\Crypto\Rsa;

$handler = new Handler(
    ['mchid' => '你的商户号'],
    function ($message, $associatedData, $nonce) use ($merchantPrivateKey) {
        // 解密数据（如有敏感信息）
        // 此处 message 是 JSON 字符串
        $data = json_decode($message, true);
        if ($data['event_type'] === 'TRANSACTION.SUCCESS') {
            $resource = $data['resource'];
            $ciphertext = $resource['ciphertext'];
            $decrypt = Rsa::decrypt($ciphertext, $merchantPrivateKey, $nonce, $associatedData);
            $order = json_decode($decrypt, true);
            
            if ($order['trade_state'] === 'SUCCESS') {
                // 更新订单状态
                error_log("微信支付成功: " . $order['out_trade_no']);
            }
        }
        return ['code' => 'SUCCESS', 'message' => '成功'];
    }
);

$response = $handler->handle(file_get_contents('php://input'));
http_response_code($response->getStatusCode());
echo $response->getBody();
```

---

## 五、安全建议

1. 所有回调 URL 必须使用 HTTPS。
2. 严格验证签名，防止伪造请求。
3. 不要信任前端传回的“支付成功”状态。
4. 敏感信息（如私钥）不要硬编码，使用环境变量或配置中心。
5. 定期更新 SDK 和证书。

---

## 六、其他支付方式

- **H5 支付**：适用于手机浏览器
- **Native 支付**：生成二维码供扫码
- **小程序支付**：需在小程序端调用

可根据业务场景选择对应接口。

---

如需完整项目模板，可参考 GitHub 上的开源项目，如：
- [ yansongda/pay](https://github.com/yansongda/pay)（支持多支付渠道，但注意维护状态）
- 官方 SDK 示例仓库

如果你提供具体场景（如：是 PC 网站？还是公众号？是否已有用户体系？），我可以给出更精准的集成方案。

# 我是在 PC 网站上集成支付功能，请给出更详细的说明，比如，我的网站上有多种商品出售，用户选了一种或多种商品之后，如何完成支付

在 PC 网站上集成支付宝和微信支付（支持多商品购物车场景），需要完成以下核心流程：

> **目标**：用户选择多个商品 → 生成订单 → 跳转到支付页面（支付宝/微信）→ 支付成功 → 异步通知更新订单状态

---

## 一、整体架构设计

```
1. 用户操作：
   商品页 → 加入购物车 → 查看购物车 → 提交订单 → 选择支付方式 → 跳转支付 → 支付成功

2. 后端逻辑：
   - 创建订单（含多个商品）
   - 调用支付接口生成支付页面
   - 接收异步通知，验证并更新订单状态
   - 提供查询接口（可选）

3. 数据库设计（简化）：
   - `orders` 表：order_id, user_id, total_amount, status, created_at
   - `order_items` 表：id, order_id, product_id, name, price, quantity
```

---

## 二、详细步骤与代码实现（PHP + MySQL）

### 步骤 1：用户提交订单（前端）

```html
<!-- 购物车页面 -->
<form action="/create_order.php" method="POST">
  <input type="hidden" name="product_ids" value="1,3,5">
  <input type="hidden" name="quantities" value="2,1,1">
  <button type="submit">去支付</button>
</form>
```

---

### 步骤 2：后端创建订单（`create_order.php`）

```php
<?php
session_start();
require_once 'db.php'; // 数据库连接

// 1. 获取用户选择的商品
$productIds = explode(',', $_POST['product_ids']);
$quantities = explode(',', $_POST['quantities']);

// 2. 查询商品信息并计算总价
$total = 0;
$orderItems = [];
$stmt = $pdo->prepare("SELECT id, name, price FROM products WHERE id = ?");
foreach ($productIds as $i => $id) {
    $stmt->execute([$id]);
    $product = $stmt->fetch();
    if (!$product) die('商品不存在');
    
    $qty = (int)$quantities[$i];
    $subtotal = $product['price'] * $qty;
    $total += $subtotal;
    
    $orderItems[] = [
        'product_id' => $id,
        'name' => $product['name'],
        'price' => $product['price'],
        'quantity' => $qty
    ];
}

// 3. 插入订单主表
$orderId = 'ORDER_' . date('YmdHis') . rand(1000, 9999);
$stmt = $pdo->prepare("
    INSERT INTO orders (order_id, user_id, total_amount, status) 
    VALUES (?, ?, ?, 'pending')
");
$stmt->execute([$orderId, $_SESSION['user_id'], $total]);

// 4. 插入订单明细
$stmt = $pdo->prepare("
    INSERT INTO order_items (order_id, product_id, name, price, quantity)
    VALUES (?, ?, ?, ?, ?)
");
foreach ($orderItems as $item) {
    $stmt->execute([
        $orderId,
        $item['product_id'],
        $item['name'],
        $item['price'],
        $item['quantity']
    ]);
}

// 5. 跳转到支付选择页
header("Location: /pay_select.php?order_id=" . urlencode($orderId));
exit;
```

---

### 步骤 3：支付方式选择页（`pay_select.php`）

```html
<h2>请选择支付方式</h2>
<p>订单号：<?= htmlspecialchars($_GET['order_id']) ?></p>
<p>应付金额：¥<?= number_format($total, 2) ?></p>

<a href="/alipay_pay.php?order_id=<?= urlencode($_GET['order_id']) ?>">支付宝支付</a>
<a href="/wechat_pay_h5.php?order_id=<?= urlencode($_GET['order_id']) ?>">微信支付（H5）</a>
```

> 💡 微信在 PC 网站上使用 **H5 支付**（生成二维码或跳转到微信支付中间页）

---

## 三、接入支付（PC 场景）

### A. 支付宝电脑网站支付（推荐）

#### 文件：`alipay_pay.php`

```php
<?php
require_once 'vendor/autoload.php';
require_once 'db.php';

use Alipay\EasySDK\Kernel\Factory;
use Alipay\EasySDK\Kernel\Config;

// 1. 获取订单信息
$orderId = $_GET['order_id'];
$stmt = $pdo->prepare("SELECT total_amount FROM orders WHERE order_id = ? AND status = 'pending'");
$stmt->execute([$orderId]);
$order = $stmt->fetch();
if (!$order) die('订单无效');

$amount = sprintf("%.2f", $order['total_amount']); // 必须保留两位小数

// 2. 配置支付宝
$config = new Config([
    'protocol' => 'https',
    'gatewayHost' => 'openapi.alipay.com',
    'signType' => 'RSA2',
    'appId' => '你的AppID',
    'merchantPrivateKey' => file_get_contents('/path/to/app_private_key.pem'),
    'alipayCertPath' => '/path/to/alipayCertPublicKey_RSA2.crt',
    'alipayRootCertPath' => '/path/to/alipayRootCert.crt',
    'merchantCertPath' => '/path/to/appCertPublicKey_*.crt',
]);
Factory::setOptions($config);

// 3. 调用支付
$result = Factory::payment()
    ->page()
    ->pay(
        '商城订单-' . $orderId,
        $orderId,
        $amount,
        'https://yourdomain.com/pay_return.php',      // 同步回调（用户支付后浏览器跳回）
        'https://yourdomain.com/alipay_notify.php'    // 异步通知（服务器间通信）
    );

// 4. 直接输出 HTML，跳转到支付宝
echo $result->body;
```

> ✅ 用户将被重定向到支付宝支付页面，支付完成后自动跳回 `pay_return.php`

---

### B. 微信支付 H5（适用于 PC 浏览器）

> 微信官方不支持 PC 网站直接调起 JSAPI，但提供 **H5 支付**：用户点击后跳转到微信支付中间页，可用手机扫码或在微信内打开。

#### 文件：`wechat_pay_h5.php`

```php
<?php
require_once 'vendor/autoload.php';
require_once 'db.php';

use WeChatPay\Builder;
use WeChatPay\Util\PemUtil;

// 1. 获取订单
$orderId = $_GET['order_id'];
$stmt = $pdo->prepare("SELECT total_amount FROM orders WHERE order_id = ? AND status = 'pending'");
$stmt->execute([$orderId]);
$order = $stmt->fetch();
if (!$order) die('订单无效');

// 2. 初始化微信支付
$merchantId = '你的商户号';
$merchantSerialNumber = '证书序列号';
$merchantPrivateKey = PemUtil::loadPrivateKey('/path/to/apiclient_key.pem');
$wechatpayCertificate = PemUtil::loadCertificate('/path/to/wechatpay_cert.pem');

$instance = Builder::factory([
    'mchid'      => $merchantId,
    'serial'     => $merchantSerialNumber,
    'privateKey' => $merchantPrivateKey,
    'certs'      => [$wechatpayCertificate],
]);

// 3. 调用 H5 支付接口
$response = $instance->chain('v3/pay/transactions/h5')
    ->post([
        'json' => [
            'appid' => '你的公众号AppID', // 注意：H5 支付需绑定公众号
            'mchid' => $merchantId,
            'description' => '商城订单-' . $orderId,
            'out_trade_no' => $orderId,
            'amount' => ['total' => (int)($order['total_amount'] * 100)], // 单位：分
            'scene_info' => [
                'payer_client_ip' => $_SERVER['REMOTE_ADDR'],
                'h5_info' => ['type' => 'Wap']
            ],
            'notify_url' => 'https://yourdomain.com/wechat_notify.php'
        ]
    ]);

if ($response->getStatusCode() === 200) {
    $data = json_decode($response->getBody(), true);
    $h5Url = $data['h5_url'];
    // 跳转到微信支付中间页
    header("Location: " . $h5Url);
    exit;
} else {
    die('微信支付请求失败');
}
```

> ⚠️ 注意：微信 H5 支付需在**商户平台配置支付域名**，且必须是备案域名。

---

## 四、处理异步通知（关键！）

### 支付宝通知：`alipay_notify.php`

```php
<?php
require_once 'vendor/autoload.php';
require_once 'db.php';

use Alipay\EasySDK\Kernel\Factory;
use Alipay\EasySDK\Kernel\Config;

$config = new Config([...]); // 同上
Factory::setOptions($config);

// 验证签名
if (Factory::payment()->common()->verify($_POST)) {
    $orderId = $_POST['out_trade_no'];
    $tradeStatus = $_POST['trade_status'];

    if ($tradeStatus === 'TRADE_SUCCESS') {
        // 更新订单状态为已支付
        $stmt = $pdo->prepare("UPDATE orders SET status = 'paid', paid_at = NOW() WHERE order_id = ? AND status = 'pending'");
        $stmt->execute([$orderId]);
        
        error_log("✅ 支付宝支付成功: $orderId");
    }
    echo 'success'; // 必须返回 success，否则支付宝会重试
} else {
    echo 'failure';
}
```

### 微信通知：`wechat_notify.php`

```php
<?php
require_once 'vendor/autoload.php';
require_once 'db.php';

use WeChatPay\Notify\Handler;
use WeChatPay\Crypto\Rsa;
use WeChatPay\Util\PemUtil;

$merchantPrivateKey = PemUtil::loadPrivateKey('/path/to/apiclient_key.pem');

$handler = new Handler(
    ['mchid' => '你的商户号'],
    function ($message, $associatedData, $nonce) use ($merchantPrivateKey) {
        $data = json_decode($message, true);
        if ($data['event_type'] === 'TRANSACTION.SUCCESS') {
            $resource = $data['resource'];
            $ciphertext = $resource['ciphertext'];
            $decrypt = Rsa::decrypt($ciphertext, $merchantPrivateKey, $nonce, $associatedData);
            $order = json_decode($decrypt, true);
            
            if ($order['trade_state'] === 'SUCCESS') {
                $orderId = $order['out_trade_no'];
                $stmt = $pdo->prepare("UPDATE orders SET status = 'paid', paid_at = NOW() WHERE order_id = ? AND status = 'pending'");
                $stmt->execute([$orderId]);
                error_log("✅ 微信支付成功: $orderId");
            }
        }
        return ['code' => 'SUCCESS', 'message' => '成功'];
    }
);

$input = file_get_contents('php://input');
$response = $handler->handle($input);
http_response_code($response->getStatusCode());
echo $response->getBody();
```

---

## 五、同步回调（用户体验优化）

`pay_return.php`（支付宝/微信均可跳回此页）：

```php
<?php
// 只用于展示结果，**不能用于判断支付是否成功！**
$orderId = $_GET['out_trade_no'] ?? '';
?>
<h2>支付结果</h2>
<?php if ($orderId): ?>
  <p>订单 <?= htmlspecialchars($orderId) ?> 已提交支付。</p>
  <p>系统正在处理，请稍后查看订单状态。</p>
  <a href="/my_orders.php">查看我的订单</a>
<?php else: ?>
  <p>支付未完成。</p>
<?php endif; ?>
```

> 🔒 **重要**：前端显示的“支付成功”只是跳转，真实状态以异步通知为准！

---

## 六、安全与最佳实践

1. **订单幂等性**：同一个 `order_id` 只能支付一次。
2. **金额校验**：支付前再次从数据库读取金额，防止前端篡改。
3. **HTTPS**：所有支付相关页面必须使用 HTTPS。
4. **日志记录**：记录所有支付请求和通知，便于排查。
5. **超时处理**：订单 30 分钟未支付自动取消。

---

## 七、测试建议

- **支付宝**：使用 [沙箱环境](https://open.alipay.com/development/sandbox)
- **微信**：使用 [微信支付 APIv3 沙箱](https://pay.weixin.qq.com/wiki/doc/apiv3/wechatpay/wechatpay6_0.shtml)（需申请）

---

如果你使用的是 Laravel、ThinkPHP 等框架，也可以使用封装好的扩展包（如 `yansongda/pay`），但务必确认其支持 APIv3 和最新安全要求。

需要我提供完整的项目结构或 Docker 配置示例吗？


# 提供完整的项目结构和 Docker 配置示例

以下是为 **PC 网站集成支付宝 + 微信支付（多商品购物车）** 的完整项目结构和 Docker 配置示例，基于 PHP 8.2 + MySQL 8 + Nginx，使用 Composer 管理依赖。

---

## 一、项目目录结构

```bash
online-shop/
├── docker/
│   ├── php/
│   │   └── Dockerfile
│   ├── nginx/
│   │   └── default.conf
│   └── mysql/
│       └── init.sql
├── src/
│   ├── config/
│   │   ├── alipay.php
│   │   ├── wechatpay.php
│   │   └── database.php
│   ├── public/
│   │   ├── index.php
│   │   ├── create_order.php
│   │   ├── pay_select.php
│   │   ├── alipay_pay.php
│   │   ├── wechat_pay_h5.php
│   │   ├── alipay_notify.php
│   │   ├── wechat_notify.php
│   │   └── pay_return.php
│   ├── lib/
│   │   ├── db.php
│   │   └── helpers.php
│   └── templates/
│       ├── cart.html
│       └── order_success.html
├── certs/
│   ├── alipay/
│   │   ├── app_private_key.pem
│   │   ├── alipayCertPublicKey_RSA2.crt
│   │   ├── alipayRootCert.crt
│   │   └── appCertPublicKey_*.crt
│   └── wechat/
│       ├── apiclient_cert.pem
│       ├── apiclient_key.pem
│       └── wechatpay_cert.pem
├── composer.json
├── docker-compose.yml
└── README.md
```

> 💡 所有敏感证书文件 **不要提交到 Git**，应通过 `.gitignore` 排除。

---

## 二、Docker 配置

### 1. `docker-compose.yml`

```yaml
version: '3.8'

services:
  web:
    build:
      context: .
      dockerfile: docker/php/Dockerfile
    container_name: shop_php
    volumes:
      - ./src:/var/www/html
      - ./certs:/var/www/certs:ro
    depends_on:
      - db
    networks:
      - shop_net

  nginx:
    image: nginx:alpine
    container_name: shop_nginx
    ports:
      - "8080:80"
    volumes:
      - ./src:/var/www/html
      - ./docker/nginx/default.conf:/etc/nginx/conf.d/default.conf
    depends_on:
      - web
    networks:
      - shop_net

  db:
    image: mysql:8.0
    container_name: shop_mysql
    environment:
      MYSQL_ROOT_PASSWORD: rootpass123
      MYSQL_DATABASE: shop_db
      MYSQL_USER: shop_user
      MYSQL_PASSWORD: shop_pass456
    volumes:
      - shop_mysql_data:/var/lib/mysql
      - ./docker/mysql/init.sql:/docker-entrypoint-initdb.d/init.sql
    networks:
      - shop_net

volumes:
  shop_mysql_data:

networks:
  shop_net:
    driver: bridge
```

---

### 2. `docker/php/Dockerfile`

```Dockerfile
FROM php:8.2-fpm

# 安装扩展
RUN apt-get update && apt-get install -y \
    libpng-dev \
    libjpeg-dev \
    libfreetype-dev \
    libzip-dev \
    unzip \
    git \
    && docker-php-ext-configure gd --with-freetype --with-jpeg \
    && docker-php-ext-install -j$(nproc) gd pdo pdo_mysql mysqli zip

# 安装 Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# 设置工作目录
WORKDIR /var/www/html

# 复制 composer.json 并安装依赖
COPY composer.json composer.lock ./
RUN composer install --no-dev --optimize-autoloader

# 启动 PHP-FPM
CMD ["php-fpm"]
```

---

### 3. `docker/nginx/default.conf`

```nginx
server {
    listen 80;
    server_name localhost;
    root /var/www/html/public;
    index index.php;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        fastcgi_pass web:9000;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }

    # 禁止访问敏感目录
    location ~ ^/(certs|docker|src/config) {
        deny all;
    }
}
```

---

### 4. `docker/mysql/init.sql`

```sql
CREATE TABLE IF NOT EXISTS products (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL UNIQUE,
    user_id INT NOT NULL,
    total_amount DECIMAL(10,2) NOT NULL,
    status ENUM('pending','paid','cancelled') DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    paid_at TIMESTAMP NULL
);

CREATE TABLE IF NOT EXISTS order_items (
    id INT AUTO_INCREMENT PRIMARY KEY,
    order_id VARCHAR(64) NOT NULL,
    product_id INT NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE
);

-- 插入测试商品
INSERT INTO products (name, price) VALUES
('PHP 编程指南', 59.90),
('Docker 实战手册', 79.00),
('Web 安全入门', 45.50);
```

---

## 三、核心配置文件

### 1. `src/config/database.php`

```php
<?php
return [
    'host' => 'db',
    'dbname' => 'shop_db',
    'user' => 'shop_user',
    'pass' => 'shop_pass456',
    'charset' => 'utf8mb4'
];
```

### 2. `src/lib/db.php`

```php
<?php
$config = require __DIR__ . '/../config/database.php';
$dsn = "mysql:host={$config['host']};dbname={$config['dbname']};charset={$config['charset']}";
$pdo = new PDO($dsn, $config['user'], $config['pass'], [
    PDO::ATTR_ERRMODE => PDO::ERRMODE_EXCEPTION,
    PDO::ATTR_DEFAULT_FETCH_MODE => PDO::FETCH_ASSOC
]);
return $pdo;
```

### 3. `composer.json`

```json
{
  "require": {
    "php": "^8.2",
    "alipaysdk/easysdk": "^2.2",
    "wechatpay/wechatpay": "^1.5"
  },
  "autoload": {
    "psr-4": {
      "Shop\\": "src/"
    }
  }
}
```

---

## 四、关键支付文件示例（简化版）

### `src/public/alipay_pay.php`

```php
<?php
require_once __DIR__ . '/../lib/db.php';
require_once __DIR__ . '/../../vendor/autoload.php';

use Alipay\EasySDK\Kernel\Factory;
use Alipay\EasySDK\Kernel\Config;

$orderId = $_GET['order_id'] ?? '';
if (!$orderId) die('缺少订单号');

// 获取订单
$stmt = $pdo->prepare("SELECT total_amount FROM orders WHERE order_id = ? AND status = 'pending'");
$stmt->execute([$orderId]);
$order = $stmt->fetch();
if (!$order) die('订单无效或已支付');

// 加载配置
$config = new Config([
    'protocol' => 'https',
    'gatewayHost' => 'openapi.alipay.com',
    'signType' => 'RSA2',
    'appId' => $_ENV['ALIPAY_APP_ID'] ?? 'your_app_id',
    'merchantPrivateKey' => file_get_contents('/var/www/certs/alipay/app_private_key.pem'),
    'alipayCertPath' => '/var/www/certs/alipay/alipayCertPublicKey_RSA2.crt',
    'alipayRootCertPath' => '/var/www/certs/alipay/alipayRootCert.crt',
    'merchantCertPath' => '/var/www/certs/alipay/appCertPublicKey_*.crt',
]);
Factory::setOptions($config);

// 调起支付
$result = Factory::payment()
    ->page()
    ->pay(
        '商城订单-' . $orderId,
        $orderId,
        sprintf("%.2f", $order['total_amount']),
        'http://localhost:8080/pay_return.php',
        'http://localhost:8080/alipay_notify.php'
    );

echo $result->body;
```

> 💡 建议将敏感参数（如 AppID）通过环境变量注入（见下文）。

---

## 五、环境变量支持（可选增强）

在 `docker-compose.yml` 中添加：

```yaml
# 在 web 服务下
environment:
  ALIPAY_APP_ID: "your_alipay_app_id"
  WECHAT_MCH_ID: "your_wechat_mch_id"
```

然后在 PHP 中用 `$_ENV['ALIPAY_APP_ID']` 读取。

---

## 六、启动项目

```bash
# 1. 构建并启动
docker-compose up -d

# 2. 安装依赖（如果 Dockerfile 未自动安装）
docker-compose exec web composer install

# 3. 访问
http://localhost:8080/
```

---

## 七、安全建议

1. **证书保护**：
   - `certs/` 目录权限设为 `600`
   - 通过 volume 挂载，避免打包进镜像

2. **生产部署**：
   - 使用 HTTPS（可加 Traefik 或 Let's Encrypt）
   - 移除测试商品和默认用户
   - 设置强密码和防火墙

3. **日志监控**：
   - 将 `error_log()` 输出到 Docker 日志：`docker-compose logs -f web`

---

## 八、测试流程

1. 访问 `http://localhost:8080/`（模拟购物车）
2. 提交订单 → 跳转到支付选择页
3. 选择支付宝 → 跳转到沙箱环境支付
4. 支付成功后，查看数据库 `orders.status` 是否变为 `paid`

---

> ✅ 此结构适合中小型电商项目，易于扩展（如加入用户系统、库存管理等）。

需要我提供前端 HTML 示例或沙箱测试账号配置指南吗？

# 我提供前端 HTML 示例

当然！以下是与上述后端逻辑配套的 **前端 HTML 示例**，涵盖：

- 商品列表页（模拟商城首页）
- 购物车页面（支持多商品选择与数量调整）
- 订单提交与支付方式选择
- 支付结果提示页

所有页面均使用原生 HTML + JavaScript（无框架），便于集成到任何 PHP 项目中。

---

## 一、商品列表页：`src/public/index.php`

```php
<?php
require_once __DIR__ . '/../lib/db.php';
$stmt = $pdo->query("SELECT * FROM products");
$products = $stmt->fetchAll();
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>在线商城</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; }
    .product { border: 1px solid #ddd; padding: 15px; margin: 10px 0; width: 300px; display: inline-block; }
    button { padding: 6px 12px; background: #4CAF50; color: white; border: none; cursor: pointer; }
  </style>
</head>
<body>
  <h1>🛒 在线商城</h1>
  <div id="products">
    <?php foreach ($products as $p): ?>
      <div class="product">
        <h3><?= htmlspecialchars($p['name']) ?></h3>
        <p>¥<?= number_format($p['price'], 2) ?></p>
        <button onclick="addToCart(<?= $p['id'] ?>, '<?= addslashes($p['name']) ?>', <?= $p['price'] ?>)">
          加入购物车
        </button>
      </div>
    <?php endforeach; ?>
  </div>

  <br><br>
  <a href="cart.php">👉 查看购物车</a>

  <script>
    let cart = JSON.parse(localStorage.getItem('cart')) || {};

    function addToCart(id, name, price) {
      if (cart[id]) {
        cart[id].quantity += 1;
      } else {
        cart[id] = { id, name, price, quantity: 1 };
      }
      localStorage.setItem('cart', JSON.stringify(cart));
      alert('已加入购物车：' + name);
    }
  </script>
</body>
</html>
```

> 💡 使用 `localStorage` 模拟购物车（实际项目建议用 Session 或数据库）

---

## 二、购物车页面：`src/public/cart.php`

```php
<?php
// 仅用于展示，真实场景应从数据库加载用户购物车
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>购物车</title>
  <style>
    table { width: 100%; border-collapse: collapse; margin: 20px 0; }
    th, td { border: 1px solid #ccc; padding: 10px; text-align: left; }
    input[type=number] { width: 60px; }
    .total { font-size: 1.2em; font-weight: bold; margin-top: 10px; }
  </style>
</head>
<body>
  <h1>🛒 购物车</h1>
  <form id="checkoutForm" action="create_order.php" method="POST">
    <table>
      <thead>
        <tr>
          <th>商品</th>
          <th>单价</th>
          <th>数量</th>
          <th>小计</th>
        </tr>
      </thead>
      <tbody id="cartItems">
        <!-- 由 JS 动态填充 -->
      </tbody>
    </table>
    <div class="total">总计：<span id="totalPrice">¥0.00</span></div>
    <br>
    <button type="submit" id="checkoutBtn" disabled>去结算</button>
  </form>

  <script>
    let cart = JSON.parse(localStorage.getItem('cart')) || {};
    const itemsContainer = document.getElementById('cartItems');
    const totalPriceEl = document.getElementById('totalPrice');
    const checkoutBtn = document.getElementById('checkoutBtn');

    function renderCart() {
      itemsContainer.innerHTML = '';
      let total = 0;
      let productIds = [], quantities = [];

      for (let id in cart) {
        const item = cart[id];
        const subtotal = (item.price * item.quantity).toFixed(2);
        total += parseFloat(subtotal);

        productIds.push(item.id);
        quantities.push(item.quantity);

        const row = `
          <tr>
            <td>${item.name}</td>
            <td>¥${item.price.toFixed(2)}</td>
            <td>
              <input type="number" min="1" value="${item.quantity}" 
                     onchange="updateQuantity(${id}, this.value)">
            </td>
            <td>¥${subtotal}</td>
          </tr>
        `;
        itemsContainer.innerHTML += row;
      }

      totalPriceEl.textContent = '¥' + total.toFixed(2);
      checkoutBtn.disabled = (productIds.length === 0);

      // 将商品 ID 和数量写入隐藏字段
      const idsInput = document.createElement('input');
      idsInput.type = 'hidden';
      idsInput.name = 'product_ids';
      idsInput.value = productIds.join(',');
      const qtyInput = document.createElement('input');
      qtyInput.type = 'hidden';
      qtyInput.name = 'quantities';
      qtyInput.value = quantities.join(',');

      // 清除旧隐藏字段
      const form = document.getElementById('checkoutForm');
      form.querySelectorAll('input[name="product_ids"], input[name="quantities"]').forEach(e => e.remove());
      form.appendChild(idsInput);
      form.appendChild(qtyInput);
    }

    function updateQuantity(id, qty) {
      qty = parseInt(qty);
      if (qty < 1) qty = 1;
      cart[id].quantity = qty;
      localStorage.setItem('cart', JSON.stringify(cart));
      renderCart();
    }

    renderCart();
  </script>
</body>
</html>
```

---

## 三、支付方式选择页：`src/public/pay_select.php`

```php
<?php
$orderId = $_GET['order_id'] ?? '';
if (!$orderId) {
    header('Location: /');
    exit;
}
// 查询订单金额（可选）
require_once __DIR__ . '/../lib/db.php';
$stmt = $pdo->prepare("SELECT total_amount FROM orders WHERE order_id = ?");
$stmt->execute([$orderId]);
$order = $stmt->fetch();
$amount = $order ? $order['total_amount'] : 0;
?>
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>选择支付方式</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; }
    .pay-option { margin: 20px 0; padding: 15px; border: 1px solid #ddd; display: inline-block; width: 200px; text-align: center; }
    .pay-option a { display: block; margin-top: 10px; padding: 8px; background: #1e90ff; color: white; text-decoration: none; }
  </style>
</head>
<body>
  <h1>💰 请选择支付方式</h1>
  <p>订单号：<strong><?= htmlspecialchars($orderId) ?></strong></p>
  <p>应付金额：<strong>¥<?= number_format($amount, 2) ?></strong></p>

  <div class="pay-option">
    <h3>支付宝</h3>
    <a href="alipay_pay.php?order_id=<?= urlencode($orderId) ?>">立即支付</a>
  </div>

  <div class="pay-option">
    <h3>微信支付</h3>
    <a href="wechat_pay_h5.php?order_id=<?= urlencode($orderId) ?>">立即支付</a>
  </div>

  <br><br>
  <a href="index.php">← 返回商城</a>
</body>
</html>
```

---

## 四、支付结果页：`src/public/pay_return.php`

```php
<!DOCTYPE html>
<html lang="zh-CN">
<head>
  <meta charset="UTF-8">
  <title>支付结果</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 40px; }
    .success { color: green; }
    .info { color: #555; }
  </style>
</head>
<body>
  <h1>✅ 支付已提交</h1>
  <p class="info">
    系统正在处理您的支付请求，请勿重复支付。
  </p>
  <p>
    订单状态将以服务器通知为准，您可随时在“我的订单”中查看。
  </p>
  <br>
  <a href="index.php">返回首页</a> |
  <a href="my_orders.php">查看我的订单</a>
</body>
</html>
```

> ⚠️ 此页面仅用于用户体验，**不表示支付成功**！

---

## 五、（可选）我的订单页：`src/public/my_orders.php`

```php
<?php
require_once __DIR__ . '/../lib/db.php';
$stmt = $pdo->query("
    SELECT order_id, total_amount, status, created_at 
    FROM orders 
    ORDER BY created_at DESC 
    LIMIT 10
");
$orders = $stmt->fetchAll();
?>
<!DOCTYPE html>
<html>
<head>
  <meta charset="UTF-8">
  <title>我的订单</title>
  <style>
    table { width: 100%; border-collapse: collapse; margin: 20px 0; }
    th, td { border: 1px solid #ccc; padding: 10px; }
    .paid { color: green; }
    .pending { color: orange; }
  </style>
</head>
<body>
  <h1>📦 我的订单</h1>
  <table>
    <thead>
      <tr>
        <th>订单号</th>
        <th>金额</th>
        <th>状态</th>
        <th>时间</th>
      </tr>
    </thead>
    <tbody>
      <?php foreach ($orders as $o): ?>
        <tr>
          <td><?= htmlspecialchars($o['order_id']) ?></td>
          <td>¥<?= number_format($o['total_amount'], 2) ?></td>
          <td class="<?= $o['status'] ?>">
            <?php
            $statusText = [
                'pending' => '待支付',
                'paid' => '已支付',
                'cancelled' => '已取消'
            ];
            echo $statusText[$o['status']] ?? $o['status'];
            ?>
          </td>
          <td><?= date('Y-m-d H:i', strtotime($o['created_at'])) ?></td>
        </tr>
      <?php endforeach; ?>
    </tbody>
  </table>
  <a href="index.php">返回商城</a>
</body>
</html>
```

---

## 六、使用说明

1. 所有页面放在 `src/public/` 目录下
2. 通过 `http://localhost:8080/` 访问
3. 购物车数据暂存于浏览器 `localStorage`（刷新不丢失）
4. 提交订单后，跳转到支付选择页
5. 支付完成后，异步通知会更新数据库状态
6. 用户可通过“我的订单”查看最终状态

---

## 七、注意事项

- **生产环境**：购物车应改用 Session 或数据库存储，避免跨设备丢失
- **安全性**：前端传入的 `product_ids` 和 `quantities` 在后端必须重新校验价格
- **微信 H5 支付**：需在微信商户平台配置 **支付域名**（如 `yourdomain.com`）
- **支付宝沙箱**：测试时使用 [沙箱账号](https://open.alipay.com/development/sandbox)

---

如果你希望将前端改为 Vue/React，或需要 AJAX 无刷新提交订单，也可以进一步扩展。需要我提供 SPA 版本吗？

