package app

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cobrich/ai-tutor-bot/internal/ai"
	"github.com/cobrich/ai-tutor-bot/internal/config"
	"github.com/cobrich/ai-tutor-bot/internal/db"
	"github.com/cobrich/ai-tutor-bot/internal/httpserver"
	"github.com/cobrich/ai-tutor-bot/internal/limiter"
	"github.com/cobrich/ai-tutor-bot/internal/service"
	"github.com/cobrich/ai-tutor-bot/internal/stats"
	"github.com/cobrich/ai-tutor-bot/internal/storage"
	"github.com/cobrich/ai-tutor-bot/internal/telegram"
	"github.com/cobrich/ai-tutor-bot/pkg/logger"
)

func Run() {
	cfg := config.Load()

	logger := logger.New(cfg.AppEnv)

	if cfg.TelegramBotToken == "" {
		logger.Error("TELEGRAM_BOT_TOKEN is required")
		os.Exit(1)
	}

	if cfg.OpenAIAPIKey == "" {
		logger.Error("OPENAI_API_KEY is required")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	pool, err := db.NewPostgres(ctx, cfg.PostgresDSN, logger)
	if err != nil {
		logger.Error("failed to connect postgres", slog.Any("error", err))
		os.Exit(1)
	}
	defer pool.Close()

	// aiClient := ai.NewOpenAIClient(
	// 	cfg.OpenAIAPIKey,
	// 	cfg.OpenAIModel,
	// 	cfg.OpenAITranscribeModel,
	// )

	historyStorage := storage.NewPostgresHistoryStorage(pool, logger)
	rateLimiter := limiter.NewPostgresRateLimiter(pool, 20, logger)
	statsStorage := stats.NewPostgresStatsStorage(pool, logger)

	llm, err := ai.NewLLM(cfg, logger)
	if err != nil {
		logger.Error("failed to get llm", slog.Any("error", err))
		os.Exit(1)
	}

	transcriber, err := ai.NewTranscriber(cfg)
	if err != nil {
		logger.Error("failed to create transcriber", "error", err)
		os.Exit(1)
	}

	tutorService := service.NewTutorService(llm, transcriber, historyStorage, rateLimiter)
	bot, err := telegram.NewBot(
		cfg.TelegramBotToken,
		tutorService,

		logger,
		cfg.AdminTelegramID,
		statsStorage,
	)
	if err != nil {
		logger.Error("failed to create telegram bot", slog.Any("error", err))
		os.Exit(1)
	}

	healthServer := httpserver.New(cfg.HTTPPort, pool, logger)
	go healthServer.Start()

	bot.Start(ctx)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("failed to shutdown http server", "error", err)
	}

	logger.Info("application stopped")
}
