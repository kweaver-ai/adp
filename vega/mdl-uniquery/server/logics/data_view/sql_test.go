// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

// // 示例和使用演示
// func main() {
// 	fmt.Println("=== SQL 构建器演示 ===\n")

// 	// 示例 1: 简单查询
// 	fmt.Println("1. 简单查询:")
// 	builder1 := NewSQLBuilder("SELECT * FROM users")
// 	builder1.AddWhere("age > 18").AddWhere("status = 'active'")
// 	fmt.Printf("结果: %s\n\n", builder1.Build())

// 	// 示例 2: 已有 WHERE 的查询
// 	fmt.Println("2. 已有 WHERE 的查询:")
// 	builder2 := NewSQLBuilder("SELECT * FROM products WHERE price > 100")
// 	builder2.AddWhere("category = 'electronics'")
// 	fmt.Printf("结果: %s\n\n", builder2.Build())

// 	// 示例 3: 子查询
// 	fmt.Println("3. 子查询:")
// 	builder3 := NewSQLBuilder("(SELECT id, name FROM employees WHERE department = 'IT')")
// 	builder3.AddWhere("id IN (1, 2, 3)")
// 	fmt.Printf("结果: %s\n\n", builder3.Build())

// 	// 示例 4: 带别名的子查询
// 	fmt.Println("4. 带别名的子查询:")
// 	builder4 := NewSQLBuilder("(SELECT u.id, u.name FROM users u JOIN profiles p ON u.id = p.user_id) AS user_profiles")
// 	builder4.AddWhere("user_profiles.id > 100")
// 	fmt.Printf("结果: %s\n\n", builder4.Build())

// 	// 示例 5: 复杂查询（含 GROUP BY）
// 	fmt.Println("5. 复杂查询（含 GROUP BY）:")
// 	builder5 := NewSQLBuilder("SELECT department, COUNT(*) as count FROM employees GROUP BY department")
// 	builder5.AddWhere("hire_date > '2020-01-01'")
// 	fmt.Printf("结果: %s\n\n", builder5.Build())

// 	// 示例 6: 批量添加条件
// 	fmt.Println("6. 批量添加条件:")
// 	conditions := []string{"status = 'active'", "deleted = false", "created_at > NOW() - INTERVAL '30 days'"}
// 	builder6 := NewSQLBuilder("SELECT * FROM orders")
// 	builder6.AddWheres(conditions)
// 	fmt.Printf("结果: %s\n\n", builder6.Build())

// 	// 示例 7: 多层嵌套子查询
// 	fmt.Println("7. 多层嵌套子查询:")
// 	complexSubQuery := `
// 		(SELECT
// 			e.id,
// 			e.name,
// 			d.name as department_name
// 		FROM employees e
// 		LEFT JOIN departments d ON e.department_id = d.id
// 		WHERE e.salary > 50000)
// 	`
// 	builder7 := NewSQLBuilder(complexSubQuery)
// 	builder7.AddWhere("department_name = 'Engineering'")
// 	fmt.Printf("结果: %s\n\n", builder7.Build())

// 	// 错误处理示例
// 	fmt.Println("8. 空条件处理:")
// 	builder8 := NewSQLBuilder("SELECT * FROM test")
// 	builder8.AddWhere("") // 空条件
// 	builder8.AddWhere("value IS NOT NULL")
// 	fmt.Printf("结果: %s\n\n", builder8.Build())
// }

// // 单元测试函数（可选）
// func testSQLBuilder() {
// 	fmt.Println("=== 单元测试 ===\n")

// 	tests := []struct {
// 		name       string
// 		baseQuery  string
// 		conditions []string
// 		expected   string
// 	}{
// 		{
// 			"简单查询添加WHERE",
// 			"SELECT * FROM users",
// 			[]string{"age > 18"},
// 			"SELECT * FROM users WHERE age > 18",
// 		},
// 		{
// 			"已有WHERE添加AND条件",
// 			"SELECT * FROM products WHERE price > 100",
// 			[]string{"category = 'electronics'"},
// 			"SELECT * FROM products WHERE price > 100 AND category = 'electronics'",
// 		},
// 		{
// 			"子查询包装",
// 			"(SELECT id FROM table1)",
// 			[]string{"id > 10"},
// 			"(SELECT id FROM table1) AS subquery WHERE id > 10",
// 		},
// 	}

// 	for _, test := range tests {
// 		builder := NewSQLBuilder(test.baseQuery)
// 		builder.AddWheres(test.conditions)
// 		result := builder.Build()

// 		status := "✓"
// 		if result != test.expected {
// 			status = "✗"
// 		}

// 		fmt.Printf("%s %s\n", status, test.name)
// 		fmt.Printf("  输入: %s + %v\n", test.baseQuery, test.conditions)
// 		fmt.Printf("  期望: %s\n", test.expected)
// 		fmt.Printf("  实际: %s\n\n", result)
// 	}
// }
