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

// IndexDir 仅扫描指定目录的直接子项 (用于按需加载)
func (ix *Indexer) IndexDir(relPath string) {
	absPath := filepath.Join(ix.rootDir, relPath)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		log.Printf("[Indexer] Error reading dir %s: %v", relPath, err)
		return
	}

	ix.db.Transaction(func(tx *gorm.DB) error {
		for _, d := range entries {
			info, err := d.Info()
			if err != nil { continue }
			
			itemPath := filepath.ToSlash(filepath.Join(relPath, d.Name()))
			fileExt := ""
			if !d.IsDir() {
				fileExt = strings.ToLower(filepath.Ext(d.Name()))
			}

			record := FileRecord{
				Parent:    relPath,
				Name:      d.Name(),
				FullPath:  itemPath,
				IsDir:     d.IsDir(),
				Size:      info.Size(),
				ModTime:   info.ModTime(),
				Extension: fileExt,
				UpdatedAt: time.Now(),
			}
			tx.Where("full_path = ?", itemPath).Assign(record).FirstOrCreate(&FileRecord{})
		}
		return nil
	})
	log.Printf("[Indexer] Incremental scan of '%s' completed.", relPath)
}

func (ix *Indexer) StartFullScan() {
	ix.mu.Lock()
	if ix.isRunning { ix.mu.Unlock(); return }
	ix.isRunning = true
	ix.mu.Unlock()
	defer func() { ix.isRunning = false }()

	log.Printf("[Indexer] Starting background full scan: %s", ix.rootDir)
	
	const batchSize = 100
	var batch []FileRecord
	
	flush := func() {
		if len(batch) == 0 { return }
		ix.db.Transaction(func(tx *gorm.DB) error {
			for _, item := range batch {
				tx.Where("full_path = ?", item.FullPath).Assign(item).FirstOrCreate(&FileRecord{})
			}
			return nil
		})
		batch = nil
	}

	filepath.WalkDir(ix.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || path == ix.rootDir { return nil }
		info, err := d.Info()
		if err != nil { return nil }
		
		rel, _ := filepath.Rel(ix.rootDir, path)
		rel = filepath.ToSlash(rel)
		parent := filepath.ToSlash(filepath.Dir(rel))
		if parent == "." { parent = "" }

		batch = append(batch, FileRecord{
			Parent: parent, Name: d.Name(), FullPath: rel, IsDir: d.IsDir(),
			Size: info.Size(), ModTime: info.ModTime(), 
			Extension: strings.ToLower(filepath.Ext(d.Name())), UpdatedAt: time.Now(),
		})
		if len(batch) >= batchSize { flush() }
		return nil
	})
	flush()
	log.Println("[Indexer] Full scan completed.")
}
