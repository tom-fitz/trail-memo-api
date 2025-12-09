# TrailMemo: Quick Start Setup Guide

This guide will walk you through setting up TrailMemo from scratch in the recommended order.

---

## Prerequisites

Before starting, make sure you have:

- [ ] Mac with Xcode installed (for iOS development)
- [ ] Go 1.21+ installed
- [ ] GitHub account
- [ ] Apple Developer account (free tier OK for development)
- [ ] Google account (for Firebase)

---

## Phase 1: Account Setup (30 minutes)

### 1. Create Firebase Project

1. Go to [Firebase Console](https://console.firebase.google.com/)
2. Click "Add project"
3. Name it `trailmemo-prod` (or similar)
4. Disable Google Analytics (not needed for MVP)
5. Click "Create project"

**Configure Authentication:**
1. In Firebase console, go to "Authentication" → "Get started"
2. Enable "Email/Password" sign-in method
3. (Optional) Enable "Google" sign-in for easier login

**Set up Cloud Storage:**
1. Go to "Storage" → "Get started"
2. Start in "production mode" (we'll add security rules later)
3. Choose a location (us-central1 recommended)
4. Note your bucket name (e.g., `trailmemo-prod.appspot.com`)

**Get iOS Configuration:**
1. Click "Add app" → iOS
2. Bundle ID: `com.yourname.trailmemo` (remember this!)
3. Download `GoogleService-Info.plist`
4. Save it for later

**Get API Service Account:**
1. Go to Project Settings → Service Accounts
2. Click "Generate new private key"
3. Save the JSON file as `serviceAccountKey.json`
4. **Keep this secure!** Never commit to Git

---

### 2. Set Up Railway.app

1. Go to [Railway.app](https://railway.app/)
2. Sign up with GitHub
3. Click "New Project"
4. Select "Deploy from GitHub repo"
5. Don't deploy yet - we'll set this up later

**Add PostgreSQL:**
1. In your Railway project, click "+ New"
2. Select "Database" → "PostgreSQL"
3. Railway automatically provisions it
4. Click on PostgreSQL → "Variables" to see connection string
5. Save the `DATABASE_URL` for later

---

## Phase 2: Backend Setup (2 hours)

### 1. Create Go API Project

```bash
# Create project directory
mkdir trailmemo-api
cd trailmemo-api

# Initialize Go module
go mod init github.com/tom-fitz/trailmemo-api

# Create directory structure
mkdir -p cmd/server
mkdir -p internal/{handlers,middleware,models,repository,services,database}
mkdir -p config
mkdir -p migrations
```

### 2. Install Dependencies

```bash
# Web framework
go get github.com/gin-gonic/gin

# Database
go get github.com/lib/pq
go get github.com/jmoiron/sqlx

# Firebase
go get firebase.google.com/go/v4
go get firebase.google.com/go/v4/auth
go get firebase.google.com/go/v4/storage

# Environment variables
go get github.com/joho/godotenv

# UUID generation
go get github.com/google/uuid

# CORS
go get github.com/gin-contrib/cors
```

### 3. Create Environment File

Create `.env` in project root:

```bash
# Database
DATABASE_URL=postgresql://user:password@host:5432/trailmemo

# Server
PORT=8080
ENV=development

# Firebase
FIREBASE_PROJECT_ID=trailmemo-prod
FIREBASE_STORAGE_BUCKET=trailmemo-prod.appspot.com
FIREBASE_SERVICE_ACCOUNT_PATH=./serviceAccountKey.json

# Security
JWT_SECRET=your_random_secret_key_here
```

**Generate a secure JWT secret:**
```bash
openssl rand -base64 32
```

### 4. Create Database Schema

Create `migrations/001_init.sql`:

```sql
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Users table
CREATE TABLE users (
    user_id VARCHAR(128) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    department VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Memos table
CREATE TABLE memos (
    memo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(128) REFERENCES users(user_id) ON DELETE CASCADE,
    user_name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    audio_url TEXT NOT NULL,
    text TEXT NOT NULL,
    duration_seconds INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location_accuracy FLOAT,
    address TEXT,
    park_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_memos_user_created ON memos(user_id, created_at DESC);
CREATE INDEX idx_memos_location ON memos(latitude, longitude);
CREATE INDEX idx_memos_park ON memos(park_name);
CREATE INDEX idx_memos_created ON memos(created_at DESC);

-- Full-text search on text
CREATE INDEX idx_memos_text_search ON memos USING gin(to_tsvector('english', text));

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Trigger to auto-update updated_at
CREATE TRIGGER update_memos_updated_at 
    BEFORE UPDATE ON memos 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();
```

**Run migration:**
```bash
# Install psql if not already installed
brew install postgresql

# Run migration
psql $DATABASE_URL -f migrations/001_init.sql
```

### 5. Create Basic API Structure

Create `cmd/server/main.go`:

```go
package main

import (
    "log"
    "os"
    
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }
    
    // Set up Gin router
    r := gin.Default()
    
    // Health check endpoint
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "status": "ok",
            "service": "trailmemo-api",
        })
    })
    
    // API v1 routes
    v1 := r.Group("/api/v1")
    {
        // Auth routes
        auth := v1.Group("/auth")
        {
            auth.POST("/register", registerHandler)
            auth.POST("/login", loginHandler)
            auth.GET("/me", authMiddleware, getMeHandler)
        }
        
        // Memo routes (protected)
        memos := v1.Group("/memos")
        memos.Use(authMiddleware)
        {
            memos.POST("", createMemoHandler)
            memos.GET("", listMemosHandler)
            memos.GET("/:id", getMemoHandler)
            memos.PUT("/:id", updateMemoHandler)
            memos.DELETE("/:id", deleteMemoHandler)
        }
    }
    
    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    
    log.Printf("Server starting on port %s", port)
    if err := r.Run(":" + port); err != nil {
        log.Fatal(err)
    }
}

// Placeholder handlers (implement these next)
func registerHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func loginHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func getMeHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func createMemoHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func listMemosHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func getMemoHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func updateMemoHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func deleteMemoHandler(c *gin.Context) { c.JSON(501, gin.H{"error": "not implemented"}) }
func authMiddleware(c *gin.Context) { c.Next() }
```

**Test the server:**
```bash
go run cmd/server/main.go

# In another terminal:
curl http://localhost:8080/health
```

### 6. Set Up Git Repository

```bash
# Create .gitignore
cat > .gitignore << EOF
.env
*.log
serviceAccountKey.json
tmp/
vendor/
EOF

# Initialize git
git init
git add .
git commit -m "Initial commit: API skeleton"

# Create GitHub repo and push
gh repo create trailmemo-api --public --source=. --remote=origin
git push -u origin main
```

### 7. Deploy to Railway

1. Go to Railway dashboard
2. Click "New Project" → "Deploy from GitHub repo"
3. Select `trailmemo-api` repository
4. Railway auto-detects Go and starts building

**Add Environment Variables in Railway:**
1. Click on your service
2. Go to "Variables" tab
3. Add all variables from your `.env` file
4. **Important:** Upload `serviceAccountKey.json` content to Railway
   - Copy entire JSON content
   - Create variable `FIREBASE_SERVICE_ACCOUNT_JSON`
   - Paste the JSON

**Update code to read service account from environment:**
```go
// In your Firebase initialization
var firebaseApp *firebase.App

if jsonStr := os.Getenv("FIREBASE_SERVICE_ACCOUNT_JSON"); jsonStr != "" {
    opt := option.WithCredentialsJSON([]byte(jsonStr))
    firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
} else {
    opt := option.WithCredentialsFile(os.Getenv("FIREBASE_SERVICE_ACCOUNT_PATH"))
    firebaseApp, err = firebase.NewApp(context.Background(), nil, opt)
}
```

Railway will auto-deploy on every git push!

---

## Phase 3: iOS App Setup (2 hours)

### 1. Create iOS Project in Xcode

1. Open Xcode
2. File → New → Project
3. Select "iOS" → "App"
4. Product Name: `TrailMemo`
5. Bundle Identifier: `com.yourname.trailmemo` (must match Firebase!)
6. Interface: SwiftUI
7. Language: Swift
8. Create

### 2. Add Firebase to iOS

1. Drag `GoogleService-Info.plist` into Xcode project
   - Make sure "Copy items if needed" is checked
   - Add to TrailMemo target

2. Install Firebase via Swift Package Manager:
   - File → Add Packages
   - URL: `https://github.com/firebase/firebase-ios-sdk`
   - Select version 10.0.0+
   - Add packages:
     - FirebaseAuth
     - FirebaseStorage

3. Initialize Firebase in `TrailMemoApp.swift`:

```swift
import SwiftUI
import FirebaseCore

@main
struct TrailMemoApp: App {
    init() {
        FirebaseApp.configure()
    }
    
    var body: some Scene {
        WindowGroup {
            ContentView()
        }
    }
}
```

### 3. Set Up Project Structure

Create these folders in Xcode:
- Models
- Views
  - Auth
  - Memos
  - Components
- ViewModels
- Services
- Utilities

### 4. Add Required Permissions

Update `Info.plist`:

```xml
<key>NSMicrophoneUsageDescription</key>
<string>TrailMemo needs microphone access to record voice memos</string>

<key>NSSpeechRecognitionUsageDescription</key>
<string>TrailMemo needs speech recognition to transcribe your voice memos</string>

<key>NSLocationWhenInUseUsageDescription</key>
<string>TrailMemo needs your location to tag memos with GPS coordinates</string>

<key>NSLocationAlwaysAndWhenInUseUsageDescription</key>
<string>TrailMemo needs your location to tag memos with GPS coordinates</string>
```

### 5. Create Basic Models

Create `Models/Memo.swift`:

```swift
import Foundation
import CoreLocation

struct Memo: Identifiable, Codable {
    let id: UUID
    let userId: String
    let userName: String  // Display name of creator
    var title: String?
    let audioURL: URL
    let text: String  // Transcribed text from iOS Speech
    let durationSeconds: Int
    let location: Location?
    let parkName: String?
    let createdAt: Date
    let updatedAt: Date
    
    struct Location: Codable {
        let latitude: Double
        let longitude: Double
        let accuracy: Double
        var address: String?
        
        var coordinate: CLLocationCoordinate2D {
            CLLocationCoordinate2D(latitude: latitude, longitude: longitude)
        }
    }
}
```

### 6. Create Basic UI

Create `Views/Auth/LoginView.swift`:

```swift
import SwiftUI
import FirebaseAuth

struct LoginView: View {
    @State private var email = ""
    @State private var password = ""
    @State private var errorMessage: String?
    @State private var isLoading = false
    
    var body: some View {
        VStack(spacing: 20) {
            Text("TrailMemo")
                .font(.largeTitle)
                .fontWeight(.bold)
            
            TextField("Email", text: $email)
                .textFieldStyle(.roundedBorder)
                .autocapitalization(.none)
                .keyboardType(.emailAddress)
            
            SecureField("Password", text: $password)
                .textFieldStyle(.roundedBorder)
            
            if let error = errorMessage {
                Text(error)
                    .foregroundColor(.red)
                    .font(.caption)
            }
            
            Button(action: login) {
                if isLoading {
                    ProgressView()
                } else {
                    Text("Sign In")
                        .frame(maxWidth: .infinity)
                }
            }
            .buttonStyle(.borderedProminent)
            .disabled(isLoading)
        }
        .padding()
    }
    
    private func login() {
        isLoading = true
        errorMessage = nil
        
        Auth.auth().signIn(withEmail: email, password: password) { result, error in
            isLoading = false
            if let error = error {
                errorMessage = error.localizedDescription
            }
        }
    }
}
```

### 7. Test the App

1. Press ⌘+R to run
2. You should see the login screen
3. Test creating an account in Firebase console to verify

---

## Phase 4: Connect iOS to API (1 hour)

### 1. Create API Client

Create `Services/APIClient.swift`:

```swift
import Foundation

class APIClient {
    static let shared = APIClient()
    private let baseURL = "https://your-app.railway.app/api/v1"
    
    private init() {}
    
    func request<T: Decodable>(
        endpoint: String,
        method: String = "GET",
        body: Data? = nil
    ) async throws -> T {
        guard let url = URL(string: baseURL + endpoint) else {
            throw URLError(.badURL)
        }
        
        var request = URLRequest(url: url)
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        
        // Add Firebase auth token
        if let token = try? await Auth.auth().currentUser?.getIDToken() {
            request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
        }
        
        if let body = body {
            request.httpBody = body
        }
        
        let (data, response) = try await URLSession.shared.data(for: request)
        
        guard let httpResponse = response as? HTTPURLResponse,
              (200...299).contains(httpResponse.statusCode) else {
            throw URLError(.badServerResponse)
        }
        
        let decoder = JSONDecoder()
        decoder.dateDecodingStrategy = .iso8601
        return try decoder.decode(T.self, from: data)
    }
}
```

### 2. Update API URL

In Xcode, add your Railway URL:
1. Select project → TrailMemo target
2. Build Settings → Search for "Info.plist"
3. Or create a `Config.swift`:

```swift
enum Config {
    static let apiBaseURL = "https://your-app.railway.app"
}
```

---

## Phase 5: Test Everything (30 minutes)

### Backend Health Check

```bash
curl https://your-app.railway.app/health
```

Expected response:
```json
{
  "status": "ok",
  "service": "trailmemo-api"
}
```

### Database Connection

Test PostgreSQL connection:
```bash
psql $DATABASE_URL -c "SELECT * FROM users LIMIT 1;"
```

### Firebase Auth

1. Go to Firebase Console → Authentication
2. Create a test user manually
3. Try logging in from iOS app

---

## Next Steps Checklist

- [ ] Implement auth handlers in Go API
- [ ] Implement memo CRUD handlers in Go API
- [ ] Add audio recording in iOS app
- [ ] Add iOS Speech recognition in iOS app
- [ ] Add location services in iOS app
- [ ] Implement file upload to Firebase Storage
- [ ] Create map view showing all memos in iOS
- [ ] Add memo detail view with audio playback
- [ ] Set up proper error handling
- [ ] Add logging and monitoring

---

## Common Issues & Solutions

### Issue: Firebase initialization fails in iOS
**Solution:** Make sure `GoogleService-Info.plist` is added to target and `FirebaseApp.configure()` is called before any Firebase usage.

### Issue: Railway deployment fails
**Solution:** Check that `go.mod` and `go.sum` are committed to git. Railway needs these to build.

### Issue: Can't connect to PostgreSQL from local machine
**Solution:** Use Railway's provided connection string. It includes necessary authentication.

### Issue: CORS errors when calling API from iOS
**Solution:** Add CORS middleware in Go:
```go
import "github.com/gin-contrib/cors"

r.Use(cors.New(cors.Config{
    AllowOrigins: []string{"*"},
    AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowHeaders: []string{"Authorization", "Content-Type"},
}))
```

### Issue: Audio files too large
**Solution:** Compress to AAC format at 64kbps in iOS before upload.

---

## Helpful Resources

- [Firebase iOS Setup](https://firebase.google.com/docs/ios/setup)
- [Gin Framework Docs](https://gin-gonic.com/docs/)
- [Railway Deployment](https://docs.railway.app/)
- [SwiftUI Tutorials](https://developer.apple.com/tutorials/swiftui)
- [Go Firebase Admin SDK](https://firebase.google.com/docs/admin/setup#go)

---

## Support

If you get stuck:
1. Check Railway logs for API errors
2. Check Xcode console for iOS errors
3. Verify Firebase configuration
4. Test API endpoints with curl or Postman
5. Check database with psql

