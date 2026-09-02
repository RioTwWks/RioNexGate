package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"rionexgate/internal/api"
	"rionexgate/internal/config"
	"rionexgate/internal/core"
	"rionexgate/internal/db"
	"rionexgate/internal/telegram"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrate(); err != nil {
			log.Fatal(err)
		}
		return
	}
	if err := runServer(); err != nil {
		log.Fatal(err)
	}
}

func runMigrate() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	if err := database.AutoMigrate(); err != nil {
		return err
	}
	if err := database.SeedDefaultNode(); err != nil {
		return err
	}
	coreMgr := core.NewManager(cfg, database)
	if err := coreMgr.Reload(); err != nil {
		log.Printf("warning: initial core reload: %v", err)
	}
	log.Println("migration complete")
	return nil
}

func runServer() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		return err
	}
	if err := database.AutoMigrate(); err != nil {
		return err
	}
	_ = database.SeedDefaultNode()

	coreMgr := core.NewManager(cfg, database)
	if err := coreMgr.Reload(); err != nil {
		log.Printf("warning: initial core reload: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go coreMgr.StartStatsCollector(ctx)

	if cfg.Telegram.BotToken != "" && cfg.Telegram.BotToken != "YOUR_BOT_TOKEN" {
		bot, err := telegram.New(cfg, database, coreMgr)
		if err != nil {
			log.Printf("telegram bot disabled: %v", err)
		} else {
			go bot.Start()
		}
	}

	router := api.NewRouter(cfg, database, coreMgr)
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: router}

	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	cancel()
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	return srv.Shutdown(shutdownCtx)
}
