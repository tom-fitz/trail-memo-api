# TrailMemo - Voice Memo Application Architecture Plan

## App Name Suggestions

Based on your friend's use case (parks department worker taking field notes):

1. **TrailMemo** - Simple, clear, emphasizes the trail/walking aspect
2. **FieldNote** - Professional, highlights the note-taking in field work
3. **PathLogger** - Tech-savvy, emphasizes location tracking
4. **ParkPulse** - Catchy, suggests staying connected to park activities
5. **WalkNote** - Direct and simple
6. **GreenPin** - Eco-friendly feel, suggests pinning locations

**Recommendation: TrailMemo** - It's memorable, describes the use case perfectly, and has a professional yet friendly feel.

---

## System Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT TIER                              │
│                                                                   │
│  ┌──────────────┐     ┌──────────────┐     ┌──────────────┐    │
│  │   iOS App    │     │  Android App │     │   Web App    │    │
│  │   (Swift)    │     │   (Future)   │     │   (Future)   │    │
│  └──────┬───────┘     └──────┬───────┘     └──────┬───────┘    │
│         │                    │                     │             │
└─────────┼────────────────────┼─────────────────────┼─────────────┘
          │                    │                     │
          └────────────────────┼─────────────────────┘
                               │
                     HTTPS/REST API
                               │
┌─────────────────────────────┼─────────────────────────────────────┐
│                         API TIER                                   │
│                              │                                     │
│                    ┌─────────▼──────────┐                         │
│                    │   Go API Server    │                         │
│                    │   (Gin Framework)  │                         │
│                    └─────────┬──────────┘                         │
│                              │                                     │
└──────────────────────────────┼─────────────────────────────────────┘
                               │
          ┌────────────────────┼────────────────────┐
          │                    │                    │
┌─────────▼───────┐  ┌─────────▼────────┐ ┌────────▼─────────┐
│  Firebase Auth  │  │ Firebase Storage │ │   PostgreSQL     │
│  (User Auth)    │  │ (Audio Files)    │ │   (Railway)      │
└─────────────────┘  └──────────────────┘ └──────────────────┘
```

---

## Technology Stack

### iOS Client
- **Language**: Swift (SwiftUI for modern UI)
- **Audio Recording**: AVFoundation
- **Speech Recognition**: iOS Speech Framework (native, on-device)
- **Location Services**: CoreLocation
- **Network**: URLSession / Alamofire
- **Local Storage**: Core Data (for offline caching)

### Backend API
- **Language**: Go (Golang)
- **Framework**: Gin (lightweight, fast)
- **Authentication**: Firebase Auth SDK for Go
- **Deployment**: Railway.app

### Storage & Services
- **Authentication**: Firebase Authentication
- **File Storage**: Firebase Cloud Storage (for audio files)
- **Database**: PostgreSQL (on Railway.app)
- **Speech-to-Text**: iOS native Speech Framework (no external service needed!)

### Deployment
- **API Hosting**: Railway.app
- **Database**: PostgreSQL on Railway.app
- **Estimated Cost**: $0-5/month for MVP

---

## Data Model

### User
```json
{
  "user_id": "string (Firebase UID)",
  "email": "string",
  "display_name": "string",
  "created_at": "timestamp",
  "department": "string (e.g., 'Parks & Recreation')",
  "role": "string (optional)"
}
```

### Memo
```json
{
  "memo_id": "uuid",
  "user_id": "string (foreign key)",
  "user_name": "string (display name of creator)",
  "title": "string (optional, auto-generated or user-provided)",
  "audio_url": "string (Cloud Storage URL)",
  "text": "string (transcribed text from iOS Speech)",
  "duration_seconds": "integer",
  "location": {
    "latitude": "float",
    "longitude": "float",
    "accuracy": "float",
    "address": "string (optional, from reverse geocoding)"
  },
  "park_name": "string (optional)",
  "created_at": "timestamp",
  "updated_at": "timestamp"
}
```

### Database Schema (PostgreSQL)

```sql
-- Users table (minimal, as most auth is in Firebase)
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
    user_name VARCHAR(255), -- Denormalized for easy display
    title VARCHAR(255),
    audio_url TEXT NOT NULL,
    text TEXT NOT NULL, -- Transcribed text from iOS Speech
    duration_seconds INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location_accuracy FLOAT,
    address TEXT,
    park_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_user_created (user_id, created_at DESC),
    INDEX idx_location (latitude, longitude),
    INDEX idx_created_at (created_at DESC)
);

