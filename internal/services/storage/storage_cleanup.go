package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"kudoboard-api/internal/config"
	"kudoboard-api/internal/log"
	"kudoboard-api/internal/models"
)

// StorageCleanupService handles cleanup of orphaned files
type StorageCleanupService struct {
	db      *gorm.DB
	storage StorageService
	cfg     *config.Config
}

// NewStorageCleanupService creates a new storage cleanup service
func NewStorageCleanupService(db *gorm.DB, storage StorageService, cfg *config.Config) *StorageCleanupService {
	return &StorageCleanupService{
		db:      db,
		storage: storage,
		cfg:     cfg,
	}
}

// CleanOrphanedFiles identifies and removes orphaned files
func (s *StorageCleanupService) CleanOrphanedFiles() error {
	log.Info("Starting orphaned file cleanup job")

	// Define common directories/prefixes to check
	prefixes := []string{
		"image/",
		"avatar/",
		"gif/",
		"video/",
		"theme/",
		"icon/",
		"general/",
	}

	var totalProcessed, totalDeleted, totalErrors int

	// Set minimum age for deletion
	minAge := time.Now().Add(-24 * time.Hour)

	// Process files by prefix to reduce memory usage
	for _, prefix := range prefixes {
		processed, deleted, errors, err := s.cleanPrefix(prefix, minAge)
		if err != nil {
			log.Error("Error cleaning prefix",
				zap.String("prefix", prefix),
				zap.Error(err))
			continue
		}

		totalProcessed += processed
		totalDeleted += deleted
		totalErrors += errors
	}

	log.Info("Orphaned file cleanup job completed",
		zap.Int("total_processed", totalProcessed),
		zap.Int("total_deleted", totalDeleted),
		zap.Int("total_errors", totalErrors))

	return nil
}

// cleanPrefix handles cleanup for a specific storage prefix
func (s *StorageCleanupService) cleanPrefix(prefix string, minAge time.Time) (int, int, int, error) {
	const batchSize = 100
	var lastKey string
	var totalProcessed, totalDeleted, totalErrors int

	for {
		// Get a batch of files
		files, err := s.listFilesBatch(prefix, lastKey, batchSize)
		if err != nil {
			return totalProcessed, totalDeleted, totalErrors, fmt.Errorf("failed to list files: %w", err)
		}

		// If no files returned, we're done with this prefix
		if len(files) == 0 {
			break
		}

		// Filter files by age
		var filesToCheck []FileInfo
		for _, file := range files {
			if file.ModTime.Before(minAge) {
				filesToCheck = append(filesToCheck, file)
			}
		}

		totalProcessed += len(filesToCheck)

		// Skip database check if no eligible files
		if len(filesToCheck) == 0 {
			// Update the last key for the next batch
			if len(files) > 0 {
				lastKey = files[len(files)-1].URL
			}
			continue
		}

		// Find orphaned files
		orphanedFiles, err := s.findOrphanedFiles(filesToCheck)
		if err != nil {
			return totalProcessed, totalDeleted, totalErrors, fmt.Errorf("failed to find orphaned files: %w", err)
		}

		// Delete orphaned files
		for _, file := range orphanedFiles {
			if err := s.storage.Delete(file.URL); err != nil {
				log.Error("Failed to delete orphaned file",
					zap.String("file_path", file.URL),
					zap.Error(err))
				totalErrors++
			} else {
				log.Info("Deleted orphaned file",
					zap.String("file_path", file.URL))
				totalDeleted++
			}
		}

		// Update the last key for the next batch
		if len(files) > 0 {
			lastKey = files[len(files)-1].URL
		}

		// If we got fewer files than the batch size, we're done with this prefix
		if len(files) < batchSize {
			break
		}
	}

	return totalProcessed, totalDeleted, totalErrors, nil
}

