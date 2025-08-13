#　总结一下支付宝API接口规范，重点讲解最常用的部分，并给出PHP代码调用例子

以下是支付宝 API 接口规范的实用总结，重点介绍常用接口、签名机制、PHP 调用示例及关键注意事项，适用于开发者快速上手集成。

---

## 一、支付宝 API 接口规范概览

### 1. 接口类型
- **公钥模式（推荐）**：使用应用私钥签名，支付宝公钥验签
- **证书模式**：适用于高安全要求场景，需使用应用证书和支付宝根证书

### 2. 请求方式
- 协议：HTTPS
- 字符集：UTF-8
- 签名算法：RSA2（SHA256WithRSA）
- 返回格式：JSON
- 接口地址：`https://openapi.alipay.com/gateway.do`

### 3. 公共请求参数
| 参数名         | 说明                     |
|----------------|--------------------------|
| app_id         | 应用ID                   |
| method         | 接口名称（如 alipay.trade.pay） |
| format         | 返回格式（固定为 JSON）   |
| charset        | 字符集（UTF-8）           |
| sign_type      | 签名类型（RSA2）          |
| sign           | 签名值                   |
| timestamp      | 时间戳（yyyy-MM-dd HH:mm:ss） |
| version        | 接口版本（默认 1.0）      |
| notify_url     | 异步通知地址（部分接口）  |
| biz_content    | 业务参数（JSON 字符串）   |

---

## 二、常用接口说明

| 接口名称                  | method 参数                   | 说明                     |
|---------------------------|-------------------------------|--------------------------|
| 手机网站支付              | alipay.trade.wap.pay          | H5 页面支付              |
| 电脑网站支付              | alipay.trade.page.pay         | PC 页面支付              |
| APP 支付                  | alipay.trade.app.pay          | 移动端原生支付           |
| 扫码支付（生成二维码）    | alipay.trade.precreate        | 生成二维码供用户扫码支付 |
| 统一收单交易支付接口      | alipay.trade.pay              | 条码/扫码支付            |
| 交易查询                  | alipay.trade.query            | 查询订单状态             |
| 交易关闭                  | alipay.trade.close            | 关闭未支付订单           |
| 交易退款                  | alipay.trade.refund           | 发起退款                 |
| 退款查询                  | alipay.trade.fastpay.refund.query | 查询退款状态         |
| 订单同步（可选）          | alipay.trade.order.settle     | 分账结算等               |

---

## 三、签名机制说明

### 1. 签名流程
1. 将所有公共参数 + biz_content 中的参数按 key 升序排序
2. 拼接为 `key=value` 形式，用 `&` 连接
3. 使用应用私钥进行 RSA2 签名
4. 将签名结果进行 URL 编码后放入 `sign` 参数

### 2. 验签流程（接收支付宝通知）
1. 去除 `sign` 和 `sign_type`
2. 按相同方式排序并拼接
3. 使用支付宝公钥验签

---

## 四、PHP 调用示例（使用官方 SDK）

### 1. 安装 SDK
```bash
composer require alipay/easysdk
```

### 2. 初始化配置
```php
use Alipay\EasySDK\Kernel\Config;
use Alipay\EasySDK\Kernel\Factory;

$config = new Config();
$config->appId = '你的APPID';
$config->merchantPrivateKey = file_get_contents('路径/应用私钥');
$config->alipayPublicKey = file_get_contents('路径/支付宝公钥');
$config->notifyUrl = 'https://yourdomain.com/notify';

Factory::setOptions($config);
```

### 3. 手机网站支付示例
```php
use Alipay\EasySDK\Payment\Wap;

$response = Factory::Payment()->wap()->pay('测试订单', 'OUT_TRADE_NO_001', 0.01, 'https://yourdomain.com/return');

// 输出 HTML 表单自动提交
echo $response->body;
```

### 4. 查询订单状态
```php
use Alipay\EasySDK\Payment\Common;

$query = Factory::Payment()->common()->query('OUT_TRADE_NO_001');

if ($query->tradeStatus === 'TRADE_SUCCESS') {
    echo "支付成功";
} else {
    echo "支付状态：" . $query->tradeStatus;
}
```

### 5. 退款示例
```php
use Alipay\EasySDK\Payment\Refund;

$refund = Factory::Payment()->refund()->refund('OUT_TRADE_NO_001', 0.01, '退款原因');

if ($refund->fundChange === 'Y') {
    echo "退款成功";
}
```

### 6. 异步通知处理
```php
use Alipay\EasySDK\Kernel\Util\VerifyUtil;

$params = $_POST;
$verify = VerifyUtil::verify($params);

if ($verify) {
    if ($_POST['trade_status'] === 'TRADE_SUCCESS') {
        // 更新订单状态
        echo 'success';
    }
} else {
    echo 'fail';
}
```

---

## 五、不使用 SDK 的签名示例（原生实现）

```php
function buildSign($params, $privateKey) {
    ksort($params);
    $stringToSign = http_build_query($params);
    $privateKey = openssl_pkey_get_private("-----BEGIN RSA PRIVATE KEY-----\n" . wordwrap($privateKey, 64, "\n", true) . "\n-----END RSA PRIVATE KEY-----");
    openssl_sign($stringToSign, $signature, $privateKey, OPENSSL_ALGO_SHA256);
    return base64_encode($signature);
}
```

---

## 六、注意事项与最佳实践

- ✅ 所有请求必须使用 HTTPS
- ✅ 私钥保存在服务器，严禁前端暴露
- ✅ 异步通知需验证签名并返回 `success`
- ✅ 商户订单号（out_trade_no）必须全局唯一
- ✅ 金额单位为元，最多两位小数
- ✅ 建议实现幂等性处理，防止重复处理通知
- ✅ 日志记录关键请求与响应，便于排查问题
- ✅ 沙箱环境可用于开发调试（https://open.alipay.com/develop/sandbox/app）

---

## 七、常见错误码参考

| 错误码       | 含义               |
|--------------|--------------------|
| ACQ.TRADE_HAS_SUCCESS | 交易已支付       |
| ACQ.TRADE_HAS_CLOSE   | 交易已关闭       |
| ACQ.INVALID_PARAMETER | 参数错误         |
| ACQ.SIGN_NOT_MATCH    | 签名不匹配       |

---

通过上述内容，你可以快速掌握支付宝 API 的核心规范与实现方式。建议优先使用官方 SDK 简化开发，同时注意签名、通知验证和订单幂等处理等关键环节，确保支付流程稳定可靠。
