# 编写一个golang函数，把给出的json格式的字符串进行紧缩，去除不必要的空格、换行等字符，返回紧缩后的json字串

你可以使用 Go 标准库中的 `encoding/json` 包来实现 JSON 字符串的紧缩（minify）。基本思路是：

1. 将输入的 JSON 字符串解析为通用的 `interface{}`（即进行反序列化）；
2. 然后再用 `json.Marshal` 重新序列化，这样会自动去除所有不必要的空白字符（如空格、换行、缩进等）；
3. 返回紧缩后的 JSON 字符串。

以下是完整的函数实现：

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CompactJSON 接收一个 JSON 字符串，返回其紧缩（minified）版本。
// 如果输入不是有效的 JSON，将返回错误。
func CompactJSON(jsonStr string) (string, error) {
	var raw interface{}

	// 解析原始 JSON 字符串
	decoder := json.NewDecoder(bytes.NewReader([]byte(jsonStr)))
	decoder.UseNumber() // 保留数字原始格式（避免 float64 转换问题）
	if err := decoder.Decode(&raw); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}

	// 重新编码为紧凑格式
	compactBytes, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal compact JSON: %w", err)
	}

	return string(compactBytes), nil
}

// 示例用法
func main() {
	prettyJSON := `{
		"name": "Alice",
		"age": 30,
		"active": true,
		"tags": [
			"go",
			"json"
		]
	}`

	compact, err := CompactJSON(prettyJSON)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Compact JSON:")
	fmt.Println(compact)
}
```

### 特点说明：

- 使用 `json.NewDecoder` + `UseNumber()` 可以避免大整数被转成 `float64`（例如 `12345678901234567890` 不会丢失精度）；
- `json.Marshal` 默认输出的就是最紧凑的格式（无空格、无换行）；
- 如果你**确定输入一定是合法 JSON**，也可以简化为直接使用 `json.Compact`（见下方替代方案）。

---

### 替代方案：使用 `json.Compact`

Go 标准库其实提供了更高效的方式——`json.Compact`，它不需要解析再重建结构，而是直接在字节层面压缩 JSON：

```go
func CompactJSONFast(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", fmt.Errorf("invalid JSON: %w", err)
	}
	return buf.String(), nil
}
```

> ✅ **推荐使用 `json.Compact`**，因为它更快、内存效率更高，且不会改变数值类型（比如不会把整数变成浮点数）。

---

### 最终推荐版本：

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// CompactJSON 使用 json.Compact 高效地压缩 JSON 字符串
func CompactJSON(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func main() {
	input := `{
		"hello": "world",
		"number": 42,
		"list": [1, 2, 3]
	}`

	output, err := CompactJSON(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(output) // {"hello":"world","number":42,"list":[1,2,3]}
}
```

这个版本简洁、高效、安全，是生产环境的首选。

# 第一种方法会改变json对象中元素的顺序，有办法避免吗？

是的，**第一种方法（使用 `map[string]interface{}` 解码）确实会改变 JSON 对象中键的顺序**，因为在 Go 1.12 之前，`map` 是无序的；即使在 Go 1.12+ 中 `map` 的遍历顺序是稳定的（但仍是随机种子决定的），**`encoding/json` 在解码到 `interface{}` 时默认使用 `map[string]interface{}`，而 `map` 本身不保证插入顺序**。

不过，从 **Go 1.18+ 开始，`encoding/json` 包引入了对保持字段顺序的支持**，但前提是：**你不能使用 `map[string]interface{}`，而应使用 `json.RawMessage` 或自定义类型，或者直接使用 `json.Compact`**。

---

### ✅ 正确答案：**使用 `json.Compact` 不会改变任何顺序！**

`json.Compact` 是**字节流级别的操作**，它只移除空白字符（空格、换行、制表符等），**完全保留原始 JSON 的结构和键顺序**。因此：

> **如果你的目标是“紧缩 JSON 且保持键顺序不变”，请务必使用 `json.Compact`，而不是先 decode 再 encode。**

---

### 为什么 `json.Compact` 能保持顺序？

- 它不解析 JSON 的语义结构（比如不构建 map 或 struct）；
- 它只是按 JSON 语法规范跳过空白字符，逐字输出有效 token；
- 因此原始输入中的 key 顺序、数字格式、字符串内容等全部原样保留（仅去除无关空白）。

---

### 示例对比

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func compactViaDecodeEncode(jsonStr string) (string, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(jsonStr), &v); err != nil {
		return "", err
	}
	b, err := json.Marshal(v)
	return string(b), err
}

func compactViaCompact(jsonStr string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(jsonStr)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func main() {
	input := `{
		"z": 1,
		"a": 2,
		"m": 3
	}`

	s1, _ := compactViaDecodeEncode(input)
	s2, _ := compactViaCompact(input)

	fmt.Println("Decode+Encode:", s1) // 可能输出 {"a":2,"m":3,"z":1}（顺序变了）
	fmt.Println("json.Compact: ", s2) // 一定输出 {"z":1,"a":2,"m":3}（顺序不变）
}
```

运行结果（典型）：
```
Decode+Encode: {"a":2,"m":3,"z":1}
json.Compact:  {"z":1,"a":2,"m":3}
```

---

### 结论

- ❌ 避免使用 “decode 到 `interface{}` 再 encode” 的方式来压缩 JSON —— **会丢失 key 顺序**；
- ✅ **始终使用 `json.Compact` 来实现 JSON 紧缩**，它高效、安全、且**100% 保持原始顺序和内容**（仅去除空白）；
- 如果你需要进一步处理 JSON（如修改字段），但又想保持顺序，可考虑使用第三方库如 [`github.com/tidwall/gjson`](https://github.com/tidwall/gjson) + [`sjson`](https://github.com/tidwall/sjson)，或解析为 `[]byte` token 流（如 `json.Decoder.Token()`）。

---

### 最终推荐函数（保持顺序 + 高效）

```go
import (
	"bytes"
	"encoding/json"
)

// CompactJSON 紧缩 JSON 字符串，保持字段顺序不变
func CompactJSON(s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(s)); err != nil {
		return "", err
	}
	return buf.String(), nil
}
```

这个函数满足你的所有需求：**紧缩、保序、高效、标准库实现**。
