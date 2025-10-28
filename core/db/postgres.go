package db

import (
	"database/sql"
	"fmt"
	"nlsql/models"
	"strings"
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
	currentTable := ""
	for rows.Next() {
		var tableName, columnName, dataType string
		if err := rows.Scan(&tableName, &columnName, &dataType); err != nil {
			return "", err
		}

		if tableName != currentTable {
			if currentTable != "" {
				schemaBuilder.WriteString(");\n")
			}
			schemaBuilder.WriteString(fmt.Sprintf("\nCREATE TABLE %s (\n", tableName))
			currentTable = tableName
		}
		schemaBuilder.WriteString(fmt.Sprintf("  %s %s,\n", columnName, dataType))
	}
	if currentTable != "" {
		// Remove trailing comma and add closing parenthesis
		schema := strings.TrimRight(schemaBuilder.String(), ",\n") + "\n);\n"
		return schema, nil
	}
	return "", fmt.Errorf("no tables found in public schema")
}
