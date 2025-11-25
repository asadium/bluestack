package blob

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/asad/bluestack/internal/core"
	"github.com/asad/bluestack/internal/logging"
)

// BlobService implements the Azure Blob Storage service emulator.
// It provides HTTP handlers for basic blob operations following Azure REST API patterns.
type BlobService struct {
	store  BlobStore
	logger logging.Logger
}

// NewBlobService creates a new blob service instance.
func NewBlobService(store BlobStore, logger logging.Logger) *BlobService {
	return &BlobService{
		store:  store,
		logger: logger,
	}
}

// Name returns the service identifier.
func (s *BlobService) Name() string {
	return "blob"
}

// RegisterRoutes sets up HTTP routes for blob operations.
// Routes follow a simplified Azure Blob Storage REST API pattern:
//   - PUT /{account}/{container} - Create container
//   - DELETE /{account}/{container} - Delete container
//   - PUT /{account}/{container}/{blobName} - Upload blob
//   - GET /{account}/{container}/{blobName} - Download blob
//   - DELETE /{account}/{container}/{blobName} - Delete blob
//   - GET /{account}/{container}?list - List blobs
func (s *BlobService) RegisterRoutes(router chi.Router) {
	// Container operations
	router.Put("/{account}/{container}", s.handleCreateContainer)
	router.Delete("/{account}/{container}", s.handleDeleteContainer)

	// Blob operations
	router.Put("/{account}/{container}/{blobName:*}", s.handlePutBlob)
	router.Get("/{account}/{container}/{blobName:*}", s.handleGetBlob)
	router.Delete("/{account}/{container}/{blobName:*}", s.handleDeleteBlob)

	// List blobs
	router.Get("/{account}/{container}", s.handleListBlobs)
}

// handleCreateContainer handles PUT /{account}/{container} to create a container.
func (s *BlobService) handleCreateContainer(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")

	if account == "" || containerName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account and container name are required")
		return
	}

	err := s.store.CreateContainer(r.Context(), account, containerName)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			s.writeError(w, http.StatusConflict, "ContainerAlreadyExists", err.Error())
		} else {
			s.logger.Error("failed to create container",
				logging.String("account", account),
				logging.String("container", containerName),
				logging.ErrorField(err),
			)
			s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to create container")
		}
		return
	}

	s.logger.Info("container created",
		logging.String("account", account),
		logging.String("container", containerName),
	)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Container %s created successfully", containerName)))
}

// handleDeleteContainer handles DELETE /{account}/{container} to delete a container.
func (s *BlobService) handleDeleteContainer(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")

	if account == "" || containerName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account and container name are required")
		return
	}

	err := s.store.DeleteContainer(r.Context(), account, containerName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			s.writeError(w, http.StatusNotFound, "ContainerNotFound", err.Error())
		} else {
			s.logger.Error("failed to delete container",
				logging.String("account", account),
				logging.String("container", containerName),
				logging.ErrorField(err),
			)
			s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to delete container")
		}
		return
	}

	s.logger.Info("container deleted",
		logging.String("account", account),
		logging.String("container", containerName),
	)
	w.WriteHeader(http.StatusNoContent)
}

// handlePutBlob handles PUT /{account}/{container}/{blobName} to upload a blob.
func (s *BlobService) handlePutBlob(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")
	blobName := chi.URLParam(r, "blobName")

	if account == "" || containerName == "" || blobName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account, container, and blob name are required")
		return
	}

	// Read request body
	content, err := io.ReadAll(r.Body)
	if err != nil {
		s.logger.Error("failed to read request body",
			logging.ErrorField(err),
		)
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Failed to read request body")
		return
	}
	defer r.Body.Close()

	// Get content type from header or default
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Extract metadata from headers (Azure uses x-ms-meta-* prefix)
	metadata := make(map[string]string)
	for key, values := range r.Header {
		if strings.HasPrefix(strings.ToLower(key), "x-ms-meta-") {
			metaKey := strings.TrimPrefix(strings.ToLower(key), "x-ms-meta-")
			if len(values) > 0 {
				metadata[metaKey] = values[0]
			}
		}
	}

	err = s.store.PutBlob(r.Context(), account, containerName, blobName, content, contentType, metadata)
	if err != nil {
		s.logger.Error("failed to put blob",
			logging.String("account", account),
			logging.String("container", containerName),
			logging.String("blob", blobName),
			logging.ErrorField(err),
		)
		s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to upload blob")
		return
	}

	s.logger.Info("blob uploaded",
		logging.String("account", account),
		logging.String("container", containerName),
		logging.String("blob", blobName),
		logging.Int("size", len(content)),
	)
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(fmt.Sprintf("Blob %s uploaded successfully", blobName)))
}

