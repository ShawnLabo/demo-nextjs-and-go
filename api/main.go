// Copyright 2022 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloud.google.com/go/spanner"
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

	ctx := context.Background()

	sc, err := spanner.NewClient(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("spanner.NewClient")
	}

	ap := &app{sc}

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
		log.Warn().Err(err).Msg(err.Error())
	}

	<-idleConnsClosed

	log.Info().Msg("bye")
}
