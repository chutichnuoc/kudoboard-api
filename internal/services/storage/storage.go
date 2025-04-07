package storage

import (
	"fmt"
	"github.com/google/uuid"
	"io"
	"kudoboard-api/internal/config"
	"mime/multipart"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	// StorageTypeLocal represents local file storage
	StorageTypeLocal string = "local"

	// StorageTypeS3 represents AWS S3 storage
	StorageTypeS3 string = "s3"
)

// FileInfo represents metadata about a stored file
type FileInfo struct {
	Filename    string
	Size        int64
	ContentType string
	URL         string
}

// StorageService defines the interface for file storage operations
type StorageService interface {
	// Save uploads a file from a multipart form
	Save(file *multipart.FileHeader, directory string) (*FileInfo, error)

	// SaveFromReader uploads a file from an io.Reader
	SaveFromReader(reader io.Reader, filename, contentType, directory string) (*FileInfo, error)

	// Get returns a reader for a stored file
	Get(fileURL string) (io.ReadCloser, error)

	// Delete removes a stored file
	Delete(fileURL string) error

	// GetURL returns the URL for a stored file
	GetURL(filename string) string
}

// NewStorageService creates a new storage service based on configuration
func NewStorageService(cfg *config.Config) (StorageService, error) {
	switch cfg.StorageType {
	case StorageTypeLocal:
		return NewLocalStorage(cfg.LocalBasePath), nil
	case StorageTypeS3:
		return NewS3Storage(cfg.S3Region, cfg.S3Bucket, cfg.S3AccessKey, cfg.S3SecretKey, cfg)
	default:
		// Default to local storage
		return NewLocalStorage(cfg.LocalBasePath), nil
	}
}

// Helper function to generate a unique filename
func generateUniqueFilename(originalFilename string) string {
	ext := filepath.Ext(originalFilename)
	name := strings.TrimSuffix(originalFilename, ext)
	timestamp := time.Now().Format("20060102150405")
	uniqueID := uuid.New().String()[0:8]

	return fmt.Sprintf("%s-%s-%s%s", name, timestamp, uniqueID, ext)
}

// Helper function to extract file path from URL
func extractPathFromURL(fileURL string) (string, error) {
	// If it looks like a URL, parse it
	if strings.HasPrefix(fileURL, "http://") || strings.HasPrefix(fileURL, "https://") {
		parsedURL, err := url.Parse(fileURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %w", err)
		}
		// Return the path portion of the URL
		return strings.TrimPrefix(parsedURL.Path, "/"), nil
	}

	// If it's a local path starting with /uploads/
	if strings.HasPrefix(fileURL, "/uploads/") {
		return strings.TrimPrefix(fileURL, "/uploads/"), nil
	}

	// If it's already a relative path, just return it
	return fileURL, nil
}
