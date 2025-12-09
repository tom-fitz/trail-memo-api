# TrailMemo MVP: Implementation Guide

## MVP Scope

**Core Features:**
1. User records voice memo while speaking
2. iOS Speech Framework transcribes in real-time
3. App captures GPS location automatically
4. Audio + text uploaded to backend
5. Map view shows ALL users' memos with pins
6. Tap pin to see memo details (who, what, where, when)
7. Play audio from any memo
8. Users can only edit/delete their own memos

**What's NOT in MVP:**
- ❌ Offline mode (can add later)
- ❌ Tags/categories (can add later)
- ❌ Search (can add later)
- ❌ Photos (can add later)
- ❌ Comments (can add later)
- ❌ Push notifications (can add later)

---

## Tech Stack (Final)

```
iOS App:
  - Swift + SwiftUI
  - AVFoundation (audio recording)
  - Speech Framework (transcription - FREE)
  - CoreLocation (GPS)
  - MapKit (map view)
  - Firebase Auth (authentication)

Backend API:
  - Go + Gin framework
  - Railway.app (hosting + PostgreSQL)
  - Firebase Auth (token verification)
  - Firebase Storage (audio files)

Cost: $0 for MVP, $5-10/month in production
```

---

## User Flow

### 1. First Launch
```
User opens app
  ↓
Login/Register screen
  ↓
Firebase authentication
  ↓
Request location permission
  ↓
Request microphone permission
  ↓
Request speech recognition permission
  ↓
Map view loads with existing memos
```

### 2. Creating a Memo
```
User taps "+" button on map
  ↓
Record screen appears
  ↓
User taps "Start Recording"
  ↓
Audio records + Speech transcribes in real-time
  ↓
User sees their words appear as they speak
  ↓
User taps "Done"
  ↓
Preview screen shows:
  - Transcribed text (editable)
  - Audio waveform
  - Location on mini-map
  - Park name (optional input)
  ↓
User taps "Save"
  ↓
Upload to backend
  ↓
Success! Navigate back to map
  ↓
New pin appears on map
```

### 3. Viewing Memos
```
User sees map with pins
  ↓
Different colored pins for different users
  ↓
User taps a pin
  ↓
Callout shows: Title, Creator, Date
  ↓
User taps callout
  ↓
Detail screen shows:
  - Full text
  - Audio player
  - Creator info
  - Location details
  - Created date
  - [Edit] [Delete] buttons (only if owner)
```

---

## iOS App Structure

### Main Views

1. **MapView** (Main Screen)
   - Shows all memos as pins
   - Colored by user
   - + button to create memo
   - Current location button

2. **RecordMemoView**
   - Start/Stop recording button
   - Real-time transcription display
   - Audio level indicator
   - Timer

3. **PreviewMemoView**
   - Editable text field
   - Audio player
   - Location info
   - Park name input
   - Save/Cancel buttons

4. **MemoDetailView**
   - Read-only text
   - Audio player
   - Map showing location
   - Creator info
   - Edit/Delete (if owner)

5. **LoginView**
   - Email/password fields
   - Sign in / Sign up toggle

### Key Services

```swift
// AudioService.swift
class AudioService {
    func startRecording(to url: URL) async throws
    func stopRecording() -> URL
    func getDuration(of url: URL) -> TimeInterval
}

// SpeechRecognitionService.swift
class SpeechRecognitionService: ObservableObject {
    @Published var transcribedText = ""
    @Published var isRecognizing = false
    
    func startRecognition() async throws
    func stopRecognition() -> String
}

// LocationService.swift
class LocationService: NSObject, ObservableObject {
    @Published var currentLocation: CLLocation?
    
    func requestPermission()
    func startUpdatingLocation()
}

// APIClient.swift
class APIClient {
    func createMemo(audio: URL, text: String, location: CLLocation?) async throws -> Memo
    func fetchAllMemos() async throws -> [Memo]
    func fetchMemo(id: UUID) async throws -> Memo
    func updateMemo(id: UUID, text: String, title: String?) async throws -> Memo
    func deleteMemo(id: UUID) async throws
}
```

---

## Backend API Structure

### Endpoints (MVP Only)

```
POST   /api/v1/auth/register          - Create user account
GET    /api/v1/auth/me                - Get current user info

POST   /api/v1/memos                  - Create memo (multipart: audio + text)
GET    /api/v1/memos                  - Get all memos (everyone's)
GET    /api/v1/memos/:id              - Get specific memo
PUT    /api/v1/memos/:id              - Update memo (owner only)
DELETE /api/v1/memos/:id              - Delete memo (owner only)
```

