# TrailMemo API

A REST API for TrailMemo - a voice memo application for parks and recreation departments to record field observations with GPS location tracking.

## Features

- 🔐 Firebase Authentication integration
- 🎙️ Voice memo management with audio file upload
- 📍 GPS location tracking and nearby memo search
- 🔍 Full-text search across memo content
- 🗺️ Collaborative map view (all users' memos)
- 🔒 Secure - users can only edit/delete their own memos
- ☁️ Firebase Storage for audio files
- 🐘 PostgreSQL database

## Tech Stack

- **Language**: Go 1.21+
- **Framework**: Gin
- **Database**: PostgreSQL
- **Authentication**: Firebase Auth
- **Storage**: Firebase Cloud Storage
- **Deployment**: Railway.app (recommended)

## Quick Start

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Firebase project with Auth and Storage enabled

### Installation

1. **Clone the repository**

```bash
git clone https://github.com/tom-fitz/trailmemo-api.git
cd trailmemo-api
```

2. **Install dependencies**

```bash
go mod download
```

3. **Set up environment variables**

Copy `.env.example` to `.env` and fill in your values:

```bash
cp .env.example .env
```

Required environment variables:
- `DATABASE_URL` - PostgreSQL connection string
- `FIREBASE_PROJECT_ID` - Your Firebase project ID
- `FIREBASE_STORAGE_BUCKET` - Your Firebase storage bucket name
- `FIREBASE_SERVICE_ACCOUNT_PATH` - Path to Firebase service account JSON file
- `JWT_SECRET` - Random secret key for JWT signing

4. **Run database migrations**

```bash
psql $DATABASE_URL -f migrations/001_init.sql
```

5. **Run the server**

```bash
go run cmd/server/main.go
```

The API will start on `http://localhost:8080`

### Test the API

```bash
# Health check
curl http://localhost:8080/health
```

## API Documentation

See the full API specification in [documentation/API_Specification.md](documentation/API_Specification.md)

### Base URL

```
Development: http://localhost:8080/api/v1
Production: https://your-app.railway.app/api/v1
```

### Authentication

All authenticated endpoints require a Firebase ID token:

```
Authorization: Bearer <firebase_id_token>
```

### Main Endpoints

#### Health Check
```
GET /health
```

#### Authentication
```
POST /api/v1/auth/register    - Create user account
GET  /api/v1/auth/me          - Get current user info
```

#### Memos
```
POST   /api/v1/memos           - Create memo (multipart upload)
GET    /api/v1/memos           - List all memos (paginated)
GET    /api/v1/memos/:id       - Get specific memo
PUT    /api/v1/memos/:id       - Update memo (owner only)
DELETE /api/v1/memos/:id       - Delete memo (owner only)
GET    /api/v1/memos/nearby    - Find memos near location
GET    /api/v1/memos/search    - Full-text search
```

## Deployment

### Railway.app (Recommended)

1. **Connect your GitHub repository to Railway**

2. **Add PostgreSQL database**
   - Click "+ New" → Database → PostgreSQL
   - Railway provides `DATABASE_URL` automatically

3. **Set environment variables**
   - Add all variables from `.env.example`
   - For `FIREBASE_SERVICE_ACCOUNT_JSON`, paste the entire JSON content

4. **Deploy**
   - Railway auto-deploys on git push to main branch

### Manual Deployment

Build the binary:

```bash
go build -o bin/server cmd/server/main.go
```

Run:

```bash
./bin/server
```

## Project Structure

```
trailmemo-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── handlers/                # HTTP request handlers
│   │   ├── auth.go              # Authentication endpoints
│   │   ├── memos.go             # Memo CRUD endpoints
│   │   └── health.go            # Health check
│   ├── middleware/              # HTTP middleware
│   │   ├── auth.go              # Firebase token verification
│   │   └── cors.go              # CORS configuration
│   ├── models/                  # Data models
│   │   ├── user.go
│   │   └── memo.go
│   ├── repository/              # Database operations
│   │   ├── user_repo.go
│   │   └── memo_repo.go
│   ├── services/                # External services
│   │   └── firebase.go          # Firebase Auth & Storage
│   └── database/
│       └── postgres.go          # Database connection
├── config/
│   └── config.go                # Configuration management
├── migrations/
│   └── 001_init.sql             # Database schema
├── documentation/               # Detailed documentation
├── go.mod
├── go.sum
├── .env.example
├── .gitignore
└── README.md
```

## Development

### Running Tests

```bash
go test ./...
```

### Running with Hot Reload

Install Air:

```bash
go install github.com/cosmtrek/air@latest
```

Run with hot reload:

```bash
air
```

### Database Migrations

To create a new migration:

```bash
# Create file: migrations/00X_description.sql
# Then run:
psql $DATABASE_URL -f migrations/00X_description.sql
```

## Configuration

### Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | Server port | No | `8080` |
| `ENV` | Environment (development/production) | No | `development` |
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `FIREBASE_PROJECT_ID` | Firebase project ID | Yes | - |
| `FIREBASE_STORAGE_BUCKET` | Firebase storage bucket | Yes | - |
| `FIREBASE_SERVICE_ACCOUNT_PATH` | Path to service account JSON | Yes* | - |
| `FIREBASE_SERVICE_ACCOUNT_JSON` | Service account JSON content | Yes* | - |
| `JWT_SECRET` | Secret for JWT signing | Yes | - |
| `MAX_UPLOAD_SIZE` | Max audio file size in bytes | No | `52428800` (50MB) |

*Either `FIREBASE_SERVICE_ACCOUNT_PATH` or `FIREBASE_SERVICE_ACCOUNT_JSON` is required

### Firebase Setup

1. Create a Firebase project at https://console.firebase.google.com/
2. Enable Authentication (Email/Password)
3. Enable Cloud Storage
4. Download service account key:
   - Project Settings → Service Accounts → Generate New Private Key
5. Set up Firebase in your iOS/Android app

## Security

- All sensitive endpoints require Firebase authentication
- Users can only modify their own memos
- File uploads are validated and size-limited
- CORS is configured (update for production)
- Environment variables for secrets

## Performance

- Database indexes on commonly queried fields
- Connection pooling for PostgreSQL
- Efficient pagination for large datasets
- Full-text search using PostgreSQL's built-in capabilities

## Troubleshooting

### Database Connection Issues

```bash
# Test database connection
psql $DATABASE_URL -c "SELECT version();"
```

### Firebase Authentication Issues

- Verify service account JSON is valid
- Check Firebase project ID matches
- Ensure token hasn't expired (1 hour validity)

### File Upload Issues

- Check Firebase Storage rules allow authenticated uploads
- Verify file size is within limits
- Ensure correct Content-Type headers

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Support

For issues and questions:
- Open an issue on GitHub
- Check the [API Specification](documentation/API_Specification.md)
- Review the [Implementation Guide](documentation/MVP_Implementation_Guide.md)

## Roadmap

See [MVP_Implementation_Guide.md](documentation/MVP_Implementation_Guide.md) for the full implementation plan.

### Current Status: MVP Complete ✅

- ✅ User authentication with Firebase
- ✅ Memo CRUD operations
- ✅ Audio file upload to Firebase Storage
- ✅ GPS location tracking
- ✅ Full-text search
- ✅ Nearby memo search
- ✅ Pagination and filtering

### Future Enhancements

- [ ] Rate limiting
- [ ] Caching layer (Redis)
- [ ] Webhooks for real-time updates
- [ ] Photo attachments
- [ ] Team/department grouping
- [ ] Export to PDF/CSV
- [ ] Analytics dashboard

---

Built with ❤️ for parks and recreation workers
