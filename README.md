# 🎓 AI Ustaz

AI-powered Telegram tutor bot built with Go.

AI Ustaz helps students solve problems, understand concepts, and learn through text, images, and voice messages using Google Gemini and OpenAI speech-to-text.

---

## ✨ Features

### 🤖 AI-Powered Learning

* Answers questions using Google Gemini
* Explains solutions step-by-step
* Adapts explanations for students
* Uses subject-specific tutoring prompts

### 📝 Text Questions

Students can send any text question and receive a detailed explanation.

### 🖼 Image Understanding

Supports photos of:

* Math problems
* Homework assignments
* Textbook pages
* Handwritten notes

The bot analyzes the image and explains the solution.

### 🎤 Voice Messages

Students can send voice messages.

Workflow:

1. Voice → Text (OpenAI Transcription)
2. AI analyzes the question
3. AI generates an explanation

### 📚 Subject Detection

Automatic subject classification:

* Mathematics
* Physics
* Chemistry
* English
* Biology
* History
* Geography

### 💬 Conversation Memory

Stores recent conversation history and supports follow-up questions.

### 📊 Analytics

* User statistics
* Request statistics
* Subject analytics
* Feedback tracking

### 👍 Feedback System

Users can rate answers:

* 👍 Helpful
* 👎 Not Helpful

### 🛡 Rate Limiting

Daily request limits protect the service and control AI costs.

### ⚙ Production Features

* PostgreSQL
* Docker
* Railway Deployment
* Prometheus Metrics
* Structured Logging
* Health Checks
* Graceful Shutdown
* Goose Migrations

---

## 📸 Screenshots

### Text Question

![Text Question](docs/screenshots/chat/text.png)

### Image Question

![Image Question](docs/screenshots/chat/image.png)

### Voice Question

![Voice Question](docs/screenshots/chat/voice.png)

### User Profile

![Profile](docs/screenshots/profile.png)

### Statistics

![Statistics](docs/screenshots/stats.png)

---

## 🏗 Architecture

```text
Telegram User
      │
      ▼
Telegram Bot
      │
      ▼
Handler Layer
      │
      ▼
Tutor Service
 ├── Gemini AI
 ├── OpenAI Transcription
 ├── Subject Detection
 ├── Conversation History
 ├── Rate Limiter
 └── Statistics
      │
      ▼
PostgreSQL
```

---

## 🛠 Tech Stack

### Backend

* Go 1.26

### Database

* PostgreSQL
* pgx
* Goose

### AI

* Google Gemini 2.5 Flash
* OpenAI Transcription API

### Infrastructure

* Docker
* Railway
* Prometheus

### Telegram

* go-telegram-bot-api

---

## 📂 Project Structure

```text
cmd/
├── bot/
└── migrate/

internal/
├── ai/
├── config/
├── history/
├── httpserver/
├── limiter/
├── metrics/
├── stats/
├── storage/
└── telegram/

migrations/
└── sql/

docs/
└── screenshots/
```

---

## 📋 Commands

```text
/start    - Start using AI Ustaz
/help     - Show bot capabilities
/limit    - Show remaining daily requests
/profile  - User profile and statistics
/stats    - Global bot statistics
/reset    - Clear conversation history
```

---

## ⚙ Environment Variables

```env
APP_ENV=local

HTTP_PORT=8082

DATABASE_URL=

TELEGRAM_BOT_TOKEN=
ADMIN_TELEGRAM_ID=

AI_PROVIDER=gemini

GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash

OPENAI_API_KEY=
OPENAI_TRANSCRIBE_MODEL=gpt-4o-mini-transcribe
```

---

## 🚀 Local Development

Start PostgreSQL:

```bash
docker compose up -d
```

Run migrations:

```bash
go run ./cmd/migrate up
```

Run bot:

```bash
go run ./cmd/bot
```

---

## ❤️ Health Check

