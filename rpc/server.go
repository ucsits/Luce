package rpc

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ucsits/Luce/blockchain"
	"github.com/ucsits/Luce/fsmgr"
)

type Server struct {
	echo   *echo.Echo
	config Config
	chain  *blockchain.Blockchain
	mu     sync.RWMutex
}

func NewServer(cfg Config, chain *blockchain.Blockchain) *Server {
	e := echo.New()
	e.HideBanner = true
	e.Server.ReadTimeout = cfg.ReadTimeout
	e.Server.WriteTimeout = cfg.WriteTimeout
	e.Use(middleware.Recover())
	e.HTTPErrorHandler = customHTTPErrorHandler

	s := &Server{
		echo:   e,
		config: cfg,
		chain:  chain,
	}

	apiGroup := e.Group("/api/v1")
	apiGroup.GET("/blocks", s.ListBlocks)
	apiGroup.GET("/blocks/:height", s.GetBlock)
	apiGroup.POST("/blocks", s.AppendBlock)
	apiGroup.GET("/chain/validate", s.ValidateChain)
	apiGroup.GET("/chain/height", s.GetHeight)

	return s
}

func (s *Server) Start() error {
	shutdownErr := make(chan error, 1)
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr <- s.Shutdown(ctx)
	}()

	addr := ":" + s.config.Port
	log.Printf("Luce RPC server listening on %s", addr)
	if err := s.echo.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	// Wait for the signal-triggered Shutdown to finish persisting the chain
	// before reporting success or failure to the caller.
	return <-shutdownErr
}

func (s *Server) Shutdown(ctx context.Context) error {
	// Stop accepting new requests and let in-flight ones drain first, so we
	// persist a quiesced chain instead of racing with concurrent appends.
	if err := s.echo.Shutdown(ctx); err != nil {
		return fmt.Errorf("stopping http server: %w", err)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := fsmgr.Dump(s.config.DataDir, *s.chain); err != nil {
		return fmt.Errorf("dumping blockchain: %w", err)
	}
	return nil
}
