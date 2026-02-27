package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"modern-fm/internal/indexer"
	"modern-fm/internal/cache"
)

func main() {
	// 1. 初始化数据库
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		dsn = "host=db user=modernfm_user password=secure_pass_123 dbname=modernfm port=5432 sslmode=disable"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}
	db.AutoMigrate(&indexer.FileRecord{})

	// 2. 初始化缓存
	rdb := cache.NewCache()

	// 3. 初始化扫描器
	rootDir := os.Getenv("ROOT_DIR")
	if rootDir == "" {
		rootDir = "/data"
	}
	ix := indexer.NewIndexer(db, rootDir)

	// 启动全量扫描（您原来的后台同步策略）
	go ix.StartFullScan()

	// 3. 路由设置
	r := gin.Default()

	// 静态文件
	r.Static("/assets", "./frontend-dist/assets")
	r.StaticFile("/favicon.ico", "./frontend-dist/favicon.ico")

	api := r.Group("/api")
	{
		// 统一的列表接口：支持按需加载目录树和文件列表
		api.GET("/files/list", func(c *gin.Context) {
			path := c.DefaultQuery("path", "")
			// 移除 Clean 以免将空路径变成 "." 导致扫描根目录失败
			if path == "/" || path == "." {
				path = ""
			}
			path = filepath.ToSlash(path)

			log.Printf("API Request: /files/list?path=%s", path)

			// 尝试从缓存读取
			cacheKey := "list:" + path
			if val, err := rdb.Get(c.Request.Context(), cacheKey); err == nil {
				c.Header("X-Cache", "HIT")
				c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(val))
				return
			}

			files, err := ix.ScanDir(path)
			if err != nil {
				log.Printf("ScanDir Error: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			
			// 写入缓存
			if jsonBytes, err := json.Marshal(files); err == nil {
				rdb.Set(c.Request.Context(), cacheKey, string(jsonBytes))
			}

			c.Header("X-Cache", "MISS")
			c.JSON(http.StatusOK, files)
		})

		// 下载
		api.GET("/files/download", func(c *gin.Context) {
			path := c.Query("path")
			full := filepath.Join(rootDir, path)
			c.FileAttachment(full, filepath.Base(full))
		})
	}

	// SPA 回退
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Status(404)
			return
		}
		c.File(filepath.Join("frontend-dist", "index.html"))
	})

	log.Println("XFileManager Backend starting on :38866")
	r.Run(":38866")
}
