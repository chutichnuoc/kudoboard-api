package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Storage implements StorageService for AWS S3 storage
type S3Storage struct {
	region     string
	bucket     string
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	s3Client   *s3.S3
}

// NewS3Storage creates a new S3 storage service
func NewS3Storage(region, bucket, accessKey, secretKey string) (*S3Storage, error) {
	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 client, uploader, and downloader
	s3Client := s3.New(sess)
	uploader := s3manager.NewUploader(sess)
	downloader := s3manager.NewDownloader(sess)

	return &S3Storage{
		region:     region,
		bucket:     bucket,
		uploader:   uploader,
		downloader: downloader,
		s3Client:   s3Client,
	}, nil
}

// Save saves a file from a multipart form to S3
func (s *S3Storage) Save(file *multipart.FileHeader, directory string) (*FileInfo, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Generate unique filename to avoid collisions
	filename := generateUniqueFilename(file.Filename)

	// Create S3 key (path)
	key := filepath.Join(directory, filename)
	key = strings.ReplaceAll(key, "\\", "/") // S3 uses forward slashes

	// Upload file to S3
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return file info
	return &FileInfo{
		Filename:    filename,
		Size:        file.Size,
		ContentType: file.Header.Get("Content-Type"),
		URL:         result.Location,
	}, nil
}

// SaveFromReader saves a file from an io.Reader to S3
func (s *S3Storage) SaveFromReader(reader io.Reader, filename, contentType, directory string) (*FileInfo, error) {
	// Generate unique filename to avoid collisions
	uniqueFilename := generateUniqueFilename(filename)

	// Create S3 key (path)
	key := filepath.Join(directory, uniqueFilename)
	key = strings.ReplaceAll(key, "\\", "/") // S3 uses forward slashes

	// Upload file to S3
	result, err := s.uploader.Upload(&s3manager.UploadInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Since we can't get the size directly from the reader,
	// we'll need to set it to 0 or fetch the object metadata later
	return &FileInfo{
		Filename:    uniqueFilename,
		Size:        0, // Size unknown when uploaded from reader
		ContentType: contentType,
		URL:         result.Location,
	}, nil
}

// Get retrieves a file from S3
func (s *S3Storage) Get(filename string) (io.ReadCloser, error) {
	// Create a temporary buffer to store the file
	key := strings.ReplaceAll(filename, "\\", "/") // S3 uses forward slashes

	// Get object from S3
	result, err := s.s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get file from S3: %w", err)
	}

	return result.Body, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(filename string) error {
	key := strings.ReplaceAll(filename, "\\", "/") // S3 uses forward slashes

	// Delete object from S3
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	// Wait for the deletion to complete
	err = s.s3Client.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to confirm file deletion from S3: %w", err)
	}

	return nil
}

// GetURL returns the URL for a stored file
func (s *S3Storage) GetURL(filename string) string {
	key := strings.ReplaceAll(filename, "\\", "/") // S3 uses forward slashes
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
}
