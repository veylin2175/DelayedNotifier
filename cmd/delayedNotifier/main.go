package main

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/lib/logger/handlers/slogpretty"
	"DelayedNotifier/internal/lib/logger/sl"
	"DelayedNotifier/internal/storage/postgres"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info("Starting delayed notifier", slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	storage, err := postgres.InitDB(cfg)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("server stopped", slog.String("signal", sign.String()))

	if err = storage.Close(); err != nil {
		log.Error("failed to close database", slog.String("error", err.Error()))
	}

	log.Info("postgres connection closed")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{
			Level: slog.LevelDebug,
		},
	}

	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
