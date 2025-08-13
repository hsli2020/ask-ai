# 编写一个golang函数，返回当前exe文件所在目录，带一个字符串参数，是一个相对路径，如果参数不为空，则将目录和参数表示的相对路径连接起来，返回这个完整文件路径

以下是符合你需求的 Go 语言实现。该函数采用 Go 惯用的错误返回模式 `(string, error)`，并自动处理跨平台路径分隔符。

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// GetExeDirPath 返回当前可执行文件所在目录。
// 若 relPath 不为空，则将其与目录拼接后返回完整路径。
func GetExeDirPath(relPath string) (string, error) {
	// 获取当前可执行文件的绝对路径
	exePath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 提取所在目录
	exeDir := filepath.Dir(exePath)

	// 如果提供了相对路径，则拼接；否则直接返回目录
	if relPath != "" {
		return filepath.Join(exeDir, relPath), nil
	}

	return exeDir, nil
}

func main() {
	// 示例 1：不传相对路径，仅返回目录
	dir, err := GetExeDirPath("")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("当前 exe 目录:", dir)

	// 示例 2：传入相对路径，返回完整路径
	fullPath, err := GetExeDirPath("config/app.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("完整路径:", fullPath)
}
```

### 🔍 核心说明
| 函数 | 作用 |
|------|------|
| `os.Executable()` | 获取当前运行程序的可执行文件路径（已解析符号链接，跨平台兼容） |
| `filepath.Dir()` | 提取路径中的目录部分（自动去除文件名） |
| `filepath.Join()` | 智能拼接路径，自动使用系统分隔符（Windows `\` / Linux/Mac `/`），并清理冗余的 `.` 或 `..` |

### 💡 注意事项
1. **路径规范化**：`filepath.Join` 会自动清理路径（如 `./config/../app.yaml` → `app.yaml`），无需额外调用 `filepath.Clean`。
2. **绝对路径保证**：`os.Executable()` 返回的通常是绝对路径，因此拼接结果也是绝对路径。若需强制转为绝对路径，可在返回前包装：`filepath.Abs(filepath.Join(exeDir, relPath))`。
3. **空参数处理**：代码显式判断了 `relPath != ""`，符合你的需求描述。实际上 `filepath.Join(exeDir, "")` 也会直接返回 `exeDir`，此处保留判断是为了语义更清晰。
4. **生产环境建议**：始终处理返回的 `error`，避免在获取路径失败时（如权限限制或程序被动态删除）引发不可预期行为。

如有需要适配特定场景（如返回不带尾斜杠的目录、或兼容软链接原始路径），可进一步调整。

