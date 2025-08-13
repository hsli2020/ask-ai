# 在PHP中，可以用echo __FILE__, __METHOD__显示当前文件名和方法名，用来作日志时定位错误很方便，在golang中有无相似的方法，或如何做到这一点？

在 Go 语言中，确实有类似的方法来获取当前文件名和函数名，主要通过 `runtime` 包来实现。以下是几种常用的方式：

## 1. 使用 `runtime.Caller()` 获取调用信息

```go
package main

import (
    "fmt"
    "runtime"
)

func getCurrentLocation() (fileName string, funcName string, line int) {
    pc, file, line, ok := runtime.Caller(1) // 1 表示调用者的层级
    if !ok {
        return "", "", 0
    }
    
    // 获取函数名
    funcName = runtime.FuncForPC(pc).Name()
    
    // 只获取文件名，不包含完整路径
    fileName = file
    
    return fileName, funcName, line
}

func myFunction() {
    file, funcName, line := getCurrentLocation()
    fmt.Printf("文件: %s, 函数: %s, 行号: %d\n", file, funcName, line)
}

func main() {
    myFunction()
}
```

## 2. 创建日志辅助函数

```go
package main

import (
    "fmt"
    "path/filepath"
    "runtime"
)

// LogWithLocation 带位置信息的日志函数
func LogWithLocation(format string, args ...interface{}) {
    _, file, line, ok := runtime.Caller(1)
    if !ok {
        file = "???"
        line = 0
    }
    
    // 只获取文件名，去掉路径
    fileName := filepath.Base(file)
    
    message := fmt.Sprintf(format, args...)
    fmt.Printf("[%s:%d] %s\n", fileName, line, message)
}

func processData() {
    LogWithLocation("开始处理数据")
    LogWithLocation("处理完成，结果: %v", "success")
}

func main() {
    processData()
}
```

## 3. 封装成结构体方法

```go
package main

import (
    "fmt"
    "path/filepath"
    "runtime"
)

type Logger struct{}

func (l *Logger) Info(format string, args ...interface{}) {
    pc, file, line, ok := runtime.Caller(2) // 2 表示跳过 Logger.Info 和调用者
    if !ok {
        fmt.Printf("无法获取调用信息\n")
        return
    }
    
    funcName := runtime.FuncForPC(pc).Name()
    fileName := filepath.Base(file)
    
    message := fmt.Sprintf(format, args...)
    fmt.Printf("[%s:%s:%d] %s\n", fileName, funcName, line, message)
}

func myBusinessLogic() {
    logger := &Logger{}
    logger.Info("业务逻辑开始执行")
    logger.Info("参数值: %d", 42)
}

func main() {
    myBusinessLogic()
}
```

## 4. 使用第三方日志库

很多第三方日志库已经内置了这些功能，比如：

### 使用 `logrus`

```go
package main

import (
    "github.com/sirupsen/logrus"
)

func init() {
    logrus.SetReportCaller(true) // 启用调用者信息
}

func main() {
    logrus.Info("这是一条带位置信息的日志")
}
```

### 使用 `zap`

```go
package main

import (
    "go.uber.org/zap"
)

func main() {
    logger, _ := zap.NewDevelopment() // Development 模式包含调用者信息
    defer logger.Sync()
    
    logger.Info("这是一条带位置信息的日志")
}
```

## 关键点说明

1. **`runtime.Caller(n)`** 中的 `n` 表示调用栈的层级：
   - `0` = 当前函数
   - `1` = 调用当前函数的函数
   - `2` = 调用者的调用者

2. **性能考虑**：`runtime.Caller()` 有一定的性能开销，在性能敏感的代码中要谨慎使用。

3. **函数名格式**：Go 的函数名格式为 `包名.函数名`，例如 `main.myFunction`。

4. **文件路径**：`runtime.Caller()` 返回的是完整路径，通常使用 `filepath.Base()` 获取文件名。

这些方法可以很好地替代 PHP 中的 `__FILE__` 和 `__METHOD__`，帮助你在日志中定位代码位置。

