package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"exchange/internal/pkg/config"
)

// CleanupConfig 清理配置
type CleanupConfig struct {
	LogDir     string
	MaxAge     int  // 保留天数
	MaxBackups int  // 最大备份文件数
	DryRun     bool // 是否为试运行
}

// LogCleanupManager 日志清理管理器
type LogCleanupManager struct {
	config *CleanupConfig
}

// NewLogCleanupManager 创建日志清理管理器
func NewLogCleanupManager(cfg *config.LogConfig) *LogCleanupManager {
	return &LogCleanupManager{
		config: &CleanupConfig{
			LogDir:     cfg.LogDir,
			MaxAge:     cfg.MaxAge,
			MaxBackups: cfg.MaxBackups,
			DryRun:     false,
		},
	}
}

// StartCleanupScheduler 启动清理调度器
func (m *LogCleanupManager) StartCleanupScheduler() {
	// 每天凌晨2点执行清理
	ticker := time.NewTicker(24 * time.Hour)
	
	// 计算到下一个凌晨2点的时间
	now := time.Now()
	next2AM := time.Date(now.Year(), now.Month(), now.Day()+1, 2, 0, 0, 0, now.Location())
	if now.Hour() < 2 {
		next2AM = time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	}
	
	// 首次执行的延迟
	initialDelay := next2AM.Sub(now)
	
	go func() {
		// 等待到凌晨2点
		time.Sleep(initialDelay)
		
		// 立即执行一次清理
		m.CleanupLogs()
		
		// 然后每24小时执行一次
		for range ticker.C {
			m.CleanupLogs()
		}
	}()
	
	Info("Log cleanup scheduler started", map[string]interface{}{
		"next_cleanup": next2AM.Format(time.RFC3339),
		"interval":     "24h",
	})
}

