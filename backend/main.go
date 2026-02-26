package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
    "fmt"
    "net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"modern-fm/internal/archive"
	"modern-fm/internal/indexer"
	"modern-fm/internal/transcode"
	"modern-fm/internal/upload"
)

func main() {
	dsn := os.Getenv("DB_URL")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	db.AutoMigrate(&indexer.FileRecord{}, &indexer.SystemConfig{})

	root := os.Getenv("ROOT_DIR")
	if root == "" { root = "/data" }
	ix := indexer.NewIndexer(db, root)
	go ix.IndexDir("")

	r := gin.Default()

	r.Static("/assets", "./frontend-dist/assets")
	r.StaticFile("/favicon.ico", "./frontend-dist/favicon.ico")

	api := r.Group("/api")
	{
		api.GET("/files/list", func(c *gin.Context) {
			relPath := c.DefaultQuery("path", "")
			relPath = filepath.ToSlash(filepath.Clean(relPath))
			if relPath == "." || relPath == "/" { relPath = "" }
			go ix.IndexDir(relPath)
			var files []indexer.FileRecord
			db.Where("parent = ?", relPath).Find(&files)
			c.JSON(200, files)
		})

		// --- 实现下载功能 ---
		api.GET("/files/download", func(c *gin.Context) {
			relPath := c.Query("path")
			fullPath := filepath.Join(root, relPath)
            // 安全检查，防止路径穿越
            if !strings.HasPrefix(filepath.Clean(fullPath), filepath.Clean(root)) {
                c.JSON(403, gin.H{"error": "非法访问"})
                return
            }
			c.FileAttachment(fullPath, filepath.Base(fullPath))
		})

		// --- 实现上传功能 (分块上传已在 internal/upload 中定义) ---
		api.POST("/files/upload", upload.HandleChunkUpload)

		// --- 实现文件管理操作 ---
		api.POST("/files/action", func(c *gin.Context) {
			var req struct {
				Action string   `json:"action"`
				Paths  []string `json:"paths"`
				Dest   string   `json:"dest"`
                NewName string  `json:"newName"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "参数错误"})
				return
			}

			switch req.Action {
			case "delete":
				for _, p := range req.Paths {
					abs := filepath.Join(root, p)
					os.RemoveAll(abs)
					db.Where("full_path = ? OR full_path LIKE ?", p, p+"/%").Delete(&indexer.FileRecord{})
				}
			case "rename":
				if len(req.Paths) > 0 {
					oldRel := req.Paths[0]
					oldAbs := filepath.Join(root, oldRel)
					newAbs := filepath.Join(filepath.Dir(oldAbs), req.NewName)
					os.Rename(oldAbs, newAbs)
                    // 更新数据库 (简化逻辑: 重新扫描父目录)
                    go ix.IndexDir(filepath.ToSlash(filepath.Dir(oldRel)))
				}
			}
			c.JSON(200, gin.H{"status": "ok"})
		})

		api.POST("/system/rescan", func(c *gin.Context) {
			go ix.StartFullScan()
			c.JSON(200, gin.H{"status": "full_scan_started"})
		})

		api.GET("/video/stream", transcode.StreamVideo)
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
