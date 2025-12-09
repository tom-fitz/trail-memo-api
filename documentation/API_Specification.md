# TrailMemo API Specification v1.0

## Base URL
```
Production: https://your-app.railway.app/api/v1
Development: http://localhost:8080/api/v1
```

## Authentication

All authenticated endpoints require a Firebase ID token in the Authorization header:

```
Authorization: Bearer <firebase_id_token>
```

To get a Firebase ID token from iOS:
```swift
let token = try await Auth.auth().currentUser?.getIDToken()
```

---

## Endpoints

### Health Check

#### GET /health
Check if the API is running.

**Authentication:** None required

**Response:**
```json
{
  "status": "ok",
  "service": "trailmemo-api",
  "version": "1.0.0"
}
```

---

## Authentication Endpoints

### Register User

#### POST /api/v1/auth/register

Create a new user account. Note: Firebase handles the actual authentication, this endpoint just creates the user record in our database.

**Authentication:** Required (Firebase token)

**Request Body:**
```json
{
  "display_name": "John Doe",
  "department": "Parks & Recreation"
}
```

**Response:** `201 Created`
```json
{
  "user_id": "firebase_uid_here",
  "email": "john@example.com",
  "display_name": "John Doe",
  "department": "Parks & Recreation",
  "created_at": "2024-12-07T10:30:00Z"
}
```

**Errors:**
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid Firebase token
- `409 Conflict` - User already exists

---

### Get Current User

#### GET /api/v1/auth/me

Get information about the currently authenticated user.

**Authentication:** Required

**Response:** `200 OK`
```json
{
  "user_id": "firebase_uid_here",
  "email": "john@example.com",
  "display_name": "John Doe",
  "department": "Parks & Recreation",
  "created_at": "2024-12-07T10:30:00Z"
}
```

**Errors:**
- `401 Unauthorized` - Invalid or missing token
- `404 Not Found` - User not found in database

---

## Memo Endpoints

### Create Memo

#### POST /api/v1/memos

Create a new voice memo with audio file upload and transcribed text.

**Authentication:** Required

**Content-Type:** `multipart/form-data`

**Form Fields:**
- `audio` (file, required) - Audio file (supported: mp3, m4a, wav, aac)
- `text` (string, required) - Transcribed text from iOS Speech
- `duration_seconds` (integer, required) - Duration in seconds
- `latitude` (float, optional) - GPS latitude
- `longitude` (float, optional) - GPS longitude
- `location_accuracy` (float, optional) - GPS accuracy in meters
- `park_name` (string, optional) - Name of the park/location
- `title` (string, optional) - Custom title for the memo

**Example cURL:**
```bash
curl -X POST https://your-app.railway.app/api/v1/memos \
  -H "Authorization: Bearer $FIREBASE_TOKEN" \
  -F "audio=@recording.m4a" \
  -F "text=Found a fallen tree blocking the main trail" \
  -F "duration_seconds=45" \
  -F "latitude=45.6789" \
  -F "longitude=-111.0123" \
  -F "location_accuracy=10.5" \
  -F "park_name=Lindley Park"
```

**Response:** `201 Created`
```json
{
  "memo_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "firebase_uid_here",
  "user_name": "John Doe",
  "title": null,
  "audio_url": "https://storage.googleapis.com/bucket/path/to/audio.m4a",
  "text": "Found a fallen tree blocking the main trail",
  "duration_seconds": 45,
  "location": {
    "latitude": 45.6789,
    "longitude": -111.0123,
    "accuracy": 10.5,
    "address": null
  },
  "park_name": "Lindley Park",
  "created_at": "2024-12-07T14:30:00Z",
  "updated_at": "2024-12-07T14:30:00Z"
}
```

**Errors:**
- `400 Bad Request` - Missing required fields or invalid file
- `401 Unauthorized` - Invalid token
- `413 Payload Too Large` - File exceeds size limit (recommend 50MB max)

---

### List Memos

#### GET /api/v1/memos

Get a list of ALL users' memos (for map view).