```http
GET /health# 🎓 AI Ustaz

AI-powered Telegram tutor bot built with Go.

AI Ustaz helps students solve problems, understand concepts, and learn through text, images, and voice messages using Google Gemini and OpenAI speech-to-text.

---

## ✨ Features

### 🤖 AI-Powered Learning

* Answers questions using Google Gemini
* Explains solutions step-by-step
* Adapts explanations for students
* Uses subject-specific tutoring prompts

### 📝 Text Questions

Students can send any text question and receive a detailed explanation.

### 🖼 Image Understanding

Supports photos of:

* Math problems
* Homework assignments
* Textbook pages
* Handwritten notes

The bot analyzes the image and explains the solution.

### 🎤 Voice Messages

Students can send voice messages.

Workflow:

1. Voice → Text (OpenAI Transcription)
2. AI analyzes the question
3. AI generates an explanation

### 📚 Subject Detection

Automatic subject classification:

* Mathematics
* Physics
* Chemistry
* English
* Biology
* History
* Geography

### 💬 Conversation Memory

Stores recent conversation history and supports follow-up questions.

### 📊 Analytics

* User statistics
* Request statistics
* Subject analytics
* Feedback tracking

### 👍 Feedback System

Users can rate answers:

* 👍 Helpful
* 👎 Not Helpful

### 🛡 Rate Limiting

Daily request limits protect the service and control AI costs.

### ⚙ Production Features

* PostgreSQL
* Docker
* Railway Deployment
* Prometheus Metrics
* Structured Logging
* Health Checks
* Graceful Shutdown
* Goose Migrations

---

## 📸 Screenshots

### Text Question

![Text Question](docs/screenshots/chat/text.png)

### Image Question

![Image Question](docs/screenshots/chat/image.png)

### Voice Question

![Voice Question](docs/screenshots/chat/voice.png)

### User Profile

![Profile](docs/screenshots/profile.png)

### Statistics

![Statistics](docs/screenshots/stats.png)

---

## 🏗 Architecture

```text
Telegram User
      │
      ▼
Telegram Bot
      │
      ▼
Handler Layer
      │
      ▼
Tutor Service
 ├── Gemini AI
 ├── OpenAI Transcription
 ├── Subject Detection
 ├── Conversation History
 ├── Rate Limiter
 └── Statistics
      │
      ▼
PostgreSQL
```

---

## 🛠 Tech Stack

### Backend

* Go 1.26

### Database

* PostgreSQL
* pgx
* Goose

### AI

* Google Gemini 2.5 Flash
* OpenAI Transcription API

### Infrastructure

* Docker
* Railway
* Prometheus

### Telegram

* go-telegram-bot-api

---

## 📂 Project Structure

```text
cmd/
├── bot/
└── migrate/

internal/
├── ai/
├── config/
├── history/
├── httpserver/
├── limiter/
├── metrics/
├── stats/
├── storage/
└── telegram/

migrations/
└── sql/

docs/
└── screenshots/
```

---

## 📋 Commands

```text
/start    - Start using AI Ustaz
/help     - Show bot capabilities
/limit    - Show remaining daily requests
/profile  - User profile and statistics
/stats    - Global bot statistics
/reset    - Clear conversation history
```

---

## ⚙ Environment Variables

```env
APP_ENV=local

HTTP_PORT=8082

DATABASE_URL=

TELEGRAM_BOT_TOKEN=
ADMIN_TELEGRAM_ID=

AI_PROVIDER=gemini

GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash

OPENAI_API_KEY=
OPENAI_TRANSCRIBE_MODEL=gpt-4o-mini-transcribe
```

---

## 🚀 Local Development

Start PostgreSQL:

```bash
docker compose up -d
```

Run migrations:

```bash
go run ./cmd/migrate up
```

Run bot:

```bash
go run ./cmd/bot
```

---

## ❤️ Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok",
  "db": "ok"
}
```

---

## 📈 Metrics

```http
GET /metrics
```

Prometheus metrics:

* telegram_requests_total
* ai_requests_total
* ai_errors_total

---

## 🔮 Future Improvements

* Exam preparation mode (UNT, IELTS, SAT)
* Personalized learning plans
* Parent dashboard
* AI-generated quizzes
* Subject progress tracking
* Admin panel

---

## 👨‍💻 Author

Bekzat Tursun
# 🎓 AI Ustaz

AI-powered Telegram tutor bot built with Go.

AI Ustaz helps students solve problems, understand concepts, and learn through text, images, and voice messages using Google Gemini and OpenAI speech-to-text.

---

## ✨ Features

### 🤖 AI-Powered Learning

* Answers questions using Google Gemini
* Explains solutions step-by-step
* Adapts explanations for students
* Uses subject-specific tutoring prompts

### 📝 Text Questions

Students can send any text question and receive a detailed explanation.

### 🖼 Image Understanding

Supports photos of:

* Math problems
* Homework assignments
* Textbook pages
* Handwritten notes

The bot analyzes the image and explains the solution.

### 🎤 Voice Messages

