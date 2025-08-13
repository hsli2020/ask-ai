# 用php代码演示refreshToken、accessToken相关授权的用法

以下是一个使用 PHP 实现 OAuth 2.0 中 accessToken 与 refreshToken 授权机制的完整示例，包括核心概念、代码实现、使用场景及最佳实践建议。

---

## 一、核心概念

- **accessToken**：短期有效，用于访问受保护资源。
- **refreshToken**：长期有效，用于在 accessToken 过期后获取新的 accessToken。
- **授权流程**：
  1. 用户授权后获取 accessToken 和 refreshToken。
  2. 使用 accessToken 访问资源。
  3. accessToken 过期后，使用 refreshToken 换取新的 accessToken。

---

## 二、OAuth 客户端类实现（含 Token 管理）

```php
<?php

class OAuthClient
{
    private $clientId;
    private $clientSecret;
    private $tokenUrl;
    private $accessToken;
    private $refreshToken;
    private $tokenExpiresAt;

    public function __construct($clientId, $clientSecret, $tokenUrl)
    {
        $this->clientId = $clientId;
        $this->clientSecret = $clientSecret;
        $this->tokenUrl = $tokenUrl;
    }

    // 初始授权获取 token
    public function authorize($code, $redirectUri)
    {
        $response = $this->requestToken([
            'grant_type' => 'authorization_code',
            'code' => $code,
            'redirect_uri' => $redirectUri,
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret,
        ]);

        $this->storeTokens($response);
        return $response;
    }

    // 使用 refreshToken 刷新 accessToken
    public function refreshAccessToken()
    {
        if (!$this->refreshToken) {
            throw new Exception('No refresh token available');
        }

        $response = $this->requestToken([
            'grant_type' => 'refresh_token',
            'refresh_token' => $this->refreshToken,
            'client_id' => $this->clientId,
            'client_secret' => $this->clientSecret,
        ]);

        $this->storeTokens($response);
        return $response;
    }

    // 获取有效 accessToken，若过期则刷新
    public function getValidAccessToken()
    {
        if ($this->isAccessTokenExpired()) {
            $this->refreshAccessToken();
        }
        return $this->accessToken;
    }

    // 使用 accessToken 调用受保护资源
    public function requestResource($url, $method = 'GET', $data = [])
    {
        $token = $this->getValidAccessToken();
        $headers = ["Authorization: Bearer $token"];

        $ch = curl_init($url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);

        if ($method === 'POST') {
            curl_setopt($ch, CURLOPT_POST, true);
            curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($data));
        }

        $response = curl_exec($ch);
        $httpCode = curl_getinfo($ch, CURLINFO_HTTP_CODE);
        curl_close($ch);

        if ($httpCode === 401) {
            // token 无效，尝试刷新后重试
            $this->refreshAccessToken();
            $token = $this->getValidAccessToken();
            $headers[0] = "Authorization: Bearer $token";
            // 此处可再次发起请求（略）
        }

        return json_decode($response, true);
    }

    // 内部方法：请求 token
    private function requestToken($params)
    {
        $ch = curl_init($this->tokenUrl);
        curl_setopt($ch, CURLOPT_POST, true);
        curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($params));
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
        $response = curl_exec($ch);
        curl_close($ch);

        $data = json_decode($response, true);
        if (!$data || isset($data['error'])) {
            throw new Exception('Token request failed: ' . ($data['error'] ?? 'Unknown error'));
        }

        return $data;
    }

    // 存储 token 信息
    private function storeTokens($data)
    {
        $this->accessToken = $data['access_token'];
        $this->refreshToken = $data['refresh_token'] ?? $this->refreshToken;
        $expiresIn = $data['expires_in'] ?? 3600;
        $this->tokenExpiresAt = time() + $expiresIn - 60; // 提前 60 秒视为过期
    }

    // 判断 accessToken 是否过期
    private function isAccessTokenExpired()
    {
        return time() >= $this->tokenExpiresAt;
    }

    // 获取当前 token 状态（用于调试）
    public function getTokenStatus()
    {
        return [
            'has_access_token' => !empty($this->accessToken),
            'has_refresh_token' => !empty($this->refreshToken),
            'expires_at' => $this->tokenExpiresAt,
            'is_expired' => $this->isAccessTokenExpired(),
        ];
    }
}
```

---

## 三、使用示例

```php
<?php
require 'OAuthClient.php';

$client = new OAuthClient(
    'your_client_id',
    'your_client_secret',
    'https://api.example.com/oauth/token'
);

// 1. 初次授权（例如从回调中获取 code）
$code = $_GET['code'] ?? null;
if ($code) {
    $client->authorize($code, 'https://yourapp.com/callback');
}

// 2. 访问受保护资源
$data = $client->requestResource('https://api.example.com/user/profile');
print_r($data);

// 3. 查看当前 token 状态
print_r($client->getTokenStatus());
```

