package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"modern-fm/internal/cache"
	"modern-fm/internal/indexer"
	"modern-fm/internal/upload"
	"modern-fm/internal/transcode"
	"modern-fm/internal/archive"
)

// FileRecord 模型 (Moved to internal/indexer)
/*
type FileRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"index"`
	FullPath  string    `gorm:"uniqueIndex"`
	IsDir     bool      `gorm:"index"`
	Size      int64
	ModTime   time.Time
	Extension string    `gorm:"index"`
}
*/

func main() {
	// 初始化 DB
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&indexer.FileRecord{})

	r := gin.Default()

	// 1. 路径遍历逻辑
	// ... (保留之前逻辑)

	// 4. 多节点同步：接收其它节点的变更通知
	r.POST("/api/internal/sync", func(c *gin.Context) {
		// 校验 X-Sync-Key
		// 根据变更事件更新本地索引或触发文件拉取
		c.JSON(200, gin.H{"status": "received"})
	})

	// 5. 远程挂载 API
	// ... (保留之前逻辑)

	// 6. 视频实时转码流
	r.GET("/api/video/stream", transcode.StreamVideo)

	// 7. 获取播放器直连列表
	// ... (保留之前逻辑)

	// 8. 压缩/解压 API
	r.POST("/api/archive/compress", func(c *gin.Context) {
		var req struct {
			Paths  []string `json:"paths"`
			Target string   `json:"target"` // zip 文件名
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
