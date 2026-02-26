package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"modern-fm/internal/archive"
	"modern-fm/internal/indexer"
	"modern-fm/internal/transcode"
	"modern-fm/internal/upload"
)

// SystemConfig 系统配置表
type SystemConfig struct {
	ID    uint   `gorm:"primaryKey"`
	Key   string `gorm:"uniqueIndex"`
	Value string
}

func main() {
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&indexer.FileRecord{}, &SystemConfig{})

	// 自动检查并触发首次全量扫描
	var initialScan SystemConfig
	result := db.Where("key = ?", "initial_scan_completed").First(&initialScan)
	if result.Error != nil {
		log.Println("[System] Detecting first run. Triggering automatic full scan...")
		root := os.Getenv("ROOT_DIR")
		if root == "" { root = "/data" }
		ix := indexer.NewIndexer(db, root)
		
		// 启动异步扫描并在完成后标记
		go func() {
			ix.StartFullScan()
			db.Create(&SystemConfig{Key: "initial_scan_completed", Value: "true"})
			log.Println("[System] Initial scan completed and recorded to database.")
		}()
	}

	r := gin.Default()

	r.Static("/assets", "./frontend-dist/assets")
	r.StaticFile("/favicon.ico", "./frontend-dist/favicon.ico")

	api := r.Group("/api")
	{
		// 1. 获取文件列表
		api.GET("/files/list", func(c *gin.Context) {
			relPath := c.DefaultQuery("path", "")
			relPath = filepath.ToSlash(filepath.Clean(relPath))
			if relPath == "." || relPath == "/" {
				relPath = ""
			}

			var files []indexer.FileRecord
			// 精确匹配父目录字段
			db.Where("parent = ?", relPath).Find(&files)
			
			log.Printf("[API] Path: '%s', Found: %d", relPath, len(files))
			c.JSON(200, files)
		})

		// 2. 强制重新扫描
		api.POST("/system/rescan", func(c *gin.Context) {
			root := os.Getenv("ROOT_DIR")
			if root == "" { root = "/data" }
			ix := indexer.NewIndexer(db, root)
			go ix.StartFullScan()
			c.JSON(200, gin.H{"status": "scan_started"})
		})

		api.POST("/files/upload", upload.HandleChunkUpload)
		api.GET("/files/search", func(c *gin.Context) {
			query := c.Query("q")
			var results []indexer.FileRecord
			db.Where("name ILIKE ?", "%"+query+"%").Limit(100).Find(&results)
			c.JSON(200, results)
		})

		api.GET("/video/stream", transcode.StreamVideo)
		api.GET("/video/link", func(c *gin.Context) {
			path := c.Query("path")
			player := c.Query("player")
			link := transcode.GeneratePlayerLink(path, c.Request.Host, player)
			c.JSON(200, gin.H{"link": link})
		})

		api.POST("/archive/compress", func(c *gin.Context) {
			var req struct {
				Paths  []string `json:"paths"`
				Target string   `json:"target"`
			}
			c.ShouldBindJSON(&req)
			targetZip := filepath.Join("/data", req.Target)
			var sources []string
			for _, p := range req.Paths {
				sources = append(sources, filepath.Join("/data", p))
			}
			if err := archive.CompressZip(sources, targetZip); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "compressed"})
		})

		api.POST("/archive/extract", func(c *gin.Context) {
			var req struct {
				Source string `json:"source"`
				Dest   string `json:"dest"`
			}
			c.ShouldBindJSON(&req)
			src := filepath.Join("/data", req.Source)
			dest := filepath.Join("/data", req.Dest)
			if err := archive.Extract(src, dest); err != nil {
				c.JSON(500, gin.H{"error": err.Error()})
				return
			}
			c.JSON(200, gin.H{"status": "extracted"})
		})
	}

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "Not Found"})
			return
		}
		c.File(filepath.Join("frontend-dist", "index.html"))
	})

	r.Run(":38866")
}
