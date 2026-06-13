package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	TelegramBotToken      string
	OpenAIAPIKey          string
	OpenAIModel           string
	OpenAITranscribeModel string
	AppEnv                string
	AdminTelegramID       int64
	PostgresDSN           string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env, using system env var-s")
	}

	return Config{
		TelegramBotToken:      getEnv("TELEGRAM_BOT_TOKEN", ""),
		OpenAIAPIKey:          getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:           getEnv("OPENAI_MODEL", "gpt-4.1-mini"),
		OpenAITranscribeModel: getEnv("OPENAI_TRANSCRIBE_MODEL", "gpt-4o-mini-transcribe"),
		AppEnv:                getEnv("APP_ENV", "local"),
		AdminTelegramID:       getEnvInt64("ADMIN_TELEGRAM_ID", 0),
		PostgresDSN:           buildPostgresDSN(),
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}

	return parsed
}

func buildPostgresDSN() string {
	host := getEnv("POSTGRES_HOST", "localhost")
	port := getEnv("POSTGRES_PORT", "5432")
	db := getEnv("POSTGRES_DB", "ai_tutor_bot")
	user := getEnv("POSTGRES_USER", "postgres")
	password := getEnv("POSTGRES_PASSWORD", "postgres")
	sslmode := getEnv("POSTGRES_SSLMODE", "disable")

	return "postgres://" + user + ":" + password + "@" + host + ":" + port + "/" + db + "?sslmode=" + sslmode
}
