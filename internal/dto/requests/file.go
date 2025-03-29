package requests

// DeleteFileRequest represents the request to delete a file
type DeleteFileRequest struct {
	FilePath string `json:"file_path" binding:"required"`
}