**Authentication:** Required

**Query Parameters:**
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 100, max: 500) - Items per page
- `park_name` (string, optional) - Filter by park name
- `start_date` (ISO 8601, optional) - Filter by creation date (inclusive)
- `end_date` (ISO 8601, optional) - Filter by creation date (inclusive)
- `user_id` (string, optional) - Filter by specific user

**Example:**
```
GET /api/v1/memos?page=1&limit=100&park_name=Lindley%20Park
```

**Response:** `200 OK`
```json
{
  "memos": [
    {
      "memo_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "firebase_uid_here",
      "user_name": "John Doe",
      "title": "Trail maintenance issue",
      "audio_url": "https://storage.googleapis.com/...",
      "text": "Found a fallen tree blocking the main trail...",
      "duration_seconds": 45,
      "location": {
        "latitude": 45.6789,
        "longitude": -111.0123,
        "accuracy": 10.5,
        "address": "123 Park Ave, Bozeman, MT"
      },
      "park_name": "Lindley Park",
      "created_at": "2024-12-07T14:30:00Z",
      "updated_at": "2024-12-07T14:32:00Z"
    }
  ],
  "pagination": {
    "current_page": 1,
    "total_pages": 5,
    "total_items": 487,
    "items_per_page": 100,
    "has_next": true,
    "has_previous": false
  }
}
```

**Notes:**
- Returns memos from ALL users for collaborative map view
- Sorted by created_at DESC by default

**Errors:**
- `401 Unauthorized` - Invalid token
- `400 Bad Request` - Invalid query parameters

---

### Get Memo

#### GET /api/v1/memos/:id

Get a specific memo by ID.

**Authentication:** Required

**Path Parameters:**
- `id` (uuid, required) - Memo ID

**Response:** `200 OK`
```json
{
  "memo_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "firebase_uid_here",
  "user_name": "John Doe",
  "title": "Trail maintenance issue",
  "audio_url": "https://storage.googleapis.com/...",
  "text": "Found a fallen tree blocking the main trail near the north entrance. Approximately 2 feet in diameter. Will need chain saw to clear.",
  "duration_seconds": 45,
  "location": {
    "latitude": 45.6789,
    "longitude": -111.0123,
    "accuracy": 10.5,
    "address": "123 Park Ave, Bozeman, MT"
  },
  "park_name": "Lindley Park",
  "created_at": "2024-12-07T14:30:00Z",
  "updated_at": "2024-12-07T14:32:00Z"
}
```

**Errors:**
- `401 Unauthorized` - Invalid token
- `404 Not Found` - Memo doesn't exist

---

### Update Memo

#### PUT /api/v1/memos/:id

Update a memo's editable fields. Only the creator can update their memos.

**Authentication:** Required

**Path Parameters:**
- `id` (uuid, required) - Memo ID

**Request Body:**
```json
{
  "title": "Updated title",
  "text": "Edited text content",
  "park_name": "Different Park"
}
```

**Notes:**
- All fields are optional
- Only include fields you want to update
- Cannot update: memo_id, user_id, user_name, audio_url, created_at, location

**Response:** `200 OK`
```json
{
  "memo_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "firebase_uid_here",
  "user_name": "John Doe",
  "title": "Updated title",
  "audio_url": "https://storage.googleapis.com/...",
  "text": "Edited text content",
  "duration_seconds": 45,
  "location": {
    "latitude": 45.6789,
    "longitude": -111.0123,
    "accuracy": 10.5,
    "address": "123 Park Ave, Bozeman, MT"
  },
  "park_name": "Different Park",
  "created_at": "2024-12-07T14:30:00Z",
  "updated_at": "2024-12-07T15:45:00Z"
}
```

**Errors:**
- `400 Bad Request` - Invalid request body
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Memo belongs to another user
- `404 Not Found` - Memo doesn't exist

---

### Delete Memo

#### DELETE /api/v1/memos/:id

Delete a memo and its associated audio file. Only the creator can delete their memos.

**Authentication:** Required

