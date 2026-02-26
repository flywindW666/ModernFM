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

func main() {
	// 初始化 DB
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&indexer.FileRecord{})

	r := gin.Default()

	// 1. 静态资源路由 (显式托管 assets 目录)
	r.Static("/assets", "./frontend-dist/assets")
	
	// 2. 托管 favicon
	r.StaticFile("/favicon.ico", "./frontend-dist/favicon.ico")

	// 3. API 路由
	api := r.Group("/api")
	{
		api.GET("/files/list", func(c *gin.Context) {
			relPath := c.DefaultQuery("path", "")
			var files []indexer.FileRecord
			
			var count int64
			db.Model(&indexer.FileRecord{}).Count(&count)
			if count == 0 {
				root := os.Getenv("ROOT_DIR")
				if root == "" { root = "/data" }
				ix := indexer.NewIndexer(db, root)
				go ix.StartFullScan()
				c.JSON(200, []indexer.FileRecord{})
				return
			}

			// 修复根目录模糊匹配问题：
			// 如果是根目录 ("")，我们只需查找所有不包含 "/" 的记录
			// 如果是子目录，则查找 path/ 开头的记录
			queryPath := relPath
			if queryPath != "" && !strings.HasSuffix(queryPath, "/") {
				queryPath += "/"
			}
			
			db.Where("full_path LIKE ?", queryPath+"%").Find(&files)
			c.JSON(200, files)
		})

		api.POST("/files/upload", upload.HandleChunkUpload)
		api.GET("/files/search", func(c *gin.Context) {
			query := c.Query("q")
			var results []indexer.FileRecord
			db.Where("name ILIKE ?", "%"+query+"%").Find(&results)
			c.JSON(200, results)
		})

		// 视频流与链接
		api.GET("/video/stream", transcode.StreamVideo)
		api.GET("/video/link", func(c *gin.Context) {
			path := c.Query("path")
			player := c.Query("player")
			link := transcode.GeneratePlayerLink(path, c.Request.Host, player)
			c.JSON(200, gin.H{"link": link})
		})

		// 归档操作
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

		// 同步接口
		api.POST("/internal/sync", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "received"})
		})
	}

	// 4. SPA 兜底路由
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File(filepath.Join("frontend-dist", "index.html"))
	})

	r.Run(":38866")
}
