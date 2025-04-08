# Kudoboard API Documentation

## Overview

This documentation covers the Kudoboard API, which powers a platform for creating collaborative appreciation boards where users can post messages, share media, and interact with others' content.

**Base URL**: `/api/v1`

## Authentication

### Authentication Methods

The API uses JWT (JSON Web Token) authentication. Include the token in the `Authorization` header:

```
Authorization: Bearer {token}
```

Certain endpoints allow anonymous access (noted in the documentation).

### Endpoints

#### Register a New User

```
POST /auth/register
```

Create a new user account.

**Request Body:**
```json
{
  "name": "string",
  "email": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "string",
    "user": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
}
```

#### Login

```
POST /auth/login
```

Authenticate a user with email and password.

**Request Body:**
```json
{
  "email": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "token": "string",
    "user": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
}
```

#### Google Login

```
POST /auth/google
```

Authenticate a user with Google OAuth.

**Request Body:**
```json
{
  "access_token": "string"
}
```

**Response:** Same as the login endpoint.

#### Facebook Login

```
POST /auth/facebook
```

Authenticate a user with Facebook OAuth.

**Request Body:**
```json
{
  "access_token": "string"
}
```

**Response:** Same as the login endpoint.

#### Get Current User

```
GET /auth/me
```

Get the authenticated user's profile.

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "name": "string",
    "email": "string",
    "profile_picture": "string",
    "is_verified": false,
    "auth_provider": "string",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Update User Profile

```
PUT /auth/me
```

Update the user's profile.

**Request Body:**
```json
{
  "name": "string",
  "profile_picture": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "name": "string",
    "email": "string",
    "profile_picture": "string",
    "is_verified": false,
    "auth_provider": "string",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Forgot Password

```
POST /auth/forgot-password
```

Initiate the password reset process.

**Request Body:**
```json
{
  "email": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "If your email exists in our system, you will receive a password reset link"
  }
}
```

#### Reset Password

```
POST /auth/reset-password
```

Reset a user's password using a reset token.

**Request Body:**
```json
{
  "token": "string",
  "password": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Password has been reset successfully"
  }
}
```

## Boards

### Endpoints

#### Create a Board

```
POST /boards
```

Create a new board.

**Authorization:** Required

**Request Body:**
```json
{
  "title": "string",
  "receiver_name": "string",
  "font_name": "string",
  "font_size": 0,
  "header_color": "string",
  "theme_id": 0,
  "effect": "string",
  "enable_intro_animation": false,
  "is_private": false,
  "allow_anonymous": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "title": "string",
    "receiver_name": "string",
    "slug": "string",
    "max_post": 0,
    "creator": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "font_name": "string",
    "font_size": 0,
    "header_color": "string",
    "show_header_color": false,
    "theme": {
      "id": 0,
      "category": "string",
      "name": "string",
      "icon_url": "string",
      "background_image_url": "string"
    },
    "effect": "string",
    "enable_intro_animation": false,
    "is_private": false,
    "is_locked": false,
    "allow_anonymous": false,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "post_count": 0
  }
}
```

#### List User Boards

```
GET /boards
```

List all boards where the user is an owner or contributor.

**Authorization:** Required

**Query Parameters:**
- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 10)
- `search`: Search term
- `sort_by`: Field to sort by (`created_at` or `title`)
- `order`: Sort order (`asc` or `desc`)

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 0,
      "title": "string",
      "receiver_name": "string",
      "slug": "string",
      "max_post": 0,
      "creator": {
        "id": 0,
        "name": "string",
        "email": "string",
        "profile_picture": "string",
        "is_verified": false,
        "auth_provider": "string",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "font_name": "string",
      "font_size": 0,
      "header_color": "string",
      "show_header_color": false,
      "theme": {
        "id": 0,
        "category": "string",
        "name": "string",
        "icon_url": "string",
        "background_image_url": "string"
      },
      "effect": "string",
      "enable_intro_animation": false,
      "is_private": false,
      "is_locked": false,
      "allow_anonymous": false,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "post_count": 0,
      "is_owner": false,
      "is_favorite": false,
      "is_archived": false
    }
  ],
  "pagination": {
    "total": 0,
    "page": 0,
    "per_page": 0,
    "total_pages": 0
  }
}
```

#### Get Board by Slug

```
GET /boards/slug/:slug
```

