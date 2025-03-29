# Kudoboard API

A Go backend for a collaborative kudos board application, similar to Kudoboard.com.

## Features

- 🔐 User authentication with JWT (including Google & Facebook login)
- 📋 Board creation with customizable themes and backgrounds
- 📝 Posts with rich text, images, GIFs, and YouTube videos
- 👥 Board collaboration with multiple users
- ❤️ Like/unlike posts
- 🔄 Reordering posts on a board
- 🌐 Anonymous posting support
- 🔍 Public and private boards
- 🖼️ Media upload and management
- 🎨 Customizable themes

## Tech Stack

- **Go**: Programming language
- **Gin**: Web framework
- **GORM**: ORM for PostgreSQL
- **PostgreSQL**: Database
- **JWT**: Authentication
- **Docker**: Containerization

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker (optional)

## Getting Started

### Environment Setup

1. Copy the example environment file:
   ```bash
   cp .env.example .env
   ```

2. Update the `.env` file with your specific configuration.

### Running with Docker

```bash
# Build and start the containers
docker-compose up -d

# Stop the containers
docker-compose down
```

### Running Locally

1. Install dependencies:
   ```bash
   go mod download
   ```

2. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

### API Documentation

The API is accessible at `http://localhost:8080/api/v1`.

#### Authentication Endpoints

- `POST /api/v1/auth/register`: Register a new user
- `POST /api/v1/auth/login`: Login
- `POST /api/v1/auth/google`: Login with Google
- `POST /api/v1/auth/facebook`: Login with Facebook
- `GET /api/v1/auth/me`: Get current user
- `PUT /api/v1/auth/me`: Update user profile

#### Board Endpoints

- `GET /api/v1/boards`: List user's boards
- `POST /api/v1/boards`: Create a new board
- `GET /api/v1/boards/public`: List public boards
- `GET /api/v1/boards/:slug`: Get board by slug
- `PUT /api/v1/boards/:id`: Update board
- `DELETE /api/v1/boards/:id`: Delete board

#### Post Endpoints

- `POST /api/v1/posts/:boardId`: Create post
- `POST /api/v1/posts/anonymous/:boardId`: Create anonymous post
- `PUT /api/v1/posts/:id`: Update post
- `DELETE /api/v1/posts/:id`: Delete post
- `POST /api/v1/posts/:id/like`: Like post
- `DELETE /api/v1/posts/:id/like`: Unlike post
- `PUT /api/v1/posts/reorder/:boardId`: Reorder posts

#### Media Endpoints

- `POST /api/v1/media/upload`: Upload media
- `POST /api/v1/media/youtube`: Add YouTube video
- `DELETE /api/v1/media/:id`: Delete media

## Project Structure

```
kudoboard-api/
├── cmd/
│   └── api/               # Application entry point
├── internal/
│   ├── api/               # API handlers and routes
│   ├── config/            # Configuration
│   ├── db/                # Database connection
│   ├── dto/               # Data Transfer Objects
│   ├── models/            # Database models
│   ├── services/          # Business logic
│   └── utils/             # Utility functions
├── migrations/            # Database migrations
├── .env                   # Environment variables
├── .env.example           # Example environment file
├── Dockerfile             # Docker configuration
├── docker-compose.yml     # Docker Compose configuration
└── README.md              # Project documentation
```

## License

MIT