-- Full-text search index on text
CREATE INDEX idx_memos_text_search ON memos USING gin(to_tsvector('english', text));
```

---

## API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `GET /api/v1/auth/me` - Get current user info

### Memos
- `POST /api/v1/memos` - Create new memo (with audio + text)
- `GET /api/v1/memos` - List all memos (all users, for map view)
- `GET /api/v1/memos/:id` - Get specific memo
- `PUT /api/v1/memos/:id` - Update memo
- `DELETE /api/v1/memos/:id` - Delete memo
- `GET /api/v1/memos/nearby` - Get memos near a location

### File Upload
- `POST /api/v1/upload/audio` - Upload audio file
- `GET /api/v1/upload/presigned-url` - Get presigned URL for direct upload

---

## Application Flow

### Creating a New Memo (Detailed Flow)

```
┌──────────┐
│   User   │
└────┬─────┘
     │
     │ 1. Opens app, taps "New Memo"
     ▼
┌─────────────────┐
│   iOS App       │
│                 │
│  - Request      │
│    location     │
│  - Request mic  │
│    permission   │
│  - Start        │
│    recording    │
└────┬────────────┘
     │
     │ 2. User speaks into phone
     │
     ▼
┌─────────────────────────────┐
│  iOS Speech Recognition     │
│                             │
│  - Real-time transcription  │
│  - Display text as they     │
│    speak                    │
└────┬────────────────────────┘
     │
     │ 3. User taps "Done"
     ▼
┌─────────────────────────────────────┐
│  iOS App Processing                 │
│                                     │
│  - Stop recording                   │
│  - Finalize transcription           │
│  - Compress audio (AAC format)      │
│  - Prepare metadata                 │
│  - Show preview with text           │
└────┬────────────────────────────────┘
     │
     │ 4. User taps "Save"
     ▼
┌─────────────────────────────────────┐
│  Upload to API                      │
│                                     │
│  POST /api/v1/memos                 │
│  Multipart:                         │
│    - audio file                     │
│    - text (transcribed)             │
│    - duration_seconds               │
│    - latitude, longitude            │
│    - park_name                      │
└────┬────────────────────────────────┘
     │
     │ 5. API processes request
     ▼
┌─────────────────────────────────────┐
│  Go API Server                      │
│                                     │
│  - Verify Firebase auth token       │
│  - Upload audio to Cloud Storage    │
│  - Save text and metadata to DB     │
│  - Return memo_id and URL           │
└────┬────────────────────────────────┘
     │
     │ 6. Response to client
     ▼
