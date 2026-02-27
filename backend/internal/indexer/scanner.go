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
	Parent    string    `gorm:"size:1024;index"` // 父目录路径
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

// ScanDir 扫描特定目录并同步数据库
func (ix *Indexer) ScanDir(relPath string) ([]FileRecord, error) {
	absPath := filepath.Join(ix.rootDir, relPath)
	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, err
	}

	var currentRecords []FileRecord
	foundMap := make(map[string]bool)

	for _, d := range entries {
		info, err := d.Info()
		if err != nil {
			continue
		}

		itemPath := filepath.ToSlash(filepath.Join(relPath, d.Name()))
		if relPath == "" || relPath == "/" {
			itemPath = d.Name()
		}
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

		// Upsert 逻辑
		ix.db.Where("full_path = ?", itemPath).Assign(record).FirstOrCreate(&FileRecord{})
	}

	// 清理数据库中已删除的文件
	var dbRecords []FileRecord
	ix.db.Where("parent = ?", relPath).Find(&dbRecords)
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
