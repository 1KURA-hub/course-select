package main

import (
	"fmt"
	"go-course/utils"
	"os"
)

func main() {
	file, err := os.Create("tokens_10000.txt")
	if err != nil {
		fmt.Println("创建文件失败", err)
		return
	}
	defer file.Close()

	for i := 1; i <= 10000; i++ {
		// 生成学生ID从 1 到 10000 的 Token
		token, err := utils.GenToken(uint(i), fmt.Sprintf("student%d", i))
		if err != nil {
			fmt.Println("生成 Token 失败", err)
			return
		}
		// 写入到文件
		_, _ = file.WriteString(token + "\n")
	}
	fmt.Println("成功生成 10000 个独立的 JWT Token 到 tokens_10000.txt")
}

