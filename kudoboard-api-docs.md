# KudoBoard API Documentation

## Overview

The KudoBoard API provides a comprehensive set of endpoints for creating and managing collaborative message boards where users can post messages, upload media, and customize the appearance. This document outlines all available API endpoints, request and response formats, authentication requirements, and expected behaviors.

## Base URL

```
https://api.kudoboard.com/api/v1
```

For local development:

```
http://localhost:8080/api/v1
```

## Authentication

Most endpoints require authentication using JWT tokens. Include the token in the Authorization header as a Bearer token:

```
Authorization: Bearer <your_token>
```

## Response Format

All API responses follow a standard format:

### Success Response

```json
{
  "success": true,
  "data": { ... },
  "pagination": { ... } // Optional, included for paginated endpoints
}
```

### Error Response

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

## Common Error Codes

- `VALIDATION_ERROR`: Invalid request parameters
- `UNAUTHORIZED`: Authentication required
- `FORBIDDEN`: Insufficient permissions
- `NOT_FOUND`: Resource not found
- `INTERNAL_SERVER_ERROR`: Server-side error

## Pagination

Paginated endpoints include pagination information in the response:

```json
"pagination": {
  "total": 100,
  "page": 2,
  "per_page": 10,
  "total_pages": 10
}
```

Request pagination parameters:

- `page`: Page number (default: 1)
- `per_page`: Items per page (default: 10, max: 100)

---

# Endpoints

## Authentication

### Register

Create a new user account.

- **URL**: `/auth/register`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "name": "User Name",
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "token": "jwt_token_here",
    "user": {
      "id": 1,
      "name": "User Name",
      "email": "user@example.com",
      "profile_picture": "",
      "is_verified": false,
      "auth_provider": "local",
      "created_at": "2023-01-01T00:00:00Z"
    }
  }
}
```

### Login

Authenticate with email and password.

- **URL**: `/auth/login`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "email": "user@example.com",
  "password": "password123"
}
```

**Response**: Same as Register response

### Google Login

Authenticate with Google OAuth.

- **URL**: `/auth/google`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "access_token": "google_access_token"
}
```

**Response**: Same as Register response

### Facebook Login

Authenticate with Facebook OAuth.

- **URL**: `/auth/facebook`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "access_token": "facebook_access_token"
}
```

**Response**: Same as Register response

### Get Current User

Get the currently authenticated user's information.

- **URL**: `/auth/me`
- **Method**: `GET`
- **Auth required**: Yes

**Response**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "User Name",
    "email": "user@example.com",
    "profile_picture": "https://example.com/avatar.jpg",
    "is_verified": true,
    "auth_provider": "google",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

### Update Profile

Update the current user's profile information.

- **URL**: `/auth/me`
- **Method**: `PUT`
- **Auth required**: Yes

**Request Body**:

```json
{
  "name": "Updated Name",
  "profile_picture": "https://example.com/new-avatar.jpg"
}
```

**Response**: Updated user object (same format as Get Current User)

### Forgot Password

Initiate the password reset process.

- **URL**: `/auth/forgot-password`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "email": "user@example.com"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "If your email exists in our system, you will receive a password reset link"
  }
}
```

### Reset Password

Reset password using token from email.

- **URL**: `/auth/reset-password`
- **Method**: `POST`
- **Auth required**: No

**Request Body**:

```json
{
  "token": "reset_token_from_email",
  "password": "new_password"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Password has been reset successfully"
  }
}
```

## Boards

### Create Board

Create a new KudoBoard.

- **URL**: `/boards`
- **Method**: `POST`
- **Auth required**: Yes

**Request Body**:

```json
{
  "title": "Happy Birthday John!",
  "receiver_name": "John Doe",
  "font_name": "Arial",
  "font_size": 16,
  "header_color": "#4285F4",
  "theme_id": 1,
  "effect": "{\"type\":\"confetti\",\"enabled\":true}",
  "enable_intro_animation": true,
  "is_private": false,
  "allow_anonymous": true
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "title": "Happy Birthday John!",
    "receiver_name": "John Doe",
    "slug": "abcdef123456789",
    "max_post": 10,
    "creator": {
      "id": 1,
      "name": "User Name",
      "email": "user@example.com",
      "profile_picture": "",
      "is_verified": true,
      "auth_provider": "local",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "font_name": "Arial",
    "font_size": 16,
    "header_color": "#4285F4",
    "show_header_color": true,
    "theme": {
      "id": 1,
      "category": "birthday",
      "name": "Birthday Theme",
      "icon_url": "https://example.com/icons/birthday.png",
      "background_image_url": "https://example.com/backgrounds/birthday.jpg"
    },
    "effect": "{\"type\":\"confetti\",\"enabled\":true}",
    "enable_intro_animation": true,
    "is_private": false,
    "is_locked": false,
    "allow_anonymous": true,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z",
    "post_count": 0
  }
}
```

### List User Boards

Get all boards created by or accessible to the current user.

- **URL**: `/boards`
- **Method**: `GET`
- **Auth required**: Yes
- **Query Parameters**:
    - `page` (optional): Page number for pagination
    - `per_page` (optional): Items per page
    - `search` (optional): Search term for board title or receiver name
    - `sort_by` (optional): Field to sort by (`created_at` or `title`)
    - `order` (optional): Sort order (`asc` or `desc`)

**Response**:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "Happy Birthday John!",
      "receiver_name": "John Doe",
      "slug": "abcdef123456789",
      "max_post": 10,
      "creator": {
        "id": 1,
        "name": "User Name",
        "email": "user@example.com",
        "profile_picture": "",
        "is_verified": true,
        "auth_provider": "local",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "font_name": "Arial",
      "font_size": 16,
      "header_color": "#4285F4",
      "show_header_color": true,
      "effect": "{\"type\":\"confetti\",\"enabled\":true}",
      "enable_intro_animation": true,
      "is_private": false,
      "is_locked": false,
      "allow_anonymous": true,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "post_count": 0,
      "is_owner": true,
      "is_favorite": false,
      "is_archived": false
    }
  ],
  "pagination": {
    "total": 1,
    "page": 1,
    "per_page": 10,
    "total_pages": 1
  }
}
```

