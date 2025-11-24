package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"nlsql/core/ai"
	core "nlsql/core/db"
	"nlsql/models"
	"sync"
)

var (
	sessionMux  sync.RWMutex
	sessionDB   *sql.DB
	sessionType string
)

// handleConnect receives DB creds, tests connection, and stores it
func HandleConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	var config models.DBConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}

	// Use the Factory to get the right adapter
	adapter, err := core.GetDBAdapter(config.DBType)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: err.Error()})
		return
	}

	// Close existing connection if any
	sessionMux.Lock()
	if sessionDB != nil {
		sessionDB.Close()
	}

	// Open new connection
	connString := adapter.GetConnectionString(config)
	db, err := sql.Open(config.DBType, connString)
	if err != nil {
		sessionMux.Unlock()
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to create connection: " + err.Error()})
		return
	}

	// Test connection
	if err := db.Ping(); err != nil {
		db.Close()
		sessionMux.Unlock()
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Connection failed: " + err.Error()})
		return
	}

	// Store connection and type for this "session"
	sessionDB = db
	sessionType = config.DBType
	sessionMux.Unlock()

	log.Println("Database connection successful for type:", config.DBType)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Connection successful"})
}

// handleGenerateQuery gets schema, calls LLM, and returns proposed SQL
func HandleGenerateQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	var req models.GenerateQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}

	sessionMux.RLock()
	db := sessionDB
	dbType := sessionType
	sessionMux.RUnlock()

	if db == nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Not connected to database"})
		return
	}

	// 1. Get adapter and fetch DDLs/schema
	adapter, err := core.GetDBAdapter(dbType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Internal error: " + err.Error()})
		return
	}

	schema, err := adapter.GetSchema(db)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to get database schema: " + err.Error()})
		return
	}

	// 2. Call LLM to generate query
	sqlQuery, err := ai.GenerateSQLQuery(req.Question, schema, dbType)
	fmt.Println(req.Question,sqlQuery)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to generate SQL query: " + err.Error()})
		return
	}

	// 3. Return proposed query for confirmation
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data:    map[string]string{"sql_query": sqlQuery},
	})
}

// handleExecuteQuery runs the user-confirmed query and returns results
func HandleExecuteQuery(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, models.APIResponse{Success: false, Message: "Method not allowed"})
		return
	}

	var req models.ExecuteQueryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Invalid request body"})
		return
	}

	sessionMux.RLock()
	db := sessionDB
	sessionMux.RUnlock()

	if db == nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{Success: false, Message: "Not connected to database"})
		return
	}

	// Execute query
	rows, err := db.Query(req.Query)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Message: "Query execution failed: " + err.Error()})
		return
	}
	defer rows.Close()

	// --- Generic result scanning ---
	// This logic dynamically scans any query result into headers and rows

	columns, err := rows.Columns()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to get columns: " + err.Error()})
		return
	}

	colTypes, err := rows.ColumnTypes()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to get column types: " + err.Error()})
		return
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	var results []map[string]interface{}
	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Failed to scan row: " + err.Error()})
			return
		}

		rowMap := make(map[string]interface{})
		for i, col := range columns {
			// Handle NULL values
			if values[i] == nil {
				rowMap[col] = nil
				continue
			}

			// Convert bytes to string for readability, otherwise use raw value
			// This is a simplification; a robust app would handle all types
			switch colTypes[i].DatabaseTypeName() {
			case "VARCHAR", "TEXT", "CHAR", "UUID", "TIMESTAMP", "DATETIME", "DATE", "TIME":
				rowMap[col] = fmt.Sprintf("%s", values[i])
			case "BYTEA":
				rowMap[col] = fmt.Sprintf("%s", values[i])
			default:
				rowMap[col] = values[i]
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Message: "Error during row iteration: " + err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"headers": columns,
			"rows":    results, // Sending as list of maps
		},
	})
}

// writeJSON sends a standard JSON response
func writeJSON(w http.ResponseWriter, status int, data models.APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
