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
	Parent    string    `gorm:"size:512;index"` 
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

	log.Printf("[Indexer] Starting high-performance bulk scan: %s", ix.rootDir)
	
	const batchSize = 100
	var batch []FileRecord
	count := 0

	flushBatch := func() {
		if len(batch) == 0 {
			return
		}
		// 使用 Claused OnConflict 实现高效大批量 Upsert
		// 这里由于不同数据库方言支持不同，先用事务包围以提升性能
		ix.db.Transaction(func(tx *gorm.DB) error {
			for _, item := range batch {
				tx.Where("full_path = ?", item.FullPath).Assign(item).FirstOrCreate(&FileRecord{})
			}
			return nil
		})
		batch = nil
	}

	filepath.WalkDir(ix.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
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
		parent := filepath.ToSlash(filepath.Dir(relPath))
		if parent == "." { parent = "" }

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

		batch = append(batch, record)
		if len(batch) >= batchSize {
			flushBatch()
		}

		count++
		if count % 1000 == 0 {
			log.Printf("[Indexer] Scanned %d items...", count)
		}
		return nil
	})

	flushBatch() // 处理最后一批
	log.Printf("[Indexer] Completed. Total items: %d", count)
}

func (ix *Indexer) calculateHash(path string) string {
	f, err := os.Open(path)
	if err != nil { return "" }
	defer f.Close()
	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil { return "" }
	return hex.EncodeToString(h.Sum(nil))
}