### Get Board by Slug

Get a board by its unique slug.

- **URL**: `/boards/slug/:slug`
- **Method**: `GET`
- **Auth required**: Optional (required for private boards)

**Response**:

```json
{
  "success": true,
  "data": {
    "board": {
      "id": 1,
      "title": "Happy Birthday John!",
      "receiver_name": "John Doe",
      "slug": "abcdef123456789",
      "max_post": 10,
      "creator": {
        "id": 1,
        "name": "User Name",
        "email": "user@example.com",
        "profile_picture": "",
        "is_verified": true,
        "auth_provider": "local",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "font_name": "Arial",
      "font_size": 16,
      "header_color": "#4285F4",
      "show_header_color": true,
      "theme": {
        "id": 1,
        "category": "birthday",
        "name": "Birthday Theme",
        "icon_url": "https://example.com/icons/birthday.png",
        "background_image_url": "https://example.com/backgrounds/birthday.jpg"
      },
      "effect": "{\"type\":\"confetti\",\"enabled\":true}",
      "enable_intro_animation": true,
      "is_private": false,
      "is_locked": false,
      "allow_anonymous": true,
      "created_at": "2023-01-01T00:00:00Z",
      "updated_at": "2023-01-01T00:00:00Z",
      "post_count": 1
    },
    "posts": [
      {
        "id": 1,
        "board_id": 1,
        "author": {
          "id": 2,
          "name": "Contributor",
          "email": "contributor@example.com",
          "profile_picture": "",
          "is_verified": true,
          "auth_provider": "local",
          "created_at": "2023-01-01T00:00:00Z"
        },
        "author_name": "Contributor",
        "content": "Happy birthday! Have a great day!",
        "background_color": "#ffffff",
        "text_color": "#000000",
        "position": 0,
        "media_path": "https://example.com/uploads/image/birthday-cake.jpg",
        "media_type": "image",
        "media_source": "internal",
        "likes_count": 2,
        "created_at": "2023-01-01T00:00:00Z",
        "updated_at": "2023-01-01T00:00:00Z"
      }
    ]
  }
}
```

### Update Board

Update a board's details.

- **URL**: `/boards/:boardId`
- **Method**: `PUT`
- **Auth required**: Yes (must be board creator)

**Request Body**:

```json
{
  "title": "Updated Board Title",
  "receiver_name": "Updated Name",
  "font_name": "Helvetica",
  "font_size": 18,
  "header_color": "#FF5722",
  "show_header_color": true,
  "theme_id": 2,
  "effect": "{\"type\":\"hearts\",\"enabled\":true}",
  "enable_intro_animation": false,
  "is_private": true,
  "allow_anonymous": false
}
```

Note: All fields are optional. Only include fields you want to update.

**Response**: Updated board object (same format as Create Board)

### Delete Board

Delete a board.

- **URL**: `/boards/:boardId`
- **Method**: `DELETE`
- **Auth required**: Yes (must be board creator)

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Board deleted successfully"
  }
}
```

### Toggle Board Lock

Lock or unlock a board to prevent new posts or modifications.

- **URL**: `/boards/:boardId/lock`
- **Method**: `PATCH`
- **Auth required**: Yes (must be board creator or admin)

**Request Body**:

```json
{
  "is_locked": true
}
```

**Response**: Updated board object (same format as Create Board)

### Update Board Preferences

Update user-specific preferences for a board (favorite/archived status).

- **URL**: `/boards/:boardId/preferences`
- **Method**: `PATCH`
- **Auth required**: Yes (must have access to the board)

**Request Body**:

```json
{
  "is_favorite": true,
  "is_archived": false
}
```

Note: Both fields are optional. Only include fields you want to update.

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Board preferences updated successfully"
  }
}
```

