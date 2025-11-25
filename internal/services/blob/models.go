package blob

import "time"

// Container represents an Azure Blob Storage container.
// This is a simplified model that captures the essential properties.
// TODO: Add more Azure-specific metadata (ETags, lease state, public access level, etc.)
type Container struct {
	// Name is the unique name of the container within an account.
	Name string

	// CreatedAt is when the container was created.
	CreatedAt time.Time

	// Metadata holds custom key-value pairs associated with the container.
	Metadata map[string]string
}

// Blob represents a blob (file) stored in Azure Blob Storage.
// This is a simplified model that captures essential properties.
// TODO: Add more Azure-specific properties (ETag, Content-MD5, Content-Type, Lease state, etc.)
type Blob struct {
	// Name is the blob name (path) within its container.
	Name string

	// Container is the name of the container this blob belongs to.
	Container string

	// Account is the storage account name (for multi-account support).
	Account string

	// Content is the actual blob data.
	Content []byte

	// ContentType is the MIME type of the blob content.
	ContentType string

	// Size is the size of the blob content in bytes.
	Size int64

	// CreatedAt is when the blob was created.
	CreatedAt time.Time

	// ModifiedAt is when the blob was last modified.
	ModifiedAt time.Time

	// Metadata holds custom key-value pairs associated with the blob.
	Metadata map[string]string
}

// BlobListResult represents the result of listing blobs in a container.
// This mimics Azure's ListBlobs response structure.
type BlobListResult struct {
	// Blobs is the list of blobs in the container.
	Blobs []BlobInfo `json:"Blobs"`

	// Prefix is the prefix used for filtering (if any).
	Prefix string `json:"Prefix,omitempty"`

	// Marker is the continuation token for pagination (if any).
	Marker string `json:"Marker,omitempty"`

	// MaxResults is the maximum number of results requested.
	MaxResults int `json:"MaxResults,omitempty"`
}

// BlobInfo is a lightweight representation of a blob used in list operations.
// It contains only metadata, not the actual content.
type BlobInfo struct {
	Name         string            `json:"Name"`
	ContentType  string            `json:"ContentType,omitempty"`
	Size         int64             `json:"ContentLength"`
	LastModified time.Time         `json:"LastModified"`
	Metadata     map[string]string `json:"Metadata,omitempty"`
}

