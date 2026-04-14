package main

import (
	"fmt"
	"go-course/utils"
	"os"
	"strconv"
)

func main() {
	count := 10000
	if len(os.Args) > 1 {
		parsed, err := strconv.Atoi(os.Args[1])
		if err != nil || parsed <= 0 {
			fmt.Println("Token 数量必须是正整数")
			return
		}
		count = parsed
	}

	filename := fmt.Sprintf("tokens_%d.txt", count)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println("创建文件失败", err)
		return
	}
	defer file.Close()

	for i := 1; i <= count; i++ {
		// 生成学生ID从 1 到 count 的 Token
		token, err := utils.GenToken(uint(i), fmt.Sprintf("student%d", i))
		if err != nil {
			fmt.Println("生成 Token 失败", err)
			return
		}
		// 写入到文件
		_, _ = file.WriteString(token + "\n")
	}
	fmt.Printf("成功生成 %d 个独立的 JWT Token 到 %s\n", count, filename)
}