---

## 四、最佳实践与安全建议

- ✅ 所有通信使用 HTTPS。
- ✅ 将 clientSecret 和 token 存储在服务器端，避免暴露给前端。
- ✅ 使用数据库或加密存储持久化 token。
- ✅ accessToken 设置较短有效期（如 1 小时），refreshToken 可更长（如 30 天）。
- ✅ 刷新 token 时验证 refreshToken 是否有效，防止重放攻击。
- ✅ 实现 token 自动刷新机制，避免用户频繁重新授权。
- ✅ 记录授权日志，便于审计和异常检测。

---

该示例展示了如何在 PHP 中安全、高效地管理 OAuth 2.0 的 accessToken 与 refreshToken，适用于大多数第三方授权场景。你可根据实际业务需求扩展存储、日志、异常处理等模块。


# 服务器端如何生成、验证、存储token呢？

在服务器端实现 OAuth 2.0 的 accessToken 和 refreshToken 机制，需要涵盖以下核心模块：Token 生成、验证、存储、刷新与安全管理。以下是一个完整、可运行的 PHP 实现示例，涵盖最佳实践和安全建议。

---

## 一、核心模块概览

| 模块         | 功能说明                          |
|--------------|-----------------------------------|
| Token 生成   | 使用 JWT 或随机字符串生成 token   |
| Token 验证   | 校验签名、有效期、权限等          |
| Token 存储   | 使用数据库或 Redis 存储 token 信息 |
| Token 刷新   | 使用 refreshToken 获取新 accessToken |
| 安全机制     | HTTPS、加密存储、防重放攻击等     |

---

## 二、数据库结构示例（MySQL）

```sql
CREATE TABLE oauth_tokens (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id INT NOT NULL,
    access_token VARCHAR(512) NOT NULL,
    refresh_token VARCHAR(512) NOT NULL,
    access_expires_at INT NOT NULL,
    refresh_expires_at INT NOT NULL,
    client_id VARCHAR(128) NOT NULL,
    scope VARCHAR(256),
    is_revoked TINYINT(1) DEFAULT 0,
    created_at INT NOT NULL,
    updated_at INT NOT NULL,
    last_used_at INT,
    INDEX idx_access_token (access_token),
    INDEX idx_refresh_token (refresh_token),
    INDEX idx_user_id (user_id)
);
```

---

## 三、Token 管理类（含生成、验证、存储、刷新）

