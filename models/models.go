package models

type DBConfig struct {
	DBType   string `json:"db_type"` // "postgres" or "mysql"
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbname"`
}

type GenerateQueryRequest struct {
	Question string `json:"question"`
}

type ExecuteQueryRequest struct {
	Query string `json:"query"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}