package db

import (
	"database/sql"
	"fmt"
	"log"
	"nlsql/models"
	"strings"
)

type MySQLAdapter struct{}

func (m *MySQLAdapter) GetConnectionString(config models.DBConfig) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		config.User, config.Password, config.Host, config.Port, config.DBName)
}

// GetSchema fetches all table names and their column definitions for MySQL
func (m *MySQLAdapter) GetSchema(db *sql.DB) (string, error) {
	// 1. Get all tables
	tableRows, err := db.Query("SHOW TABLES")
	if err != nil {
		return "", err
	}
	defer tableRows.Close()

	var tableNames []string
	for tableRows.Next() {
		var tableName string
		if err := tableRows.Scan(&tableName); err != nil {
			return "", err
		}
		tableNames = append(tableNames, tableName)
	}

	var schemaBuilder strings.Builder
	// 2. For each table, get its structure
	for _, tableName := range tableNames {
		schemaBuilder.WriteString(fmt.Sprintf("\nCREATE TABLE %s (\n", tableName))

		descRows, err := db.Query(fmt.Sprintf("DESCRIBE %s", tableName))
		if err != nil {
			log.Printf("Warning: could not describe table %s: %v", tableName, err)
			continue
		}

		for descRows.Next() {
			var field, typeCol, null, key, defaultCol, extra interface{}
			if err := descRows.Scan(&field, &typeCol, &null, &key, &defaultCol, &extra); err != nil {
				descRows.Close()
				return "", err
			}
			// Format as "column_name type,"
			schemaBuilder.WriteString(fmt.Sprintf("  %s %s,\n", field, typeCol))
		}
		descRows.Close()

		schema := strings.TrimRight(schemaBuilder.String(), ",\n") + "\n);\n"
		schemaBuilder.Reset() // Clear for next loop
		schemaBuilder.WriteString(schema)
	}

	return schemaBuilder.String(), nil
}
