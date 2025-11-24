package db

import (
	"database/sql"
	"fmt"
	"nlsql/models"
	"strings"

	_ "github.com/lib/pq"
)

type PostgresAdapter struct{}

func (p *PostgresAdapter) GetConnectionString(config models.DBConfig) string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName)
}

// GetSchema fetches all table names and their column definitions for Postgres
func (p *PostgresAdapter) GetSchema(db *sql.DB) (string, error) {
	query := `
		SELECT 
			c.table_name, 
			c.column_name, 
			c.data_type 
		FROM 
			information_schema.columns c
		WHERE 
			c.table_schema = 'public'
		ORDER BY 
			c.table_name, c.ordinal_position;
	`
	rows, err := db.Query(query)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	var schemaBuilder strings.Builder
	var currentTable string
	var currentColumns []string

	for rows.Next() {
		var tableName, columnName, dataType string
		if err := rows.Scan(&tableName, &columnName, &dataType); err != nil {
			return "", err
		}

		if tableName != currentTable {
			if currentTable != "" {
				schemaBuilder.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", currentTable))
				schemaBuilder.WriteString(strings.Join(currentColumns, ",\n"))
				schemaBuilder.WriteString("\n);\n\n")
			}
			currentTable = tableName
			currentColumns = []string{}
		}
		currentColumns = append(currentColumns, fmt.Sprintf("  %s %s", columnName, dataType))
	}

	if currentTable != "" {
		schemaBuilder.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", currentTable))
		schemaBuilder.WriteString(strings.Join(currentColumns, ",\n"))
		schemaBuilder.WriteString("\n);\n")
	} else {
		return "", fmt.Errorf("no tables found in public schema")
	}

	return schemaBuilder.String(), nil
}
