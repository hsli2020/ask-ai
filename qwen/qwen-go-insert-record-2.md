# 🤵 加入对复杂类型的处理，若为复杂类型，先转换为JSON格式

Qwen3-235B-A22B-2507 
好的，我来完善这个函数，加入对复杂类型的处理，将复杂类型转换为JSON格式存储。 

package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

// InsertRecord 向指定表插入一条记录
// db: 数据库连接
// tableName: 表名
// data: 字段名 => 值的映射
func InsertRecord(db *sql.DB, tableName string, data map[string]interface{}) error {
	if len(data) == 0 {
		return fmt.Errorf("数据为空，无法插入")
	}

	// 构建字段名和占位符
	var columns []string
	var placeholders []string
	var values []interface{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?") // MySQL 使用 ?

		// 处理复杂类型，转换为JSON
		processedValue, err := processValue(v)
		if err != nil {
			return fmt.Errorf("处理字段 %s 的值时出错: %w", k, err)
		}
		values = append(values, processedValue)
	}

	// 构造 SQL 语句
	sqlStmt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// 执行插入
	_, err := db.Exec(sqlStmt, values...)
	if err != nil {
		return fmt.Errorf("插入记录失败: %w", err)
	}

	fmt.Printf("记录插入成功到表 %s\n", tableName)
	return nil
}

// processValue 处理值，如果是复杂类型则转换为JSON字符串
func processValue(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}

	// 获取值的反射类型
	v := reflect.ValueOf(value)
	kind := v.Kind()

	// 基本类型直接返回
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String:
		return value, nil

	case reflect.Slice, reflect.Array, reflect.Map, reflect.Struct:
		// 复杂类型：切片、数组、映射、结构体，转换为JSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("序列化为JSON失败: %w", err)
		}
		return string(jsonData), nil

	case reflect.Ptr:
		// 处理指针类型
		if v.IsNil() {
			return nil, nil
		}
		return processValue(v.Elem().Interface())

	default:
		// 其他类型也尝试转换为JSON
		jsonData, err := json.Marshal(value)
		if err != nil {
			return nil, fmt.Errorf("序列化为JSON失败: %w", err)
		}
		return string(jsonData), nil
	}
}

// InsertRecordWithJSONConversion 带有更多选项的版本，允许指定JSON字段
func InsertRecordWithJSONConversion(db *sql.DB, tableName string, data map[string]interface{}, jsonFields ...string) error {
	if len(data) == 0 {
		return fmt.Errorf("数据为空，无法插入")
	}

	// 创建JSON字段集合
	jsonFieldSet := make(map[string]bool)
	for _, field := range jsonFields {
		jsonFieldSet[field] = true
	}

	// 构建字段名和占位符
	var columns []string
	var placeholders []string
	var values []interface{}

	for k, v := range data {
		columns = append(columns, k)
		placeholders = append(placeholders, "?")

		// 如果指定了该字段需要JSON转换，或者自动判断为复杂类型
		var processedValue interface{}
		var err error
		
		if jsonFieldSet[k] {
			// 强制JSON转换
			processedValue, err = forceJSONConversion(v)
		} else {
			// 自动判断
			processedValue, err = processValue(v)
		}
		
		if err != nil {
			return fmt.Errorf("处理字段 %s 的值时出错: %w", k, err)
		}
		values = append(values, processedValue)
	}

	// 构造 SQL 语句
	sqlStmt := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(placeholders, ", "),
	)

	// 执行插入
	result, err := db.Exec(sqlStmt, values...)
	if err != nil {
		return fmt.Errorf("插入记录失败: %w", err)
	}

	// 获取插入的ID（如果有）
	lastInsertID, err := result.LastInsertId()
	if err == nil {
		fmt.Printf("记录插入成功，ID: %d\n", lastInsertID)
	} else {
		fmt.Printf("记录插入成功到表 %s\n", tableName)
	}
	
	return nil
}

// forceJSONConversion 强制将任何值转换为JSON字符串
func forceJSONConversion(value interface{}) (interface{}, error) {
	if value == nil {
		return "null", nil
	}

	jsonData, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("序列化为JSON失败: %w", err)
	}
	return string(jsonData), nil
}

// 示例结构体
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	ZipCode string `json:"zip_code"`
}

type User struct {
	Name     string            `json:"name"`
	Age      int               `json:"age"`
	Email    string            `json:"email"`
	Address  Address           `json:"address"`
	Hobbies  []string          `json:"hobbies"`
	Metadata map[string]string `json:"metadata"`
}

// 示例用法
func main() {
	// 替换为你的数据库连接信息
	dsn := "user:password@tcp(localhost:3306)/yourdb"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// 测试1: 基本类型和复杂类型混合
	data1 := map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
		"address": Address{
			Street:  "123 Main St",
			City:    "New York",
			ZipCode: "10001",
		},
		"hobbies": []string{"reading", "swimming", "coding"},
		"metadata": map[string]string{
			"source": "website",
			"level":  "premium",
		},
	}

	err = InsertRecord(db, "users", data1)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}

	// 测试2: 使用指定JSON字段的版本
	user := User{
		Name:  "Bob",
		Age:   25,
		Email: "bob@example.com",
		Address: Address{
			Street:  "456 Oak Ave",
			City:    "Los Angeles",
			ZipCode: "90210",
		},
		Hobbies: []string{"gaming", "music"},
		Metadata: map[string]string{
			"source": "app",
			"level":  "standard",
		},
	}

	data2 := map[string]interface{}{
		"user_info": user,           // 整个结构体
		"status":    "active",       // 基本类型
		"score":     95.5,           // 基本类型
		"tags":      []string{"A", "B"}, // 切片
	}

	// 指定哪些字段需要JSON转换
	err = InsertRecordWithJSONConversion(db, "user_profiles", data2, "user_info", "tags")
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}
}

主要改进： 

  1 processValue 函数： 
      - 使用反射判断值的类型
      - 基本类型（bool, int, float, string等）直接返回
      - 复杂类型（struct, slice, map, array等）转换为JSON字符串
      - 处理nil值和指针类型
         

  2 forceJSONConversion 函数： 
      - 强制将任何值转换为JSON字符串
      - 即使是基本类型也会被JSON化（如字符串会变成带引号的JSON字符串）
         

  3 增强版本 InsertRecordWithJSONConversion： 
      - 支持指定哪些字段需要强制JSON转换
      - 提供更灵活的控制
         

  4 错误处理： 
      - 详细的错误信息，包含字段名
      - JSON序列化失败的处理
         

  5 返回信息： 
      - 显示插入的记录ID（如果有自增主键）
         

使用场景： 

  - 复杂对象存储：用户配置、地址信息、元数据等
  - 动态字段：不确定结构的数据
  - NoSQL风格存储：在关系型数据库中存储JSON数据
     

注意事项： 

  - 数据库字段需要是支持JSON或TEXT类型的（如MySQL的JSON、TEXT，PostgreSQL的JSON等）
  - JSON转换会增加存储空间
  - 查询JSON字段时需要使用数据库特定的JSON函数
  - 性能考虑：频繁的JSON序列化/反序列化会有开销
