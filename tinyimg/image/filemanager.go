package image

import (
	"fmt"
	"github.com/dedicatedcloud/tinyimg/tinyimg/cache"
	"github.com/dedicatedcloud/tinyimg/tinyimg/config"
	"github.com/dedicatedcloud/tinyimg/tinyimg/stat"
	"log/slog"
	"runtime/debug"
	"time"
)

// FileManager handles collections of Files for conversion.
type FileManager struct {
	File *File

	Logger *slog.Logger

	config *config.Config
	stats  *stat.Stat
	cache  *cache.Cache[string, CompressResult]
}

// NewFileManager creates a new FileManager.
func NewFileManager() *FileManager {
	logger := slog.Default()
	logger.Info("FileManager initialized...")
	cache_ := cache.NewCache[string, CompressResult]()

	fm := &FileManager{
		config: config.GetConfig(),
		stats:  stat.NewStat(),
		Logger: logger,
		cache:  cache_,
	}
	fm.startCacheClearing()

	return fm
}

func (fm *FileManager) startCacheClearing() {
	go func() {
		for {
			fm.Logger.Info("Clearing cache...")
			time.Sleep(10 * time.Minute)
			fm.cache.Clear()
		}
	}()
}

// HandleFile processes a file from the client.
func (fm *FileManager) HandleFile(file *File) (err error) {
	if err = file.Decode(); err != nil {
		return err
	}
	fm.File = file
	fm.File.cache = fm.cache
	fm.Logger.Info("added file to file manager", "filename", file.Name)

	return nil
}

// Clear removes the files in the FileManager.
func (fm *FileManager) Clear() {
	fm.File = nil
	debug.FreeOSMemory()
}

// Convert runs the conversion on all files in the FileManager.
func (fm *FileManager) Convert() (fileResults []CompressResult, files []string, errs []error) {
	startTime := time.Now()
	file := fm.File
	fileResults, files, errs = file.Write(fm.config)

	for _, f := range fileResults {
		fm.stats.IncreaseByteCount(f.SavedBytes)
		fm.stats.IncreaseTimeCount(f.Time)
		fm.stats.IncreaseImageCount(1)
	}
	fm.Clear()

	took := time.Since(startTime).Seconds()
	fmt.Println("Conversion took", took, "seconds")

	return fileResults, files, errs
}

func (fm *FileManager) ZipFiles(files []string) (string, error) {
	zippedFile, err := zipFiles(files, fm.config)
	if err != nil {
		return "", err
	}
	return zippedFile, nil
}
