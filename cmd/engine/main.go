package main

import (
	"FunPay-Core/internal/engine"
	"FunPay-Core/internal/fp"
	"FunPay-Core/internal/rest"
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	debug := flag.Bool("debug", false, "enable debug-level logging")
	flag.Parse()

	level := slog.LevelInfo
	if *debug {
		level = slog.LevelDebug
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	})))

	err := godotenv.Load()
	if err != nil {
		slog.Error("Error loading .env file", "err", err)
		os.Exit(1)
	}

	goldenKey := os.Getenv("FP_GOLDEN_KEY")
	sessionID := os.Getenv("FP_PHPSESSID")
	goldenSeal := os.Getenv("FP_GOLDEN_SEAL")

	slog.Info("engine starting")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client := fp.NewClient(goldenKey, sessionID, goldenSeal, 800*time.Millisecond, 600*time.Millisecond)
	account, err := client.GetAccount(ctx)
	if err != nil {
		slog.Error("get account failed", "err", err)
		return
	}
	slog.Info("account loaded", "login", account.Login, "balance", account.Balance)

	userID := os.Getenv("FP_USER_ID")
	csrfToken := os.Getenv("FP_CSRF_TOKEN")

	objectTypes := []string{"orders_counters", "chat_bookmarks"}

	runner := fp.NewRunner(client, userID, csrfToken, objectTypes)

	if err := runner.Init(ctx); err != nil {
		slog.Error("runner init failed", "err", err)
		return
	}

	buf := engine.NewBuffer()
	engineToken := os.Getenv("ENGINE_TOKEN")
	if engineToken == "" {
		slog.Error("ENGINE_TOKEN not set")
		os.Exit(1)
	}

	listenAddr := os.Getenv("ENGINE_LISTEN")
	if listenAddr == "" {
		listenAddr = "127.0.0.1:8731"
	}

	srv := rest.NewServer(buf, engineToken)
	go func() {
		if err := srv.Start(ctx, listenAddr); err != nil {
			slog.Error("rest server stopped", "err", err)
			cancel()
		}
	}()
	slog.Info("rest listening", "addr", listenAddr)
	for {
		ev, err := runner.Poll(ctx)
		if err != nil {
			if errors.Is(err, fp.ErrAuthLost) {
				slog.Error("auth lost: golden_seal expired", "err", err)
				buf.Push([]engine.Event{{
					Type: engine.EngineStatus,
					At:   time.Now(),
					Payload: engine.EngineStatusPayload{
						State: "auth_lost",
						Error: err.Error(),
					},
				}})
				srv.SetState("auth_lost")
				slog.Info("polling paused, waiting for restart or signal")
				<-ctx.Done()
				return
			}
			slog.Error("poll failed", "err", err)
			return
		}

		events := engine.WrapEvents(ev)
		buf.Push(events)
		buf.EvictExpired(time.Now())

		for _, e := range events {
			slog.Info("event", "event_id", e.EventID, "type", e.Type, "at", e.At)
		}

		slog.Debug("poll cycle", "events", len(events))
		select {
		case <-ctx.Done():
			return
		case <-time.After(2 * time.Second):
		}
	}
}
