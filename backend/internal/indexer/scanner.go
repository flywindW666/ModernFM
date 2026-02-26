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
	Parent    string    `gorm:"size:512;index"` // 父目录路径，用于快速查询直接子项
	Name      string    `gorm:"size:255;index"`
	FullPath  string    `gorm:"uniqueIndex;column:full_path"`
	IsDir     bool      `gorm:"index"`
	Size      int64
	ModTime   time.Time
	Extension string    `gorm:"size:100;index"`
	Hash      string    `gorm:"size:40;index"` 
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Indexer struct {
	db       *gorm.DB
	rootDir  string
	mu       sync.Mutex
	isRunning bool
}

func NewIndexer(db *gorm.DB, root string) *Indexer {
	return &Indexer{db: db, rootDir: filepath.Clean(root)}
}

func (ix *Indexer) StartFullScan() {
	ix.mu.Lock()
	if ix.isRunning {
		ix.mu.Unlock()
		return
	}
	ix.isRunning = true
	ix.mu.Unlock()

	defer func() { ix.isRunning = false }()

	log.Printf("[Indexer] Starting deep scan of: %s", ix.rootDir)
	
	count := 0
	err := filepath.WalkDir(ix.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("[Indexer] Error accessing %s: %v", path, err)
			return nil
		}

		if path == ix.rootDir {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		
		relPath, _ := filepath.Rel(ix.rootDir, path)
		relPath = filepath.ToSlash(relPath)
		
		// 计算父路径
		parent := filepath.ToSlash(filepath.Dir(relPath))
		if parent == "." {
			parent = ""
		}

		fileExt := ""
		if !d.IsDir() {
			fileExt = strings.ToLower(filepath.Ext(d.Name()))
		}

		record := FileRecord{
			Parent:    parent,
			Name:      d.Name(),
			FullPath:  relPath,
			IsDir:     d.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: fileExt,
			UpdatedAt: time.Now(),
		}

		// 使用更稳定的 Upsert 逻辑 (修复 Save 操作在冲突时的行为)
		if err := ix.db.Where("full_path = ?", relPath).Assign(record).FirstOrCreate(&FileRecord{}).Error; err != nil {
			// 如果 FirstOrCreate 依然报错，尝试强制 Save (Update or Create)
			if err := ix.db.Save(&record).Error; err != nil {
				log.Printf("[Indexer] DB Error for %s: %v", relPath, err)
			}
		}
		
		count++
		if count % 1000 == 0 {
			log.Printf("[Indexer] Indexed %d items...", count)
		}
		return nil
	})
	
	if err != nil {
		log.Printf("[Indexer] Scan finished with fatal error: %v", err)
	}
	log.Printf("[Indexer] Completed. Total items indexed: %d", count)
}

func (ix *Indexer) calculateHash(path string) string {
	f, err := os.Open(path)
	if err != nil { return "" }
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil { return "" }
	return hex.EncodeToString(h.Sum(nil))
}
