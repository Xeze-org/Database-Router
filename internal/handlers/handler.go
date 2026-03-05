// Package handlers contains all HTTP handler methods and shared middleware.
package handlers

import "db-router/internal/db"

// Handler wraps the database manager so all HTTP handlers can access
// database connections through a single receiver.
type Handler struct {
	DB *db.Manager
}

// New creates a Handler from an active database manager.
func New(dm *db.Manager) *Handler {
	return &Handler{DB: dm}
}
