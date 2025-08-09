package main

import (
	"DelayedNotifier/internal/config"
	"DelayedNotifier/internal/http-server/handlers/notify/createNotify"
	"DelayedNotifier/internal/http-server/handlers/notify/deleteNotify"
	"DelayedNotifier/internal/http-server/handlers/notify/getStatus"
	"DelayedNotifier/internal/http-server/middleware/mwlogger"
	"DelayedNotifier/internal/lib/logger/handlers/slogpretty"
	"DelayedNotifier/internal/lib/logger/sl"
	"DelayedNotifier/internal/rabbitMQ/broker"
	"DelayedNotifier/internal/service"
	"DelayedNotifier/internal/storage/postgres"
	"DelayedNotifier/internal/telegram/notifier"
	"DelayedNotifier/internal/worker"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
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

	rabbitURL := fmt.Sprintf("amqp://%s:%s@%s:%d/", cfg.Rabbit.User, cfg.Rabbit.Password, cfg.Rabbit.Host, cfg.Rabbit.Port)
	mqBroker, err := broker.New(rabbitURL)
	if err != nil {
		log.Error("failed to init RabbitMQ broker", sl.Err(err))
		os.Exit(1)
	}

	if err = mqBroker.DeclareQueue(cfg.Rabbit.QueueName); err != nil {
		log.Error("failed to declare RabbitMQ queue", sl.Err(err))
		os.Exit(1)
	}

	tgNotifier, err := notifier.New(cfg.TGToken)
	if err != nil {
		log.Error("failed to init Telegram notifier", sl.Err(err))
		os.Exit(1)
	}

	appService := service.New(storage, mqBroker, cfg, tgNotifier)

	msgs, err := mqBroker.Consume(cfg.Rabbit.QueueName)
	if err != nil {
		log.Error("failed to consume from RabbitMQ", sl.Err(err))
		os.Exit(1)
	}

	workerHandler := worker.New(appService, log)
	go workerHandler.Start(msgs)

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(mwlogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.URLFormat)

	fileServer := http.FileServer(http.Dir("./static"))

	router.Handle("/static/*", http.StripPrefix("/static", fileServer))

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	router.Post("/notify", createNotify.New(log, appService))
	router.Get("/notify/{id}", getStatus.New(log, appService))
	router.Delete("/notify/{id}", deleteNotify.New(log, appService))

	log.Info("starting server", slog.String("address", cfg.HTTPServer.Address))

	srv := &http.Server{
		Addr:         cfg.HTTPServer.Address,
		Handler:      router,
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	if err = srv.ListenAndServe(); err != nil {
		log.Error("failed to start server", sl.Err(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop

	log.Info("server stopped", slog.String("signal", sign.String()))

	if err = storage.Close(); err != nil {
		log.Error("failed to close database", slog.String("error", err.Error()))
	}

	if err = mqBroker.Close(); err != nil {
		log.Error("failed to close RabbitMQ broker", slog.String("error", err.Error()))
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
