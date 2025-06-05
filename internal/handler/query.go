package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// QueryRequest represents the incoming JSON to /query.
type QueryRequest struct {
	Model      string   `json:"model"`      // e.g. "items"
	Fields     []string `json:"fields"`     // which columns to SELECT
	QueryText  string   `json:"queryText"`  // free text to search for
	MaxResults int      `json:"maxResults"` // LIMIT
	Fuzzy      bool     `json:"fuzzy"`      // enable fuzzy matching (optional)
}

// MakeQueryHandler returns an http.HandlerFunc that performs a safe text search
// on only the text columns, and returns all requested fields.
func MakeQueryHandler(db *sql.DB, allowed map[string][]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only POST is allowed
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Decode JSON body
		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		// Basic validation
		if req.Model == "" {
			http.Error(w, "'model' is required", http.StatusBadRequest)
			return
		}
		if len(req.Fields) == 0 {
			http.Error(w, "'fields' array cannot be empty", http.StatusBadRequest)
			return
		}
		if req.QueryText == "" {
			http.Error(w, "'queryText' is required", http.StatusBadRequest)
			return
		}
		if req.MaxResults <= 0 || req.MaxResults > 100 {
			req.MaxResults = 10 // default fallback
		}

		// Check if this model is allowed, and which columns we can SELECT
		allowedFields, ok := allowed[req.Model]
		if !ok {
			http.Error(w, "model not found", http.StatusBadRequest)
			return
		}

		// Ensure all requested fields are in allowedFields
		allowedSet := make(map[string]bool, len(allowedFields))
		for _, f := range allowedFields {
			allowedSet[f] = true
		}
		for _, f := range req.Fields {
			if !allowedSet[f] {
				http.Error(w, fmt.Sprintf("field '%s' not allowed for model '%s'", f, req.Model), http.StatusBadRequest)
				return
			}
		}

		// Build SELECT clause
		selectCols := strings.Join(req.Fields, ", ")

		// === Build WHERE clause using only "content" terms ===
		searchCols := []string{"name", "category"}
		terms := prepareSearchTerms(req.QueryText)

		var whereClauses []string
		var args []interface{}
		argIdx := 1

		for _, t := range terms {
			likePattern := "%" + t + "%"
			var parts []string
			for _, col := range searchCols {
				parts = append(parts, fmt.Sprintf("%s ILIKE $%d", col, argIdx))
				args = append(args, likePattern)
				argIdx++
			}
			whereClauses = append(whereClauses, "("+strings.Join(parts, " OR ")+")")
		}

		whereClause := ""
		if len(whereClauses) > 0 {
			whereClause = strings.Join(whereClauses, " AND ")
		} else {
			whereClause = "1=1"
		}

		// === Simplified ORDER BY to prioritize exact substring matches ===
		exactPattern := "%" + strings.ToLower(req.QueryText) + "%"
		orderClause := fmt.Sprintf(`
            CASE
                WHEN LOWER(name) LIKE LOWER($%d) THEN 1
                WHEN LOWER(category) LIKE LOWER($%d) THEN 2
                ELSE 3
            END, name ASC
        `, argIdx, argIdx+1)
		args = append(args, exactPattern, exactPattern)
		argIdx += 2

		// Append LIMIT argument
		args = append(args, req.MaxResults)
		limitIdx := argIdx

		// Final SQL
		queryStr := fmt.Sprintf(`
            SELECT %s
              FROM %s
             WHERE %s
          ORDER BY %s
             LIMIT $%d
        `, selectCols, req.Model, whereClause, orderClause, limitIdx)

		// Execute the query
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		rows, err := db.QueryContext(ctx, queryStr, args...)
		if err != nil {
			http.Error(w, "database query error", http.StatusInternalServerError)
			fmt.Printf("MakeQueryHandler error: %v\nQuery: %s\nArgs: %v\n", err, queryStr, args)
			return
		}
		defer rows.Close()

		// Fetch all rows into []map[string]interface{}
		results := make([]map[string]interface{}, 0)
		cols, _ := rows.Columns()
		for rows.Next() {
			values := make([]interface{}, len(cols))
			for i := range values {
				var tmp interface{}
				values[i] = &tmp
			}
			if err := rows.Scan(values...); err != nil {
				http.Error(w, "row scan error", http.StatusInternalServerError)
				return
			}
			rowMap := make(map[string]interface{}, len(cols))
			for i, colName := range cols {
				valPtr := values[i].(*interface{})
				rowMap[colName] = *valPtr
			}
			results = append(results, rowMap)
		}
		if err := rows.Err(); err != nil {
			http.Error(w, "row iteration error", http.StatusInternalServerError)
			return
		}

		// Return JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

// prepareSearchTerms cleans and splits the search query into individual terms.
// We expanded the stop‐word list to include common “ordering” verbs.
func prepareSearchTerms(queryText string) []string {
	cleaned := strings.ToLower(strings.TrimSpace(queryText))
	rawTerms := strings.FieldsFunc(cleaned, func(c rune) bool {
		return c == ' ' || c == ',' || c == ';' || c == '\t' || c == '\n'
	})

	stopWords := map[string]bool{
		"the":    true,
		"a":      true,
		"an":     true,
		"and":    true,
		"or":     true,
		"but":    true,
		"in":     true,
		"on":     true,
		"at":     true,
		"to":     true,
		"for":    true,
		"of":     true,
		"with":   true,
		"by":     true,
		"is":     true,
		"are":    true,
		"was":    true,
		"were":   true,
		"be":     true,
		"been":   true,
		"have":   true,
		"has":    true,
		"had":    true,
		"do":     true,
		"does":   true,
		"did":    true,
		"will":   true,
		"would":  true,
		"could":  true,
		"should": true,
		"want":   true,
		"like":   true,
		"buy":    true,
		"need":   true,
		"please": true,
		"give":   true,
	}

	var filtered []string
	for _, term := range rawTerms {
		if len(term) < 2 {
			continue
		}
		if stopWords[term] {
			continue
		}
		filtered = append(filtered, term)
	}

	if len(filtered) == 0 {
		return []string{cleaned}
	}
	return filtered
}