// CleanupLogs 清理日志文件
func (m *LogCleanupManager) CleanupLogs() error {
	if m.config.LogDir == "" {
		return fmt.Errorf("log directory not configured")
	}
	
	start := time.Now()
	
	Info("Starting log cleanup", map[string]interface{}{
		"log_dir":     m.config.LogDir,
		"max_age":     m.config.MaxAge,
		"max_backups": m.config.MaxBackups,
		"dry_run":     m.config.DryRun,
	})
	
	// 获取所有日志文件
	files, err := m.getLogFiles()
	if err != nil {
		Error("Failed to get log files", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	// 按年龄清理
	deletedByAge, err := m.cleanupByAge(files)
	if err != nil {
		Error("Failed to cleanup by age", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	// 按数量清理
	deletedByCount, err := m.cleanupByCount(files)
	if err != nil {
		Error("Failed to cleanup by count", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}
	
	duration := time.Since(start)
	
	Info("Log cleanup completed", map[string]interface{}{
		"duration":        duration.String(),
		"deleted_by_age":  deletedByAge,
		"deleted_by_count": deletedByCount,
		"total_deleted":   deletedByAge + deletedByCount,
	})
	
	return nil
}

// getLogFiles 获取所有日志文件
func (m *LogCleanupManager) getLogFiles() ([]LogFileInfo, error) {
	var files []LogFileInfo
	
	err := filepath.Walk(m.config.LogDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if info.IsDir() {
			return nil
		}
		
		// 只处理日志文件
		if m.isLogFile(info.Name()) {
			files = append(files, LogFileInfo{
				Path:    path,
				Name:    info.Name(),
				Size:    info.Size(),
				ModTime: info.ModTime(),
			})
		}
		
		return nil
	})
	
	return files, err
}

// isLogFile 判断是否为日志文件
func (m *LogCleanupManager) isLogFile(filename string) bool {
	// 匹配 .log 文件和压缩的日志文件
	return strings.HasSuffix(filename, ".log") ||
		   strings.HasSuffix(filename, ".log.gz") ||
		   strings.HasSuffix(filename, ".log.bz2") ||
		   strings.Contains(filename, ".log.")
}

// cleanupByAge 按年龄清理文件
func (m *LogCleanupManager) cleanupByAge(files []LogFileInfo) (int, error) {
	if m.config.MaxAge <= 0 {
		return 0, nil
	}
	
	cutoff := time.Now().AddDate(0, 0, -m.config.MaxAge)
	deleted := 0
	
	for _, file := range files {
		if file.ModTime.Before(cutoff) {
			if m.config.DryRun {
				Info("Would delete old log file", map[string]interface{}{
					"file":    file.Path,
					"age":     time.Since(file.ModTime).String(),
					"dry_run": true,
				})
			} else {
				if err := os.Remove(file.Path); err != nil {
					Warn("Failed to delete old log file", map[string]interface{}{
						"file":  file.Path,
						"error": err.Error(),
					})
				} else {
					Info("Deleted old log file", map[string]interface{}{
						"file": file.Path,
						"age":  time.Since(file.ModTime).String(),
					})
					deleted++
				}
			}
		}
	}
	
	return deleted, nil
}

// cleanupByCount 按数量清理文件
func (m *LogCleanupManager) cleanupByCount(files []LogFileInfo) (int, error) {
	if m.config.MaxBackups <= 0 {
		return 0, nil
	}
	
	// 按基础文件名分组
	fileGroups := make(map[string][]LogFileInfo)
	for _, file := range files {
		baseName := m.getBaseName(file.Name)
		fileGroups[baseName] = append(fileGroups[baseName], file)
	}
	
	deleted := 0
	
	for baseName, groupFiles := range fileGroups {
		if len(groupFiles) <= m.config.MaxBackups {
			continue
		}
		
		// 按修改时间排序（最新的在前）
		sort.Slice(groupFiles, func(i, j int) bool {
			return groupFiles[i].ModTime.After(groupFiles[j].ModTime)
		})
		
		// 删除超出数量限制的文件
		for i := m.config.MaxBackups; i < len(groupFiles); i++ {
			file := groupFiles[i]
			
			if m.config.DryRun {
				Info("Would delete excess log file", map[string]interface{}{
					"file":      file.Path,
					"base_name": baseName,
					"dry_run":   true,
				})
			} else {
				if err := os.Remove(file.Path); err != nil {
					Warn("Failed to delete excess log file", map[string]interface{}{
						"file":  file.Path,
						"error": err.Error(),
					})
				} else {
					Info("Deleted excess log file", map[string]interface{}{
						"file":      file.Path,
						"base_name": baseName,
					})
					deleted++
				}
			}
		}
	}
	
	return deleted, nil
}

// getBaseName 获取日志文件的基础名称
func (m *LogCleanupManager) getBaseName(filename string) string {
	// 移除时间戳和压缩后缀
	name := filename
	
	// 移除压缩后缀
	if strings.HasSuffix(name, ".gz") {
		name = strings.TrimSuffix(name, ".gz")
	}
	if strings.HasSuffix(name, ".bz2") {
		name = strings.TrimSuffix(name, ".bz2")
	}
	
	// 移除时间戳部分（假设格式为 filename.log.2023-01-01T00-00-00.000）
	parts := strings.Split(name, ".")
	if len(parts) >= 3 {
		// 检查是否有时间戳格式
		for i := len(parts) - 1; i >= 0; i-- {
			if len(parts[i]) >= 10 && strings.Contains(parts[i], "-") {
				// 可能是时间戳，移除它
				parts = parts[:i]
				break
			}
		}
	}
	
	return strings.Join(parts, ".")
}

// LogFileInfo 日志文件信息
type LogFileInfo struct {
	Path    string
	Name    string
	Size    int64
	ModTime time.Time
}

// GetLogStats 获取日志统计信息
func (m *LogCleanupManager) GetLogStats() (map[string]interface{}, error) {
	files, err := m.getLogFiles()
	if err != nil {
		return nil, err
	}
	
	var totalSize int64
	var oldestFile, newestFile time.Time
	filesByType := make(map[string]int)
	
	for i, file := range files {
		totalSize += file.Size
		
		if i == 0 {
			oldestFile = file.ModTime
			newestFile = file.ModTime
		} else {
			if file.ModTime.Before(oldestFile) {
				oldestFile = file.ModTime
			}
			if file.ModTime.After(newestFile) {
				newestFile = file.ModTime
			}
		}
		
		// 按文件类型分类
		ext := filepath.Ext(file.Name)
		filesByType[ext]++
	}
	
	stats := map[string]interface{}{
		"total_files":    len(files),
		"total_size":     totalSize,
		"total_size_mb":  float64(totalSize) / 1024 / 1024,
		"files_by_type":  filesByType,
	}
	
	if len(files) > 0 {
		stats["oldest_file"] = oldestFile.Format(time.RFC3339)
		stats["newest_file"] = newestFile.Format(time.RFC3339)
		stats["age_range"] = newestFile.Sub(oldestFile).String()
	}
	
	return stats, nil
}

// ForceCleanup 强制清理（用于手动触发）
func (m *LogCleanupManager) ForceCleanup(dryRun bool) error {
	originalDryRun := m.config.DryRun
	m.config.DryRun = dryRun
	
	defer func() {
		m.config.DryRun = originalDryRun
	}()
	
	return m.CleanupLogs()
}