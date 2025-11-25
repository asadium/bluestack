package blob

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// BlobStore defines the interface for blob storage operations.
// This abstraction allows for different storage backends (in-memory, file-based, SQLite, etc.)
// to be swapped in without changing the service implementation.
type BlobStore interface {
	// CreateContainer creates a new container with the given name in the specified account.
	CreateContainer(ctx context.Context, account, containerName string) error

	// DeleteContainer deletes a container and all its blobs.
	DeleteContainer(ctx context.Context, account, containerName string) error

	// ContainerExists checks if a container exists.
	ContainerExists(ctx context.Context, account, containerName string) (bool, error)

	// PutBlob stores a blob in the specified container.
	PutBlob(ctx context.Context, account, containerName, blobName string, content []byte, contentType string, metadata map[string]string) error

	// GetBlob retrieves a blob from storage.
	GetBlob(ctx context.Context, account, containerName, blobName string) (*Blob, error)

	// DeleteBlob removes a blob from storage.
	DeleteBlob(ctx context.Context, account, containerName, blobName string) error

	// ListBlobs returns a list of blobs in the specified container.
	// prefix can be used to filter blob names, and maxResults limits the number returned.
	ListBlobs(ctx context.Context, account, containerName, prefix string, maxResults int) ([]BlobInfo, error)
}

// FileBlobStore is a file-based implementation of BlobStore.
// It stores blobs as files under DATA_DIR/blob/<account>/<container>/<blobName>.
// This is a simple but effective approach for local development and testing.
type FileBlobStore struct {
	baseDir string
	mu      sync.RWMutex
	// In-memory index for quick lookups (could be replaced with SQLite later)
	containers map[string]bool // key: account/container
}

// NewFileBlobStore creates a new file-based blob store.
func NewFileBlobStore(baseDir string) (*FileBlobStore, error) {
	blobDir := filepath.Join(baseDir, "blob")
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create blob directory: %w", err)
	}

	return &FileBlobStore{
		baseDir:    blobDir,
		containers: make(map[string]bool),
	}, nil
}

// containerPath returns the filesystem path for a container.
func (s *FileBlobStore) containerPath(account, containerName string) string {
	return filepath.Join(s.baseDir, account, containerName)
}

// blobPath returns the filesystem path for a blob.
func (s *FileBlobStore) blobPath(account, containerName, blobName string) string {
	return filepath.Join(s.containerPath(account, containerName), blobName)
}

// containerKey returns a unique key for a container.
func (s *FileBlobStore) containerKey(account, containerName string) string {
	return fmt.Sprintf("%s/%s", account, containerName)
}

func (s *FileBlobStore) CreateContainer(ctx context.Context, account, containerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.containerKey(account, containerName)
	if s.containers[key] {
		return fmt.Errorf("container %s already exists", containerName)
	}

	path := s.containerPath(account, containerName)
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("failed to create container directory: %w", err)
	}

	s.containers[key] = true
	return nil
}

func (s *FileBlobStore) DeleteContainer(ctx context.Context, account, containerName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := s.containerKey(account, containerName)
	if !s.containers[key] {
		return fmt.Errorf("container %s does not exist", containerName)
	}

	path := s.containerPath(account, containerName)
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to delete container directory: %w", err)
	}

	delete(s.containers, key)
	return nil
}

func (s *FileBlobStore) ContainerExists(ctx context.Context, account, containerName string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := s.containerKey(account, containerName)
	return s.containers[key], nil
}

func (s *FileBlobStore) PutBlob(ctx context.Context, account, containerName, blobName string, content []byte, contentType string, metadata map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure container exists
	key := s.containerKey(account, containerName)
	if !s.containers[key] {
		path := s.containerPath(account, containerName)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to ensure container directory: %w", err)
		}
		s.containers[key] = true
	}

	// Write blob file
	blobPath := s.blobPath(account, containerName, blobName)
	blobDir := filepath.Dir(blobPath)
	if err := os.MkdirAll(blobDir, 0755); err != nil {
		return fmt.Errorf("failed to create blob directory: %w", err)
	}

	if err := os.WriteFile(blobPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write blob: %w", err)
	}

	// TODO: Store metadata and content type in a separate metadata file or SQLite
	return nil
}

func (s *FileBlobStore) GetBlob(ctx context.Context, account, containerName, blobName string) (*Blob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	blobPath := s.blobPath(account, containerName, blobName)
	content, err := os.ReadFile(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("blob %s does not exist", blobName)
		}
		return nil, fmt.Errorf("failed to read blob: %w", err)
	}

	info, err := os.Stat(blobPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat blob: %w", err)
	}

	// TODO: Load metadata and content type from metadata store
	return &Blob{
		Name:        blobName,
		Container:   containerName,
		Account:     account,
		Content:     content,
		ContentType: "application/octet-stream", // Default
		Size:        info.Size(),
		CreatedAt:   info.ModTime(),
		ModifiedAt:  info.ModTime(),
		Metadata:    make(map[string]string),
	}, nil
}

func (s *FileBlobStore) DeleteBlob(ctx context.Context, account, containerName, blobName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	blobPath := s.blobPath(account, containerName, blobName)
	if err := os.Remove(blobPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("blob %s does not exist", blobName)
		}
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}

func (s *FileBlobStore) ListBlobs(ctx context.Context, account, containerName, prefix string, maxResults int) ([]BlobInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	containerPath := s.containerPath(account, containerName)
	if _, err := os.Stat(containerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("container %s does not exist", containerName)
	}

	var results []BlobInfo
	err := filepath.Walk(containerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil // Skip directories
		}

		// Get relative blob name
		relPath, err := filepath.Rel(containerPath, path)
		if err != nil {
			return err
		}

		blobName := filepath.ToSlash(relPath) // Normalize path separators

		// Apply prefix filter
		if prefix != "" && !filepath.HasPrefix(blobName, prefix) {
			return nil
		}

		// Apply max results limit
		if maxResults > 0 && len(results) >= maxResults {
			return filepath.SkipAll // Stop walking
		}

		// TODO: Load metadata and content type from metadata store
		results = append(results, BlobInfo{
			Name:         blobName,
			ContentType:  "application/octet-stream",
			Size:         info.Size(),
			LastModified: info.ModTime(),
			Metadata:     make(map[string]string),
		})

		return nil
	})

	return results, err
}

