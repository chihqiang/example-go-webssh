package main

import (
	"chihqiang/webssh/ws"
	"context"
	"embed"
	"errors"
	"github.com/chihqiang/logx"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed static/*
var embedFs embed.FS

const (
	serverAddr   = ":8080"
	readTimeout  = 15 * time.Second
	writeTimeout = 15 * time.Second
	idleTimeout  = 60 * time.Second
)

var (
	staticFs fs.FS
	err      error
)

func init() {
	staticFs, err = fs.Sub(embedFs, "static")
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	registerRoutes()
	server := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	logx.Info("HTTP server starting at %s", serverAddr)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logx.Error(
				"HTTP server startup failed | addr: %s | error: %v | time: %s",
				serverAddr,
				err,
				time.Now().Format("2006-01-02 15:04:05"),
			)
			return
		}
	}()
	// 4. 优雅关闭处理（监听系统信号）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logx.Warn("Shutting down HTTP server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logx.Error("HTTP server forced shutdown | error: %v", err)
	} else {
		logx.Info("HTTP server exited normally")
	}
}
func registerRoutes() {
	http.Handle("/", http.FileServer(http.FS(staticFs)))
	http.HandleFunc("/ws", ws.SSHHandle)
}
