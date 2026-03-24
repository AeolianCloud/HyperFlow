package main

// @title           Hyperflow API
// @version         1.0
// @description     Hyperflow PVE 管理接口
// @host            localhost:8080
// @BasePath        /api/pve

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "hyperflow/docs"
	"hyperflow/internal/operations"
	"hyperflow/internal/pve"
)

func main() {
	// 自动加载 .env 文件（文件不存在时忽略）
	_ = godotenv.Load()

	// 初始化 PveClient，缺失配置快速失败
	client, err := pve.NewClient()
	if err != nil {
		log.Fatalf("PVE configuration error: %v", err)
	}

	nodesSvc := pve.NewNodesService(client)
	vmsSvc := pve.NewVmsService(client)
	storageSvc := pve.NewStorageService(client)

	// 初始化 MySQL 连接
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN environment variable is required")
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open MySQL connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}

	// 初始化 operations 持久化层并确保表存在
	store := operations.NewMySQLStore(db)
	if err := store.CreateTable(); err != nil {
		log.Fatalf("failed to create operations table: %v", err)
	}
	operationsSvc := operations.NewService(store, vmsSvc)

	r := gin.Default()

	// 注册所有 /api/pve/* 路由
	apipve := r.Group("/api/pve")
	{
		registerNodesRoutes(apipve.Group("/nodes"), nodesSvc)

		nodes := apipve.Group("/nodes/:node")
		registerVmsRoutes(nodes.Group("/vms"), vmsSvc)

		registerStorageRoutes(apipve.Group("/storage"), storageSvc)

		registerOperationsRoutes(apipve.Group("/operations"), operationsSvc)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run(":8080")
}
