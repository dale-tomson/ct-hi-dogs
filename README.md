# Dogs API

A full-stack REST API for managing dog breeds and sub-breeds, built with Go and React.

## Purpose

Dogs API provides a simple yet complete example of a modern full-stack web application. It demonstrates best practices for:
- REST API design with proper error handling
- SQLite database management with transactions
- Request authentication and rate limiting
- Frontend-backend integration
- Docker containerization and cloud deployment

## Tech Stack

### Backend
- **Language**: Go 1.26
- **Router**: [chi](https://github.com/go-chi/chi) — lightweight HTTP router
- **Database**: SQLite with [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite)
- **Middleware**: 
  - CORS support
  - HTTP rate limiting (100 req/min per IP)
  - API key authentication
  - Request logging and recovery
  - gzip compression

### Frontend
- **Framework**: React 19
- **Build Tool**: Vite
- **Package Manager**: pnpm

### Deployment
- **Container**: Docker (multi-stage build)
- **Hosting**: [Render](https://render.com)

## Quick Start

### Prerequisites
- Go 1.26+
- Node.js 22+ with pnpm
- SQLite (included in most systems)

### Development Setup

**Build the project:**
```bash
# Install frontend dependencies and build
cd frontend
pnpm install --frozen-lockfile
pnpm run build
cd ..

# Build Go binary
go build -o dogs-api .
```

**Run the API server:**
```bash
DB_PATH=./dogs.db API_KEY=dev-secret go run main.go
```

**Run the frontend dev server (hot reload):**
```bash
cd frontend
pnpm run dev
```

**Run tests:**
```bash
go test ./...
```

### Environment Variables

Create `.env.local` (see `.env.example`):
```
PORT=8080
DB_PATH=./dogs.db
API_KEY=your-secure-key
VITE_API_KEY=your-secure-key
```

## API Documentation

Base URL: `http://localhost:8080/api`

All `/api/dogs/*` endpoints require the `X-API-Key` header.

### Health Check
```
GET /api/health
```
Returns server and database status.

**Response (200 OK):**
```json
{
  "status": "ok",
  "db": "ok"
}
```

---

### List All Breeds
```
GET /api/dogs
X-API-Key: <api-key>
```

**Response (200 OK):**
```json
[
  {
    "id": 1,
    "breed": "labrador",
    "sub_breeds": ["black", "yellow", "chocolate"]
  },
  {
    "id": 2,
    "breed": "poodle",
    "sub_breeds": ["standard", "miniature", "toy"]
  }
]
```

---

### Get Single Breed
```
GET /api/dogs/{breed}
X-API-Key: <api-key>
```

**Example:**
```
GET /api/dogs/labrador
```

**Response (200 OK):**
```json
{
  "id": 1,
  "breed": "labrador",
  "sub_breeds": ["black", "yellow", "chocolate"]
}
```

**Response (404 Not Found):**
```json
{
  "error": "breed not found"
}
```

---

### Create Breed
```
POST /api/dogs
X-API-Key: <api-key>
Content-Type: application/json
```

**Request body:**
```json
{
  "breed": "corgi",
  "sub_breeds": ["pembroke", "cardigan"]
}
```

**Response (201 Created):**
```json
{
  "id": 3,
  "breed": "corgi",
  "sub_breeds": ["pembroke", "cardigan"]
}
```

**Validation rules:**
- `breed`: required, lowercase letters only (a-z)
- `sub_breeds`: optional, each lowercase letters only (a-z)

**Error responses:**
- `400 Bad Request` — invalid JSON or validation failed
- `409 Conflict` — breed already exists

---

### Update Breed Sub-breeds
```
PUT /api/dogs/{breed}
X-API-Key: <api-key>
Content-Type: application/json
```

**Request body:**
```json
{
  "sub_breeds": ["pembroke", "cardigan"]
}
```

**Response (200 OK):**
```json
{
  "id": 3,
  "breed": "corgi",
  "sub_breeds": ["pembroke", "cardigan"]
}
```

---

### Delete Breed
```
DELETE /api/dogs/{breed}
X-API-Key: <api-key>
```

**Response (204 No Content)** — no body

**Error response (404 Not Found):**
```json
{
  "error": "breed not found"
}
```

---

## Rate Limiting

API endpoints are rate-limited to **100 requests per minute per IP address**.

**Rate limit headers** in responses:
```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1234567890
```

---

## Project Structure

```
.
├── main.go              # Server setup, routing, static file serving
├── handlers/            # HTTP request handlers
│   ├── dogs.go         # CRUD operations for breeds
│   └── dogs_test.go    # Unit tests
├── models/             # Database queries and business logic
│   └── dog.go
├── db/                 # Database initialization and seeding
│   └── db.go
├── middleware/         # Authentication and request processing
│   └── auth.go
├── frontend/           # React + Vite application
│   ├── src/
│   ├── package.json
│   └── vite.config.ts
├── docker-compose.yml  # Local Docker setup
├── Dockerfile          # Multi-stage production build
└── render.yaml         # Render.com deployment config
```

## Deployment

### Docker

Build and run locally:
```bash
docker-compose up
```

### Render

Push to GitHub and connect the repository to Render. The `render.yaml` file defines build and deployment steps automatically.

---

## Testing

Run all tests:
```bash
go test ./...
```

Run tests with verbose output:
```bash
go test ./... -v
```

---

## License

MIT
