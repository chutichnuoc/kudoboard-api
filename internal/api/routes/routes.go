package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"kudoboard-api/internal/api/handlers"
	"kudoboard-api/internal/api/middleware"
	"kudoboard-api/internal/config"
	"kudoboard-api/internal/services"
	"net/http"
)

// Setup configures all API routes
func Setup(
	router *gin.Engine,
	db *gorm.DB,
	cfg *config.Config,
	authService *services.AuthService,
	boardService *services.BoardService,
	postService *services.PostService,
	mediaService *services.MediaService,
) {
	// Apply global middleware
	router.Use(middleware.ErrorHandler())
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

	// 404 handler
	router.NoRoute(middleware.NotFoundHandler)

	// Create handler instances with services
	authHandler := handlers.NewAuthHandler(authService, cfg)
	boardHandler := handlers.NewBoardHandler(boardService, postService, authService, cfg)
	postHandler := handlers.NewPostHandler(postService, boardService, authService, cfg)
	mediaHandler := handlers.NewMediaHandler(mediaService, boardService, postService, cfg)

	// Create middleware instances
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
			authProtected.POST("/logout", authHandler.Logout)
		}
	}

	// Board routes
	boards := v1.Group("/boards")
	{
		// Public board endpoints
		boards.GET("/public", boardHandler.ListPublicBoards)
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

			// Board contributors
			boardsAuth.GET("/:boardId/contributors", boardHandler.ListBoardContributors)
			boardsAuth.POST("/:boardId/contributors", boardHandler.AddContributor)
			boardsAuth.PUT("/:boardId/contributors/:userId", boardHandler.UpdateContributor)
			boardsAuth.DELETE("/:boardId/contributors/:userId", boardHandler.RemoveContributor)

			// Posts within a board
			boardsAuth.POST("/:boardId/posts", postHandler.CreatePost)
			boardsAuth.PUT("/:boardId/posts/reorder", postHandler.ReorderPosts)
		}
	}

	// Anonymous post endpoints
	anonymous := v1.Group("/anonymous")
	{
		anonymous.POST("/boards/:boardId/posts", postHandler.CreateAnonymousPost)
		anonymous.POST("/boards/:boardId/media/upload", mediaHandler.UploadAnonymousMedia)
		anonymous.POST("/boards/:boardId/media/youtube", mediaHandler.AddYoutubeAnonymous)
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

	// Media routes
	media := v1.Group("/media")
	{
		mediaAuth := media.Group("")
		mediaAuth.Use(authMiddleware.RequireAuth())
		{
			mediaAuth.POST("/upload/:postId", mediaHandler.UploadMedia)
			mediaAuth.POST("/youtube/:postId", mediaHandler.AddYoutube)
			mediaAuth.DELETE("/:mediaId", mediaHandler.DeleteMedia)
		}
	}

	// Theme routes
	themes := v1.Group("/themes")
	{
		themes.GET("", boardHandler.ListThemes)
		themes.GET("/:themeId", boardHandler.GetTheme)
	}
}
