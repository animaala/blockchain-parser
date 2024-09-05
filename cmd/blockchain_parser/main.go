package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	p "github.com/animaala/blockchain-parser/internal/parser"
)

func main() {
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	URL := "https://cloudflare-eth.com"
	if os.Getenv("ETH_URL") != "" {
		URL = os.Getenv("ETH_URL")
	}

	p := p.NewEthereumParser(URL, http.Client{
		Timeout: 5 * time.Second, // 5 seconds http timeout
	})

	initApiServer(p, l)
}

func initApiServer(parser *p.EthereumParser, l *slog.Logger) {
	// Create a new HTTP server and handlers
	mux := http.NewServeMux()

	mux.HandleFunc("/subscribe", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}

		success := parser.Subscribe(address)
		if success {
			fmt.Fprintf(w, "Successfully subscribed to address: %s", address)
			l.Debug("Subscribed to address", "address", address)
		} else {
			http.Error(w, "Address already subscribed", http.StatusConflict)
			l.Debug("Address already subscribed", "address", address)
		}
	})

	mux.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Query().Get("address")
		if address == "" {
			http.Error(w, "Address parameter is required", http.StatusBadRequest)
			return
		}
		transactions := parser.GetTransactions(address) // cached txs
		l.Debug("Transactions", "address", address, "transactions", transactions)
		json.NewEncoder(w).Encode(transactions)
	})

	mux.HandleFunc("/current-block", func(w http.ResponseWriter, r *http.Request) {
		currentBlock := parser.GetCurrentBlock()
		l.Debug("Current block", "block", currentBlock)
		fmt.Fprintf(w, "Current block: %d", currentBlock)
	})

	mux.HandleFunc("/parse-block", func(w http.ResponseWriter, r *http.Request) {
		blockNumStr := r.URL.Query().Get("block")
		blockNum, err := strconv.ParseUint(blockNumStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid block number", http.StatusBadRequest)
			l.Error("Invalid block number", "block", blockNumStr)
			return
		}

		err = parser.ParseBlock(blockNum)
		if err != nil {
			http.Error(w, "Failed to parse block", http.StatusInternalServerError)
			l.Error("Failed to parse block", "block", blockNum, "error", err)
			return
		}
		l.Info("Block parsed successfully", "block", blockNum)
		w.Write([]byte("Block parsed successfully"))
	})

	// creating http server which gracefully shutdowns on OS interrupt signals
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	// Channel to listen for OS interrupt signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Run the server in a goroutine and use channels to signal when to shutdown
	go func() {
		l.Info("Server is starting on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("Could not listen on :8080", "error", err)
		}
	}()

	// Block until we receive an OS signal
	<-stop

	// Shutdown the server with a context and a timeout
	l.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		l.Error("Server forced to shutdown", "error", err)
	} else {
		l.Info("Server gracefully stopped")
	}
}
