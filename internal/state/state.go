package state

// This package provides state management for the emulator.
// Currently, each service manages its own state (e.g., blob service uses FileBlobStore).
// This package can be extended in the future to provide:
//   - Centralized state management
//   - SQLite-based persistence
//   - In-memory state with optional persistence
//   - State snapshots and restoration
//   - Multi-account state isolation

// TODO: Implement centralized state management if needed.
// For now, services manage their own state through their store interfaces.