### Go Project Structure

```
trailmemo-api/
├── cmd/
│   └── server/
│       └── main.go                   # Entry point
├── internal/
│   ├── handlers/
│   │   ├── auth.go                   # Register, GetMe
│   │   ├── memos.go                  # CRUD operations
│   │   └── health.go                 # Health check
│   ├── middleware/
│   │   ├── auth.go                   # Firebase token verification
│   │   └── cors.go                   # CORS configuration
│   ├── models/
│   │   ├── user.go                   # User struct
│   │   └── memo.go                   # Memo struct
│   ├── repository/
│   │   ├── user_repo.go              # User DB operations
│   │   └── memo_repo.go              # Memo DB operations
│   └── services/
│       ├── firebase.go               # Firebase auth & storage
│       └── storage.go                # File upload helpers
├── config/
│   └── config.go                     # Environment config
├── migrations/
│   └── 001_init.sql                  # Database schema
├── go.mod
├── go.sum
└── .env                               # Environment variables
```

---

## Database Schema (PostgreSQL)

```sql
-- Users table
CREATE TABLE users (
    user_id VARCHAR(128) PRIMARY KEY,  -- Firebase UID
    email VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    department VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Memos table
CREATE TABLE memos (
    memo_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id VARCHAR(128) REFERENCES users(user_id) ON DELETE CASCADE,
    user_name VARCHAR(255) NOT NULL,   -- Denormalized for performance
    title VARCHAR(255),
    audio_url TEXT NOT NULL,
    text TEXT NOT NULL,                -- From iOS Speech
    duration_seconds INTEGER,
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    location_accuracy FLOAT,
    park_name VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_memos_created ON memos(created_at DESC);
CREATE INDEX idx_memos_location ON memos(latitude, longitude);
CREATE INDEX idx_memos_user ON memos(user_id);
```

---

## Implementation Steps

### Week 1: Setup (2-3 days)

**Day 1: Infrastructure**
- [ ] Create Firebase project
  - Enable Email/Password auth
  - Create iOS app in Firebase
  - Set up Cloud Storage
  - Download service account key
- [ ] Create Railway account
  - Create new project
  - Add PostgreSQL database
  - Note connection string
- [ ] Create GitHub repositories
  - `trailmemo-api` (backend)
  - `trailmemo-ios` (iOS app)

**Day 2: Backend Skeleton**
- [ ] Initialize Go project
- [ ] Install dependencies (gin, firebase, sqlx, etc.)
- [ ] Create project structure
- [ ] Set up environment variables
- [ ] Run database migrations
- [ ] Deploy to Railway
- [ ] Test health endpoint

**Day 3: iOS Project**
- [ ] Create iOS project in Xcode
- [ ] Add Firebase SDK
- [ ] Configure Firebase (GoogleService-Info.plist)
- [ ] Add Info.plist permissions
- [ ] Create basic project structure
- [ ] Build and run on simulator

---

### Week 2: Authentication (2-3 days)

**Backend:**
- [ ] Implement Firebase auth middleware
- [ ] Create auth handlers (register, getMe)
- [ ] Test with Postman/cURL

**iOS:**
- [ ] Create LoginView UI
- [ ] Implement Firebase auth flow
- [ ] Create AuthViewModel
- [ ] Handle login states
- [ ] Navigate to MapView on success

**Testing:**
- [ ] Test registration
- [ ] Test login/logout
- [ ] Test token persistence

---

### Week 3: Recording & Speech (4-5 days)

**iOS:**
- [ ] Request microphone permission
- [ ] Request speech recognition permission
- [ ] Create AudioService
  - Start/stop recording
  - Save to file
  - Get duration
- [ ] Create SpeechRecognitionService
  - Real-time transcription
  - Handle permissions
  - Error handling
- [ ] Create RecordMemoView UI
  - Record button
  - Live transcription display
  - Timer
  - Audio level indicator
- [ ] Create PreviewMemoView
  - Show transcribed text (editable)
  - Audio preview
  - Location display
  - Park name input

**Testing:**
- [ ] Test recording on device (not simulator!)
- [ ] Test transcription accuracy
- [ ] Test permissions flow

---

### Week 4: Location & Upload (3-4 days)

**iOS:**
- [ ] Request location permission
- [ ] Create LocationService
- [ ] Capture location during recording
- [ ] Display location on preview