// findOrphanedFiles efficiently identifies files not referenced in the database
func (s *StorageCleanupService) findOrphanedFiles(files []FileInfo) ([]FileInfo, error) {
	// Extract all file paths
	filePaths := make([]string, len(files))
	for i, file := range files {
		filePaths[i] = file.URL
	}

	// Create a map for efficient lookups
	existingPathsMap := make(map[string]bool)

	// Check posts table
	var existingPostPaths []string
	if err := s.db.Model(&models.Post{}).
		Where("media_path IN ? AND media_source = 'internal'", filePaths).
		Pluck("media_path", &existingPostPaths).Error; err != nil {
		return nil, err
	}
	for _, path := range existingPostPaths {
		existingPathsMap[path] = true
	}

	// Check themes table - icon_url
	var iconPaths []string
	if err := s.db.Model(&models.Theme{}).
		Where("icon_url IN ?", filePaths).
		Pluck("icon_url", &iconPaths).Error; err != nil {
		return nil, err
	}
	for _, path := range iconPaths {
		existingPathsMap[path] = true
	}

	// Check themes table - background_image_url
	var bgPaths []string
	if err := s.db.Model(&models.Theme{}).
		Where("background_image_url IN ?", filePaths).
		Pluck("background_image_url", &bgPaths).Error; err != nil {
		return nil, err
	}
	for _, path := range bgPaths {
		existingPathsMap[path] = true
	}

	// Check users table - profile_picture
	var profilePaths []string
	if err := s.db.Model(&models.User{}).
		Where("profile_picture IN ?", filePaths).
		Pluck("profile_picture", &profilePaths).Error; err != nil {
		return nil, err
	}
	for _, path := range profilePaths {
		existingPathsMap[path] = true
	}

	// Find orphaned files
	var orphanedFiles []FileInfo
	for _, file := range files {
		if !existingPathsMap[file.URL] {
			orphanedFiles = append(orphanedFiles, file)
		}
	}

	return orphanedFiles, nil
}

// listFilesBatch retrieves a batch of files from the appropriate storage
func (s *StorageCleanupService) listFilesBatch(prefix string, startAfter string, batchSize int) ([]FileInfo, error) {
	if s.cfg.StorageType == StorageTypeS3 {
		return s.listS3FilesBatch(prefix, startAfter, batchSize)
	}
	return s.listLocalFilesBatch(prefix, startAfter, batchSize)
}

// listLocalFilesBatch lists files from local storage with pagination
func (s *StorageCleanupService) listLocalFilesBatch(prefix string, startAfter string, batchSize int) ([]FileInfo, error) {
	basePath := s.cfg.LocalBasePath
	prefixPath := filepath.Join(basePath, prefix)

	// Make sure the directory exists
	if _, err := os.Stat(prefixPath); os.IsNotExist(err) {
		// Directory doesn't exist, return empty result
		return []FileInfo{}, nil
	}

	var files []FileInfo
	var skipUntilAfter bool

	if startAfter == "" {
		skipUntilAfter = false
	} else {
		skipUntilAfter = true
	}

	// Walk through the directory
	err := filepath.Walk(prefixPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Get the relative path and convert to URL format
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		relPath = strings.ReplaceAll(relPath, "\\", "/")
		url := "/uploads/" + relPath

		// Skip files until we reach the startAfter marker
		if skipUntilAfter {
			if url <= startAfter {
				return nil
			}
			skipUntilAfter = false
		}

		// Add file to results
		files = append(files, FileInfo{
			URL:     url,
			ModTime: info.ModTime(),
		})

		// Stop if we've reached the batch size
		if len(files) >= batchSize {
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return files, nil
}

// listS3FilesBatch lists files from S3 with pagination
func (s *StorageCleanupService) listS3FilesBatch(prefix string, startAfter string, batchSize int) ([]FileInfo, error) {
	// Type assertion to access S3 client
	s3Storage, ok := s.storage.(*S3Storage)
	if !ok {
		return nil, fmt.Errorf("storage is not an S3Storage")
	}

	// Get S3 client and bucket
	svc := s3Storage.GetS3Client()
	bucket := s3Storage.GetBucketName()

	// Create request input
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(int64(batchSize)),
	}

	// If startAfter is provided, use it for pagination
	if startAfter != "" {
		// Convert URL back to S3 key
		key, err := ExtractPathFromURL(startAfter)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start marker: %w", err)
		}
		input.StartAfter = aws.String(key)
	}

	// List objects
	result, err := svc.ListObjectsV2(input)
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 objects: %w", err)
	}

	// Process results
	files := make([]FileInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		url := s3Storage.GetURL(*obj.Key)
		files = append(files, FileInfo{
			URL:     url,
			ModTime: *obj.LastModified,
		})
	}

	return files, nil
}
