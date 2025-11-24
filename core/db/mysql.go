package db

import (
	"database/sql"
	"fmt"
	"log"
	"nlsql/models"
	"strings"
	_ "github.com/go-sql-driver/mysql"
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
		schemaBuilder.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", tableName))

		descRows, err := db.Query(fmt.Sprintf("DESCRIBE %s", tableName))
		if err != nil {
			log.Printf("Warning: could not describe table %s: %v", tableName, err)
			continue
		}

		var columns []string
		for descRows.Next() {
			var field, typeCol, null, key, defaultCol, extra interface{}
			if err := descRows.Scan(&field, &typeCol, &null, &key, &defaultCol, &extra); err != nil {
				descRows.Close()
				return "", err
			}
			// Format as "column_name type"
			columns = append(columns, fmt.Sprintf("  %s %s", field, typeCol))
		}
		descRows.Close()

		schemaBuilder.WriteString(strings.Join(columns, ",\n"))
		schemaBuilder.WriteString("\n);\n\n")
	}

	return schemaBuilder.String(), nil
}
