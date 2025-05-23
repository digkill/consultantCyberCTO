# 🤖 Loan Checker Daemon

![Loan Bot](./assets/loan-bot-banner.png)

## Overview

**Loan Checker** is a backend daemon service written in Go. It processes free-form input messages (e.g. from Telegram), extracts client names, phone numbers, or emails using AI (ChatGPT), then checks a MySQL database for relevant loan applications.

If a record is found and marked as *Status = 2* (i.e. **contract signed, funds sent**), it triggers a webhook call to:

```
https://erp.code-class.ru/loan-application/check?order_id=$receipt_uuid
```

## Features

- ✅ Fuzzy matching against names, emails and phone numbers
- 🤖 GPT-based message parsing via `ExtractClientsFromAnyMessage`
- 🌐 Sends webhooks for confirmed loans (Status 2)
- 🔄 Automatically reloads loan info after webhook call
- 🧾 Logs SQL query, webhook attempts and results
- 💥 Fault-tolerant webhook delivery (up to 3 retries)
- 🐧 Designed to run as a daemon (e.g. systemd)

## Environment Variables

```env
MYSQL_DSN=user:password@tcp(host:3306)/codeclass?charset=utf8mb4&parseTime=true&loc=Local
```

## Running

```bash
go build -o loan-daemon cmd/loan-daemon/main.go
./loan-daemon
```

Or as a `systemd` service:

```bash
sudo cp loan-daemon.service /etc/systemd/system/
sudo systemctl enable loan-daemon
sudo systemctl start loan-daemon
```

## Directory Structure

```
.
├── cmd/loan-daemon/main.go     # Entry point
├── loan-daemon.service         # systemd unit
├── README.md                   # You're reading it
```

## License

MIT License © 2025 Mediarise
