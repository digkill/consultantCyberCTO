package gpt

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
)

type clientExtractionResponse struct {
	Clients []ParsedClient `json:"clients"`
}

func ExtractClientsFromAnyMessage(input string) []ParsedClient {
	prompt := `
Ты получишь текст, содержащий любые сообщения, включая ФИО, телефоны и email клиентов.

Твоя задача — найти всех явно указанных клиентов и вернуть их в формате JSON, где каждый клиент содержит:
- first_name
- last_name
- middle_name (если есть)
- phone
- email

Пример ответа:
{
  "clients": [
    {
      "first_name": "Татьяна",
      "last_name": "Арсланова",
      "middle_name": "Игоревна",
      "phone": "+7 (985) 757-83-49",
      "email": "palladagrand@yandex.ru"
    },
    ...
  ]
}

Не пиши никаких комментариев, только JSON.
Текст:
` + input

	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": "Ты AI, который извлекает контактные данные клиентов из любого текста и возвращает JSON-массив клиентов."},
			{"role": "user", "content": prompt},
		},
	}

	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.Getenv("OPENAI_API_KEY"))

	resp, err := http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return nil
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil
	}

	// Чистим json от мусора (иногда GPT добавляет ```)
	jsonRaw := strings.Trim(result.Choices[0].Message.Content, " \n`")

	var extracted clientExtractionResponse
	if err := json.Unmarshal([]byte(jsonRaw), &extracted); err != nil {
		return nil
	}

	return extracted.Clients
}
