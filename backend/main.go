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

	// 1. 静态资源路由 (必须在 NoRoute 之前)
	// 托管 assets 目录
	r.Static("/assets", "./frontend-dist/assets")
	// 托管 favicon
	r.StaticFile("/favicon.ico", "./frontend-dist/favicon.ico")

	// 2. SPA 路由支持 (兜底所有 404)
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		// 如果是 API 请求但没匹配到路由，返回 404 而不是 index.html
		if strings.HasPrefix(path, "/api") {
			c.JSON(404, gin.H{"error": "API not found"})
			return
		}
		// 返回前端入口
		c.File(filepath.Join("frontend-dist", "index.html"))
	})

	// 3. 路径遍历逻辑 (API)
	r.GET("/api/files/list", func(c *gin.Context) {
		relPath := c.DefaultQuery("path", "")
		var files []indexer.FileRecord
		
		// 统一处理路径前缀
		prefix := relPath
		if prefix != "" && !strings.HasSuffix(prefix, "/") {
			prefix += "/"
		}

		// 如果数据库为空，启动一次扫描 (临时补救)
		var count int64
		db.Model(&indexer.FileRecord{}).Count(&count)
		if count == 0 {
			root := os.Getenv("ROOT_DIR")
			if root == "" {
				root = "/data"
			}
			ix := indexer.NewIndexer(db, root)
			go ix.StartFullScan()
			c.JSON(200, []indexer.FileRecord{})
			return
		}

		// 修复 SQL：支持精准匹配当前目录及所有子项
		// 前端现在会处理层级过滤
		db.Where("full_path LIKE ?", relPath+"%").Find(&files)
		
		log.Printf("Query path: [%s], Found items: %d", relPath, len(files))
		c.JSON(200, files)
	})

	// 2. 文件上传 (分块)
	r.POST("/api/files/upload", upload.HandleChunkUpload)

	// 3. 搜索 API
	r.GET("/api/files/search", func(c *gin.Context) {
		query := c.Query("q")
		var results []indexer.FileRecord
		db.Where("name ILIKE ?", "%"+query+"%").Find(&results)
		c.JSON(200, results)
	})

	// 4. 多节点同步
	r.POST("/api/internal/sync", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "received"})
	})

	// 6. 视频实时转码流
	r.GET("/api/video/stream", transcode.StreamVideo)

	// 7. 获取播放器直连列表
	r.GET("/api/video/link", func(c *gin.Context) {
		path := c.Query("path")
		player := c.Query("player")
		link := transcode.GeneratePlayerLink(path, c.Request.Host, player)
		c.JSON(200, gin.H{"link": link})
	})

	// 8. 压缩/解压 API
	r.POST("/api/archive/compress", func(c *gin.Context) {
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

	r.POST("/api/archive/extract", func(c *gin.Context) {
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

	r.Run(":38866")
}
