package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
)

const shutdownTimeout = 10

func main() {
	cfg, err := initConfig()
	if err != nil {
		panic(err)
	}

	if err := initLogger(cfg.LogLevel, cfg.LogPretty); err != nil {
		panic(err)
	}

	ap := &app{}

	serve(ap.handler(), cfg.Port)
}

func serve(app http.Handler, port string) {
	s := &http.Server{
		Addr:    ":" + port,
		Handler: app,
	}

	idleConnsClosed := make(chan struct{})

	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

		sig := <-ch
		log.Info().Str("signal", sig.String()).Msg("received signal")
		log.Info().Msg("terminating")

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout*time.Second)
		defer cancel()

		if err := s.Shutdown(ctx); err != nil {
			log.Err(err).Msg(err.Error())
		}

		log.Info().Msg("shutdown completed")

		close(idleConnsClosed)
	}()

	log.Info().Msg("started")

	if err := s.ListenAndServe(); err != nil {
		log.Fatal().Err(err).Msg(err.Error())
	}

	<-idleConnsClosed

	log.Info().Msg("bye")
}