**Backend:**
- [ ] Create memo endpoints
  - POST /api/v1/memos (multipart upload)
  - Verify Firebase token
  - Upload audio to Firebase Storage
  - Save memo to PostgreSQL
  - Return memo object
- [ ] Test file upload

**iOS:**
- [ ] Create APIClient
- [ ] Implement createMemo function
- [ ] Handle upload progress
- [ ] Handle errors
- [ ] Test full flow: Record → Preview → Upload

**Testing:**
- [ ] Test with real device
- [ ] Test with different locations
- [ ] Test file size limits
- [ ] Test error cases (no network, etc.)

---

### Week 5: Map View (4-5 days)

**Backend:**
- [ ] Implement GET /api/v1/memos
  - Return all memos from all users
  - Include user_name, location, etc.
- [ ] Test endpoint

**iOS:**
- [ ] Create MapView with MapKit
- [ ] Fetch all memos from API
- [ ] Create custom map annotations
  - Different colors per user
  - Show user name
- [ ] Add pin taps
  - Show callout with basic info
- [ ] Add + button to create memo
- [ ] Add current location button
- [ ] Create MapViewModel
  - Fetch memos
  - Handle map state
  - Refresh memos

**Testing:**
- [ ] Test with multiple memos
- [ ] Test pin clustering
- [ ] Test different map types
- [ ] Test location accuracy

---

### Week 6: Memo Details & Playback (3-4 days)

**Backend:**
- [ ] Implement GET /api/v1/memos/:id
- [ ] Implement PUT /api/v1/memos/:id
- [ ] Implement DELETE /api/v1/memos/:id
- [ ] Add authorization (owner check)

**iOS:**
- [ ] Create MemoDetailView
  - Show memo text
  - Show creator info
  - Show location on map
  - Show timestamp
- [ ] Create AudioPlayerView
  - Play/pause button
  - Progress bar
  - Time labels
- [ ] Implement edit functionality
  - Only for memo owner
  - Update text
- [ ] Implement delete functionality
  - Only for memo owner
  - Confirmation dialog

**Testing:**
- [ ] Test audio playback
- [ ] Test edit flow
- [ ] Test delete flow
- [ ] Test authorization (can't edit others' memos)

---

### Week 7: Polish & Testing (5-7 days)

**Error Handling:**
- [ ] Network errors
- [ ] Permission denials
- [ ] Invalid data
- [ ] File too large
- [ ] Rate limiting

**UI/UX Improvements:**
- [ ] Loading indicators
- [ ] Empty states
- [ ] Error messages
- [ ] Success animations
- [ ] Pull to refresh on map

**Performance:**
- [ ] Optimize map annotations
- [ ] Cache memos locally
- [ ] Compress audio files
- [ ] Lazy load audio

**Testing:**
- [ ] Test on multiple devices
- [ ] Test with poor network
- [ ] Test with many memos
- [ ] Test edge cases
- [ ] User acceptance testing with parks department

**Documentation:**
- [ ] API documentation
- [ ] iOS code documentation
- [ ] Deployment guide
- [ ] User guide

---

### Week 8: Deployment & Launch (2-3 days)

**Backend:**
- [ ] Environment variables set in Railway
- [ ] Database migrations run
- [ ] Health check passing
- [ ] SSL configured
- [ ] Logs configured

**iOS:**
- [ ] App icons
- [ ] Launch screen
- [ ] Privacy policy
- [ ] Terms of service
- [ ] TestFlight build
- [ ] Beta testing
- [ ] App Store submission

**Final Testing:**
- [ ] End-to-end testing
- [ ] Multiple users simultaneously
- [ ] Different locations
- [ ] Performance under load

---

## MVP Success Metrics

**Technical:**
- [ ] App loads in < 3 seconds
- [ ] Recording starts in < 1 second
- [ ] Transcription accuracy > 90%
- [ ] Upload completes in < 10 seconds (on good network)
- [ ] Map loads with 100+ pins smoothly
- [ ] No crashes in 1 hour of use

**User Experience:**
- [ ] Can create memo in < 30 seconds
- [ ] Can find own memos on map
- [ ] Can view any user's memos
- [ ] Audio plays back clearly
- [ ] Location is accurate within 50 meters

**Reliability:**
- [ ] 99% uptime for API
- [ ] No data loss
- [ ] Handles network interruptions gracefully

---

## Common Issues & Solutions

### Issue: Speech recognition not working
**Solution:** 
- Must run on real device (not simulator)
- Check microphone permission
- Check speech recognition permission
- Make sure device language matches recognition language