┌─────────────────────────────────────┐
│  iOS App                            │
│                                     │
│  - Show success message             │
│  - Navigate to map view             │
│  - Display new memo on map          │
└─────────────────────────────────────┘
```

---

## Key Features to Implement

### Phase 1: MVP (Minimum Viable Product)
1. ✅ User authentication (Firebase)
2. ✅ Record voice memo with real-time transcription (iOS Speech)
3. ✅ Capture GPS location
4. ✅ Upload audio + text to cloud
5. ✅ Map view showing ALL users' memos
6. ✅ View memo details (text, audio, location, creator)
7. ✅ Delete own memos

### Phase 2: Enhanced Features
1. ⭐ Search memos by text content
2. ⭐ Filter by date, location, or park
3. ⭐ Edit memo text
4. ⭐ Offline mode (record and sync later)
5. ⭐ Add tags/categories
6. ⭐ User profiles with avatar
7. ⭐ Audio playback controls

### Phase 3: Advanced Features
1. 🚀 Team/department grouping
2. 🚀 Photo attachments
3. 🚀 Export to PDF/CSV
4. 🚀 Push notifications for nearby memos
5. 🚀 Comment on memos
6. 🚀 Analytics dashboard for department

---

## Security Considerations

### Authentication & Authorization
- Use Firebase Authentication for secure user management
- Verify Firebase ID tokens on every API request
- Implement role-based access control (RBAC) if needed
- Use HTTPS only (TLS 1.2+)

### Data Protection
- Encrypt audio files at rest (Cloud Storage does this automatically)
- Encrypt data in transit (HTTPS)
- Implement rate limiting on API endpoints
- Validate and sanitize all inputs
- Use prepared statements for database queries (prevent SQL injection)

### Privacy
- Don't store exact GPS coordinates publicly
- Allow users to delete their data (GDPR compliance)
- Clear privacy policy
- Optional: fuzzing location data for public sharing

---

## Cost Estimation (Free Tiers)

### Development Phase (Free)
- **Railway.app**: $5/month credit (free to start)
- **Firebase Auth**: 50K monthly active users free
- **Firebase Storage**: 5GB free, 1GB/day downloads
- **PostgreSQL (Railway)**: Included in $5 credit
- **iOS Speech Recognition**: Completely free (built into iOS)

### Expected Costs at Scale
- If your friend's department has ~10 users creating 50 memos/month:
  - Audio storage: ~500MB/month ≈ Free tier
  - Database: Well within free tier
  - API hosting: Free tier should cover
  - Speech recognition: $0 (iOS native)

**Total estimated cost**: $0-5/month for small team

### At Higher Scale (50 users, 200 memos/month)
- Railway: $5-10/month
- Firebase Storage: $0-2/month
- Database: Included

**Total estimated cost**: $5-12/month

---

## Development Roadmap

### Week 1-2: Setup & Infrastructure
- [ ] Set up Firebase project (Auth + Storage)
- [ ] Create Go API skeleton (Gin framework)
- [ ] Deploy to Railway with PostgreSQL
- [ ] Create iOS project structure
- [ ] Implement authentication flow

### Week 3-4: Core Functionality
- [ ] iOS: Audio recording with AVFoundation
- [ ] iOS: Real-time speech recognition with iOS Speech Framework
- [ ] iOS: Location capture with CoreLocation
- [ ] API: File upload endpoint
- [ ] API: Firebase Storage integration
- [ ] Database schema implementation
- [ ] API: CRUD operations for memos

### Week 5-6: Map View & UI
- [ ] iOS: Map view showing all memos
- [ ] iOS: Custom map annotations with user info
- [ ] iOS: Memo detail view
- [ ] iOS: Audio playback
- [ ] API: Get all memos endpoint
- [ ] API: Nearby memos endpoint

### Week 7-8: Polish & Testing
- [ ] Error handling and retry logic
- [ ] UI/UX improvements
- [ ] Permission handling (mic, location)
- [ ] Testing (unit + integration)
- [ ] Beta testing with parks department
- [ ] App Store submission preparation

---

## iOS Project Structure

```
TrailMemo/
├── App/
│   ├── TrailMemoApp.swift
│   └── AppDelegate.swift
├── Models/
│   ├── User.swift
│   ├── Memo.swift
│   └── Location.swift
├── Views/
│   ├── Auth/
│   │   ├── LoginView.swift
│   │   └── RegisterView.swift
│   ├── Memos/
│   │   ├── MapView.swift              ← Main view
│   │   ├── MemoAnnotationView.swift   ← Map pin
│   │   ├── MemoDetailView.swift
│   │   └── RecordMemoView.swift
│   └── Components/
│       ├── AudioPlayerView.swift
│       └── AudioWaveformView.swift
├── ViewModels/
│   ├── AuthViewModel.swift
│   ├── MapViewModel.swift
│   └── RecordViewModel.swift
├── Services/
│   ├── AudioService.swift
│   ├── SpeechRecognitionService.swift  ← iOS Speech
│   ├── LocationService.swift
│   ├── APIClient.swift
│   └── AuthService.swift
├── Utilities/
│   ├── Constants.swift
│   └── Extensions.swift
└── Resources/
    ├── Assets.xcassets
    └── Info.plist
