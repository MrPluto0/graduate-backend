# Go Backend Project

This is a basic Go backend project structure that includes various modules for handling user-related requests, middleware for authentication and logging, and configuration management.

## Project Structure

```
go-backend
├── cmd
│   └── server
│       └── main.go          # Entry point of the application
├── internal
│   ├── api
│   │   ├── handlers
│   │   │   ├── user_handler.go    # User-related request handlers
│   │   │   └── health_handler.go   # Health check handler
│   │   ├── middleware
│   │   │   ├── auth.go            # Authentication middleware
│   │   │   └── logging.go         # Logging middleware
│   │   └── routes.go              # API routes and middleware registration
│   ├── config
│   │   └── config.go              # Configuration management
│   ├── models
│   │   └── user.go                # User model definition
│   ├── repository
│   │   └── user_repository.go      # User data storage interaction
│   └── service
│       └── user_service.go        # Business logic for user operations
├── pkg
│   ├── database
│   │   └── connection.go          # Database connection functions
│   └── utils
│       └── helper.go              # Utility functions
├── configs
│   └── config.yaml                # Application configuration file
├── go.mod                         # Go module configuration
├── go.sum                         # Go module dependency checksums
└── README.md                      # Project documentation
```

## Getting Started

1. **Clone the repository:**
   ```
   git clone <repository-url>
   cd go-backend
   ```

2. **Install dependencies:**
   ```
   go mod tidy
   ```

3. **Run the application:**
   ```
   go run cmd/server/main.go
   ```

## API Endpoints

- **User Endpoints**
  - `POST /users` - Create a new user
  - `GET /users/{id}` - Retrieve user information

- **Health Check**
  - `GET /health` - Check the health status of the service

## Configuration

The application configuration can be found in `configs/config.yaml`. Modify this file to set up your application parameters.

## License

This project is licensed under the MIT License. See the LICENSE file for details.