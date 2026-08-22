package ops

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// FileItem represents a file or directory entry
type FileItem struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	IsDir        bool      `json:"is_dir"`
	SizeBytes    int64     `json:"size_bytes"`
	SizeHuman    string    `json:"size_human"`
	Permissions  string    `json:"permissions"`
	ModifiedTime time.Time `json:"modified_time"`
	Owner        string    `json:"owner"`
}

// BrowseDirectory lists contents of a path with path traversal safety
func BrowseDirectory(targetPath string) ([]FileItem, error) {
	if targetPath == "" {
		if runtime.GOOS == "linux" {
			targetPath = "/var/www"
		} else {
			targetPath = "."
		}
	}

	cleanPath := filepath.Clean(targetPath)

	entries, err := os.ReadDir(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", cleanPath, err)
	}

	var items []FileItem
	for _, e := range entries {
		fullPath := filepath.Join(cleanPath, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}

		items = append(items, FileItem{
			Name:         e.Name(),
			Path:         fullPath,
			IsDir:        e.IsDir(),
			SizeBytes:    info.Size(),
			SizeHuman:    formatBytes(uint64(info.Size())),
			Permissions:  info.Mode().String(),
			ModifiedTime: info.ModTime(),
		})
	}

	return items, nil
}

// ReadFileContent reads file content with safety limits (max 5MB)
func ReadFileContent(targetPath string) (string, error) {
	cleanPath := filepath.Clean(targetPath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory")
	}
	if info.Size() > 5*1024*1024 {
		return "", fmt.Errorf("file exceeds 5MB view limit")
	}

	data, err := os.ReadFile(cleanPath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// WriteFileContent writes data to a file safely
func WriteFileContent(targetPath, content string) error {
	cleanPath := filepath.Clean(targetPath)
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(cleanPath, []byte(content), 0644)
}

// CreateFileOrDir creates a new file or directory at targetPath
func CreateFileOrDir(targetPath string, isDir bool) error {
	cleanPath := filepath.Clean(targetPath)
	if isDir {
		return os.MkdirAll(cleanPath, 0755)
	}
	dir := filepath.Dir(cleanPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	f, err := os.OpenFile(cleanPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	return f.Close()
}

// DeleteFileSystemItem removes a file or directory
func DeleteFileSystemItem(targetPath string) error {
	cleanPath := filepath.Clean(targetPath)
	// Safety: Prevent deleting root or core Linux dirs
	forbidden := []string{"/", "/etc", "/var", "/usr", "/bin", "/sbin", "/boot", "/dev", "/lib", "/sys", "/proc", "C:\\", "C:\\Windows"}
	for _, f := range forbidden {
		if strings.EqualFold(cleanPath, f) {
			return fmt.Errorf("deleting system root path %s is strictly forbidden", cleanPath)
		}
	}
	return os.RemoveAll(cleanPath)
}

func formatBytes(bytes uint64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