Get a board by its unique slug.

**Authorization:** Optional

**Response:**
```json
{
  "success": true,
  "data": {
    "board": {
      "id": 0,
      "title": "string",
      "receiver_name": "string",
      "slug": "string",
      "max_post": 0,
      "creator": {
        "id": 0,
        "name": "string",
        "email": "string",
        "profile_picture": "string",
        "is_verified": false,
        "auth_provider": "string",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "font_name": "string",
      "font_size": 0,
      "header_color": "string",
      "show_header_color": false,
      "theme": {
        "id": 0,
        "category": "string",
        "name": "string",
        "icon_url": "string",
        "background_image_url": "string"
      },
      "effect": "string",
      "enable_intro_animation": false,
      "is_private": false,
      "is_locked": false,
      "allow_anonymous": false,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "post_count": 0
    },
    "posts": [
      {
        "id": 0,
        "board_id": 0,
        "author": {
          "id": 0,
          "name": "string",
          "email": "string",
          "profile_picture": "string",
          "is_verified": false,
          "auth_provider": "string",
          "created_at": "2023-01-01T00:00:00Z"
        },
        "author_name": "string",
        "content": "string",
        "background_color": "string",
        "text_color": "string",
        "position": 0,
        "media_path": "string",
        "media_type": "string",
        "media_source": "string",
        "likes_count": 0,
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ]
  }
}
```

#### Update Board

```
PUT /boards/:boardId
```

Update a board.

**Authorization:** Required

**Request Body:**
```json
{
  "title": "string",
  "receiver_name": "string",
  "font_name": "string",
  "font_size": 0,
  "header_color": "string",
  "show_header_color": false,
  "theme_id": 0,
  "effect": "string",
  "enable_intro_animation": false,
  "is_private": false,
  "allow_anonymous": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "title": "string",
    "receiver_name": "string",
    "slug": "string",
    "max_post": 0,
    "creator": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "font_name": "string",
    "font_size": 0,
    "header_color": "string",
    "show_header_color": false,
    "theme": {
      "id": 0,
      "category": "string",
      "name": "string",
      "icon_url": "string",
      "background_image_url": "string"
    },
    "effect": "string",
    "enable_intro_animation": false,
    "is_private": false,
    "is_locked": false,
    "allow_anonymous": false,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "post_count": 0
  }
}
```

#### Delete Board

```
DELETE /boards/:boardId
```

