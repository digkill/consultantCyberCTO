package handlers

import (
	"fmt"
	"strings"

	"github.com/digkill/consultantCyberCTO/internal/components/db"
	"github.com/digkill/consultantCyberCTO/internal/components/gpt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func formatLoanInfo(record *db.LoanRecord) string {
	confirm := "не указана"
	created := "не указано"

	if record.ConfirmDate.Valid {
		confirm = record.ConfirmDate.Time.Format("2006-01-02 15:04:05")
	}
	if record.CreatedAt.Valid {
		created = record.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}

	return fmt.Sprintf(`✅ Чек найден:

👤 Клиент: %s %s %s
📞 Телефон: %s
📧 Email: %s
💰 Сумма: %.2f ₽
📦 Статус: %d
🧾 Ссылка на чек: %s
🕒 Дата подтверждения: %s
🗓️ Создано: %s`,
		record.LastName, record.FirstName, record.MiddleName,
		record.ClientPhone, record.ClientEmail,
		record.Amount, record.Status,
		record.ReceiptLink, confirm, created,
	)
}

func HandleMessage(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	parsedClients := gpt.ExtractClientsFromAnyMessage(msg.Text)

	if len(parsedClients) == 0 {
		bot.Send(tgbotapi.NewMessage(msg.Chat.ID, "❌ Не удалось извлечь ни одного клиента. Пожалуйста, укажите хотя бы ФИО, телефон или email."))
		return
	}

	for _, client := range parsedClients {
		queryText := strings.Join([]string{
			client.FirstName,
			client.LastName,
			client.MiddleName,
			client.Phone,
			client.Email,
		}, " ")

		records := db.FindClientsFromMessage(queryText)

		if len(records) == 0 {
			reply := fmt.Sprintf("❌ Чек не найден по клиенту %s %s %s", client.LastName, client.FirstName, client.MiddleName)
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, reply))
			continue
		}

		for _, record := range records {
			reply := formatLoanInfo(record)
			bot.Send(tgbotapi.NewMessage(msg.Chat.ID, reply))
		}
	}
}
