package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements StorageService for local file system storage
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage service
func NewLocalStorage(basePath string) *LocalStorage {
	// Create base directory if it doesn't exist
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		_ = os.MkdirAll(basePath, 0755)
	}

	return &LocalStorage{
		basePath: basePath,
	}
}

// Save saves a file from a multipart form to local storage
func (s *LocalStorage) Save(file *multipart.FileHeader, directory string) (*FileInfo, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename to avoid collisions
	filename := generateUniqueFilename(file.Filename)

	// Create directory if it doesn't exist
	dirPath := filepath.Join(s.basePath, directory)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create destination file
	fullPath := filepath.Join(dirPath, filename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Get content type
	contentType := file.Header.Get("Content-Type")

	// Return file info
	relativePath := filepath.Join(directory, filename)
	return &FileInfo{
		Filename:    filename,
		Size:        file.Size,
		ContentType: contentType,
		URL:         s.GetURL(relativePath),
	}, nil
}

// SaveFromReader saves a file from an io.Reader to local storage
func (s *LocalStorage) SaveFromReader(reader io.Reader, filename, contentType, directory string) (*FileInfo, error) {
	// Generate unique filename to avoid collisions
	uniqueFilename := generateUniqueFilename(filename)

	// Create directory if it doesn't exist
	dirPath := filepath.Join(s.basePath, directory)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Create destination file
	fullPath := filepath.Join(dirPath, uniqueFilename)
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents and get size
	written, err := io.Copy(dst, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file: %w", err)
	}

	// Return file info
	relativePath := filepath.Join(directory, uniqueFilename)
	return &FileInfo{
		Filename:    uniqueFilename,
		Size:        written,
		ContentType: contentType,
		URL:         s.GetURL(relativePath),
	}, nil
}

// Get retrieves a file from local storage
func (s *LocalStorage) Get(fileURL string) (io.ReadCloser, error) {
	// Extract file path from URL
	relativePath, err := extractPathFromURL(fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file URL: %w", err)
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", relativePath)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(fileURL string) error {
	// Extract file path from URL
	relativePath, err := extractPathFromURL(fileURL)
	if err != nil {
		return fmt.Errorf("failed to parse file URL: %w", err)
	}

	fullPath := filepath.Join(s.basePath, relativePath)

	// Check if file exists before attempting to delete
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Return success if file doesn't exist - idempotent deletion
		return nil
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns the URL for a stored file
func (s *LocalStorage) GetURL(filename string) string {
	// For local storage, use a relative URL path with forwards slashes
	return "/uploads/" + strings.ReplaceAll(filename, "\\", "/")
}