// handleGetBlob handles GET /{account}/{container}/{blobName} to download a blob.
func (s *BlobService) handleGetBlob(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")
	blobName := chi.URLParam(r, "blobName")

	if account == "" || containerName == "" || blobName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account, container, and blob name are required")
		return
	}

	blob, err := s.store.GetBlob(r.Context(), account, containerName, blobName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			s.writeError(w, http.StatusNotFound, "BlobNotFound", err.Error())
		} else {
			s.logger.Error("failed to get blob",
				logging.String("account", account),
				logging.String("container", containerName),
				logging.String("blob", blobName),
				logging.ErrorField(err),
			)
			s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to retrieve blob")
		}
		return
	}

	// Set response headers
	w.Header().Set("Content-Type", blob.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(blob.Size, 10))
	w.Header().Set("Last-Modified", blob.ModifiedAt.Format(http.TimeFormat))

	// Set metadata headers
	for key, value := range blob.Metadata {
		w.Header().Set("x-ms-meta-"+key, value)
	}

	s.logger.Info("blob downloaded",
		logging.String("account", account),
		logging.String("container", containerName),
		logging.String("blob", blobName),
		logging.Int64("size", blob.Size),
	)

	w.WriteHeader(http.StatusOK)
	w.Write(blob.Content)
}

// handleDeleteBlob handles DELETE /{account}/{container}/{blobName} to delete a blob.
func (s *BlobService) handleDeleteBlob(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")
	blobName := chi.URLParam(r, "blobName")

	if account == "" || containerName == "" || blobName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account, container, and blob name are required")
		return
	}

	err := s.store.DeleteBlob(r.Context(), account, containerName, blobName)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			s.writeError(w, http.StatusNotFound, "BlobNotFound", err.Error())
		} else {
			s.logger.Error("failed to delete blob",
				logging.String("account", account),
				logging.String("container", containerName),
				logging.String("blob", blobName),
				logging.ErrorField(err),
			)
			s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to delete blob")
		}
		return
	}

	s.logger.Info("blob deleted",
		logging.String("account", account),
		logging.String("container", containerName),
		logging.String("blob", blobName),
	)
	w.WriteHeader(http.StatusNoContent)
}

// handleListBlobs handles GET /{account}/{container}?list to list blobs in a container.
func (s *BlobService) handleListBlobs(w http.ResponseWriter, r *http.Request) {
	account := chi.URLParam(r, "account")
	containerName := chi.URLParam(r, "container")

	if account == "" || containerName == "" {
		s.writeError(w, http.StatusBadRequest, "InvalidRequest", "Account and container name are required")
		return
	}

	// Parse query parameters
	prefix := r.URL.Query().Get("prefix")
	maxResultsStr := r.URL.Query().Get("maxresults")
	maxResults := 0
	if maxResultsStr != "" {
		if val, err := strconv.Atoi(maxResultsStr); err == nil && val > 0 {
			maxResults = val
		}
	}

	blobs, err := s.store.ListBlobs(r.Context(), account, containerName, prefix, maxResults)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			s.writeError(w, http.StatusNotFound, "ContainerNotFound", err.Error())
		} else {
			s.logger.Error("failed to list blobs",
				logging.String("account", account),
				logging.String("container", containerName),
				logging.ErrorField(err),
			)
			s.writeError(w, http.StatusInternalServerError, "InternalError", "Failed to list blobs")
		}
		return
	}

	result := BlobListResult{
		Blobs:      blobs,
		Prefix:     prefix,
		MaxResults: maxResults,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(result); err != nil {
		s.logger.Error("failed to encode response",
			logging.ErrorField(err),
		)
	}
}

// writeError writes an error response in a consistent format.
// TODO: Match Azure Blob Storage error response format more closely.
func (s *BlobService) writeError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// Ensure BlobService implements the Service interface.
var _ core.Service = (*BlobService)(nil)

