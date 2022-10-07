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
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func initLogger(level string, pretty bool) error {
	var l zerolog.Level

	if level != "" {
		lv, err := zerolog.ParseLevel(level)
		if err != nil {
			return err
		}

		l = lv
	} else {
		l = zerolog.InfoLevel
	}

	var w io.Writer
	if pretty {
		w = zerolog.ConsoleWriter{Out: os.Stdout}
	} else {
		w = os.Stdout
	}

	zerolog.TimeFieldFormat = time.RFC3339Nano
	zerolog.SetGlobalLevel(l)
	log.Logger = zerolog.New(w).With().Timestamp().Caller().Logger().Hook(severityHook{})

	return nil
}

type severityHook struct{}

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
func (h severityHook) Run(e *zerolog.Event, l zerolog.Level, msg string) {
	if l != zerolog.NoLevel {
		var s string
		switch l {
		case zerolog.TraceLevel:
			s = "DEFAULT"
		case zerolog.DebugLevel:
			s = "DEBUG"
		case zerolog.InfoLevel:
			s = "INFO"
		case zerolog.WarnLevel:
			s = "WARNING"
		case zerolog.ErrorLevel:
			s = "ERROR"
		case zerolog.FatalLevel:
			s = "CRITICAL"
		case zerolog.PanicLevel:
			s = "EMERGENCY"
		}
		e.Str("severity", s)
	}
}