Students can send voice messages.

Workflow:

1. Voice → Text (OpenAI Transcription)
2. AI analyzes the question
3. AI generates an explanation

### 📚 Subject Detection

Automatic subject classification:

* Mathematics
* Physics
* Chemistry
* English
* Biology
* History
* Geography

### 💬 Conversation Memory

Stores recent conversation history and supports follow-up questions.

### 📊 Analytics

* User statistics
* Request statistics
* Subject analytics
* Feedback tracking

### 👍 Feedback System

Users can rate answers:

* 👍 Helpful
* 👎 Not Helpful

### 🛡 Rate Limiting

Daily request limits protect the service and control AI costs.

### ⚙ Production Features

* PostgreSQL
* Docker
* Railway Deployment
* Prometheus Metrics
* Structured Logging
* Health Checks
* Graceful Shutdown
* Goose Migrations

---

## 📸 Screenshots

### Text Question

![Text Question](docs/screenshots/chat/text.png)

### Image Question

![Image Question](docs/screenshots/chat/image.png)

### Voice Question

![Voice Question](docs/screenshots/chat/voice.png)

### User Profile

![Profile](docs/screenshots/profile.png)

### Statistics

![Statistics](docs/screenshots/stats.png)

---

## 🏗 Architecture

```text
Telegram User
      │
      ▼
Telegram Bot
      │
      ▼
Handler Layer
      │
      ▼
Tutor Service
 ├── Gemini AI
 ├── OpenAI Transcription
 ├── Subject Detection
 ├── Conversation History
 ├── Rate Limiter
 └── Statistics
      │
      ▼
PostgreSQL
```

---

## 🛠 Tech Stack

### Backend

* Go 1.26

### Database

* PostgreSQL
* pgx
* Goose

### AI

* Google Gemini 2.5 Flash
* OpenAI Transcription API

### Infrastructure

* Docker
* Railway
* Prometheus

### Telegram

* go-telegram-bot-api

---

## 📂 Project Structure

```text
cmd/
├── bot/
└── migrate/

internal/
├── ai/
├── config/
├── history/
├── httpserver/
├── limiter/
├── metrics/
├── stats/
├── storage/
└── telegram/

migrations/
└── sql/

docs/
└── screenshots/
```

---

## 📋 Commands

```text
/start    - Start using AI Ustaz
/help     - Show bot capabilities
/limit    - Show remaining daily requests
/profile  - User profile and statistics
/stats    - Global bot statistics
/reset    - Clear conversation history
```

---

## ⚙ Environment Variables

```env
APP_ENV=local

HTTP_PORT=8082

DATABASE_URL=

TELEGRAM_BOT_TOKEN=
ADMIN_TELEGRAM_ID=

AI_PROVIDER=gemini

GEMINI_API_KEY=
GEMINI_MODEL=gemini-2.5-flash

OPENAI_API_KEY=
OPENAI_TRANSCRIBE_MODEL=gpt-4o-mini-transcribe
```

---

## 🚀 Local Development

Start PostgreSQL:

```bash
docker compose up -d
```

Run migrations:

```bash
go run ./cmd/migrate up
```

Run bot:

```bash
go run ./cmd/bot
```

---

## ❤️ Health Check

```http
GET /health
```

Response:

```json
{
  "status": "ok",
  "db": "ok"
}
```

---

## 📈 Metrics

```http
GET /metrics
```

Prometheus metrics:

* telegram_requests_total
* ai_requests_total
* ai_errors_total

---

## 🔮 Future Improvements

* Exam preparation mode (UNT, IELTS, SAT)
* Personalized learning plans
* Parent dashboard
* AI-generated quizzes
* Subject progress tracking
* Admin panel

---

## 👨‍💻 Author

Bekzat Tursun

Backend Developer (Go)

GitHub: https://github.com/cobrich

Backend Developer (Go)

GitHub: https://github.com/cobrich

```

Response:

```json
{
  "status": "ok",
  "db": "ok"
}
```

---

## 📈 Metrics

```http
GET /metrics
```

Prometheus metrics:

* telegram_requests_total
* ai_requests_total
* ai_errors_total

---

## 🔮 Future Improvements

* Exam preparation mode (UNT, IELTS, SAT)
* Personalized learning plans
* Parent dashboard
* AI-generated quizzes
* Subject progress tracking
* Admin panel

---

## 👨‍💻 Author

Bekzat Tursun

Backend Developer (Go)

GitHub: https://github.com/cobrich
