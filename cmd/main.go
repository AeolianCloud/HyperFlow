package main

// @title           Hyperflow API
// @version         1.0
// @description     Hyperflow PVE 管理接口
// @host            localhost:8080
// @BasePath        /api/pve

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "hyperflow/docs"
	"hyperflow/internal/pve"
)

func main() {
	// 自动加载 .env 文件（文件不存在时忽略）
	_ = godotenv.Load()

	// 6.2: 启动时初始化 PveClient，缺失配置快速失败
	client, err := pve.NewClient()
	if err != nil {
		log.Fatalf("PVE configuration error: %v", err)
	}

	nodesSvc := pve.NewNodesService(client)
	vmsSvc := pve.NewVmsService(client)
	storageSvc := pve.NewStorageService(client)

	r := gin.Default()

	// 6.1: 注册所有 /api/pve/* 路由
	apipve := r.Group("/api/pve")
	{
		registerNodesRoutes(apipve.Group("/nodes"), nodesSvc)

		nodes := apipve.Group("/nodes/:node")
		registerVmsRoutes(nodes.Group("/vms"), vmsSvc)

		registerStorageRoutes(apipve.Group("/storage"), storageSvc)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}
