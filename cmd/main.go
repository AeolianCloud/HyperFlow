package main

// @title           Hyperflow API
// @version         1.0
// @description     Hyperflow PVE 管理接口
// @host            localhost:8080
// @BasePath        /api/pve

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "hyperflow/docs"
	"hyperflow/internal/logger"
	"hyperflow/internal/operations"
	"hyperflow/internal/pve"
)

func main() {
	// 自动加载 .env 文件（文件不存在时忽略）
	_ = godotenv.Load()

	// 初始化 PveClient，缺失配置快速失败
	// 初始化 MySQL 连接
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		log.Fatal("MYSQL_DSN environment variable is required")
	}
	var err error
	dsn, err = normalizeMySQLDSN(dsn)
	if err != nil {
		log.Fatalf("invalid MySQL DSN: %v", err)
	}
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("failed to open MySQL connection: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	logWriter := logger.NewMySQLLogger(db)
	if err := logWriter.CreateTable(); err != nil {
		log.Fatalf("failed to create logs table: %v", err)
	}

	client, err := pve.NewClient(logWriter)
	if err != nil {
		log.Fatalf("PVE configuration error: %v", err)
	}

	nodesSvc := pve.NewNodesService(client)
	vmsSvc := pve.NewVmsService(client)
	storageSvc := pve.NewStorageService(client)

	// 初始化 operations 持久化层并确保表存在
	store := operations.NewMySQLStore(db)
	if err := store.CreateTable(); err != nil {
		log.Fatalf("failed to create operations table: %v", err)
	}
	operationsSvc := operations.NewService(store, vmsSvc, logWriter)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(requestLoggingMiddleware(logWriter))

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

	requestLoggerGlobal = logWriter

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	<-signals

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}

	logDrainCtx, logDrainCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer logDrainCancel()
	logWriter.Shutdown(logDrainCtx)
}
