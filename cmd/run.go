package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/urfave/cli/v2"
	"github.com/w-h-a/golens/internal/client/sender"
	v1 "github.com/w-h-a/golens/internal/client/sender/v1"
	roothttphandler "github.com/w-h-a/golens/internal/handler/http/root"
	"github.com/w-h-a/golens/internal/server"
	httpserver "github.com/w-h-a/golens/internal/server/http"
	"github.com/w-h-a/golens/internal/service/wire"
)

func Run(c *cli.Context) error {
	ctx := c.Context

	stopChannels := map[string]chan struct{}{}

	s, err := InitV1Sender(ctx, "https://api.openai.com")
	if err != nil {
		return err
	}

	p := wire.New(s)
	stopChannels["proxy"] = make(chan struct{})

	httpSrv, err := InitHttpServer(ctx, ":8090", p)
	if err != nil {
		return err
	}
	stopChannels["httpserver"] = make(chan struct{})

	var wg sync.WaitGroup
	errCh := make(chan error, len(stopChannels))
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- p.Run(stopChannels["proxy"])
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		errCh <- httpSrv.Run(stopChannels["httpserver"])
	}()

	select {
	case err := <-errCh:
		if err != nil {
			return err
		}
	case <-sigChan:
		for _, stop := range stopChannels {
			close(stop)
		}
	}

	wg.Wait()

	close(errCh)
	for err := range errCh {
		if err != nil {
			// log
		}
	}

	return nil
}

// TODO: if we end up with wrappers around v1, the user can configure that here
func InitV1Sender(ctx context.Context, baseURL string) (sender.V1Sender, error) {
	return v1.NewSender(
		sender.WithBaseURL(baseURL),
	), nil
}

// TODO: if we end up with wrappers/middleware or also being fronted by grpc, the user can configure that here
func InitHttpServer(ctx context.Context, httpAddr string, w *wire.Wire) (server.Server, error) {
	srv := httpserver.NewServer(
		server.WithAddress(httpAddr),
	)

	router := mux.NewRouter()

	rootHandler := roothttphandler.New(w)

	router.PathPrefix("/").HandlerFunc(rootHandler.Handle)

	if err := srv.Handle(router); err != nil {
		return nil, fmt.Errorf("failed to attach handler: %w", err)
	}

	return srv, nil
}
