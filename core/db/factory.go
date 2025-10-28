package db

import "fmt"

func GetDBAdapter(dbType string) (DBAdapter, error) {
	switch dbType {
	case "postgres":
		return &PostgresAdapter{}, nil
	case "mysql":
		return &MySQLAdapter{}, nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}
