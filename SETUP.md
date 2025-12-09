# TrailMemo API - Complete Setup Guide

This guide will walk you through setting up the TrailMemo API from scratch.

## Prerequisites

Before you begin, ensure you have:

- ✅ Go 1.21 or higher installed
- ✅ PostgreSQL database (local or Railway)
- ✅ Firebase account (free tier is fine)
- ✅ Git installed

## Step 1: Firebase Setup (15 minutes)

### 1.1 Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Add project"
3. Enter project name: `trailmemo-prod`
4. Disable Google Analytics (optional for MVP)
5. Click "Create project"

### 1.2 Enable Authentication

1. In Firebase Console, go to **Authentication**
2. Click "Get started"
3. Enable **Email/Password** sign-in method
4. Save

### 1.3 Set Up Cloud Storage

1. Go to **Storage** in Firebase Console
2. Click "Get started"
3. Start in **production mode**
4. Choose location: `us-central1` (or closest to you)
5. Note your bucket name: `trailmemo-prod.appspot.com`

### 1.4 Get Service Account Key

1. Go to **Project Settings** (gear icon) → **Service accounts**
2. Click "Generate new private key"
3. Click "Generate key"
4. Save the JSON file as `serviceAccountKey.json` in your project root
5. **IMPORTANT**: Add this file to `.gitignore` (already done)

### 1.5 Note Your Project Details

You'll need:
- Project ID: (e.g., `trailmemo-prod`)
- Storage Bucket: (e.g., `trailmemo-prod.appspot.com`)
- Service Account JSON file path

## Step 2: Database Setup

### Option A: Railway (Recommended for Production)

1. Go to [Railway.app](https://railway.app/)
2. Sign up with GitHub
3. Click "New Project"
4. Select "Provision PostgreSQL"
5. Copy the `DATABASE_URL` from the Variables tab

### Option B: Local PostgreSQL

1. Install PostgreSQL:
   ```bash
   # macOS
   brew install postgresql@15
   brew services start postgresql@15

   # Ubuntu
   sudo apt-get install postgresql postgresql-contrib
   sudo systemctl start postgresql
   ```

2. Create database:
   ```bash
   createdb trailmemo
   ```

3. Your DATABASE_URL will be:
   ```
   postgresql://localhost:5432/trailmemo?sslmode=disable
   ```

## Step 3: Project Setup

### 3.1 Clone and Install Dependencies

```bash
# Clone the repository
cd trail-memo-api

# Install Go dependencies
go mod download
go mod tidy
```

### 3.2 Configure Environment Variables

Create a `.env` file in the project root:

```bash
# Copy the example
cp .env.example .env

# Edit .env with your values
nano .env  # or use your preferred editor
```

Fill in your `.env` file:

```env
# Server
PORT=8080
ENV=development

# Database (use your Railway or local PostgreSQL URL)
DATABASE_URL=postgresql://user:password@host:5432/trailmemo

# Firebase
FIREBASE_PROJECT_ID=trailmemo-prod
FIREBASE_STORAGE_BUCKET=trailmemo-prod.appspot.com
FIREBASE_SERVICE_ACCOUNT_PATH=./serviceAccountKey.json

# Security (generate a random secret)
JWT_SECRET=your_random_secret_key_here

# Upload limits
MAX_UPLOAD_SIZE=52428800
```

### 3.3 Generate JWT Secret

```bash
# Generate a secure random secret
openssl rand -base64 32
```

Copy the output and paste it as your `JWT_SECRET` in `.env`

## Step 4: Run Database Migrations

```bash
# Make sure your DATABASE_URL is set
export DATABASE_URL="your_database_url_here"

# Run migrations
make migrate

# Or manually:
psql $DATABASE_URL -f migrations/001_init.sql
```

You should see output confirming tables were created.

## Step 5: Test the API

### 5.1 Start the Server

```bash
# Using Make
make run

# Or directly
go run cmd/server/main.go
```

You should see:
```
🚀 TrailMemo API server starting on port 8080
📍 Environment: development
🗄️  Database connected
🔥 Firebase initialized
```

### 5.2 Test Health Endpoint

In a new terminal:

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "trailmemo-api",
  "version": "1.0.0"
}
```

✅ If you see this, your API is running!

## Step 6: Create a Test User

### 6.1 Create User in Firebase Console

1. Go to Firebase Console → Authentication → Users
2. Click "Add user"
3. Email: `test@example.com`
4. Password: `Test123!`
5. Click "Add user"

### 6.2 Get Firebase ID Token

You'll need to get an ID token from Firebase. The easiest way is:

**Option A: Use the iOS app** (when it's ready)

**Option B: Use Firebase REST API**

```bash
# Get your Firebase Web API Key from Project Settings
WEB_API_KEY="your_web_api_key"

# Sign in and get ID token
curl -X POST "https://identitytoolkit.googleapis.com/v1/accounts:signInWithPassword?key=$WEB_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "Test123!",
    "returnSecureToken": true
  }'
```

Copy the `idToken` from the response.

### 6.3 Register User in API

```bash
# Set your token
TOKEN="your_firebase_id_token_here"

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "display_name": "Test User",
    "department": "Parks & Recreation"
  }'
