package data

import (
	"database/sql"

	"sitecraft/internal/cache"
	"sitecraft/internal/db"
)

// Repository manages database queries and in-memory cache synchronization.
type Repository struct {
	queries *db.Queries
	rawDB   *sql.DB
	cache   *cache.Store
}

var globalRepo *Repository

// SetGlobalRepository configures the singleton repository used by package functions.
func SetGlobalRepository(repo *Repository) {
	globalRepo = repo
}

// GetGlobalRepository returns the current configured repository.
func GetGlobalRepository() *Repository {
	return globalRepo
}

// NewRepository creates a new SQL repository instance.
func NewRepository(database *sql.DB, queries *db.Queries) *Repository {
	return &Repository{
		queries: queries,
		rawDB:   database,
		cache:   cache.GetStore(),
	}
}

// WithCache overrides the cache store for this repository instance.
func (r *Repository) WithCache(c *cache.Store) *Repository {
	r.cache = c
	return r
}

// Cache returns the configured cache store or falls back to default.
func (r *Repository) Cache() *cache.Store {
	if r.cache != nil {
		return r.cache
	}
	return cache.GetStore()
}
