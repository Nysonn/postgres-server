package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"database/sql"
	// "encoding/json"
	// "net/http"

	"github.com/lib/pq"
)

// RegisterModelRequest is the expected JSON payload.
type RegisterModelRequest struct {
	Name   string          `json:"name"`   // e.g. "Product"
	Schema json.RawMessage `json:"schema"` // JSON schema or metadata
}

// ModelInfo represents a row in the models table.
type ModelInfo struct {
	ID        int             `json:"id"`
	Name      string          `json:"name"`
	Schema    json.RawMessage `json:"schema"`
	Version   int             `json:"version"`
	CreatedAt string          `json:"createdAt"`
}

// RegisterModelHandler inserts a new model row.
func RegisterModelHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req RegisterModelRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.Name == "" {
			http.Error(w, "name is required", http.StatusBadRequest)
			return
		}
		if len(req.Schema) == 0 {
			http.Error(w, "schema is required", http.StatusBadRequest)
			return
		}

		// Perform the insert
		const q = `
            INSERT INTO models (name, schema)
            VALUES ($1, $2)
            RETURNING id, version, created_at
        `
		var info ModelInfo
		info.Name = req.Name
		info.Schema = req.Schema

		err := db.QueryRowContext(r.Context(), q, req.Name, req.Schema).
			Scan(&info.ID, &info.Version, &info.CreatedAt)
		if err != nil {
			// handle unique-constraint violation
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
				http.Error(w, "model already exists", http.StatusConflict)
				return
			}
			http.Error(w, "database insert error", http.StatusInternalServerError)
			fmt.Printf("RegisterModel error: %v\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(info)
	}
}

// ReadModelHandler fetches a single model by name.
func ReadModelHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect GET /admin/models?name=Product
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "query parameter 'name' is required", http.StatusBadRequest)
			return
		}

		const q = `
            SELECT id, name, schema, version, created_at
              FROM models
             WHERE name = $1
        `
		var info ModelInfo
		err := db.QueryRowContext(r.Context(), q, name).
			Scan(&info.ID, &info.Name, &info.Schema, &info.Version, &info.CreatedAt)
		if err == sql.ErrNoRows {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "database query error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(info)
	}
}

// ListModelsHandler returns metadata for all registered models.
func ListModelsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect GET /admin/models/list
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		const q = `
            SELECT id, name, schema, version, created_at
              FROM models
             ORDER BY name
        `
		rows, err := db.QueryContext(r.Context(), q)
		if err != nil {
			http.Error(w, "database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []ModelInfo
		for rows.Next() {
			var info ModelInfo
			if err := rows.Scan(&info.ID, &info.Name, &info.Schema, &info.Version, &info.CreatedAt); err != nil {
				http.Error(w, "row scan error", http.StatusInternalServerError)
				return
			}
			list = append(list, info)
		}
		if rows.Err() != nil {
			http.Error(w, "row iteration error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	}
}

// DeleteModelHandler removes a model by name.
func DeleteModelHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Expect DELETE /admin/models?name=Product
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		name := r.URL.Query().Get("name")
		if name == "" {
			http.Error(w, "query parameter 'name' is required", http.StatusBadRequest)
			return
		}

		const q = `DELETE FROM models WHERE name = $1`
		res, err := db.ExecContext(r.Context(), q, name)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
				// foreign key violation
				http.Error(w, "cannot delete: model in use", http.StatusConflict)
				return
			}
			http.Error(w, "database delete error", http.StatusInternalServerError)
			return
		}
		count, _ := res.RowsAffected()
		if count == 0 {
			http.Error(w, "model not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
