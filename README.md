# TransJakarta Routes API

A RESTful API application built with Go (Gin framework) to manage TransJakarta BRT route information, stops/terminals, vehicles, and user reports.

## Features

- **User Authentication**:

  - Traditional email/password authentication
  - OAuth authentication (Google)
  - JWT token-based authentication

- **Role-Based Access Control**:

  - Common users can view data and submit reports
  - Admins can perform full CRUD operations

- **Route Management**:

  - Routes with multiple stops in sequence
  - Track route changes and modifications
  - Support for temporary route changes

- **Stop/Terminal Management**:

  - Detailed location information
  - Facilities information
  - Status tracking

- **Vehicle Management**:

  - Vehicle assignment to routes
  - Capacity and type tracking

- **Reporting System**:
  - Users can submit reports about route/stop issues
  - Status tracking (pending → reviewed → resolved)
  - Admin notes and resolution tracking

- **Real-Time Chat**:
  - Direct messaging with follow-relationship validation
  - Group chat with roles and permissions
  - Forum chat with 24-hour auto-clear
  - Message reactions and threaded replies
  - Typing indicators and read receipts
  - Online presence tracking

## Technology Stack

- **Framework**: Gin
- **Database**: PostgreSQL with GORM
- **Authentication**: JWT + OAuth2
- **Password Hashing**: bcrypt

## Getting Started

### Prerequisites

- Go 1.25 or higher
- PostgreSQL 12 or higher
- Google OAuth credentials (for OAuth authentication)

### Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd tj-routes
```

2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables:
   Create a `.env` file in the root directory with the following variables (see Environment Variables section below for details):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=tj_routes
DB_SSLMODE=disable

# Server
SERVER_HOST=localhost
SERVER_PORT=8080
ENVIRONMENT=development

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRATION_HOURS=24
JWT_REFRESH_EXPIRATION_HOURS=168

# OAuth
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/v1/auth/oauth/google/callback

# Logging
LOG_LEVEL=info
```

4. Set up the database:

```bash
# Create PostgreSQL database
createdb tj_routes

# Or use your preferred database setup method
```

5. Run migrations (automatic on startup):
   The application will automatically run migrations when it starts.

6. Start the server:

```bash
go run cmd/api/main.go
```

The API will be available at `http://localhost:8080`

### Accessing from Mobile Device (Same WiFi Network)

To access the API from your phone or other devices on the same WiFi network:

1. **Find your computer's local IP address:**

   - **macOS/Linux**: Run `ifconfig | grep "inet " | grep -v 127.0.0.1` or `ipconfig getifaddr en0` (macOS)
   - **Windows**: Run `ipconfig` and look for "IPv4 Address" under your WiFi adapter
   - The IP will typically look like `192.168.x.x` or `10.0.x.x`

2. **Update your `.env` file** to bind to all network interfaces:

   ```bash
   SERVER_HOST=0.0.0.0
   SERVER_PORT=8080
   ```

   (Change `SERVER_HOST` from `localhost` to `0.0.0.0`)

3. **Restart the server** with the updated configuration

4. **Access from your phone:**
   - Open a browser or API client on your phone
   - Use: `http://YOUR_LOCAL_IP:8080`
   - Example: `http://192.168.1.100:8080`
   - Test the health endpoint: `http://YOUR_LOCAL_IP:8080/health`
   - API docs: `http://YOUR_LOCAL_IP:8080/api/docs`

**Note**: Make sure your phone and computer are on the same WiFi network, and that your firewall allows incoming connections on port 8080.

### API Documentation

Once the server is running, you can access the interactive API documentation:

- OpenAPI: `http://localhost:8080/api/docs`

## API Endpoints

### Health Check

- `GET /health` - Health check endpoint (public)

### Authentication

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login with email/password
- `GET /api/v1/auth/oauth/:provider` - Initiate OAuth (e.g., `/api/v1/auth/oauth/google`)
- `GET /api/v1/auth/oauth/:provider/callback` - OAuth callback (e.g., `/api/v1/auth/oauth/google/callback`)

**Note**: Most endpoints below `/api/v1/auth` require authentication via JWT token in the `Authorization` header: `Bearer <token>`. However, some endpoints are available to guest users (non-authenticated) as indicated below.

### Guest Access

Guest users (non-authenticated) can access:

- **Read-only endpoints**: Routes, Stops, and Vehicles (GET requests)
- **Report creation**: Create reports (assigned to system user account)

All write operations (POST, PUT, DELETE) require authentication, and admin operations require both authentication and admin role.

### Stops/Terminals

- `GET /api/v1/stops` - List all stops (paginated) - **Guest accessible**
- `GET /api/v1/stops/:id` - Get stop details - **Guest accessible**
- `POST /api/v1/stops` - Create stop (admin only, requires authentication)
- `PUT /api/v1/stops/:id` - Update stop (admin only, requires authentication)
- `DELETE /api/v1/stops/:id` - Delete stop (admin only, requires authentication)

### Routes