**Path Parameters:**
- `id` (uuid, required) - Memo ID

**Response:** `204 No Content`

**Errors:**
- `401 Unauthorized` - Invalid token
- `403 Forbidden` - Memo belongs to another user
- `404 Not Found` - Memo doesn't exist

---

### Get Nearby Memos

#### GET /api/v1/memos/nearby

Find memos near a specific location (from all users).

**Authentication:** Required

**Query Parameters:**
- `latitude` (float, required) - Center latitude
- `longitude` (float, required) - Center longitude
- `radius_meters` (integer, default: 1000, max: 50000) - Search radius in meters
- `limit` (integer, default: 50, max: 200) - Maximum results

**Example:**
```
GET /api/v1/memos/nearby?latitude=45.6789&longitude=-111.0123&radius_meters=5000&limit=50
```

**Response:** `200 OK`
```json
{
  "memos": [
    {
      "memo_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_name": "John Doe",
      "title": "Trail issue",
      "park_name": "Lindley Park",
      "location": {
        "latitude": 45.6789,
        "longitude": -111.0123
      },
      "distance_meters": 234,
      "created_at": "2024-12-07T14:30:00Z"
    }
  ],
  "center": {
    "latitude": 45.6789,
    "longitude": -111.0123
  },
  "radius_meters": 5000,
  "total_found": 12
}
```

**Notes:**
- Results are sorted by distance (closest first)
- Returns memos from ALL users
- Distance calculated using Haversine formula

**Errors:**
- `400 Bad Request` - Missing or invalid coordinates
- `401 Unauthorized` - Invalid token

---

### Search Memos

#### GET /api/v1/memos/search

Full-text search across memo text content (searches all users' memos).

**Authentication:** Required

**Query Parameters:**
- `q` (string, required) - Search query
- `page` (integer, default: 1) - Page number
- `limit` (integer, default: 20, max: 100) - Items per page

**Example:**
```
GET /api/v1/memos/search?q=fallen+tree&page=1&limit=20
```

**Response:** `200 OK`
```json
{
  "results": [
    {
      "memo_id": "550e8400-e29b-41d4-a716-446655440000",
      "user_name": "John Doe",
      "title": "Trail maintenance issue",
      "text": "Found a **fallen tree** blocking the main trail...",
      "park_name": "Lindley Park",
      "relevance_score": 0.95,
      "created_at": "2024-12-07T14:30:00Z"
    }
  ],
  "query": "fallen tree",
  "pagination": {
    "current_page": 1,
    "total_pages": 2,
    "total_items": 23
  }
}
```

**Notes:**
- Uses PostgreSQL full-text search
- Supports basic operators: AND, OR, NOT
- Returns relevance score
- Searches all users' memos

**Errors:**
- `400 Bad Request` - Missing or empty query
- `401 Unauthorized` - Invalid token

---

## Tag Endpoints (Future Phase)

---

## File Upload Endpoints

### Get Presigned Upload URL

#### GET /api/v1/upload/presigned-url

Get a presigned URL for direct upload to Cloud Storage (alternative to multipart upload).

**Authentication:** Required

**Query Parameters:**
- `filename` (string, required) - Original filename
- `content_type` (string, required) - MIME type (e.g., audio/m4a)

**Response:** `200 OK`
```json
{
  "upload_url": "https://storage.googleapis.com/bucket/uploads/user123/memo456.m4a?signature=...",
  "file_path": "uploads/user123/memo456.m4a",
  "expires_at": "2024-12-07T15:30:00Z"
}
```

**Usage Flow:**
1. Client requests presigned URL
2. Client uploads file directly to Cloud Storage using the URL
3. Client calls POST /api/v1/memos with `file_path` instead of multipart upload

**Errors:**
- `400 Bad Request` - Missing or invalid parameters
- `401 Unauthorized` - Invalid token

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Human-readable error message",
    "details": {
      "field": "email",
      "reason": "Invalid email format"
    }
  }
}
```

**Common Error Codes:**
- `VALIDATION_ERROR` - Invalid request data
- `AUTHENTICATION_ERROR` - Invalid or missing auth token
- `AUTHORIZATION_ERROR` - Insufficient permissions
- `NOT_FOUND` - Resource doesn't exist
- `CONFLICT` - Resource already exists
- `RATE_LIMIT_EXCEEDED` - Too many requests
- `INTERNAL_ERROR` - Server error

---

## Rate Limiting

- **Authenticated requests:** 1000 requests/hour per user
- **Unauthenticated requests:** 100 requests/hour per IP
- **File uploads:** 100 uploads/day per user

Rate limit headers included in all responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 987
X-RateLimit-Reset: 1638889200
```

