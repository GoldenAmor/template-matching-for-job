package main

import (
	"fmt"
)

func main() {
	router := RoutesController()
	err := router.Run(":6060")
	if err != nil {
		fmt.Println("启动服务失败:", err)
	}
}
