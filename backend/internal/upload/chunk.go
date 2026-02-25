package upload

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ChunkUploadRequest 处理分块上传参数
type ChunkUploadRequest struct {
	Filename    string `form:"filename"`
	Identifier  string `form:"identifier"` // 文件唯一标识 (通常是 MD5)
	ChunkNumber int    `form:"chunkNumber"`
	TotalChunks int    `form:"totalChunks"`
	Path        string `form:"path"`       // 目标相对路径
}

const TempDir = "./uploads_temp"

// HandleChunkUpload 处理单个分块
func HandleChunkUpload(c *gin.Context) {
	var req ChunkUploadRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(400, gin.H{"error": "参数错误"})
		return
	}

	file, _ := c.FormFile("file")
	
	// 创建临时文件夹保存该文件的分块
	chunkDir := filepath.Join(TempDir, req.Identifier)
	os.MkdirAll(chunkDir, 0755)

	// 保存分块文件
	chunkPath := filepath.Join(chunkDir, strconv.Itoa(req.ChunkNumber))
	if err := c.SaveUploadedFile(file, chunkPath); err != nil {
		c.JSON(500, gin.H{"error": "保存分块失败"})
		return
	}

	// 检查是否所有分块都已到达
	if isUploadComplete(chunkDir, req.TotalChunks) {
		go mergeChunks(req) // 异步合并
		c.JSON(200, gin.H{"status": "merging", "message": "所有分块已上传，正在合并"})
	} else {
		c.JSON(200, gin.H{"status": "chunk_received", "chunk": req.ChunkNumber})
	}
}

func isUploadComplete(dir string, total int) bool {
	files, _ := os.ReadDir(dir)
	return len(files) == total
}

func mergeChunks(req ChunkUploadRequest) {
	targetPath := filepath.Join("/data", req.Path, req.Filename)
	os.MkdirAll(filepath.Dir(targetPath), 0755)
	
	destFile, _ := os.Create(targetPath)
	defer destFile.Close()

	for i := 1; i <= req.TotalChunks; i++ {
		chunkPath := filepath.Join(TempDir, req.Identifier, strconv.Itoa(i))
		sourceFile, _ := os.Open(chunkPath)
		io.Copy(destFile, sourceFile)
		sourceFile.Close()
	}

	// 清理临时文件
	os.RemoveAll(filepath.Join(TempDir, req.Identifier))
	fmt.Printf("File %s merged successfully\n", req.Filename)
}
