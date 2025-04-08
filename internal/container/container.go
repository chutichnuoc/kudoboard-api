package container

import (
	"gorm.io/gorm"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/services"
	"kudoboard-api/internal/services/storage"
)

// Container holds all application services and dependencies
type Container struct {
	Config         *config.Config
	DB             *gorm.DB
	StorageService storage.StorageService

	// Services
	AuthService     *services.AuthService
	BoardService    *services.BoardService
	PostService     *services.PostService
	ThemeService    *services.ThemeService
	FileService     *services.FileService
	GiphyService    *services.GiphyService
	UnsplashService *services.UnsplashService
}

// NewContainer creates and initializes a new dependency container
func NewContainer(cfg *config.Config, db *gorm.DB) (*Container, error) {
	container := &Container{
		Config: cfg,
		DB:     db,
	}

	// Initialize storage service
	storageService, err := storage.NewStorageService(cfg)
	if err != nil {
		return nil, err
	}
	container.StorageService = storageService

	// Initialize services in the correct order (respect dependencies)
	container.AuthService = services.NewAuthService(db, storageService, cfg)
	container.BoardService = services.NewBoardService(db, storageService, cfg)
	container.ThemeService = services.NewThemeService(db, storageService, cfg)
	container.FileService = services.NewFileService(storageService, cfg)
	container.GiphyService = services.NewGiphyService(cfg)
	container.UnsplashService = services.NewUnsplashService(cfg)

	// Services with dependencies on other services
	container.PostService = services.NewPostService(
		db,
		storageService,
		cfg,
		container.BoardService,
	)

	return container, nil
}
