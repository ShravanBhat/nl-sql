package ai

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"nlsql/models"
	"os"
	"strings"
	"time"
)

func GenerateSQLQuery(question string, schema string, dbType string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: GEMINI_API_KEY is not set. Using dummy query.")
		return "SELECT 'Error: API Key not set', 'Please add your Gemini API key in main.go' AS 'Status'", nil
	}

	apiUrl := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash-preview-09-2025:generateContent?key=" + apiKey

	systemPrompt := fmt.Sprintf(`
You are an expert %s SQL assistant. Given the following database schema, write a single,
executable SQL query to answer the user's question.
Only output the raw SQL query and nothing else. Do not wrap it in markdown or add any explanation.

--- SCHEMA ---
%s
--- END SCHEMA ---
`, dbType, schema)

	payload := models.GeminiRequestPayload{
		SystemInstruction: &models.GeminiInstruction{
			Parts: []models.GeminiPart{{Text: systemPrompt}},
		},
		Contents: []models.GeminiContent{
			{Parts: []models.GeminiPart{{Text: question}}},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("error marshalling payload: %w", err)
	}

	req, err := http.NewRequest("POST", apiUrl, strings.NewReader(string(payloadBytes)))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request to Gemini: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var geminiResp models.GeminiResponsePayload
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		sqlQuery := geminiResp.Candidates[0].Content.Parts[0].Text
		// Clean up potential markdown
		sqlQuery = strings.TrimSpace(sqlQuery)
		sqlQuery = strings.TrimPrefix(sqlQuery, "```sql")
		sqlQuery = strings.TrimPrefix(sqlQuery, "```")
		sqlQuery = strings.TrimSuffix(sqlQuery, "```")
		return sqlQuery, nil
	}

	return "", fmt.Errorf("no SQL query generated, or unexpected response format: %s", string(body))
}
