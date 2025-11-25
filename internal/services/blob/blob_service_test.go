package blob

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/asad/bluestack/internal/logging"
)

// setupTestService creates a test blob service with a temporary store.
func setupTestService(t *testing.T) (*BlobService, BlobStore, func()) {
	// Create temporary directory for test data
	tmpDir, err := os.MkdirTemp("", "bluestack-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create logger (using info level to reduce noise in tests)
	logger, err := logging.NewLogger("error")
	if err != nil {
		t.Fatalf("failed to create logger: %v", err)
	}

	// Create blob store
	store, err := NewFileBlobStore(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create blob store: %v", err)
	}

	// Create service
	service := NewBlobService(store, logger)

	// Cleanup function
	cleanup := func() {
		os.RemoveAll(tmpDir)
	}

	return service, store, cleanup
}

// TestBlobService_CreateContainer tests container creation.
func TestBlobService_CreateContainer(t *testing.T) {
	service, store, cleanup := setupTestService(t)
	defer cleanup()

	// Create router and register routes
	router := chi.NewRouter()
	service.RegisterRoutes(router)

	// Test creating a container
	req := httptest.NewRequest("PUT", "/blob/testaccount/testcontainer", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Verify container exists
	exists, err := store.ContainerExists(context.Background(), "testaccount", "testcontainer")
	if err != nil {
		t.Fatalf("failed to check container existence: %v", err)
	}
	if !exists {
		t.Error("container should exist after creation")
	}
}

// TestBlobService_PutGetBlob tests blob upload and download.
func TestBlobService_PutGetBlob(t *testing.T) {
	service, store, cleanup := setupTestService(t)
	defer cleanup()

	// Create container first
	err := store.CreateContainer(context.Background(), "testaccount", "testcontainer")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// Create router and register routes
	router := chi.NewRouter()
	service.RegisterRoutes(router)

	// Test uploading a blob
	blobContent := []byte("test blob content")
	req := httptest.NewRequest("PUT", "/blob/testaccount/testcontainer/testblob.txt", bytes.NewReader(blobContent))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	// Test downloading the blob
	req = httptest.NewRequest("GET", "/blob/testaccount/testcontainer/testblob.txt", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if !bytes.Equal(w.Body.Bytes(), blobContent) {
		t.Errorf("expected content %q, got %q", string(blobContent), w.Body.String())
	}
}

// TestBlobService_DeleteBlob tests blob deletion.
func TestBlobService_DeleteBlob(t *testing.T) {
	service, store, cleanup := setupTestService(t)
	defer cleanup()

	// Create container and blob
	err := store.CreateContainer(context.Background(), "testaccount", "testcontainer")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	err = store.PutBlob(context.Background(), "testaccount", "testcontainer", "testblob.txt", []byte("content"), "text/plain", nil)
	if err != nil {
		t.Fatalf("failed to put blob: %v", err)
	}

	// Create router and register routes
	router := chi.NewRouter()
	service.RegisterRoutes(router)

	// Test deleting the blob
	req := httptest.NewRequest("DELETE", "/blob/testaccount/testcontainer/testblob.txt", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}

	// Verify blob is deleted
	_, err = store.GetBlob(context.Background(), "testaccount", "testcontainer", "testblob.txt")
	if err == nil {
		t.Error("blob should not exist after deletion")
	}
}

// TestBlobService_ListBlobs tests blob listing.
func TestBlobService_ListBlobs(t *testing.T) {
	service, store, cleanup := setupTestService(t)
	defer cleanup()

	// Create container and multiple blobs
	err := store.CreateContainer(context.Background(), "testaccount", "testcontainer")
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	blobs := []string{"blob1.txt", "blob2.txt", "prefix/blob3.txt"}
	for _, blobName := range blobs {
		err = store.PutBlob(context.Background(), "testaccount", "testcontainer", blobName, []byte("content"), "text/plain", nil)
		if err != nil {
			t.Fatalf("failed to put blob %s: %v", blobName, err)
		}
	}

	// Create router and register routes
	router := chi.NewRouter()
	service.RegisterRoutes(router)

	// Test listing blobs
	req := httptest.NewRequest("GET", "/blob/testaccount/testcontainer?list", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var result BlobListResult
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(result.Blobs) < len(blobs) {
		t.Errorf("expected at least %d blobs, got %d", len(blobs), len(result.Blobs))
	}
}

// TODO: Add integration tests using official Azure SDKs pointing at the local endpoint.
// Example:
//   - Use azure-sdk-for-go to create a blob client pointing to http://localhost:4566
//   - Test that the SDK can successfully interact with our emulator
//   - Verify that common SDK operations work correctly

