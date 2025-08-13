# 用php语言编写演示程序，展示JWT的工作原理

以下是一个使用 PHP 编写的完整演示程序，展示 JWT（JSON Web Token）的工作原理，包括结构说明、生成、验证、解码及完整流程演示。

---

## 一、JWT 结构说明

JWT 由三部分组成，使用 `.` 分隔：

1. **Header（头部）**：指定算法和令牌类型
2. **Payload（负载）**：包含声明（claims），如用户信息、过期时间等
3. **Signature（签名）**：使用密钥对前两部分进行签名，防止篡改

示例结构：
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c
```

---

## 二、核心类：JWT 实现（不使用第三方库）

```php
<?php
class JWT {
    private static $algorithm = 'HS256';
    private static $secret = 'your-256-bit-secret';

    // Base64Url 编码
    private static function base64UrlEncode($data) {
        return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
    }

    // Base64Url 解码
    private static function base64UrlDecode($data) {
        return base64_decode(strtr($data, '-_', '+/'));
    }

    // 生成 JWT
    public static function generate($payload) {
        $header = [
            'alg' => self::$algorithm,
            'typ' => 'JWT'
        ];

        $headerEncoded = self::base64UrlEncode(json_encode($header));
        $payloadEncoded = self::base64UrlEncode(json_encode($payload));

        $signature = hash_hmac('sha256', "$headerEncoded.$payloadEncoded", self::$secret, true);
        $signatureEncoded = self::base64UrlEncode($signature);

        return "$headerEncoded.$payloadEncoded.$signatureEncoded";
    }

    // 验证 JWT
    public static function verify($token) {
        $parts = explode('.', $token);
        if (count($parts) !== 3) return false;

        [$headerEncoded, $payloadEncoded, $signatureEncoded] = $parts;

        $signature = hash_hmac('sha256', "$headerEncoded.$payloadEncoded", self::$secret, true);
        $signatureExpected = self::base64UrlEncode($signature);

        if (!hash_equals($signatureExpected, $signatureEncoded)) return false;

        $payload = json_decode(self::base64UrlDecode($payloadEncoded), true);
        if (isset($payload['exp']) && $payload['exp'] < time()) return false;

        return $payload;
    }

    // 解码 JWT（不验证签名）
    public static function decode($token) {
        $parts = explode('.', $token);
        if (count($parts) !== 3) return null;

        $header = json_decode(self::base64UrlDecode($parts[0]), true);
        $payload = json_decode(self::base64UrlDecode($parts[1]), true);

        return ['header' => $header, 'payload' => $payload];
    }
}
```

---

## 三、使用示例：生成、验证、解码

```php
<?php
require_once 'JWT.php';

// 1. 生成 JWT
$payload = [
    'sub' => '1234567890',
    'name' => 'John Doe',
    'iat' => time(),
    'exp' => time() + 3600 // 1小时后过期
];

$token = JWT::generate($payload);
echo "生成的 JWT: $token\n\n";

// 2. 验证 JWT
$verified = JWT::verify($token);
if ($verified) {
    echo "✓ JWT 验证成功\n";
    echo "用户: " . $verified['name'] . "\n";
} else {
    echo "✗ JWT 验证失败\n";
}

// 3. 解码 JWT（不验证）
$decoded = JWT::decode($token);
echo "\nJWT 结构:\n";
echo "Header: " . json_encode($decoded['header'], JSON_PRETTY_PRINT) . "\n";
echo "Payload: " . json_encode($decoded['payload'], JSON_PRETTY_PRINT) . "\n";
```

---

## 四、篡改演示：签名验证机制

```php
<?php
// 篡改 payload
$parts = explode('.', $token);
$payload = json_decode(JWT::base64UrlDecode($parts[1]), true);
$payload['name'] = 'Hacker';
$parts[1] = JWT::base64UrlEncode(json_encode($payload));
$tamperedToken = implode('.', $parts);

echo "\n篡改后的 Token: $tamperedToken\n";
if (JWT::verify($tamperedToken)) {
    echo "✗ 验证通过（不安全）\n";
} else {
    echo "✓ 验证失败（签名被篡改）\n";
}
```

---

## 五、完整演示流程（可选网页版）

你可以将上述代码整合进一个 `demo.php` 文件，通过浏览器访问展示完整的生成、验证、篡改流程，并输出结构化信息。

---

## 六、注意事项与最佳实践

- 使用强密钥（至少 256 位）
- 始终验证签名和过期时间（exp）
- 不要将敏感信息放入 payload（JWT 可被解码）
- 使用 HTTPS 传输
- 推荐使用成熟库（如 `firebase/php-jwt`）

安装示例：
```bash
composer require firebase/php-jwt
```

---

该演示程序完整展示了 JWT 的生成、验证、解码及防篡改机制，适合用于教学、原型开发或理解 JWT 原理。
