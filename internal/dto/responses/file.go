package responses

import "time"

// FileInfo represents information about an uploaded file
type FileInfo struct {
	FileName    string    `json:"file_name"`
	FilePath    string    `json:"file_path"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