---

## Pagination

All list endpoints support pagination with consistent parameters:

**Query Parameters:**
- `page` - Page number (1-indexed)
- `limit` - Items per page

**Response Structure:**
```json
{
  "items": [...],
  "pagination": {
    "current_page": 1,
    "total_pages": 10,
    "total_items": 197,
    "items_per_page": 20,
    "has_next": true,
    "has_previous": false
  }
}
```

---

## Webhooks (Future)

For real-time updates, webhooks can be configured to notify about:
- Transcription completion
- Failed transcription
- Memo created by team member
- Memo deleted

Configure webhooks in user settings.

---

## Data Types Reference

### Memo Object
```typescript
interface Memo {
  memo_id: string;          // UUID
  user_id: string;          // Firebase UID
  user_name: string;        // Display name of creator
  title: string | null;
  audio_url: string;        // HTTPS URL
  text: string;             // Transcribed from iOS Speech
  duration_seconds: number;
  location: Location | null;
  park_name: string | null;
  created_at: string;       // ISO 8601
  updated_at: string;       // ISO 8601
}
```

### Location Object
```typescript
interface Location {
  latitude: number;   // Decimal degrees (-90 to 90)
  longitude: number;  // Decimal degrees (-180 to 180)
  accuracy: number;   // Meters
  address: string | null;
}
```

### User Object
```typescript
interface User {
  user_id: string;      // Firebase UID
  email: string;
  display_name: string | null;
  department: string | null;
  created_at: string;   // ISO 8601
}
```

---

## iOS Integration Examples

### Recording with Speech Recognition

```swift
import Speech

class RecordViewModel: ObservableObject {
    @Published var transcribedText = ""
    @Published var isRecording = false
    
    private var audioEngine = AVAudioEngine()
    private var speechRecognizer = SFSpeechRecognizer()
    private var recognitionRequest: SFSpeechAudioBufferRecognitionRequest?
    private var recognitionTask: SFSpeechRecognitionTask?
    private var audioRecorder: AVAudioRecorder?
    
    func startRecording() async throws {
        // Request speech recognition permission
        guard await SFSpeechRecognizer.requestAuthorization() == .authorized else {
            throw RecordingError.speechNotAuthorized
        }
        
        // Set up audio session
        let audioSession = AVAudioSession.sharedInstance()
        try audioSession.setCategory(.record, mode: .measurement)
        try audioSession.setActive(true, options: .notifyOthersOnDeactivation)
        
        // Start audio recording for file
        let audioURL = getAudioURL()
        let settings: [String: Any] = [
            AVFormatIDKey: Int(kAudioFormatMPEG4AAC),
            AVSampleRateKey: 44100,
            AVNumberOfChannelsKey: 1,
            AVEncoderAudioQualityKey: AVAudioQuality.high.rawValue
        ]
        audioRecorder = try AVAudioRecorder(url: audioURL, settings: settings)
        audioRecorder?.record()
        
        // Start speech recognition
        recognitionRequest = SFSpeechAudioBufferRecognitionRequest()
        guard let recognitionRequest = recognitionRequest else { return }
        recognitionRequest.shouldReportPartialResults = true
        
        let inputNode = audioEngine.inputNode
        recognitionTask = speechRecognizer?.recognitionTask(with: recognitionRequest) { [weak self] result, error in
            if let result = result {
                DispatchQueue.main.async {
                    self?.transcribedText = result.bestTranscription.formattedString
                }
            }
        }
        
        let recordingFormat = inputNode.outputFormat(forBus: 0)
        inputNode.installTap(onBus: 0, bufferSize: 1024, format: recordingFormat) { buffer, _ in
            recognitionRequest.append(buffer)
        }
        
        audioEngine.prepare()
        try audioEngine.start()
        isRecording = true
    }
    
    func stopRecording() {
        audioEngine.stop()
        audioEngine.inputNode.removeTap(onBus: 0)
        recognitionRequest?.endAudio()
        audioRecorder?.stop()
        isRecording = false
    }
}
```

