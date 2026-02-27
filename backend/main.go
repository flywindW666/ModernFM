package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"modern-fm/internal/indexer"
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

	// 2. 初始化扫描器
	rootDir := os.Getenv("ROOT_DIR")
	if rootDir == "" {
		rootDir = "/data"
	}
	ix := indexer.NewIndexer(db, rootDir)

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
			path = filepath.ToSlash(filepath.Clean(path))
			if path == "." || path == "/" {
				path = ""
			}

			files, err := ix.ScanDir(path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
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

	log.Println("ModernFM Backend starting on :38866")
	r.Run(":38866")
}