- `GET /api/v1/routes` - List all routes (paginated) - **Guest accessible**
- `GET /api/v1/routes/:id` - Get route details with stops - **Guest accessible**
- `POST /api/v1/routes` - Create route with stops (admin only, requires authentication)
- `PUT /api/v1/routes/:id` - Update route metadata (admin only, requires authentication)
- `PUT /api/v1/routes/:id/stops` - Update route stops (admin only, requires authentication)
- `DELETE /api/v1/routes/:id` - Delete route (admin only, requires authentication)

### Vehicles

- `GET /api/v1/vehicles` - List all vehicles (paginated) - **Guest accessible**
- `GET /api/v1/vehicles/:id` - Get vehicle details - **Guest accessible**
- `POST /api/v1/vehicles` - Create vehicle (admin only, requires authentication)
- `PUT /api/v1/vehicles/:id` - Update vehicle (admin only, requires authentication)
- `DELETE /api/v1/vehicles/:id` - Delete vehicle (admin only, requires authentication)

### Reports

- `GET /api/v1/reports` - List reports (authenticated users see only their own, admins see all, requires authentication)
- `GET /api/v1/reports/:id` - Get report details (authenticated users see own, admins see all, requires authentication)
- `POST /api/v1/reports` - Create report - **Guest accessible** (guest reports are assigned to system user account)
- `PUT /api/v1/reports/:id/status` - Update report status (admin only, requires authentication)
- `DELETE /api/v1/reports/:id` - Delete report (admin only, requires authentication)

### Users (Admin only)

- `GET /api/v1/users` - List users
- `GET /api/v1/users/:id` - Get user details
- `PUT /api/v1/users/:id/role` - Update user role

### Chat System

**All chat endpoints require authentication.**

#### Direct Chat
- `POST /api/v1/chat/requests` - Create a chat request
- `GET /api/v1/chat/requests/sent` - List sent chat requests
- `GET /api/v1/chat/requests/received` - List received chat requests
- `POST /api/v1/chat/requests/:id/accept` - Accept chat request
- `POST /api/v1/chat/requests/:id/reject` - Reject chat request

#### Conversations
- `POST /api/v1/chat/conversations` - Create conversation
- `GET /api/v1/chat/conversations` - List user conversations
- `GET /api/v1/chat/conversations/:id` - Get conversation details
- `GET /api/v1/chat/conversations/:id/messages` - Get conversation messages
- `PUT /api/v1/chat/conversations/:id/read` - Mark conversation as read
- `DELETE /api/v1/chat/conversations/:id` - Delete conversation

#### Group Chat
- `POST /api/v1/chat/groups` - Create group
- `GET /api/v1/chat/groups` - List user's groups
- `GET /api/v1/chat/groups/:id` - Get group details
- `PUT /api/v1/chat/groups/:id` - Update group
- `DELETE /api/v1/chat/groups/:id` - Delete group
- `PUT /api/v1/chat/groups/:id/avatar` - Update group avatar

#### Group Members
- `POST /api/v1/chat/groups/:id/members` - Add member
- `GET /api/v1/chat/groups/:id/members` - List group members
- `PUT /api/v1/chat/groups/:id/members/:userId/role` - Update member role
- `DELETE /api/v1/chat/groups/:id/members/:userId` - Remove member

#### Group Invites
- `POST /api/v1/chat/groups/:id/invites` - Create invite
- `GET /api/v1/chat/groups/:id/invites` - List group invites
- `POST /api/v1/chat/invites/:id/accept` - Accept invite
- `POST /api/v1/chat/invites/:id/reject` - Reject invite

#### Messages
- `POST /api/v1/chat/messages` - Send message
- `GET /api/v1/chat/messages/:id` - Get message details
- `GET /api/v1/chat/messages/conversation/:id` - List conversation messages
- `GET /api/v1/chat/messages/group/:id` - List group messages
- `PUT /api/v1/chat/messages/:id/status` - Update message status
- `DELETE /api/v1/chat/messages/:id` - Delete message

#### Message Reactions
- `POST /api/v1/chat/messages/:id/reactions` - Add reaction
- `GET /api/v1/chat/messages/:id/reactions` - Get message reactions
- `DELETE /api/v1/chat/messages/:id/reactions` - Remove reaction

#### Forum Chat
- `GET /api/v1/forums/:id/messages` - List forum messages (24-hour auto-clear)
- `POST /api/v1/forums/:id/messages` - Send forum message
- `DELETE /api/v1/forums/:id/messages/:msg_id` - Delete forum message

#### WebSocket Connection
- `WS /api/v1/ws` - WebSocket endpoint for real-time chat
  - Query params: `?token=<JWT_TOKEN>`

**Example WebSocket connection:**
```bash
# Using wscat
wscat -c "ws://localhost:8080/api/v1/ws?token=YOUR_JWT_TOKEN"

# Using JavaScript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?token=' + YOUR_JWT_TOKEN);
```

**WebSocket message types:**
- `send_message` - Send a message
- `mark_read` - Mark messages as read
- `typing` - Send typing indicator
- `ping` / `pong` - Keepalive