```

Expected response:
```json
{
  "user_id": "firebase_uid",
  "email": "test@example.com",
  "display_name": "Test User",
  "department": "Parks & Recreation",
  "created_at": "2024-12-09T..."
}
```

### 6.4 Test Getting User Info

```bash
curl http://localhost:8080/api/v1/auth/me \
  -H "Authorization: Bearer $TOKEN"
```

## Step 7: Test Memo Creation

### 7.1 Create a Test Audio File

```bash
# Create a small test audio file (optional)
# Or use any .m4a, .mp3, .wav file you have
```

### 7.2 Create a Memo

```bash
curl -X POST http://localhost:8080/api/v1/memos \
  -H "Authorization: Bearer $TOKEN" \
  -F "audio=@path/to/your/audio.m4a" \
  -F "text=This is a test memo about trail conditions" \
  -F "duration_seconds=30" \
  -F "latitude=45.6789" \
  -F "longitude=-111.0123" \
  -F "location_accuracy=10.5" \
  -F "park_name=Test Park"
```

### 7.3 List Memos

```bash
curl http://localhost:8080/api/v1/memos \
  -H "Authorization: Bearer $TOKEN"
```

## Step 8: Deploy to Railway (Production)

### 8.1 Connect Repository

1. Go to Railway dashboard
2. Click "New Project"
3. Select "Deploy from GitHub repo"
4. Choose `trailmemo-api` repository
5. Railway will auto-detect Go and start building

### 8.2 Add PostgreSQL

1. Click "+ New" → Database → PostgreSQL
2. Railway automatically provisions and connects it
3. The `DATABASE_URL` is automatically set

### 8.3 Set Environment Variables

In Railway project settings → Variables tab, add:

```
ENV=production
FIREBASE_PROJECT_ID=trailmemo-prod
FIREBASE_STORAGE_BUCKET=trailmemo-prod.appspot.com
FIREBASE_SERVICE_ACCOUNT_JSON=<paste entire JSON content>
JWT_SECRET=<your_secret>
MAX_UPLOAD_SIZE=52428800
```

**For FIREBASE_SERVICE_ACCOUNT_JSON:**
- Open your `serviceAccountKey.json`
- Copy the ENTIRE JSON content
- Paste it as the value (Railway handles this correctly)

### 8.4 Run Migrations on Railway

```bash
# Get Railway PostgreSQL URL from dashboard
DATABASE_URL="railway_postgres_url"

# Run migrations
psql $DATABASE_URL -f migrations/001_init.sql
```

### 8.5 Deploy

```bash
git add .
git commit -m "Initial deployment"
git push origin main
```

Railway automatically deploys on push!

### 8.6 Get Your Production URL

Railway provides a URL like: `https://your-app.railway.app`

Test it:
```bash
curl https://your-app.railway.app/health
```

## Troubleshooting

### Issue: "Failed to connect to database"

**Solution:**
- Check your `DATABASE_URL` is correct
- Ensure PostgreSQL is running (if local)
- Test connection: `psql $DATABASE_URL -c "SELECT 1;"`

### Issue: "Error initializing Firebase"

**Solution:**
- Verify `serviceAccountKey.json` exists and is valid JSON
- Check `FIREBASE_PROJECT_ID` matches your Firebase project
- Ensure file path in `.env` is correct

### Issue: "Invalid or expired token"

**Solution:**
- Firebase ID tokens expire after 1 hour
- Generate a new token using the sign-in endpoint
- Verify you're using the correct Web API Key

### Issue: "Error uploading audio file"

**Solution:**
- Check Firebase Storage is enabled
- Verify storage bucket name is correct
- Check file size is under 50MB
- Ensure file format is supported (mp3, m4a, wav, aac)

### Issue: Railway deployment fails

**Solution:**
- Ensure `go.mod` and `go.sum` are committed
- Check Railway build logs for specific errors
- Verify all environment variables are set
- Make sure migrations have been run

## Next Steps

1. ✅ API is running locally
2. ✅ Test user created
3. ✅ Database connected
4. ✅ Firebase integrated

Now you can:
- Start building the iOS app
- Add more test data
- Configure production settings
- Set up monitoring and logging

## Useful Commands

```bash
# Development
make run              # Run server
make build            # Build binary
make test             # Run tests
make migrate          # Run migrations

# With hot reload
make dev              # Requires: go install github.com/cosmtrek/air@latest

# Docker
make docker-build     # Build Docker image
make docker-run       # Run in Docker

# Database
psql $DATABASE_URL    # Connect to database
```

## Support

- 📖 [API Specification](documentation/API_Specification.md)
- 📋 [Implementation Guide](documentation/MVP_Implementation_Guide.md)
- 🏗️ [Architecture Plan](documentation/TrailMemo_Architecture_Plan.md)

## Security Checklist

Before going to production:

- [ ] Change all default secrets
- [ ] Use strong JWT secret (32+ characters)
- [ ] Set up Firebase Security Rules
- [ ] Configure CORS for your domain only
- [ ] Enable HTTPS (Railway does this automatically)
- [ ] Set up monitoring and logging
- [ ] Review Firebase Storage security rules
- [ ] Enable rate limiting (future enhancement)

---

🎉 **Congratulations!** Your TrailMemo API is now set up and running!