### List Board Contributors

List all contributors for a board.

- **URL**: `/boards/:boardId/contributors`
- **Method**: `GET`
- **Auth required**: Yes (must have access to the board)

**Response**:

```json
{
  "success": true,
  "data": [
    {
      "board_id": 1,
      "user": {
        "id": 1,
        "name": "User Name",
        "email": "user@example.com",
        "profile_picture": "",
        "is_verified": true,
        "auth_provider": "local",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "role": "admin",
      "created_at": "2023-01-01T00:00:00Z"
    },
    {
      "board_id": 1,
      "user": {
        "id": 2,
        "name": "Contributor",
        "email": "contributor@example.com",
        "profile_picture": "",
        "is_verified": true,
        "auth_provider": "local",
        "created_at": "2023-01-01T00:00:00Z"
      },
      "role": "contributor",
      "created_at": "2023-01-01T00:00:00Z"
    }
  ]
}
```

### Add Contributor

Add a new contributor to a board.

- **URL**: `/boards/:boardId/contributors`
- **Method**: `POST`
- **Auth required**: Yes (must be board creator)

**Request Body**:

```json
{
  "email": "contributor@example.com",
  "role": "contributor" // "viewer", "contributor", or "admin"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "board_id": 1,
    "user": {
      "id": 2,
      "name": "Contributor",
      "email": "contributor@example.com",
      "profile_picture": "",
      "is_verified": true,
      "auth_provider": "local",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "role": "contributor",
    "created_at": "2023-01-01T00:00:00Z"
  }
}
```

### Update Contributor

Update a contributor's role.

- **URL**: `/boards/:boardId/contributors/:contributorId`
- **Method**: `PUT`
- **Auth required**: Yes (must be board creator)

**Request Body**:

```json
{
  "role": "admin" // "viewer", "contributor", or "admin"
}
```

**Response**: Updated contributor object (same format as Add Contributor)

### Remove Contributor

Remove a contributor from a board.

- **URL**: `/boards/:boardId/contributors/:contributorId`
- **Method**: `DELETE`
- **Auth required**: Yes (must be board creator)

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Contributor removed successfully"
  }
}
```

## Posts

### Create Post

Create a new post on a board.

- **URL**: `/boards/:boardId/posts`
- **Method**: `POST`
- **Auth required**: Optional (anonymous posts allowed if board settings permit)

**Request Body**:

```json
{
  "content": "Happy birthday! Have a great day!",
  "author_name": "Anonymous", // Required for anonymous posts
  "background_color": "#ffffff",
  "text_color": "#000000",
  "media_path": "https://example.com/uploads/image/birthday-cake.jpg",
  "media_type": "image", // "image", "gif", "video", "youtube"
  "media_source": "internal" // "internal" or "external"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "board_id": 1,
    "author": {
      "id": 2,
      "name": "Contributor",
      "email": "contributor@example.com",
      "profile_picture": "",
      "is_verified": true,
      "auth_provider": "local",
      "created_at": "2023-01-01T00:00:00Z"
    },
    "author_name": "Contributor",
    "content": "Happy birthday! Have a great day!",
    "background_color": "#ffffff",
    "text_color": "#000000",
    "position": 0,
    "media_path": "https://example.com/uploads/image/birthday-cake.jpg",
    "media_type": "image",
    "media_source": "internal",
    "likes_count": 0,
    "created_at": "2023-01-01T00:00:00Z",
    "updated_at": "2023-01-01T00:00:00Z"
  }
}
```

### Update Post

Update an existing post.

- **URL**: `/posts/:postId`
- **Method**: `PUT`
- **Auth required**: Yes (must be post author, board creator, or board admin)

**Request Body**:

```json
{
  "content": "Updated content",
  "author_name": "Updated Name",
  "background_color": "#f0f0f0",
  "text_color": "#333333",
  "media_path": "https://example.com/uploads/image/new-image.jpg",
  "media_type": "image",
  "media_source": "internal"
}
```

Note: All fields are optional. Only include fields you want to update.

**Response**: Updated post object (same format as Create Post)

### Delete Post

Delete a post.

- **URL**: `/posts/:postId`
- **Method**: `DELETE`
- **Auth required**: Yes (must be post author, board creator, or board admin)

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Post deleted successfully"
  }
}
```

### Like Post

Add a like to a post.