Delete a board.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Board deleted successfully"
  }
}
```

#### Toggle Board Lock

```
PATCH /boards/:boardId/lock
```

Lock or unlock a board.

**Authorization:** Required

**Request Body:**
```json
{
  "is_locked": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "title": "string",
    "receiver_name": "string",
    "slug": "string",
    "max_post": 0,
    "creator": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "font_name": "string",
    "font_size": 0,
    "header_color": "string",
    "show_header_color": false,
    "theme": {
      "id": 0,
      "category": "string",
      "name": "string",
      "icon_url": "string",
      "background_image_url": "string"
    },
    "effect": "string",
    "enable_intro_animation": false,
    "is_private": false,
    "is_locked": true,
    "allow_anonymous": false,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "post_count": 0
  }
}
```

#### Update Board Preferences

```
PATCH /boards/:boardId/preferences
```

Update a user's preferences for a board.

**Authorization:** Required

**Request Body:**
```json
{
  "is_favorite": true,
  "is_archived": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Board preferences updated successfully"
  }
}
```

#### List Board Contributors

```
GET /boards/:boardId/contributors
```

List all contributors for a board.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "board_id": 0,
      "user": {
        "id": 0,
        "name": "string",
        "email": "string",
        "profile_picture": "string",
        "is_verified": false,
        "auth_provider": "string",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "role": "viewer|contributor|admin",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

#### Add Contributor

```
POST /boards/:boardId/contributors
```

Add a contributor to a board.

**Authorization:** Required

**Request Body:**
```json
{
  "email": "string",
  "role": "viewer|contributor|admin"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "board_id": 0,
    "user": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "role": "viewer|contributor|admin",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Update Contributor

```
PUT /boards/:boardId/contributors/:contributorId
```

Update a contributor's role.

**Authorization:** Required

**Request Body:**
```json
{
  "role": "viewer|contributor|admin"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "board_id": 0,
    "user": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "role": "viewer|contributor|admin",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Remove Contributor

```
DELETE /boards/:boardId/contributors/:contributorId
```

Remove a contributor from a board.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Contributor removed successfully"
  }
}
```

#### Reorder Posts

```
PUT /boards/:boardId/posts/reorder
```

Update the order of posts on a board.

**Authorization:** Required

**Request Body:**
```json
{
  "post_positions": [
    {
      "id": 0,
      "position": 0
    }
  ]
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Posts reordered successfully"
  }
}
```

## Posts

### Endpoints

#### Create Post

```
POST /boards/:boardId/posts
```

Create a new post on a board.

**Authorization:** Optional (anonymous allowed if board settings permit)

**Request Body:**
```json
{
  "content": "string",
  "author_name": "string",
  "background_color": "string",
  "text_color": "string",
  "media_path": "string",
  "media_type": "string",
  "media_source": "internal|external"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "board_id": 0,
    "author": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "author_name": "string",
    "content": "string",
    "background_color": "string",
    "text_color": "string",
    "position": 0,
    "media_path": "string",
    "media_type": "string",
    "media_source": "string",
    "likes_count": 0,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Update Post

```
PUT /posts/:postId
```

Update a post.

**Authorization:** Required

**Request Body:**
```json
{
  "content": "string",
  "author_name": "string",
  "background_color": "string",
  "text_color": "string",
  "media_path": "string",
  "media_type": "string",
  "media_source": "internal|external"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "board_id": 0,
    "author": {
      "id": 0,
      "name": "string",
      "email": "string",
      "profile_picture": "string",
      "is_verified": false,
      "auth_provider": "string",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "author_name": "string",
    "content": "string",
    "background_color": "string",
    "text_color": "string",
    "position": 0,
    "media_path": "string",
    "media_type": "string",
    "media_source": "string",
    "likes_count": 0,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Delete Post

```
DELETE /posts/:postId
```

Delete a post.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Post deleted successfully"
  }
}
```

#### Like Post

```
POST /posts/:postId/like
```

Add a like to a post.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Post liked successfully",
    "likes_count": 0
  }
}
```

#### Unlike Post

```
DELETE /posts/:postId/like
```

Remove a like from a post.

**Authorization:** Required

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Post unliked successfully",
    "likes_count": 0
  }
}
```

## Themes

### Endpoints

#### List Themes

```
GET /themes
```

List all available themes.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 0,
      "category": "string",
      "name": "string",
      "icon_url": "string",
      "background_image_url": "string"
    }
  ]
}
```

#### Get Theme

```
GET /themes/:themeId
```

Get a theme by ID.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "category": "string",
    "name": "string",
    "icon_url": "string",
    "background_image_url": "string"
  }
}
```

#### Create Theme (Admin Only)

```
POST /themes
```

Create a new theme.

**Authorization:** Admin Only

**Request Body:**
```json
{
  "category": "string",
  "name": "string",
  "icon_url": "string",
  "background_image_url": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "category": "string",
    "name": "string",
    "icon_url": "string",
    "background_image_url": "string"
  }
}
```

#### Update Theme (Admin Only)

```
PUT /themes/:themeId
```

Update an existing theme.

**Authorization:** Admin Only

**Request Body:**
```json
{
  "category": "string",
  "name": "string",
  "icon_url": "string",
  "background_image_url": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": 0,
    "category": "string",
    "name": "string",
    "icon_url": "string",
    "background_image_url": "string"
  }
}
```

#### Delete Theme (Admin Only)

```
DELETE /themes/:themeId
```

Delete a theme.

**Authorization:** Admin Only

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Theme deleted successfully"
  }
}
```

## Files

### Endpoints

#### Upload File

```
POST /files/upload
```

Upload a file.

**Authorization:** Optional

**Content-Type:** multipart/form-data

**Form Data:**
- `file`: The file to upload
- `category`: File category (optional, default: "general")

**Response:**
```json
{
  "success": true,
  "data": {
    "file_name": "string",
    "file_path": "string",
    "file_type": "string",
    "file_size": 0,
    "content_type": "string",
    "uploaded_at": "2023-01-01T00:00:00Z"
  }
}
```

#### Delete File

```
DELETE /files
```

Delete a file.

**Authorization:** Required

**Request Body:**
```json
{
  "file_path": "string"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "File deleted successfully"
  }
}
```

