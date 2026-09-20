package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/librearchive/librearchive/internal/bootstrap"
	"github.com/librearchive/librearchive/internal/config"
	"github.com/librearchive/librearchive/internal/database"
	api "github.com/librearchive/librearchive/internal/http"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load configuration: %v", err)
	}

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	initialAdministrator, err := bootstrap.EnsureInitialAdministrator(db)
	if err != nil {
		log.Fatalf("initialize administrator: %v", err)
	}
	if initialAdministrator.Created {
		fmt.Fprintf(os.Stderr, "\nLibreArchive Initial Administrator Created\nUsername: %s\nPassphrase (shown once): %s\nStore this passphrase securely before continuing.\n\n", initialAdministrator.Username, initialAdministrator.Passphrase)
	}

	server, err := api.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("initialize server: %v", err)
	}

	httpServer := &http.Server{Addr: cfg.ListenAddress, Handler: server.Router(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		log.Printf("LibreArchive API listening on %s", cfg.ListenAddress)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("serve API: %v", err)
		}
	}()
	shutdownSignal := make(chan os.Signal, 1)
	signal.Notify(shutdownSignal, syscall.SIGINT, syscall.SIGTERM)
	<-shutdownSignal
	context, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(context); err != nil {
		log.Fatalf("serve API: %v", err)
	}
}
