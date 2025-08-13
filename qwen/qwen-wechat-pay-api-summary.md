# 总结一下微信支付API接口规范，重点讲解最常用的部分，并给出PHP代码调用例子

微信支付目前主要支持两个 API 版本：V2（XML + MD5/RSA 签名）和 V3（JSON + RSA 签名）。推荐优先使用 V3，因其更安全、结构更清晰。以下是微信支付 API 的核心规范、常用接口说明及完整的 PHP 调用示例，帮助你快速集成。

---

## 一、API 版本对比

| 特性         | V2 接口                        | V3 接口                        |
|--------------|--------------------------------|--------------------------------|
| 数据格式     | XML                            | JSON                           |
| 签名方式     | MD5 / RSA                      | RSA（SHA256 with RSA）         |
| 证书使用     | 部分接口需要                   | 所有接口都需要                 |
| 推荐使用     | 旧项目兼容                     | 新项目推荐                     |

---

## 二、常用接口概览（V3）

| 接口名称       | 功能说明             | 接口路径                                 |
|----------------|----------------------|------------------------------------------|
| 统一下单       | 创建支付订单         | `/v3/pay/transactions/jsapi`             |
| 查询订单       | 查询支付结果         | `/v3/pay/transactions/id/{transaction_id}` |
| 关闭订单       | 主动关闭未支付订单   | `/v3/pay/transactions/id/{transaction_id}/close` |
| 申请退款       | 发起退款请求         | `/v3/refund/domestic/refunds`            |
| 查询退款       | 查询退款状态         | `/v3/refund/domestic/refunds/{refund_id}` |
| 下载对账单     | 获取交易对账文件     | `/v3/bill/tradebill`                     |

---

## 三、签名机制（V3）

V3 接口使用 RSA 签名，流程如下：

1. 构造签名串（请求方法、URL、时间戳、随机串、请求体）
2. 使用商户私钥对签名串进行 SHA256 with RSA 签名
3. 将签名结果放入请求头 `Authorization` 中

签名串格式示例：
```
POST
/v3/pay/transactions/jsapi
1717020800
random_nonce_str
{"appid":"wx123456","amount":{"total":1}}
```

---

## 四、PHP 调用示例（V3）

### 1. 配置类（WeChatPayConfig.php）

```php
class WeChatPayConfig {
    const APPID = 'wx1234567890';
    const MCHID = '1234567890';
    const API_V3_KEY = 'your_api_v3_key';
    const PRIVATE_KEY_PATH = '/path/to/apiclient_key.pem';
    const CERT_SERIAL_NO = 'your_cert_serial_no';
    const NOTIFY_URL = 'https://yourdomain.com/notify';
}
```

### 2. 签名工具类（WeChatPayUtil.php）

```php
class WeChatPayUtil {
    public static function generateSignature($method, $url, $timestamp, $nonce, $body) {
        $signStr = "{$method}\n{$url}\n{$timestamp}\n{$nonce}\n{$body}\n";
        $privateKey = file_get_contents(WeChatPayConfig::PRIVATE_KEY_PATH);
        openssl_sign($signStr, $signature, $privateKey, OPENSSL_ALGO_SHA256);
        return base64_encode($signature);
    }

    public static function getAuthorizationHeader($method, $url, $body) {
        $timestamp = time();
        $nonce = bin2hex(random_bytes(16));
        $signature = self::generateSignature($method, $url, $timestamp, $nonce, $body);
        return sprintf(
            'WECHATPAY2-SHA256-RSA2048 mchid="%s",nonce_str="%s",timestamp="%d",serial_no="%s",signature="%s"',
            WeChatPayConfig::MCHID,
            $nonce,
            $timestamp,
            WeChatPayConfig::CERT_SERIAL_NO,
            $signature
        );
    }
}
```

### 3. 发起支付（createOrder.php）

