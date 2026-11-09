package auth

import (
	"database/sql"
	"time"
)

// SQLStore implements scs.Store for both SQLite and Postgres.
type SQLStore struct {
	db         *sql.DB
	isPostgres bool
	stopChan   chan struct{}
}

// NewSQLStore creates a new SQLStore and starts a background cleanup worker.
func NewSQLStore(db *sql.DB, isPostgres bool) *SQLStore {
	s := &SQLStore{
		db:         db,
		isPostgres: isPostgres,
		stopChan:   make(chan struct{}),
	}
	go s.startCleanup(5 * time.Minute)
	return s
}

// Find retrieves the session data for the given token from the sessions table.
func (s *SQLStore) Find(token string) ([]byte, bool, error) {
	var data []byte
	var query string
	if s.isPostgres {
		query = "SELECT data FROM sessions WHERE token = $1 AND expiry > CURRENT_TIMESTAMP"
	} else {
		query = "SELECT data FROM sessions WHERE token = ? AND expiry > unixepoch()"
	}

	err := s.db.QueryRow(query, token).Scan(&data)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return data, true, nil
}

// Commit inserts or updates session data and expiry.
func (s *SQLStore) Commit(token string, b []byte, expiry time.Time) error {
	var query string
	if s.isPostgres {
		query = `INSERT INTO sessions (token, data, expiry)
			VALUES ($1, $2, $3)
			ON CONFLICT (token) DO UPDATE SET data = EXCLUDED.data, expiry = EXCLUDED.expiry`
		_, err := s.db.Exec(query, token, b, expiry)
		return err
	}

	// SQLite
	query = `INSERT INTO sessions (token, data, expiry)
		VALUES (?, ?, ?)
		ON CONFLICT (token) DO UPDATE SET data = excluded.data, expiry = excluded.expiry`
	_, err := s.db.Exec(query, token, b, float64(expiry.Unix()))
	return err
}

// Delete removes session token from the database.
func (s *SQLStore) Delete(token string) error {
	var query string
	if s.isPostgres {
		query = "DELETE FROM sessions WHERE token = $1"
	} else {
		query = "DELETE FROM sessions WHERE token = ?"
	}
	_, err := s.db.Exec(query, token)
	return err
}

// startCleanup periodically removes expired sessions.
func (s *SQLStore) startCleanup(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			_ = s.deleteExpired()
		case <-s.stopChan:
			return
		}
	}
}

func (s *SQLStore) deleteExpired() error {
	var query string
	if s.isPostgres {
		query = "DELETE FROM sessions WHERE expiry < CURRENT_TIMESTAMP"
	} else {
		query = "DELETE FROM sessions WHERE expiry < unixepoch()"
	}
	_, err := s.db.Exec(query)
	return err
}

// Close stops the cleanup worker.
func (s *SQLStore) Close() {
	close(s.stopChan)
}
