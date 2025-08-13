# 帮我编写两个golang函数，功能要求如下：
filename := AppDir(Datestamp("filename-%Y%M%D-%h%m%s.txt"))
// filename is like "/path/to/app/filename-20260521-152708.txt"

# Go 时间戳格式化和应用目录函数

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Datestamp 根据格式字符串生成时间戳
// 支持的格式占位符:
//
//	%Y - 4位年份 (2026)
//	%M - 2位月份 (01-12)
//	%D - 2位日期 (01-31)
//	%h - 2位小时, 24小时制 (00-23)
//	%m - 2位分钟 (00-59)
//	%s - 2位秒 (00-59)
func Datestamp(format string) string {
	now := time.Now()

	replacer := strings.NewReplacer(
		"%Y", fmt.Sprintf("%04d", now.Year()),
		"%M", fmt.Sprintf("%02d", now.Month()),
		"%D", fmt.Sprintf("%02d", now.Day()),
		"%h", fmt.Sprintf("%02d", now.Hour()),
		"%m", fmt.Sprintf("%02d", now.Minute()),
		"%s", fmt.Sprintf("%02d", now.Second()),
	)

	return replacer.Replace(format)
}

// AppDir 返回应用目录下的文件完整路径
// 默认使用可执行文件所在目录，可通过环境变量 APP_DIR 覆盖
func AppDir(filename string) string {
	// 优先使用环境变量指定的目录
	if dir := os.Getenv("APP_DIR"); dir != "" {
		return filepath.Join(dir, filename)
	}

	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		// 降级使用当前工作目录
		return filepath.Join(".", filename)
	}

	return filepath.Join(filepath.Dir(execPath), filename)
}
```

## 使用示例

```go
func main() {
	// 生成带时间戳的文件名
	filename := AppDir(Datestamp("filename-%Y%M%D-%h%m%s.txt"))
	fmt.Println(filename)
	// 输出示例: /path/to/app/filename-20260521-152708.txt

	// 其他格式示例
	fmt.Println(Datestamp("log-%Y%M%D.txt"))
	// 输出: log-20260521.txt

	fmt.Println(Datestamp("backup-%Y%M%D-%h%m%s.tar.gz"))
	// 输出: backup-20260521-152708.tar.gz
}
```

## 扩展：支持更多占位符的版本

```go
// DatestampEx 扩展版，支持更多占位符
//
//	%Y - 4位年份
//	%y - 2位年份
//	%M - 2位月份
//	%D - 2位日期
//	%h - 2位小时(24h)
//	%H - 2位小时(12h)
//	%m - 2位分钟
//	%s - 2位秒
//	%a - 星期缩写
//	%A - 星期全称
//	%b - 月份缩写
//	%B - 月份全称
func DatestampEx(format string) string {
	now := time.Now()
	replacer := strings.NewReplacer(
		"%Y", fmt.Sprintf("%04d", now.Year()),
		"%y", fmt.Sprintf("%02d", now.Year()%100),
		"%M", fmt.Sprintf("%02d", now.Month()),
		"%D", fmt.Sprintf("%02d", now.Day()),
		"%h", fmt.Sprintf("%02d", now.Hour()),
		"%H", fmt.Sprintf("%02d", now.Hour()%12+1),
		"%m", fmt.Sprintf("%02d", now.Minute()),
		"%s", fmt.Sprintf("%02d", now.Second()),
		"%a", now.Weekday().String()[:3],
		"%A", now.Weekday().String(),
		"%b", now.Month().String()[:3],
		"%B", now.Month().String(),
	)
	return replacer.Replace(format)
}
```

## 函数说明

| 函数 | 功能 | 返回值示例 |
|------|------|-----------|
| `Datestamp(format)` | 格式化当前时间为字符串 | `"filename-20260521-152708.txt"` |
| `AppDir(filename)` | 拼接应用目录路径 | `"/usr/local/app/filename.txt"` |

**注意**: `strings.NewReplacer` 是并发安全的，可以放心在多 goroutine 中调用。

# 编写另外一个golang函数：ArchiveFile(filename, dir string) string；这个函数把文件名改为“filename-YYYYMMDD-hhmmss.ext”，日期时间是文件的创建时间，并把文件移动到dir所指的目录中，返回值是新的文件名

在 Go 语言中，涉及文件系统的操作（如重命名、移动）通常需要返回 `error` 以处理文件不存在、权限不足等异常情况。因此，我将函数签名设计为 `ArchiveFile(filename, dir string) (string, error)`。

此外，考虑到实际开发中的几个常见坑（如 `.tar.gz` 等复合后缀的处理、跨磁盘分区的移动），我为你编写了一个**工业级健壮版本**：

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ArchiveFile 将文件归档（移动并重命名）到指定目录，添加时间戳
// filename: 原始文件的路径
// dir: 归档目标目录
// 返回值: 归档后的新文件完整路径，以及可能的错误
func ArchiveFile(filename, dir string) (string, error) {
	// 1. 提取文件名并分离基础名与后缀
	// 注意：这里使用自定义的 splitExt 而不是 filepath.Ext
	// 是为了正确处理类似 "backup.tar.gz" 变成 "backup-时间戳.tar.gz" 的情况
	// 如果用 filepath.Ext，会错误地变成 "backup.tar-时间戳.gz"
	base := filepath.Base(filename)
	name, ext := splitExt(base)

	// 2. 生成时间戳 (格式: YYYYMMDD-hhmmss)
	timestamp := Datestamp("%Y%M%D-%h%m%s")

	// 3. 拼接新的文件名
	newFilename := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

	// 4. 构建目标完整路径
	destPath := filepath.Join(dir, newFilename)

	// 5. 确保目标目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建归档目录失败: %w", err)
	}

	// 6. 移动文件 (注意处理跨文件系统/分区的移动)
	err := os.Rename(filename, destPath)
	if err != nil {
		// 如果是因为跨设备导致的错误 (例如 Linux 下的 EXDEV: invalid cross-device link)
		// 则降级使用 复制+删除 的方式
		if strings.Contains(err.Error(), "invalid cross-device link") {
			err = copyAndDelete(filename, destPath)
			if err != nil {
				return "", fmt.Errorf("跨磁盘归档文件失败: %w", err)
			}
		} else {
			return "", fmt.Errorf("归档文件失败: %w", err)
		}
	}

	return destPath, nil
}

// splitExt 分离文件名的基础部分和扩展名部分
// 例如: "file.tar.gz" -> "file", ".tar.gz"
// 例如: "file.txt"    -> "file", ".txt"
// 例如: "file"        -> "file", ""
func splitExt(filename string) (name, ext string) {
	idx := strings.Index(filename, ".")
	if idx == -1 {
		return filename, ""
	}
	return filename[:idx], filename[idx:]
}

// copyAndDelete 跨磁盘移动文件：先复制后删除
func copyAndDelete(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 执行复制
	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// 确保数据刷入磁盘
	if err = dstFile.Sync(); err != nil {
		return err
	}

	// 复制成功后删除原文件
	return os.Remove(src)
}

// ---------------- 以下是复用之前的 Datestamp 函数 ----------------

func Datestamp(format string) string {
	now := time.Now()
	replacer := strings.NewReplacer(
		"%Y", fmt.Sprintf("%04d", now.Year()),
		"%M", fmt.Sprintf("%02d", now.Month()),
		"%D", fmt.Sprintf("%02d", now.Day()),
		"%h", fmt.Sprintf("%02d", now.Hour()),
		"%m", fmt.Sprintf("%02d", now.Minute()),
		"%s", fmt.Sprintf("%02d", now.Second()),
	)
	return replacer.Replace(format)
}
```

