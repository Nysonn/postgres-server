package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// QueryRequest represents a generic search request against any model/table.
type QueryRequest struct {
	Model      string   `json:"model"`      // e.g., "items", "users", "orders"
	QueryText  string   `json:"queryText"`  // Text to search within specified fields
	Fields     []string `json:"fields"`     // Columns to apply the search on
	MaxResults int      `json:"maxResults"` // Limit of results
}

// Record is a generic map representation of a DB row.
type Record map[string]interface{}

// MakeQueryHandler returns an http.HandlerFunc that executes a generic table search.
func MakeQueryHandler(db *sql.DB, allowed map[string][]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Validate model
		cols, ok := allowed[req.Model]
		if !ok {
			http.Error(w, "model not allowed", http.StatusBadRequest)
			return
		}

		// Validate fields
		var fieldsToSearch []string
		for _, f := range req.Fields {
			for _, ac := range cols {
				if f == ac {
					fieldsToSearch = append(fieldsToSearch, f)
				}
			}
		}
		if len(fieldsToSearch) == 0 {
			http.Error(w, "no valid fields to search", http.StatusBadRequest)
			return
		}

		if len(req.QueryText) < 1 {
			http.Error(w, "queryText must be non-empty", http.StatusBadRequest)
			return
		}
		if req.MaxResults <= 0 || req.MaxResults > 100 {
			req.MaxResults = 10
		}

		// Build WHERE clauses dynamically
		var clauses []string
		for _, f := range fieldsToSearch {
			// ILIKE for case-insensitive match
			clauses = append(clauses, fmt.Sprintf("%s ILIKE '%%' || $1 || '%%'", f))
		}
		where := strings.Join(clauses, " OR ")

		query := fmt.Sprintf(
			"SELECT * FROM %s WHERE %s LIMIT $2",
			req.Model, where,
		)

		rows, err := db.QueryContext(r.Context(), query, req.QueryText, req.MaxResults)
		if err != nil {
			http.Error(w, "database query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Dynamically read columns
		colsNames, err := rows.Columns()
		if err != nil {
			http.Error(w, "error fetching columns", http.StatusInternalServerError)
			return
		}

		results := make([]Record, 0)
		for rows.Next() {
			// Create a slice of pointers to empty interfaces
			values := make([]interface{}, len(colsNames))
			scanArgs := make([]interface{}, len(colsNames))
			for i := range values {
				scanArgs[i] = &values[i]
			}

			if err := rows.Scan(scanArgs...); err != nil {
				http.Error(w, "error scanning row", http.StatusInternalServerError)
				return
			}

			rec := make(Record)
			for i, col := range colsNames {
				rec[col] = values[i]
			}
			results = append(results, rec)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "row iteration error", http.StatusInternalServerError)
			return
		}

		// Write JSON response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": results,
			"count":   len(results),
		})
	}
}