See [CHAT_FEATURES.md](CHAT_FEATURES.md) for detailed user documentation and [CHAT_DEVELOPER_GUIDE.md](CHAT_DEVELOPER_GUIDE.md) for technical details.

## Testing

Run unit tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -cover ./...
```

Generate coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Run tests with race detector:

```bash
go test -race ./...
```

## Project Structure

```
tj-routes/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/                  # Configuration management
│   ├── models/                  # Database models
│   ├── repository/              # Data access layer
│   ├── service/                 # Business logic
│   │   └── mocks/               # Mock implementations for testing
│   ├── handler/                 # HTTP handlers
│   ├── middleware/              # Auth, RBAC middleware
│   └── utils/                   # Utilities (JWT, password, OAuth, database)
├── docs/                        # API documentation (OpenAPI spec)
└── README.md                    # This file
```

## Environment Variables

The following environment variables are required (with default values shown):

### Database Configuration

- `DB_HOST` - Database host (default: `localhost`)
- `DB_PORT` - Database port (default: `5432`)
- `DB_USER` - Database user (default: `postgres`)
- `DB_PASSWORD` - Database password (default: `postgres`)
- `DB_NAME` - Database name (default: `tj_routes`)
- `DB_SSLMODE` - SSL mode for database connection (default: `disable`)

### Server Configuration

- `SERVER_HOST` - Server host (default: `localhost`)
- `SERVER_PORT` - Server port (default: `8080`)
- `ENVIRONMENT` - Environment mode: `development` or `production` (default: `development`)

### JWT Configuration

- `JWT_SECRET` - Secret key for JWT token signing (default: `your-super-secret-jwt-key-change-in-production`)
- `JWT_EXPIRATION_HOURS` - JWT token expiration time in hours (default: `24`)
- `JWT_REFRESH_EXPIRATION_HOURS` - Refresh token expiration time in hours (default: `168`)

### OAuth Configuration

- `GOOGLE_CLIENT_ID` - Google OAuth client ID (required for OAuth)
- `GOOGLE_CLIENT_SECRET` - Google OAuth client secret (required for OAuth)
- `GOOGLE_REDIRECT_URL` - OAuth redirect URL (default: `http://localhost:8080/api/v1/auth/oauth/google/callback`)

### Logging Configuration

- `LOG_LEVEL` - Logging level (default: `info`)

**Important**: In production, always set a strong `JWT_SECRET` and use secure database credentials.

## Building and Deployment

### Build the application:

```bash
# Windows
go build -o api.exe cmd/api/main.go

# Linux/Mac
go build -o api cmd/api/main.go
```

### Run the compiled binary:

```bash
# Windows
./api.exe

# Linux/Mac
./api
```

### Docker Deployment

The project includes Docker support for easy deployment:

```bash
# Development
docker-compose up -d

# Production
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

See [DEPLOYMENT.md](DEPLOYMENT.md) for detailed deployment instructions to Google Cloud or other platforms.

### Production Readiness

⚠️ **Important**: Before deploying to production, review [PRODUCTION_READINESS.md](PRODUCTION_READINESS.md) for critical security and configuration issues that need to be addressed.

Key items to address:

- CORS configuration (restrict to specific domains)
- Error message sanitization
- Database migrations (replace AutoMigrate)
- Graceful shutdown
- Structured logging
- Environment variable validation

## Development

### Code Structure

- **Models**: Database models using GORM
- **Repository**: Data access layer with repository pattern
- **Service**: Business logic layer
- **Handler**: HTTP request handlers (Gin)
- **Middleware**: Authentication and authorization middleware
- **Utils**: Helper utilities for JWT, password hashing, OAuth, and database

### Database Migrations

The application uses GORM's AutoMigrate feature, which automatically creates/updates database tables on startup based on the models defined in `internal/models/`.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details (if applicable).

## Contributing

Contributions are welcome! Please follow these guidelines:

1. **Fork the repository** and create your feature branch (`git checkout -b feature/amazing-feature`)
2. **Follow Go best practices** and maintain code style consistency
3. **Write tests** for new features and ensure all tests pass (`go test ./...`)
4. **Update documentation** if you're adding new features or changing existing behavior
5. **Commit your changes** with clear, descriptive commit messages
6. **Push to your branch** (`git push origin feature/amazing-feature`)
7. **Open a Pull Request** with a detailed description of your changes

### Code Style

- Follow Go standard formatting (`gofmt` or `goimports`)
- Write meaningful variable and function names
- Add comments for exported functions and types
- Keep functions focused and maintainable

### Testing

- Write unit tests for service layer logic
- Use mocks for repository dependencies (see `internal/service/mocks/`)
- Aim for good test coverage
- Run tests before submitting PRs

## Support

For issues, questions, or contributions, please open an issue on the repository.

## Acknowledgments

- Built with [Gin](https://gin-gonic.com/) web framework
- Database management with [GORM](https://gorm.io/)
- Authentication with [JWT](https://jwt.io/) and OAuth2