### 使用示例

```go
func main() {
	// 假设当前目录下有一个 test.txt 文件
	// 先创建一个测试文件
	os.WriteFile("test.txt", []byte("hello world"), 0644)
	
	// 定义归档目录
	archiveDir := "./archives"

	// 执行归档
	newPath, err := ArchiveFile("test.txt", archiveDir)
	if err != nil {
		fmt.Printf("归档失败: %v\n", err)
		return
	}

	fmt.Println("归档成功，新文件路径为:", newPath)
	// 输出示例: 归档成功，新文件路径为: archives/test-20260521-152708.txt
}
```

### 代码亮点说明：

1. **为什么返回值加了 `error`？** 
   Go 的惯例是“不要忽略错误”。如果文件本来就不存在，或者没有写入权限，只返回 `string` 会让调用方产生“归档成功”的错觉。调用时只需简单判断 `if err != nil` 即可。
2. **智能后缀处理 (`splitExt`)**
   标准库的 `filepath.Ext("backup.tar.gz")` 只会返回 `.gz`。如果直接拼接，会生成难看的 `backup.tar-20260521-152708.gz`。自定义的 `splitExt` 会将其完美处理为 `backup-20260521-152708.tar.gz`。
3. **跨磁盘/分区移动 (`copyAndDelete`)**
   在 Linux 系统中，`os.Rename` 只能在**同一个文件系统（分区）**下重命名。如果你想把文件从 `/tmp` 移动到 `/data`（不同分区），`os.Rename` 会报错 `invalid cross-device link`。代码中自动捕获了这个错误，并优雅降级为“复制+删除”的策略，保证业务不中断。


# 时间戳不是当前系统的时间，而是文件的创建时间

在操作系统层面，获取文件的“创建时间”存在一个跨平台的痛点：
* **Windows** 和 **macOS** 原生支持获取文件真实的创建时间。
* **Linux** 传统文件系统（如 ext4 早期版本）**没有**真正的创建时间，只有访问时间、修改时间和状态改变时间。

