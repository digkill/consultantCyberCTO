package gpt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
)

type ParsedClient struct {
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
	MiddleName string `json:"middle_name"`
	Phone      string `json:"phone"`
	Email      string `json:"email"`
}

func ExtractClientsBatch(input string) []ParsedClient {
	prompt := `
Ты получишь текст с данными нескольких клиентов, среди прочего текста. 
Извлеки из него список клиентов в JSON-формате. На каждого клиента: 
ФИО (разделённое на фамилию, имя, отчество), телефон, email. 
Пример ответа:

{
  "clients": [
    {
      "first_name": "Имя",
      "last_name": "Фамилия",
      "middle_name": "Отчество",
      "phone": "Телефон",
      "email": "Email"
    },
    ...
  ]
}

Текст:
` + input

	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": "Ты — AI, который парсит список клиентов из свободного текста и возвращает JSON."},
			{"role": "user", "content": prompt},
		},
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	var gptResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}
	}
	json.NewDecoder(resp.Body).Decode(&gptResp)

	var result struct {
		Clients []ParsedClient `json:"clients"`
	}
	json.Unmarshal([]byte(gptResp.Choices[0].Message.Content), &result)

	return result.Clients
}