### Creating a Memo

```swift
func uploadMemo(
    audioURL: URL, 
    text: String,
    location: CLLocation?, 
    parkName: String?
) async throws -> Memo {
    let url = URL(string: "\(Config.apiBaseURL)/api/v1/memos")!
    var request = URLRequest(url: url)
    request.httpMethod = "POST"
    
    // Get Firebase token
    let token = try await Auth.auth().currentUser?.getIDToken()
    request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
    
    // Create multipart form data
    let boundary = UUID().uuidString
    request.setValue("multipart/form-data; boundary=\(boundary)", 
                     forHTTPHeaderField: "Content-Type")
    
    var body = Data()
    
    // Add audio file
    let audioData = try Data(contentsOf: audioURL)
    body.append("--\(boundary)\r\n")
    body.append("Content-Disposition: form-data; name=\"audio\"; filename=\"memo.m4a\"\r\n")
    body.append("Content-Type: audio/m4a\r\n\r\n")
    body.append(audioData)
    body.append("\r\n")
    
    // Add transcribed text
    body.append("--\(boundary)\r\n")
    body.append("Content-Disposition: form-data; name=\"text\"\r\n\r\n")
    body.append("\(text)\r\n")
    
    // Add duration
    let duration = Int(AVURLAsset(url: audioURL).duration.seconds)
    body.append("--\(boundary)\r\n")
    body.append("Content-Disposition: form-data; name=\"duration_seconds\"\r\n\r\n")
    body.append("\(duration)\r\n")
    
    // Add metadata
    if let location = location {
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"latitude\"\r\n\r\n")
        body.append("\(location.coordinate.latitude)\r\n")
        
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"longitude\"\r\n\r\n")
        body.append("\(location.coordinate.longitude)\r\n")
        
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"location_accuracy\"\r\n\r\n")
        body.append("\(location.horizontalAccuracy)\r\n")
    }
    
    if let parkName = parkName {
        body.append("--\(boundary)\r\n")
        body.append("Content-Disposition: form-data; name=\"park_name\"\r\n\r\n")
        body.append("\(parkName)\r\n")
    }
    
    body.append("--\(boundary)--\r\n")
    request.httpBody = body
    
    let (data, _) = try await URLSession.shared.data(for: request)
    let decoder = JSONDecoder()
    decoder.dateDecodingStrategy = .iso8601
    return try decoder.decode(Memo.self, from: data)
}
```

### Fetching Memos

```swift
func fetchMemos(page: Int = 1, limit: Int = 20) async throws -> [Memo] {
    let url = URL(string: "\(Config.apiBaseURL)/api/v1/memos?page=\(page)&limit=\(limit)")!
    var request = URLRequest(url: url)
    
    let token = try await Auth.auth().currentUser?.getIDToken()
    request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")
    
    let (data, _) = try await URLSession.shared.data(for: request)
    
    struct Response: Codable {
        let memos: [Memo]
    }
    
    let decoder = JSONDecoder()
    decoder.dateDecodingStrategy = .iso8601
    let response = try decoder.decode(Response.self, from: data)
    return response.memos
}
```

---

## Changelog

### v1.0.0 (2024-12-07)
- Initial API specification
- Basic CRUD operations for memos
- Authentication with Firebase
- File upload support
- Speech-to-text integration
- Location-based features
- Full-text search
