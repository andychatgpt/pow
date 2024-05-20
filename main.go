package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"pow/controllers"
)

func main() {
	// 创建一个默认的Gin引擎
	r := gin.Default()
	// 定义一个GET请求的路由处理函数
	r.POST("/pow/get", controllers.Get)

	// 启动HTTP服务，监听在本地的8080端口
	log.Println(8081)
	r.Run(":9125")
}
