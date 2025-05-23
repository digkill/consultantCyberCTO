package db

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/digkill/consultantCyberCTO/internal/components/gpt"
	_ "github.com/go-sql-driver/mysql"
)

type LoanRecord struct {
	FirstName   sql.NullString
	LastName    sql.NullString
	MiddleName  sql.NullString
	ClientPhone sql.NullString
	ClientEmail sql.NullString
	Amount      float64
	Status      int
	ReceiptLink sql.NullString
	ConfirmDate sql.NullTime
	CreatedAt   sql.NullTime
	ReceiptUUID sql.NullString
}

type Payload struct {
	Values struct {
		Contact struct {
			Phone string `json:"phone"`
			Email string `json:"email"`
			Fio   struct {
				FirstName  string `json:"firstName"`
				LastName   string `json:"lastName"`
				MiddleName string `json:"middleName"`
			} `json:"fio"`
		} `json:"contact"`
	} `json:"values"`
}

func FindClientsFromMessage(input string) []*LoanRecord {
	dbConn, err := sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	parsedClients := gpt.ExtractClientsFromAnyMessage(input)
	var results []*LoanRecord

	for _, client := range parsedClients {
		log.Printf("Поиск по: firstName=%s, lastName=%s, email=%s\n", client.FirstName, client.LastName, client.Email)

		var where []string
		var args []interface{}

		fullName := strings.TrimSpace(client.FirstName + " " + client.LastName + " " + client.MiddleName)
		if fullName != "" {
			where = append(where, "LOWER(CONCAT_WS(' ', first_name, last_name, middle_name)) LIKE LOWER(?)")
			args = append(args, "%"+strings.ToLower(fullName)+"%")
		}
		if client.Phone != "" {
			where = append(where, "REPLACE(client_phone, ' ', '') LIKE REPLACE(?, ' ', '')")
			args = append(args, "%"+strings.ReplaceAll(client.Phone, " ", "")+"%")
		}
		if client.Email != "" {
			where = append(where, "LOWER(client_email) LIKE LOWER(?)")
			args = append(args, "%"+strings.ToLower(strings.TrimSpace(client.Email))+"%")
		}

		if len(where) == 0 {
			continue
		}

		query := `SELECT first_name, last_name, middle_name, client_phone, client_email,
       amount, status, receipt_link, confirm_date, created_at, receipt_uuid
FROM loan_application
WHERE ` + strings.Join(where, " OR ") + `
ORDER BY created_at DESC
LIMIT 1`

		logQuery := fmt.Sprintf("\nSQL: %s\nBINDINGS: %#v\n", query, args)
		log.Println(logQuery)

		row := dbConn.QueryRow(query, args...)

		var record LoanRecord
		err := row.Scan(
			&record.FirstName, &record.LastName, &record.MiddleName,
			&record.ClientPhone, &record.ClientEmail,
			&record.Amount, &record.Status, &record.ReceiptLink,
			&record.ConfirmDate, &record.CreatedAt,
			&record.ReceiptUUID,
		)
		if err != nil {
			log.Printf("❌ Ошибка при Scan: %v", err)
			continue
		}

		// Если статус == 2, отправляем вебхук
		if record.Status == 2 && record.ReceiptUUID.Valid {
			webhookURL := fmt.Sprintf("https://erp.code-class.ru/loan-application/check?order_id=%s", record.ReceiptUUID.String)

			success := false
			for attempt := 1; attempt <= 3; attempt++ {
				resp, err := http.Get(webhookURL)
				if err != nil {
					log.Printf("❌ Попытка %d — ошибка при отправке вебхука: %v", attempt, err)
					continue
				}
				bodyBytes, _ := io.ReadAll(resp.Body)
				log.Printf("✅ Вебхук отправлен: %s [HTTP %d] — Ответ: %s", webhookURL, resp.StatusCode, string(bodyBytes))
				resp.Body.Close()
				success = true
				break
			}

			if !success {
				log.Printf("❗ Все попытки отправки вебхука не удались: %s", webhookURL)
			}
		}

		// Повторно получаем актуальные данные из БД после отправки вебхука
		row = dbConn.QueryRow(query, args...)
		err = row.Scan(
			&record.FirstName, &record.LastName, &record.MiddleName,
			&record.ClientPhone, &record.ClientEmail,
			&record.Amount, &record.Status, &record.ReceiptLink,
			&record.ConfirmDate, &record.CreatedAt,
			&record.ReceiptUUID,
		)
		if err != nil {
			log.Printf("❌ Ошибка при повторном Scan после вебхука: %v", err)
			continue
		}

		results = append(results, &record)
	}

	return results
}
