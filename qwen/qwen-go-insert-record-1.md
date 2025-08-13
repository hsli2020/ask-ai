# 🤵 编写一个golang函数，实现向数据库表中插入一条记录，参数一为表名，参数二为“字段名 => 值"的map
 
Qwen3-235B-A22B-2507

在 Go 中，你可以使用 database/sql 包结合一个数据库驱动（如 github.com/go-sql-driver/mysql 或 github.com/lib/pq）来实现向数据库表中插入一条记录。以下是一个通用的函数，接受表名和一个字段名到值的 map[string]interface{}，并动态生成 SQL 插入语句。 
示例：使用 MySQL 的通用插入函数 

package main

import (
	"database/sql"
	"fmt"
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
		values = append(values, v)
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

	fmt.Println("记录插入成功")
	return nil
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

	// 测试数据
	data := map[string]interface{}{
		"name":  "Alice",
		"age":   30,
		"email": "alice@example.com",
	}

	err = InsertRecord(db, "users", data)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}
}


注意事项： 

  1 SQL 注入风险：表名和字段名不能通过参数化方式传入，因此直接拼接存在风险。确保 tableName 和 data 的键是可信的（例如来自内部配置，而非用户输入）。
  2 数据库驱动：
      - 上面示例使用的是 MySQL，占位符是 ?。
      - 如果使用 PostgreSQL，则占位符应为 $1, $2, ...，需要修改 placeholders 的生成逻辑。
         
  3 字段类型兼容性：interface{} 可以接受大多数基本类型（string, int, bool 等），但复杂类型（如 struct）需要序列化（如 JSON）后再插入。
  4 错误处理：建议根据实际业务添加更详细的日志或重试机制。

PostgreSQL 版本（占位符不同） 

如果你使用的是 PostgreSQL，将占位符改为 $1, $2, ...： 
// PostgreSQL 版本的占位符生成
var placeholders []string
for i := 1; i <= len(data); i++ {
    placeholders = append(placeholders, fmt.Sprintf("$%d", i))
}

其余逻辑不变。 

这个函数是通用的，适合简单的 CRUD 场景。如需更强大功能（如 ORM），可考虑使用 GORM 等库。 

 