```php
$body = json_encode([
    'appid' => WeChatPayConfig::APPID,
    'mchid' => WeChatPayConfig::MCHID,
    'description' => '测试商品',
    'out_trade_no' => 'ORDER_' . time(),
    'notify_url' => WeChatPayConfig::NOTIFY_URL,
    'amount' => ['total' => 1, 'currency' => 'CNY'],
    'payer' => ['openid' => 'user_openid']
]);

$url = 'https://api.mch.weixin.qq.com/v3/pay/transactions/jsapi';
$auth = WeChatPayUtil::getAuthorizationHeader('POST', parse_url($url, PHP_URL_PATH), $body);

$ch = curl_init($url);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
    'Accept: application/json',
    'Authorization: ' . $auth
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
curl_close($ch);

echo $response;
```

### 4. 查询订单（queryOrder.php）

```php
$transactionId = '4200001234567890';
$url = "https://api.mch.weixin.qq.com/v3/pay/transactions/id/{$transactionId}";
$auth = WeChatPayUtil::getAuthorizationHeader('GET', parse_url($url, PHP_URL_PATH), '');

$ch = curl_init($url);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Accept: application/json',
    'Authorization: ' . $auth
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
curl_close($ch);

echo $response;
```

### 5. 申请退款（refund.php）

```php
$body = json_encode([
    'out_trade_no' => 'ORDER_' . time(),
    'out_refund_no' => 'REFUND_' . time(),
    'reason' => '用户申请退款',
    'amount' => ['refund' => 1, 'total' => 1, 'currency' => 'CNY']
]);

$url = 'https://api.mch.weixin.qq.com/v3/refund/domestic/refunds';
$auth = WeChatPayUtil::getAuthorizationHeader('POST', parse_url($url, PHP_URL_PATH), $body);

$ch = curl_init($url);
curl_setopt($ch, CURLOPT_POST, true);
curl_setopt($ch, CURLOPT_POSTFIELDS, $body);
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Content-Type: application/json',
    'Accept: application/json',
    'Authorization: ' . $auth
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
$response = curl_exec($ch);
curl_close($ch);

echo $response;
```

### 6. 支付回调验证（notify.php）

```php
$signature = $_SERVER['HTTP_WECHATPAY_SIGNATURE'] ?? '';
$timestamp = $_SERVER['HTTP_WECHATPAY_TIMESTAMP'] ?? '';
$nonce = $_SERVER['HTTP_WECHATPAY_NONCE'] ?? '';
$body = file_get_contents('php://input');
$serial = $_SERVER['HTTP_WECHATPAY_SERIAL'] ?? '';

$signStr = "{$timestamp}\n{$nonce}\n{$body}\n";
$certPath = '/path/to/wechatpay_cert.pem';
$cert = file_get_contents($certPath);
$pubKey = openssl_get_publickey($cert);
$verify = openssl_verify($signStr, base64_decode($signature), $pubKey, OPENSSL_ALGO_SHA256);

if ($verify === 1) {
    $data = json_decode($body, true);
    // 处理业务逻辑...
    echo json_encode(['code' => 'SUCCESS', 'message' => 'OK']);
} else {
    http_response_code(400);
    echo json_encode(['code' => 'FAIL', 'message' => 'Signature verify failed']);
}
```

---

## 五、注意事项与最佳实践

- ✅ 使用 HTTPS 保证通信安全
- ✅ 所有请求必须携带签名和证书序列号
- ✅ 商户私钥严禁泄露，建议存放于服务器安全目录
- ✅ 回调通知需验证签名，防止伪造请求
- ✅ 所有金额单位为“分”，注意转换
- ✅ 建议封装统一请求类，便于维护和扩展
- ✅ 使用日志记录所有请求与响应，便于排查问题
- ✅ 退款、关单等操作建议加幂等控制，防止重复请求

---

## 六、推荐工具与资源

- 官方文档：https://pay.weixin.qq.com/wiki/doc/apiv3/
- SDK 推荐：
  - 官方：`wechatpay/wechatpay-php`
  - 社区：`yansongda/pay`
- 调试工具：微信支付沙箱环境、签名验证工具

---

通过以上整理，你可以快速掌握微信支付 V3 接口的核心规范与 PHP 实现方式。如需进一步封装为框架或集成到现有系统，建议结合日志、异常处理、配置管理等机制完善整体支付流程。