- **URL**: `/posts/:postId/like`
- **Method**: `POST`
- **Auth required**: Yes

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Post liked successfully",
    "likes_count": 3
  }
}
```

### Unlike Post

Remove a like from a post.

- **URL**: `/posts/:postId/like`
- **Method**: `DELETE`
- **Auth required**: Yes

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Post unliked successfully",
    "likes_count": 2
  }
}
```

### Reorder Posts

Update the order of posts on a board.

- **URL**: `/boards/:boardId/posts/reorder`
- **Method**: `PUT`
- **Auth required**: Yes (must be board creator or admin)

**Request Body**:

```json
{
  "post_positions": [
    {
      "id": 1,
      "position": 2
    },
    {
      "id": 2,
      "position": 1
    },
    {
      "id": 3,
      "position": 3
    }
  ]
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Posts reordered successfully"
  }
}
```

## Themes

### List Themes

Get all available themes.

- **URL**: `/themes`
- **Method**: `GET`
- **Auth required**: No

**Response**:

```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "category": "birthday",
      "name": "Birthday Theme",
      "icon_url": "https://example.com/icons/birthday.png",
      "background_image_url": "https://example.com/backgrounds/birthday.jpg"
    },
    {
      "id": 2,
      "category": "congratulations",
      "name": "Congratulations Theme",
      "icon_url": "https://example.com/icons/congrats.png",
      "background_image_url": "https://example.com/backgrounds/congrats.jpg"
    }
  ]
}
```

### Get Theme

Get a theme by ID.

- **URL**: `/themes/:themeId`
- **Method**: `GET`
- **Auth required**: No

**Response**:

```json
{
  "success": true,
  "data": {
    "id": 1,
    "category": "birthday",
    "name": "Birthday Theme",
    "icon_url": "https://example.com/icons/birthday.png",
    "background_image_url": "https://example.com/backgrounds/birthday.jpg"
  }
}
```

### Create Theme (Admin Only)

Create a new theme.

- **URL**: `/themes`
- **Method**: `POST`
- **Auth required**: Yes (admin only)

**Request Body**:

```json
{
  "category": "farewell",
  "name": "Farewell Theme",
  "icon_url": "https://example.com/icons/farewell.png",
  "background_image_url": "https://example.com/backgrounds/farewell.jpg"
}
```

**Response**: Created theme object (same format as Get Theme)

### Update Theme (Admin Only)

Update an existing theme.

- **URL**: `/themes/:themeId`
- **Method**: `PUT`
- **Auth required**: Yes (admin only)

**Request Body**:

```json
{
  "category": "updated_category",
  "name": "Updated Theme Name",
  "icon_url": "https://example.com/icons/updated.png",
  "background_image_url": "https://example.com/backgrounds/updated.jpg"
}
```

Note: All fields are optional. Only include fields you want to update.

**Response**: Updated theme object (same format as Get Theme)

### Delete Theme (Admin Only)

Delete a theme.

- **URL**: `/themes/:themeId`
- **Method**: `DELETE`
- **Auth required**: Yes (admin only)

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "Theme deleted successfully"
  }
}
```

## Files

### Upload File

Upload a file (image, GIF, or video).

- **URL**: `/files/upload`
- **Method**: `POST`
- **Auth required**: Optional
- **Content-Type**: `multipart/form-data`

**Request Body**:

- `file`: The file to upload (required)
- `category`: Category of the file (optional, default: "general")

**Response**:

```json
{
  "success": true,
  "data": {
    "file_name": "image-20230101123456-abcdef.jpg",
    "file_path": "/uploads/image/user_1/image-20230101123456-abcdef.jpg",
    "file_type": "image",
    "file_size": 102400,
    "content_type": "image/jpeg",
    "uploaded_at": "2023-01-01T12:34:56Z"
  }
}
```

### Delete File

Delete a file.

- **URL**: `/files`
- **Method**: `DELETE`
- **Auth required**: Yes

**Request Body**:

```json
{
  "file_path": "/uploads/image/user_1/image-20230101123456-abcdef.jpg"
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "message": "File deleted successfully"
  }
}
```

## Misc

### Health Check

Check API health status.

- **URL**: `/health`
- **Method**: `GET`
- **Auth required**: No

**Response**:

```json
{
  "status": "ok"
}
```

## Role-Based Access Control

The API enforces different access levels for board contributors:

1. **Viewer**: Can only view a board and its posts
2. **Contributor**: Can view the board and add posts
3. **Admin**: Can view, add posts, modify board settings, and manage contributors
4. **Creator**: Full control over the board (same as admin plus the ability to delete the board)

## Errors

If an error occurs, the API will return an appropriate HTTP status code and a standardized error response:

```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

Common HTTP status codes:

- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required
- `403 Forbidden`: Insufficient permissions
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Server error

## Rate Limiting

The API enforces rate limits to prevent abuse. When rate limited, the API will return a `429 Too Many Requests` response.