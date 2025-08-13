# 编写一个golang函数，将一个整数转换为36进制字符串

你可以使用 Go 语言编写一个函数，将一个整数转换为 36 进制的字符串。36 进制使用数字 `0-9` 和字母 `a-z`（小写）来表示 0 到 35 的值。

以下是实现该功能的 Go 函数：

```go
package main

import (
	"fmt"
	"strings"
)

// IntToBase36 将一个整数转换为36进制字符串
func IntToBase36(n int64) string {
	if n == 0 {
		return "0"
	}

	negative := false
	if n < 0 {
		negative = true
		n = -n
	}

	digits := "0123456789abcdefghijklmnopqrstuvwxyz"
	var result strings.Builder

	for n > 0 {
		remainder := n % 36
		result.WriteByte(digits[remainder])
		n /= 36
	}

	// 反转字符串
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	if negative {
		return "-" + string(runes)
	}
	return string(runes)
}

// 示例使用
func main() {
	fmt.Println(IntToBase36(0))       // 输出: 0
	fmt.Println(IntToBase36(35))      // 输出: z
	fmt.Println(IntToBase36(36))      // 输出: 10
	fmt.Println(IntToBase36(12345))   // 输出: 9ix
	fmt.Println(IntToBase36(-12345))  // 输出: -9ix
}
```

### 说明：

- 使用 `int64` 类型以支持较大的整数。
- 处理负数情况，添加负号。
- 使用 `strings.Builder` 提高性能（可选，也可用 `[]byte` 或 `string` 拼接）。
- 最后反转字符串，因为我们在计算时是从低位到高位的。

### 输出示例：

```
0
z
10
9ix
-9ix
```

这个函数是健壮的，支持正数、负数和零。你可以根据需要将其集成到你的项目中。