为了不引入沉重的第三方 C 库，下面的代码采用**最佳妥协方案**：
1. 优先获取真实的创建时间。
2. 如果系统不支持（如大部分 Linux），自动降级使用文件的**修改时间** 作为时间戳来源。

### 代码实现

```go
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

// ArchiveFile 将文件归档（移动并重命名）到指定目录，时间戳取自文件本身的创建/修改时间
func ArchiveFile(filename, dir string) (string, error) {
	// 1. 获取文件的时间信息 (优先创建时间，降级修改时间)
	fileTime, err := getFileTime(filename)
	if err != nil {
		return "", fmt.Errorf("获取文件时间失败: %w", err)
	}

	// 2. 提取文件名并分离基础名与后缀 (处理 .tar.gz 等复合后缀)
	base := filepath.Base(filename)
	name, ext := splitExt(base)

	// 3. 根据文件的时间生成时间戳
	timestamp := FormatTimestamp(fileTime, "%Y%M%D-%h%m%s")

	// 4. 拼接新的文件名
	newFilename := fmt.Sprintf("%s-%s%s", name, timestamp, ext)

	// 5. 构建目标完整路径
	destPath := filepath.Join(dir, newFilename)

	// 6. 确保目标目录存在
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建归档目录失败: %w", err)
	}

	// 7. 移动文件 (处理跨磁盘分区的移动)
	err = os.Rename(filename, destPath)
	if err != nil {
		if strings.Contains(err.Error(), "invalid cross-device link") {
			err = copyAndDelete(filename, destPath)
			if err != nil {
				return "", fmt.Errorf("跨磁盘归档文件失败: %w", err)
			}
		} else {
			return "", fmt.Errorf("归档文件失败: %w", err)
		}
	}

	return destPath, nil
}

// getFileTime 获取文件的创建时间
// 注意：在 Linux 系统中如果无法获取创建时间，会降级返回修改时间
func getFileTime(filename string) (time.Time, error) {
	info, err := os.Stat(filename)
	if err != nil {
		return time.Time{}, err
	}

	// 尝试通过底层系统调用获取创建时间
	sysInfo := info.Sys()
	if sysInfo != nil {
		switch runtime.GOOS {
		case "darwin": // macOS
			if stat, ok := sysInfo.(*syscall.Stat_t); ok {
				return time.Unix(stat.Birthtimespec.Sec, int64(stat.Birthtimespec.Nsec)), nil
			}
		// Windows 和 Linux 的处理见下方注释说明
		}
	}

	// 降级方案：返回文件的修改时间
	return info.ModTime(), nil
}

// FormatTimestamp 根据格式字符串格式化指定的时间
func FormatTimestamp(t time.Time, format string) string {
	replacer := strings.NewReplacer(
		"%Y", fmt.Sprintf("%04d", t.Year()),
		"%M", fmt.Sprintf("%02d", t.Month()),
		"%D", fmt.Sprintf("%02d", t.Day()),
		"%h", fmt.Sprintf("%02d", t.Hour()),
		"%m", fmt.Sprintf("%02d", t.Minute()),
		"%s", fmt.Sprintf("%02d", t.Second()),
	)
	return replacer.Replace(format)
}

// splitExt 分离文件名的基础部分和扩展名部分 (解决 .tar.gz 问题)
func splitExt(filename string) (name, ext string) {
	idx := strings.Index(filename, ".")
	if idx == -1 {
		return filename, ""
	}
	return filename[:idx], filename[idx:]
}

// copyAndDelete 跨磁盘移动文件：先复制后删除
func copyAndDelete(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err = io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	if err = dstFile.Sync(); err != nil {
		return err
	}

	return os.Remove(src)
}
```

### 💡 关于 Windows 真实创建时间的扩展说明

在上面的纯标准库代码中，Windows 默认也会走降级逻辑（使用修改时间）。如果你**明确知道你的程序只运行在 Windows**，或者你愿意引入一个极其轻量的官方包，可以通过 `golang.org/x/sys/windows` 获取真实的 Windows 创建时间：

**如果你需要 Windows 精确创建时间，请做如下替换：**

1. 运行 `go get golang.org/x/sys/windows`
2. 在 `getFileTime` 的 `switch runtime.GOOS` 中增加 Windows 分支：

```go
import "golang.org/x/sys/windows"

// 在 switch runtime.GOOS 中增加：
case "windows":
    if stat, ok := sysInfo.(*windows.Win32FileAttributeData); ok {
       	// 将 Windows 特有的 Filetime 转换为 Go 的 time.Time
        return stat.CreationTime.Time(), nil
    }
```

对于 macOS，标准库原生支持，无需额外引入任何包即可完美获取 Birth Time（如上面代码所示）。对于 Linux 服务器环境，使用修改时间（`ModTime`）是业界的标准通用做法。
