package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/example/multitenant-search/internal/api"
	"github.com/example/multitenant-search/internal/platform"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, e := platform.LoadConfig("configs/config.yaml")
	if e != nil {
		panic(e)
	}
	log := platform.NewLogger(cfg.LogLevel)
	store, e := platform.NewStore(cfg.DataDir)
	if e != nil {
		panic(e)
	}
	srv := api.NewServer(store, log)
	httpSrv := &http.Server{Addr: cfg.HTTPAddr, Handler: srv.Routes(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 20 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		log.Info("http server starting", "addr", cfg.HTTPAddr)
		if e := httpSrv.ListenAndServe(); e != nil && !errors.Is(e, http.ErrServerClosed) {
			log.Error("http server failed", "error", e)
		}
	}()
	ln, e := net.Listen("tcp", cfg.GRPCAddr)
	if e == nil {
		ln.Close()
		log.Info("grpc endpoint reserved", "addr", cfg.GRPCAddr)
	}
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if e := httpSrv.Shutdown(ctx); e != nil {
		fmt.Println("shutdown:", e)
	}
	log.Info("server stopped")
}
