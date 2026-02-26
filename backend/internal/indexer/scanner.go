package indexer

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// FileRecord 定义文件系统的数据库记录模型
type FileRecord struct {
	ID        uint      `gorm:"primaryKey"`
	ParentID  uint      `gorm:"index"` // 父目录ID，加速目录遍历
	Name      string    `gorm:"size:255;index"`
	FullPath  string    `gorm:"uniqueIndex;column:full_path"`
	IsDir     bool      `gorm:"index"`
	Size      int64
	ModTime   time.Time
	Extension string    `gorm:"size:20;index"`
	Hash      string    `gorm:"size:40;index"` // SHA-1，用于排重或校验
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Indexer 索引管理器
type Indexer struct {
	db       *gorm.DB
	rootDir  string
	mu       sync.Mutex
	isRunning bool
}

// NewIndexer 创建索引器
func NewIndexer(db *gorm.DB, root string) *Indexer {
	return &Indexer{db: db, rootDir: root}
}

// StartFullScan 开始全量扫描索引
func (ix *Indexer) StartFullScan() {
	ix.mu.Lock()
	if ix.isRunning {
		ix.mu.Unlock()
		return
	}
	ix.isRunning = true
	ix.mu.Unlock()

	defer func() { ix.isRunning = false }()

	log.Printf("Starting full system scan: %s", ix.rootDir)
	
	// 使用 WalkDir 进行高效遍历 (Go 1.16+)
	err := filepath.WalkDir(ix.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("Error accessing path %q: %v\n", path, err)
			return nil // 忽略无法访问的目录
		}

		// 忽略根目录自身在数据库中的记录
		if path == ix.rootDir {
			return nil
		}

		info, _ := d.Info()
		relPath, _ := filepath.Rel(ix.rootDir, path)
		// 规范化路径分隔符为正斜杠，防止 Windows 环境干扰
		relPath = filepath.ToSlash(relPath)
		
		// 计算文件 Hash (可选，针对大文件可优化为只计算首尾)
		fileHash := ""
		if !d.IsDir() && info.Size() < 50*1024*1024 { // 仅对 50MB 以下文件生成完整 Hash
			fileHash = ix.calculateHash(path)
		}

		record := FileRecord{
			Name:      d.Name(),
			FullPath:  relPath,
			IsDir:     d.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: strings.ToLower(filepath.Ext(d.Name())),
			Hash:      fileHash,
			UpdatedAt: time.Now(),
		}

		// 使用 Upsert (更新或插入) 逻辑
		return ix.db.Where("full_path = ?", relPath).
			Assign(record).
			FirstOrCreate(&FileRecord{}).Error
	})
	
	if err != nil {
		log.Printf("WalkDir finished with error: %v", err)
	}
	
	log.Println("Full scan completed.")
}

// calculateHash 生成文件 SHA-1
func (ix *Indexer) calculateHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
