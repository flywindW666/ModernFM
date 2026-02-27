package indexer

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type FileRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Parent    string    `gorm:"size:1024;index"`
	Name      string    `gorm:"size:255;index"`
	FullPath  string    `gorm:"uniqueIndex;column:full_path;size:2048"`
	IsDir     bool      `gorm:"index"`
	Size      int64
	ModTime   time.Time
	Extension string `gorm:"size:100;index"`
	UpdatedAt time.Time
}

type Indexer struct {
	db      *gorm.DB
	rootDir string
}

func NewIndexer(db *gorm.DB, root string) *Indexer {
	return &Indexer{db: db, rootDir: filepath.Clean(root)}
}

func (ix *Indexer) StartFullScan() {
	log.Printf("[Indexer] Starting background full scan: %s", ix.rootDir)
	filepath.WalkDir(ix.rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(ix.rootDir, path)
		if rel == "." {
			return nil
		}
		
		rel = filepath.ToSlash(rel)
		parent := filepath.ToSlash(filepath.Dir(rel))
		if parent == "." {
			parent = ""
		}

		info, _ := d.Info()
		record := FileRecord{
			Parent:    parent,
			Name:      d.Name(),
			FullPath:  rel,
			IsDir:     d.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: strings.ToLower(filepath.Ext(d.Name())),
			UpdatedAt: time.Now(),
		}
		ix.db.Where("full_path = ?", rel).Assign(record).FirstOrCreate(&FileRecord{})
		return nil
	})
	log.Println("[Indexer] Full scan completed.")
}

func (ix *Indexer) ScanDir(relPath string) ([]FileRecord, error) {
	absPath := filepath.Join(ix.rootDir, relPath)
	log.Printf("[Indexer] Scanning path: %s", absPath)
	
	entries, err := os.ReadDir(absPath)
	if err != nil {
		log.Printf("[Indexer] Error reading dir %s: %v", absPath, err)
		return nil, err
	}

	currentRecords := make([]FileRecord, 0)
	foundMap := make(map[string]bool)

	for _, d := range entries {
		info, err := d.Info()
		if err != nil {
			continue
		}

		itemPath := filepath.ToSlash(filepath.Join(relPath, d.Name()))
		foundMap[itemPath] = true

		record := FileRecord{
			Parent:    filepath.ToSlash(relPath),
			Name:      d.Name(),
			FullPath:  itemPath,
			IsDir:     d.IsDir(),
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			Extension: strings.ToLower(filepath.Ext(d.Name())),
			UpdatedAt: time.Now(),
		}
		currentRecords = append(currentRecords, record)

		ix.db.Where("full_path = ?", itemPath).Assign(record).FirstOrCreate(&FileRecord{})
		
		// 递归策略：如果是目录，且您原来的设定是需要深度扫描，我们在这里触发异步扫描
		if d.IsDir() {
			go ix.ScanDir(itemPath)
		}
	}

	var dbRecords []FileRecord
	ix.db.Where("parent = ?", filepath.ToSlash(relPath)).Find(&dbRecords)
	for _, dr := range dbRecords {
		if !foundMap[dr.FullPath] {
			ix.db.Delete(&dr)
			if dr.IsDir {
				ix.db.Where("full_path LIKE ?", dr.FullPath+"/%").Delete(&FileRecord{})
			}
		}
	}

	return currentRecords, nil
}
