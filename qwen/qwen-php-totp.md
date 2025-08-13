# 🤵 编写一个PHP函数，生成TOTP(Time-based one-time password)，生成的TOTP必须能得到Google Authenticator的确认

你可以使用 PHP 编写一个符合 RFC 6238 标准的 TOTP（基于时间的一次性密码）生成函数，使其与 Google Authenticator 兼容。Google Authenticator 使用 SHA-1 哈希算法、30 秒的时间步长和 6 位数字的密码长度。

以下是一个完整的、可运行的 PHP 函数来生成 TOTP：

```php
<?php

/**
 * 生成与 Google Authenticator 兼容的 TOTP
 *
 * @param string $secret Base32 编码的密钥（通常来自二维码）
 * @param int $time_step 时间步长（默认 30 秒）
 * @param int $digits 生成的密码位数（默认 6）
 * @param string $algorithm 哈希算法（默认 sha1）
 * @return string 6 位数字的 TOTP
 */
function generateTOTP($secret, $time_step = 30, $digits = 6, $algorithm = 'sha1') {
    // 1. 将 Base32 编码的密钥转换为原始字节
    $secret = strtoupper(str_replace(['=', ' '], '', $secret)); // 去掉填充和空格
    $base32chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
    $base32charsFlipped = array_flip(str_split($base32chars));

    $binaryString = "";
    for ($i = 0; $i < strlen($secret); $i++) {
        $value = $base32charsFlipped[$secret[$i]];
        $binaryString .= str_pad(decbin($value), 5, '0', STR_PAD_LEFT);
    }

    $binaryArray = str_split($binaryString, 8);
    $secretKey = "";
    foreach ($binaryArray as $byte) {
        if (strlen($byte) == 8) {
            $secretKey .= chr(bindec($byte));
        }
    }

    // 2. 计算时间戳对应的时间计数器
    $timeCounter = floor(time() / $time_step);

    // 3. 将时间计数器编码为 8 字节的大端整数
    $timeBytes = pack('N*', 0, $timeCounter); // 8 字节：高 32 位为 0，低 32 位为时间计数器

    // 4. 使用 HMAC-SHA1 计算哈希值
    $hash = hash_hmac($algorithm, $timeBytes, $secretKey, true); // raw output

    // 5. 动态截断（Dynamic Truncation）
    $offset = ord($hash[19]) & 0xf; // 最后一个字节的低 4 位作为偏移量
    $binary =  (
        ((ord($hash[$offset + 0]) & 0x7f) << 24) |
        ((ord($hash[$offset + 1]) & 0xff) << 16) |
        ((ord($hash[$offset + 2]) & 0xff) << 8) |
        (ord($hash[$offset + 3]) & 0xff)
    );

    // 6. 取模得到指定位数的数字
    $totp = $binary % pow(10, $digits);

    // 7. 格式化为固定长度的字符串（补前导零）
    return str_pad($totp, $digits, '0', STR_PAD_LEFT);
}

// 示例使用
$secret = "JBSWY3DPEHPK3PXP"; // 示例密钥（Base32），通常来自二维码
$totp = generateTOTP($secret);
echo "TOTP: " . $totp . "\n";

?>
```

---

### 🔍 说明：

1. **密钥格式**：
   - Google Authenticator 使用 Base32 编码的密钥。
   - 示例：`JBSWY3DPEHPK3PXP` 是一个合法的 Base32 密钥。

2. **时间步长**：
   - 默认为 30 秒，这是 Google Authenticator 的标准。

3. **哈希算法**：
   - Google Authenticator 默认使用 `SHA-1`，所以 `$algorithm = 'sha1'`。

4. **动态截断（RFC 6238）**：
   - 从 HMAC 结果中提取 4 字节整数，然后取模生成 6 位数字。

5. **时间同步**：
   - 确保服务器时间与 UTC 同步，否则 TOTP 会不匹配。

---

### ✅ 如何与 Google Authenticator 配对？

1. 在 Google Authenticator 中添加手动账户。
2. 输入账户名（如 `user@example.com`）。
3. 输入 `$secret`（如 `JBSWY3DPEHPK3PXP`）。
4. 选择时间模式（基于时间）。
5. 生成的 TOTP 应该与你的 PHP 函数输出一致。

---

### 🔧 可选：生成二维码（便于扫码添加）

你可以使用 `endroid/qr-code` Composer 包生成二维码：

```bash
composer require endroid/qr-code
```

然后生成二维码 URL：

```php
$issuer = 'MyApp';
$account = 'user@example.com';
$secret = 'JBSWY3DPEHPK3PXP';

$uri = urlencode("otpauth://totp/{$issuer}:{$account}?secret={$secret}&issuer={$issuer}");

echo "Scan this URL in Google Authenticator:\n";
echo "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data={$uri}\n";
```

---

### ✅ 测试建议

- 使用在线 TOTP 生成器（如 https://www.google.com/chart/interactive/docs/qr_tutorial.html）验证输出。
- 或使用另一个已知正确的实现进行比对。

---

这个函数完全符合 RFC 6238，并与 Google Authenticator 兼容，无需外部依赖（除了 `hash_hmac` 和 `pack`，这些是 PHP 内置函数）。