```php
<?php

class TokenManager
{
    private $pdo;
    private $jwtSecret;
    private $accessTokenExpire = 3600;      // 1 小时
    private $refreshTokenExpire = 2592000;  // 30 天

    public function __construct(PDO $pdo, $jwtSecret)
    {
        $this->pdo = $pdo;
        $this->jwtSecret = $jwtSecret;
    }

    // 生成 token 对
    public function generateTokens($userId, $clientId, $scope = '')
    {
        $accessToken = $this->generateJWT($userId, $this->accessTokenExpire, $scope);
        $refreshToken = bin2hex(random_bytes(32));

        $this->storeTokens([
            'user_id' => $userId,
            'client_id' => $clientId,
            'access_token' => $accessToken,
            'refresh_token' => $refreshToken,
            'access_expires_at' => time() + $this->accessTokenExpire,
            'refresh_expires_at' => time() + $this->refreshTokenExpire,
            'scope' => $scope,
        ]);

        return [
            'access_token' => $accessToken,
            'refresh_token' => $refreshToken,
            'expires_in' => $this->accessTokenExpire,
            'token_type' => 'Bearer',
        ];
    }

    // 验证 accessToken
    public function validateAccessToken($token)
    {
        try {
            $payload = $this->verifyJWT($token);
            if (!$payload) return false;

            $stmt = $this->pdo->prepare("SELECT * FROM oauth_tokens WHERE access_token = ? AND is_revoked = 0 AND access_expires_at > ?");
            $stmt->execute([$token, time()]);
            $record = $stmt->fetch(PDO::FETCH_ASSOC);

            if (!$record) return false;

            // 更新最后使用时间
            $this->pdo->prepare("UPDATE oauth_tokens SET last_used_at = ? WHERE id = ?")
                ->execute([time(), $record['id']]);

            return $payload;
        } catch (Exception $e) {
            return false;
        }
    }

    // 刷新 accessToken
    public function refreshAccessToken($refreshToken, $clientId)
    {
        $stmt = $this->pdo->prepare("SELECT * FROM oauth_tokens WHERE refresh_token = ? AND client_id = ? AND is_revoked = 0 AND refresh_expires_at > ?");
        $stmt->execute([$refreshToken, $clientId, time()]);
        $record = $stmt->fetch(PDO::FETCH_ASSOC);

        if (!$record) {
            throw new Exception('Invalid or expired refresh token');
        }

        // 可选：刷新后使旧 token 失效（提升安全性）
        $this->revokeToken($record['id']);

        return $this->generateTokens($record['user_id'], $clientId, $record['scope']);
    }

    // 撤销 token
    public function revokeToken($tokenId)
    {
        $this->pdo->prepare("UPDATE oauth_tokens SET is_revoked = 1, updated_at = ? WHERE id = ?")
            ->execute([time(), $tokenId]);
    }

    // 撤销用户所有 token（如登出）
    public function revokeAllUserTokens($userId)
    {
        $this->pdo->prepare("UPDATE oauth_tokens SET is_revoked = 1, updated_at = ? WHERE user_id = ?")
            ->execute([time(), $userId]);
    }

    // 清理过期 token
    public function cleanupExpiredTokens()
    {
        $this->pdo->exec("DELETE FROM oauth_tokens WHERE access_expires_at < " . (time() - 86400));
    }

    // 生成 JWT
    private function generateJWT($userId, $expire, $scope)
    {
        $header = json_encode(['alg' => 'HS256', 'typ' => 'JWT']);
        $payload = json_encode([
            'user_id' => $userId,
            'iat' => time(),
            'exp' => time() + $expire,
            'scope' => $scope,
        ]);

        $base64Header = $this->base64UrlEncode($header);
        $base64Payload = $this->base64UrlEncode($payload);
        $signature = hash_hmac('sha256', "$base64Header.$base64Payload", $this->jwtSecret, true);
        $base64Signature = $this->base64UrlEncode($signature);

        return "$base64Header.$base64Payload.$base64Signature";
    }

    // 验证 JWT
    private function verifyJWT($token)
    {
        $parts = explode('.', $token);
        if (count($parts) !== 3) return false;

        [$header, $payload, $signature] = $parts;
        $validSignature = hash_hmac('sha256', "$header.$payload", $this->jwtSecret, true);
        if ($this->base64UrlDecode($signature) !== $validSignature) return false;

        $payloadData = json_decode($this->base64UrlDecode($payload), true);
        if (!$payloadData || ($payloadData['exp'] ?? 0) < time()) return false;

        return $payloadData;
    }

    // 存储 token 到数据库
    private function storeTokens($data)
    {
        $stmt = $this->pdo->prepare("INSERT INTO oauth_tokens 
            (user_id, client_id, access_token, refresh_token, access_expires_at, refresh_expires_at, scope, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)");
        $stmt->execute([
            $data['user_id'],
            $data['client_id'],
            $data['access_token'],
            $data['refresh_token'],
            $data['access_expires_at'],
            $data['refresh_expires_at'],
            $data['scope'],
            time(),
            time(),
        ]);
    }

    // Base64 URL 编码/解码
    private function base64UrlEncode($data)
    {
        return rtrim(strtr(base64_encode($data), '+/', '-_'), '=');
    }

    private function base64UrlDecode($data)
    {
        return base64_decode(strtr($data, '-_', '+/'));
    }
}
```

---

## 四、使用示例

```php
<?php
$pdo = new PDO('mysql:host=localhost;dbname=oauth', 'root', 'password');
$tokenManager = new TokenManager($pdo, 'your_jwt_secret_key');

// 生成 token
$tokens = $tokenManager->generateTokens(123, 'client_abc', 'read write');

// 验证 accessToken
$payload = $tokenManager->validateAccessToken($tokens['access_token']);

// 刷新 token
$newTokens = $tokenManager->refreshAccessToken($tokens['refresh_token'], 'client_abc');

// 撤销 token
$tokenManager->revokeAllUserTokens(123);
```

---

## 五、安全最佳实践

- ✅ 所有通信使用 HTTPS
- ✅ JWT 密钥使用强随机字符串，并定期更换
- ✅ refreshToken 一次性使用或绑定设备指纹
- ✅ 设置合适的 token 有效期（access: 1h，refresh: 7~30 天）
- ✅ 存储时加密敏感字段（如 refresh_token）
- ✅ 实现登录设备管理、异常检测、审计日志
- ✅ 定期清理过期/撤销的 token 记录

---

该实现方案适用于大多数 OAuth 2.0 授权服务器场景，具备良好的安全性、可扩展性和可维护性。你可根据业务需求进一步集成 Redis 缓存、多设备支持、权限控制等模块。