```

---

## Go API Project Structure

```
trailmemo-api/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── handlers/
│   │   ├── auth.go
│   │   ├── memos.go
│   │   └── upload.go
│   ├── middleware/
│   │   ├── auth.go
│   │   └── logging.go
│   ├── models/
│   │   ├── user.go
│   │   └── memo.go
│   ├── repository/
│   │   ├── user_repo.go
│   │   └── memo_repo.go
│   ├── services/
│   │   ├── storage.go
│   │   ├── transcription.go
│   │   └── firebase.go
│   └── database/
│       └── postgres.go
├── config/
│   └── config.go
├── migrations/
│   └── 001_init.sql
├── go.mod
├── go.sum
├── Dockerfile
└── .env.example
```

---

## Environment Variables

### iOS App
```
FIREBASE_API_KEY=your_key
API_BASE_URL=https://your-api.railway.app
```

### Go API
```
# Database
DATABASE_URL=postgresql://user:password@host:5432/trailmemo

# Firebase
FIREBASE_PROJECT_ID=your-project-id
FIREBASE_SERVICE_ACCOUNT_JSON=path/to/serviceAccountKey.json

# Cloud Storage
STORAGE_BUCKET=your-bucket-name
STORAGE_PROVIDER=firebase # or 's3'

# Speech-to-Text
STT_PROVIDER=google # or 'assemblyai', 'whisper'
STT_API_KEY=your_api_key

# Server
PORT=8080
ENV=development # or 'production'

# Security
JWT_SECRET=your_secret_key
```

---

## Next Steps

1. **Decision Points**:
   - Choose database: PostgreSQL (recommended) vs Firestore
   - Choose transcription service: Google vs AssemblyAI vs Whisper
   - Choose deployment: Railway vs Render vs Fly.io

2. **Set Up Accounts**:
   - Firebase project
   - Railway/Render account
   - Cloud storage provider
   - Speech-to-text service

3. **Initial Development**:
   - Start with Go API skeleton
   - Set up database and migrations
   - Create basic authentication flow
   - Build iOS recording interface

4. **Testing Strategy**:
   - Unit tests for business logic
   - Integration tests for API endpoints
   - Manual testing with parks department staff
   - TestFlight for iOS beta testing

---

## Questions to Consider

1. **Audio Format**: What quality is acceptable? (Higher quality = larger files)
   - Recommendation: AAC at 64kbps (good balance)

2. **Offline Mode**: How critical is offline functionality?
   - Parks may have poor cell service
   - Recommendation: Queue uploads for later

3. **Collaboration**: Will team members need to see each other's memos?
   - Might need team/department grouping

4. **Data Retention**: How long should memos be kept?
   - Storage costs consideration

5. **Map Integration**: Would a map view of all memos be useful?
   - Very useful for parks department!

---

## Recommended Tech Stack Summary

**For simplicity, zero cost, and native iOS capabilities:**

- **iOS**: Swift + SwiftUI
- **Speech Recognition**: iOS Speech Framework (native, free, on-device)
- **Backend**: Go + Gin framework
- **Database**: PostgreSQL on Railway
- **Deployment**: Railway.app (includes PostgreSQL)
- **Auth**: Firebase Authentication
- **Storage**: Firebase Cloud Storage (5GB free)

This stack gives you a professional, scalable application with **$0 cost** during development and minimal costs in production. No external speech-to-text service needed!
