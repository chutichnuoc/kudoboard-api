package routes

import (
	"github.com/gin-gonic/gin"
	"kudoboard-api/internal/api/handlers"
	"kudoboard-api/internal/api/middleware"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/services"
	"net/http"
)

// Setup configures all API routes
func Setup(
	router *gin.Engine,
	cfg *config.Config,
	authService *services.AuthService,
	boardService *services.BoardService,
	postService *services.PostService,
	themeService *services.ThemeService,
	fileService *services.FileService,
	giphyService *services.GiphyService,
	unsplashService *services.UnsplashService,
) {
	// Create error middleware with debug mode based on environment
	errorMiddleware := middleware.NewErrorMiddleware(cfg.Environment != "production")

	// Apply global middleware
	router.Use(errorMiddleware.ErrorHandler())
	router.Use(middleware.CorsMiddleware(cfg))

	// Serve uploaded files in development mode
	if cfg.Environment != "production" && cfg.StorageType == "local" {
		router.Static("/uploads", cfg.LocalBasePath)
	}

	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// 404 and 405 handlers
	router.NoRoute(errorMiddleware.NotFoundHandler)
	router.NoMethod(errorMiddleware.MethodNotAllowedHandler)

	// Create handler instances with services
	authHandler := handlers.NewAuthHandler(authService, cfg)
	boardHandler := handlers.NewBoardHandler(boardService, postService, themeService, authService, cfg)
	postHandler := handlers.NewPostHandler(postService, boardService, authService, cfg)
	themeHandler := handlers.NewThemeHandler(themeService, cfg)
	fileHandler := handlers.NewFileHandler(fileService, cfg)
	giphyHandler := handlers.NewGiphyHandler(giphyService, cfg)
	unsplashHandler := handlers.NewUnsplashHandler(unsplashService, cfg)

	authMiddleware := middleware.NewAuthMiddleware(authService, cfg)

	// API v1 routes
	v1 := router.Group("/api/v1")

	// Auth routes
	auth := v1.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/google", authHandler.GoogleLogin)
		auth.POST("/facebook", authHandler.FacebookLogin)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)

		// Auth routes requiring authentication
		authProtected := auth.Group("")
		authProtected.Use(authMiddleware.RequireAuth())
		{
			authProtected.GET("/me", authHandler.GetMe)
			authProtected.PUT("/me", authHandler.UpdateProfile)
		}
	}

	// Board routes
	boards := v1.Group("/boards")
	{
		// Public board endpoints
		boards.GET("/slug/:slug", authMiddleware.OptionalAuth(), boardHandler.GetBoardBySlug)

		// Board endpoints requiring authentication
		boardsAuth := boards.Group("")
		boardsAuth.Use(authMiddleware.RequireAuth())
		{
			// Board CRUD operations
			boardsAuth.GET("", boardHandler.ListUserBoards)
			boardsAuth.POST("", boardHandler.CreateBoard)
			boardsAuth.PUT("/:boardId", boardHandler.UpdateBoard)
			boardsAuth.DELETE("/:boardId", boardHandler.DeleteBoard)
			boardsAuth.PATCH("/:boardId/lock", boardHandler.ToggleBoardLock)

			// Board preferences
			boardsAuth.PATCH("/:boardId/preferences", boardHandler.UpdateBoardPreferences)

			// Board contributors
			boardsAuth.GET("/:boardId/contributors", boardHandler.ListBoardContributors)
			boardsAuth.POST("/:boardId/contributors", boardHandler.AddContributor)
			boardsAuth.PUT("/:boardId/contributors/:contributorId", boardHandler.UpdateContributor)
			boardsAuth.DELETE("/:boardId/contributors/:contributorId", boardHandler.RemoveContributor)

			// Posts within a board
			boardsAuth.PUT("/:boardId/posts/reorder", postHandler.ReorderPosts)
		}

		// Posts within a board
		boards.POST("/:boardId/posts", authMiddleware.OptionalAuth(), postHandler.CreatePost)
	}

	// Post operations
	posts := v1.Group("/posts")
	{
		// Posts require authentication
		postsAuth := posts.Group("")
		postsAuth.Use(authMiddleware.RequireAuth())
		{
			postsAuth.PUT("/:postId", postHandler.UpdatePost)
			postsAuth.DELETE("/:postId", postHandler.DeletePost)
			postsAuth.POST("/:postId/like", postHandler.LikePost)
			postsAuth.DELETE("/:postId/like", postHandler.UnlikePost)
		}
	}

	// Theme routes
	themes := v1.Group("/themes")
	{
		themes.GET("", themeHandler.ListThemes)
		themes.GET("/:themeId", themeHandler.GetTheme)

		// Protected theme routes (only for admins)
		themesAdmin := themes.Group("")
		themesAdmin.Use(authMiddleware.RequireAuth(), middleware.AdminOnly())
		{
			themesAdmin.POST("", themeHandler.CreateTheme)
			themesAdmin.PUT("/:themeId", themeHandler.UpdateTheme)
			themesAdmin.DELETE("/:themeId", themeHandler.DeleteTheme)
		}
	}

	// File routes
	files := v1.Group("/files")
	{
		// Public upload endpoint (works for both authenticated and anonymous users)
		files.POST("/upload", authMiddleware.OptionalAuth(), fileHandler.UploadFile)

		// Authenticated endpoints
		filesAuth := files.Group("")
		filesAuth.Use(authMiddleware.RequireAuth())
		{
			filesAuth.DELETE("", fileHandler.DeleteFile)
		}
	}

	giphy := v1.Group("/giphy")
	{
		giphy.GET("/search", giphyHandler.Search)
		giphy.GET("/trending", giphyHandler.Trending)
		giphy.GET("/random", giphyHandler.Random)
		giphy.GET("/:gifId", giphyHandler.GetById)
	}

	unsplash := v1.Group("/unsplash")
	{
		unsplash.GET("/search", unsplashHandler.Search)
		unsplash.GET("/random", unsplashHandler.Random)
		unsplash.GET("/:photoId", unsplashHandler.GetById)
	}
}