### Issue: Location not accurate
**Solution:**
- Use `desiredAccuracy = kCLLocationAccuracyBest`
- Wait for location accuracy < 50m before recording
- Test outdoors (GPS doesn't work well indoors)

### Issue: Audio files too large
**Solution:**
- Use AAC compression
- Sample rate: 44.1kHz (not 48kHz)
- Mono (not stereo)
- Quality: 64kbps

### Issue: Map slow with many pins
**Solution:**
- Enable pin clustering
- Only load visible pins
- Cache pin annotations
- Use custom annotation views

### Issue: Upload fails
**Solution:**
- Check file size limit
- Verify Firebase Storage rules
- Check network connectivity
- Implement retry logic

---

## What to Build AFTER MVP

### Phase 2 Features (Priority Order)
1. **Offline Mode** - Record without network, sync later
2. **Search** - Full-text search through memo text
3. **Filters** - Filter by date, park, user
4. **Tags** - Categorize memos
5. **User Profiles** - See user info, their memos

### Phase 3 Features
1. **Photos** - Attach photos to memos
2. **Comments** - Discuss memos
3. **Teams** - Department grouping
4. **Export** - Download as PDF/CSV
5. **Analytics** - Usage dashboard

---

## Cost Breakdown (Real Numbers)

### MVP Development (Free)
- Railway: $0 (using free credits)
- Firebase Auth: $0 (under 50K users)
- Firebase Storage: $0 (under 5GB)
- iOS: $0 (development)

### First Year Production
- Railway: $60/year ($5/month)
- Firebase: $0-10/year (very light usage)
- Apple Developer: $99/year (required for App Store)
- **Total: ~$160-170/year**

### At Scale (100 users, 1000 memos/month)
- Railway: $240/year ($20/month)
- Firebase: $50/year
- Apple Developer: $99/year
- **Total: ~$390/year**

---

## Quick Reference: Key Commands

### Run iOS App
```bash
open trailmemo-ios/TrailMemo.xcodeproj
# Press ⌘+R to run
```

### Run Backend Locally
```bash
cd trailmemo-api
go run cmd/server/main.go
```

### Deploy Backend
```bash
git push origin main
# Railway auto-deploys on push
```

### Test API
```bash
# Health check
curl https://your-app.railway.app/health

# Create memo (with auth)
curl -X POST https://your-app.railway.app/api/v1/memos \
  -H "Authorization: Bearer $TOKEN" \
  -F "audio=@test.m4a" \
  -F "text=Test memo"
```

### Database Commands
```bash
# Connect to Railway PostgreSQL
psql $DATABASE_URL

# Run migration
psql $DATABASE_URL -f migrations/001_init.sql

# Query memos
psql $DATABASE_URL -c "SELECT * FROM memos LIMIT 5;"
```

---

## Timeline Summary

```
Week 1:  Setup & Infrastructure           [████████░░] 80%
Week 2:  Authentication                   [██████████] 100%
Week 3:  Recording & Speech Recognition   [██████████] 100%
Week 4:  Location & Upload                [██████████] 100%
Week 5:  Map View                         [██████████] 100%
Week 6:  Memo Details & Playback          [██████████] 100%
Week 7:  Polish & Testing                 [██████████] 100%
Week 8:  Deployment & Launch              [██████████] 100%

Total: 8 weeks to MVP
```

**Realistic Timeline:** 8-10 weeks working part-time, 4-6 weeks full-time

---

## Final Checklist Before Launch

### Backend
- [ ] API deployed to Railway
- [ ] Database migrations complete
- [ ] Environment variables configured
- [ ] Firebase Storage rules set
- [ ] SSL/HTTPS working
- [ ] Rate limiting enabled
- [ ] Logs configured

### iOS
- [ ] All permissions requested properly
- [ ] Error handling complete
- [ ] No compiler warnings
- [ ] App icons added
- [ ] Launch screen added
- [ ] Privacy policy linked
- [ ] TestFlight build uploaded
- [ ] Beta testers invited

### Testing
- [ ] End-to-end flow works
- [ ] Tested on multiple devices
- [ ] Tested with poor network
- [ ] Tested with 100+ memos
- [ ] Parks department approved

### Documentation
- [ ] README with setup instructions
- [ ] API documentation
- [ ] User guide
- [ ] Admin guide

---

## Success! 🎉

You now have a fully functional TrailMemo app that:
- Records voice memos with real-time transcription
- Captures GPS location automatically
- Uploads to cloud storage
- Shows all team members' memos on a map
- Allows playback and editing
- Costs < $15/month to run

