# 在Laravel应用中，取得当前用户有哪些方法？

在 Laravel 应用中，获取当前用户有多种方法，以下是常用的方式：

## 1. 使用 Auth Facade

```php
use Illuminate\Support\Facades\Auth;

// 获取当前认证用户
$user = Auth::user();

// 检查用户是否已认证
if (Auth::check()) {
    $user = Auth::user();
}

// 获取用户 ID
$userId = Auth::id();
```

## 2. 使用 auth() 辅助函数

```php
// 获取当前用户
$user = auth()->user();

// 或者直接调用
$user = auth('web')->user(); // 指定 guard

// 检查是否已认证
if (auth()->check()) {
    $user = auth()->user();
}

// 获取用户 ID
$userId = auth()->id();
```

## 3. 在控制器中通过依赖注入

```php
use Illuminate\Http\Request;

public function index(Request $request)
{
    // 通过请求对象获取用户
    $user = $request->user();
    
    // 或指定 guard
    $user = $request->user('api');
}
```

## 4. 在中间件中获取用户

```php
public function handle($request, Closure $next)
{
    $user = $request->user();
    // 或
    $user = Auth::user();
    
    return $next($request);
}
```

## 5. 在 Blade 模板中使用

```blade
{{-- 检查用户是否已认证 --}}
@auth
    <p>欢迎, {{ Auth::user()->name }}!</p>
@endauth

{{-- 或者 --}}
@if(auth()->check())
    <p>用户ID: {{ auth()->id() }}</p>
@endif

{{-- 直接访问用户属性 --}}
{{ auth()->user()?->name }}
```

## 6. 在模型或服务类中

```php
use Illuminate\Support\Facades\Auth;

class UserService 
{
    public function getCurrentUser()
    {
        return Auth::user();
    }
}
```

## 7. 处理多 Guard 场景

```php
// Web guard
$user = Auth::guard('web')->user();

// API guard
$user = Auth::guard('api')->user();

// 使用辅助函数指定 guard
$user = auth('api')->user();
```

## 8. 安全检查和错误处理

```php
// 安全的方式 - 避免调用 null 对象的方法
$user = Auth::user();
if ($user) {
    echo $user->name;
}

// 使用可选链（PHP 8.0+）
echo Auth::user()?->name;

// 或使用辅助函数
echo optional(Auth::user())->name;
```

## 9. 在 API 中使用 Sanctum/Passport

```php
// 使用 Sanctum
use Laravel\Sanctum\Http\Middleware\EnsureFrontendRequestsAreStateful;

// 在 API 控制器中
public function profile(Request $request)
{
    return $request->user(); // 返回当前 API 用户
}
```

## 最佳实践建议

1. **在控制器中**：推荐使用 `$request->user()` 或 `auth()->user()`
2. **在视图中**：使用 `@auth` 指令或 `auth()->user()`
3. **在服务类中**：使用 `Auth` Facade
4. **总是进行 null 检查**：确保用户已认证后再访问用户属性
5. **明确指定 guard**：在多 guard 应用中明确指定要使用的 guard

```php
// 推荐的安全写法
if (auth()->check()) {
    $name = auth()->user()->name;
    $email = auth()->user()->email;
}
```

这些方法可以根据你的具体使用场景选择最适合的方式。

