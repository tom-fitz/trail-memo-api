package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"path/filepath"
	"time"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/storage"
	"github.com/google/uuid"
	"google.golang.org/api/option"
)

// FirebaseService handles Firebase operations
type FirebaseService struct {
	app     *firebase.App
	auth    *auth.Client
	storage *storage.Client
	bucket  string
}

// NewFirebaseService creates a new Firebase service instance
func NewFirebaseService(projectID, bucketName, serviceAccountPath, serviceAccountJSON string) (*FirebaseService, error) {
	ctx := context.Background()

	var opt option.ClientOption
	if serviceAccountJSON != "" {
		// Use JSON from environment variable (production)
		opt = option.WithCredentialsJSON([]byte(serviceAccountJSON))
	} else {
		// Use file path (local development)
		opt = option.WithCredentialsFile(serviceAccountPath)
	}

	// Initialize Firebase app
	config := &firebase.Config{
		ProjectID:     projectID,
		StorageBucket: bucketName,
	}

	app, err := firebase.NewApp(ctx, config, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase app: %v", err)
	}

	// Initialize Auth client
	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase auth: %v", err)
	}

	// Initialize Storage client
	storageClient, err := app.Storage(ctx)
	if err != nil {
		return nil, fmt.Errorf("error initializing Firebase storage: %v", err)
	}

	return &FirebaseService{
		app:     app,
		auth:    authClient,
		storage: storageClient,
		bucket:  bucketName,
	}, nil
}

// VerifyIDToken verifies a Firebase ID token and returns the user ID
func (fs *FirebaseService) VerifyIDToken(ctx context.Context, idToken string) (string, error) {
	token, err := fs.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", fmt.Errorf("error verifying ID token: %v", err)
	}
	return token.UID, nil
}

// GetUserByUID retrieves user information from Firebase Auth
func (fs *FirebaseService) GetUserByUID(ctx context.Context, uid string) (*auth.UserRecord, error) {
	user, err := fs.auth.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("error getting user: %v", err)
	}
	return user, nil
}

// UploadAudioFile uploads an audio file to Firebase Storage
func (fs *FirebaseService) UploadAudioFile(ctx context.Context, file *multipart.FileHeader, userID string) (string, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("memos/%s/%s%s", userID, uuid.New().String(), ext)

	// Get bucket
	bucket, err := fs.storage.Bucket(fs.bucket)
	if err != nil {
		return "", fmt.Errorf("error getting bucket: %v", err)
	}

	// Create object writer
	obj := bucket.Object(fileName)
	writer := obj.NewWriter(ctx)
	writer.ContentType = file.Header.Get("Content-Type")
	writer.Metadata = map[string]string{
		"uploaded_by": userID,
		"uploaded_at": time.Now().Format(time.RFC3339),
	}

	// Copy file data to storage
	if _, err := io.Copy(writer, src); err != nil {
		writer.Close()
		return "", fmt.Errorf("error uploading file: %v", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("error closing writer: %v", err)
	}

	// Make file publicly accessible
	// if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
	// 	return "", fmt.Errorf("error setting ACL: %v", err)
	// }

	// Generate public URL
	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", fs.bucket, fileName)
	return url, nil
}

// DeleteAudioFile deletes an audio file from Firebase Storage
func (fs *FirebaseService) DeleteAudioFile(ctx context.Context, audioURL string) error {
	// Extract file path from URL
	// URL format: https://storage.googleapis.com/bucket-name/path/to/file
	bucket, err := fs.storage.Bucket(fs.bucket)
	if err != nil {
		return fmt.Errorf("error getting bucket: %v", err)
	}

	// Parse file path from URL
	// This is a simplified version - you might need more robust URL parsing
	prefix := fmt.Sprintf("https://storage.googleapis.com/%s/", fs.bucket)
	if len(audioURL) <= len(prefix) {
		return fmt.Errorf("invalid audio URL")
	}
	filePath := audioURL[len(prefix):]

	// Delete the file
	obj := bucket.Object(filePath)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("error deleting file: %v", err)
	}

	return nil
}
