package handler

import (
	"encoding/json"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
)

// MigrationsStatus holds migration metadata.
type MigrationsStatus struct {
	Version uint `json:"version"`
	Dirty   bool `json:"dirty"`
}

// MakeMigrationsStatusHandler returns an endpoint that reports current migration state.
func MakeMigrationsStatusHandler(m *migrate.Migrate) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		version, dirty, err := m.Version()
		if err == migrate.ErrNilVersion {
			// No migrations have been applied yet.
			version = 0
			dirty = false
		} else if err != nil {
			http.Error(w, "failed to fetch migration version", http.StatusInternalServerError)
			return
		}

		status := MigrationsStatus{
			Version: version,
			Dirty:   dirty,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(status)
	}
}