#### Clean Orphaned Files (Admin Only)

```
POST /files/cleanup-orphaned
```

Trigger a cleanup of orphaned files.

**Authorization:** Admin Only

**Response:**
```json
{
  "success": true,
  "data": {
    "message": "Orphaned file cleanup job started successfully"
  }
}
```

## Giphy Integration

### Endpoints

#### Search GIFs

```
GET /giphy/search
```

Search for GIFs.

**Authorization:** None

**Query Parameters:**
- `q`: Search query (required)
- `limit`: Maximum number of results (default: 25)
- `offset`: Results offset (default: 0)
- `rating`: Content rating filter
- `lang`: Language filter

**Response:**
```json
{
  "success": true,
  "data": {
    "data": [],
    "pagination": {},
    "meta": {}
  }
}
```

#### Trending GIFs

```
GET /giphy/trending
```

Get trending GIFs.

**Authorization:** None

**Query Parameters:**
- `limit`: Maximum number of results (default: 25)
- `offset`: Results offset (default: 0)
- `rating`: Content rating filter

**Response:**
```json
{
  "success": true,
  "data": {
    "data": [],
    "pagination": {},
    "meta": {}
  }
}
```

#### Get GIF by ID

```
GET /giphy/:gifId
```

Get a specific GIF by ID.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {
    "data": {},
    "meta": {}
  }
}
```

#### Random GIF

```
GET /giphy/random
```

Get a random GIF.

**Authorization:** None

**Query Parameters:**
- `tag`: Tag to limit results to
- `rating`: Content rating filter

**Response:**
```json
{
  "success": true,
  "data": {
    "data": {},
    "meta": {}
  }
}
```

## Unsplash Integration

### Endpoints

#### Search Photos

```
GET /unsplash/search
```

Search for photos.

**Authorization:** None

**Query Parameters:**
- `query`: Search query (required)
- `page`: Page number (default: 1)
- `per_page`: Results per page (default: 10)
- `order_by`: Order results by (default: "relevant")

**Response:**
```json
{
  "success": true,
  "data": {
    "total": 0,
    "total_pages": 0,
    "results": []
  }
}
```

#### Random Photos

```
GET /unsplash/random
```

Get random photos.

**Authorization:** None

**Query Parameters:**
- `count`: Number of photos (default: 1, max: 30)
- `query`: Filter results by search term
- `topics`: Filter by topics
- `username`: Filter by username
- `collections`: Filter by collections
- `featured`: Featured photos only (default: false)

**Response:**
```json
{
  "success": true,
  "data": {
    "results": []
  }
}
```

#### Get Photo by ID

```
GET /unsplash/:photoId
```

Get a specific photo by ID.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {}
}
```

## Health Check

### Endpoints

#### Liveness Check

```
GET /health
```

Basic liveness probe.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "UP",
    "version": "string",
    "environment": "string",
    "timestamp": "2023-01-01T00:00:00Z",
    "uptime": "string"
  }
}
```

#### Readiness Check

```
GET /health/readiness
```

Readiness probe.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "UP",
    "version": "string",
    "environment": "string",
    "timestamp": "2023-01-01T00:00:00Z",
    "components": {
      "database": "UP",
      "storage": "UP",
      "giphy": "CONFIGURED",
      "unsplash": "CONFIGURED"
    },
    "uptime": "string"
  }
}
```

#### Detailed Health Check

```
GET /health/detailed
```

Comprehensive health check.

**Authorization:** None

**Response:**
```json
{
  "success": true,
  "data": {
    "status": "UP",
    "version": "string",
    "environment": "string",
    "timestamp": "2023-01-01T00:00:00Z",
    "components": {
      "database": "UP",
      "database_open_connections": "string",
      "database_in_use": "string",
      "database_idle": "string"
    },
    "uptime": "string"
  }
}
```

## Error Handling

All API endpoints return a standardized error response format:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": "Additional error details (only in development mode)"
  }
}
```

### Common Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `NOT_FOUND` | 404 | Resource not found |
| `UNAUTHORIZED` | 401 | Authentication required or invalid |
| `FORBIDDEN` | 403 | User lacks permission for the action |
| `BAD_REQUEST` | 400 | Invalid request parameters |
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `INTERNAL_ERROR` | 500 | Server error |
| `RATE_LIMIT_EXCEEDED` | 429 | Too many requests |