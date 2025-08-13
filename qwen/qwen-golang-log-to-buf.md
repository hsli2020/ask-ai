# 如何为golang的log模块增加一个功能，它的默认输出是一个内存缓冲区，当程序结束后，才将缓冲区的内容写入文件

在 Go 中，标准库的 `log` 包默认输出到 `os.Stderr`。要实现一个内存缓冲区并在程序结束时将内容写入文件，你可以通过以下步骤：

1. **创建一个内存缓冲区**（例如使用 `bytes.Buffer`）。
2. **将 `log` 的输出重定向到这个缓冲区**。
3. **在程序退出时（使用 `defer` 或 `os.Exit` 的钩子）将缓冲区内容写入文件**。

下面是一个完整的示例：

---

### ✅ 示例代码

```go
package main

import (
	"bytes"
	"log"
	"os"
)

func main() {
	// 1. 创建一个内存缓冲区
	var buf bytes.Buffer

	// 2. 将 log 的输出设置为这个缓冲区
	log.SetOutput(&buf)

	// 3. 注册一个 defer 函数，在程序退出前将日志写入文件
	defer func() {
		// 打开或创建日志文件
		file, err := os.Create("app.log")
		if err != nil {
			// 如果写入文件失败，至少把日志输出到 stderr，避免丢失
			log.SetOutput(os.Stderr)
			log.Printf("无法创建日志文件: %v", err)
			log.Print(buf.String()) // 输出缓冲内容到 stderr
			return
		}
		defer file.Close()

		// 将缓冲区内容写入文件
		if _, err := buf.WriteTo(file); err != nil {
			log.SetOutput(os.Stderr)
			log.Printf("写入日志文件失败: %v", err)
		}
	}()

	// 4. 正常使用 log
	log.Println("这是第一条日志")
	log.Println("这是第二条日志")

	// 程序结束时，defer 会自动执行，将日志写入 app.log
}
```

---

### 🔍 说明

- `log.SetOutput(io.Writer)` 允许你自定义日志输出目标。
- `bytes.Buffer` 实现了 `io.Writer` 接口，因此可以直接作为 `log` 的输出。
- 使用 `defer` 确保即使发生 panic，日志也会被写入（除非程序被 `os.Exit()` 强制退出）。
- 如果你使用 `os.Exit()`，**defer 不会执行**！如果你的程序可能调用 `os.Exit()`，可以考虑使用 `runtime.SetFinalizer` 或在退出前手动调用 flush 函数。

---

### ⚠️ 注意：`os.Exit()` 会跳过 defer

如果你的程序可能调用 `os.Exit(1)`，上面的 `defer` 不会执行。为了解决这个问题，可以封装一个退出函数：

```go
var logBuffer = &bytes.Buffer{}

func init() {
	log.SetOutput(logBuffer)
}

func safeExit(code int) {
	flushLog()
	os.Exit(code)
}

func flushLog() {
	file, err := os.Create("app.log")
	if err != nil {
		os.Stderr.WriteString("无法创建日志文件\n")
		os.Stderr.WriteString(logBuffer.String())
		return
	}
	defer file.Close()
	file.WriteString(logBuffer.String())
}
```

然后在需要退出的地方调用 `safeExit(1)` 而不是 `os.Exit(1)`。

---

### ✅ 总结

通过 `log.SetOutput()` + `bytes.Buffer` + `defer`，你可以轻松实现“延迟写入日志到文件”的功能。这是对标准 `log` 包的轻量级扩展，无需引入第三方日志库。

如需更高级功能（如日志级别、格式化、轮转等），建议使用 `log/slog`（Go 1.21+）或第三方库如 `zap`、`logrus`。