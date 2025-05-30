# PostgreSQL MCP Server

A Model Context Protocol (MCP) server implementation in Go that provides a flexible and secure way to manage and query data models through a RESTful API.

## Features

- 🔐 Secure model management with JWT authentication
- 📝 Dynamic model registration with JSON schema validation
- 🔍 Flexible query system for searching across models
- 🗄️ PostgreSQL integration with automatic migrations
- 🔒 Role-based access control (admin vs public endpoints)
- 🚀 RESTful API design

## Prerequisites

- Go 1.16 or higher
- PostgreSQL 12 or higher
- Make (optional, for using Makefile commands)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/Nysonn/postgres-server.git
cd postgres-server
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file in the project root:
```env
DATABASE_URL=postgresql://user:password@localhost:5432/your_database
JWT_SECRET=your_jwt_secret_key
SERVER_ADDRESS=:8080  # Optional, defaults to :8080
```

## Database Setup

1. Create a PostgreSQL database:
```sql
CREATE DATABASE your_database;
```

2. The server will automatically run migrations to create necessary tables.

## Running the Server

```bash
go run cmd/server/main.go
```

The server will:
- Load environment variables
- Connect to the database
- Run any pending migrations
- Start the HTTP server

## API Endpoints

### Admin Endpoints (JWT Required)

#### Register a Model
```http
POST /admin/models/register
Authorization: Bearer <your_jwt_token>
Content-Type: application/json

{
    "name": "items",
    "schema": {
        "type": "object",
        "properties": {
            "name": {"type": "string"},
            "category": {"type": "string"},
            "price_ugx": {"type": "number"},
            "available": {"type": "boolean"}
        }
    }
}
```

#### List All Models
```http
GET /admin/models/list
Authorization: Bearer <your_jwt_token>
```

#### Get Model Details
```http
GET /admin/models/get?name=<model_name>
Authorization: Bearer <your_jwt_token>
```

#### Delete Model
```http
DELETE /admin/models/delete?name=<model_name>
Authorization: Bearer <your_jwt_token>
```

### Public Endpoints

#### Query Models
```http
POST /query
Content-Type: application/json

{
    "model": "items",
    "queryText": "search term",
    "fields": ["name", "category"],
    "maxResults": 10
}
```

## Generating JWT Tokens

Use the included token generator:

```bash
go run cmd/token/main.go
```

This will output a JWT token that can be used for admin operations.

## Project Structure

```
.
├── cmd/
│   ├── server/        # Main server application
│   └── token/         # JWT token generator
├── internal/
│   ├── config/        # Configuration management
│   ├── db/           # Database connection and utilities
│   ├── handler/      # HTTP request handlers
│   └── middleware/   # HTTP middleware (auth, etc.)
├── migrations/       # Database migrations
├── .env             # Environment variables (not in git)
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```

## Development

### Adding New Migrations

1. Create new migration files in the `migrations` directory:
   - `XXXX_description.up.sql` for applying changes
   - `XXXX_description.down.sql` for reverting changes

2. The server will automatically run new migrations on startup.

### Testing

```bash
go test ./...
```

## Security Considerations

- JWT tokens should be kept secure
- Database credentials should be properly managed
- The server should be run behind a reverse proxy in production
- Regular security audits are recommended

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [Go](https://golang.org/)
- [PostgreSQL](https://www.postgresql.org/)
- [JWT-Go](https://github.com/golang-jwt/jwt)
- [Golang Migrate](https://github.com/golang-migrate/migrate) 