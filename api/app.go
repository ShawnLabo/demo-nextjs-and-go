package main

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/spanner"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
)

type errorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var (
	errResInternalServerError = &errorResponse{http.StatusInternalServerError, "internal server error"}
)

type app struct {
	spanner *spanner.Client
}

func (ap *app) handler() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(requestLogger(log.Logger))
	r.Use(logRequest())

	r.Get("/", ap.root)
	r.Route("/api", func(r chi.Router) {
		r.Use(render.SetContentType(render.ContentTypeJSON))

		// TODO: Use CORS middleware only for local development.
		r.Use(cors.Handler(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Content-Type"},
			AllowCredentials: false,
		}))

		r.Get("/", ap.apiRoot)

		r.Route("/accounts", func(r chi.Router) {
			r.Get("/", ap.getAccounts)
			r.Post("/", ap.createAccount)
		})

		// A secure endpoint called with an API token.
		r.Get("/secure", ap.secureEndpoint)
	})

	return r
}

func (ap *app) root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "ok")
}

func (ap *app) apiRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `{"message": "Hello, API!"}`)
}

type getAccountsResponse struct {
	Accounts []*account `json:"accounts"`
}

func (ap *app) getAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := getLogger(ctx)

	// https://pkg.go.dev/cloud.google.com/go/spanner#Client.Single
	iter := ap.spanner.Single().Read(ctx, accountsTable, spanner.AllKeys(), allAccountColumns)
	defer iter.Stop()

	accounts := []*account{}

	for {
		row, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			logger.Err(err).Msg("iter.Next")
			render.JSON(w, r, errResInternalServerError)
			return
		}

		a := &account{}

		// https://pkg.go.dev/cloud.google.com/go/spanner#Row.ToStruct
		if err := row.ToStruct(a); err != nil {
			logger.Err(err).Msg("row.ToStruct")
			render.JSON(w, r, errResInternalServerError)
			return
		}

		accounts = append(accounts, a)
	}

	render.JSON(w, r, getAccountsResponse{accounts})
}

type createAccountRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type createAccountResponse struct {
	*account
}

func (ap *app) createAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := getLogger(ctx)

	req := &createAccountRequest{}

	// https://pkg.go.dev/github.com/go-chi/render#DecodeJSON
	if err := render.DecodeJSON(r.Body, req); err != nil {
		logger.Err(err).Msg("render.DecodeJSON")
		render.JSON(w, r, &errorResponse{http.StatusBadRequest, "invalid request"})
		return
	}

	a := &account{
		AccountID:    uuid.NewString(),
		APIToken:     uuid.NewString(),
		Email:        req.Email,
		Name:         req.Name,
		LastAccessed: nil,
	}

	// https://pkg.go.dev/cloud.google.com/go/spanner#InsertStruct
	m, err := spanner.InsertStruct(accountsTable, a)
	if err != nil {
		logger.Err(err).Msg("spanner.UpdateStruct")
		render.JSON(w, r, errResInternalServerError)
		return
	}

	// https://pkg.go.dev/cloud.google.com/go/spanner#Client.Apply
	if _, err := ap.spanner.Apply(ctx, []*spanner.Mutation{m}); err != nil {
		// https://pkg.go.dev/cloud.google.com/go/spanner#ErrCode
		if code := spanner.ErrCode(err); code == codes.AlreadyExists {
			render.JSON(w, r, &errorResponse{
				http.StatusConflict, fmt.Sprintf("%s is already used.", a.Email),
			})
			return
		}

		logger.Err(err).Msg("ap.spanner.Apply")
		render.JSON(w, r, errResInternalServerError)
		return
	}

	render.JSON(w, r, &createAccountResponse{a})
}

type secureEndpointResponse struct {
	Message string `json:"message"`
}

func (ap *app) secureEndpoint(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := getLogger(ctx)

	// API tokens are handed through Authorization Bearer scheme.
	auth := r.Header.Get("Authorization")
	if len(auth) <= 7 || auth[:7] != "Bearer " {
		render.JSON(w, r, &errorResponse{http.StatusBadRequest, "invalid header"})
		return
	}

	token := r.Header.Get("Authorization")[7:]

	fn := func(ctx context.Context, tx *spanner.ReadWriteTransaction) error {
		key := spanner.Key{token}

		// https://pkg.go.dev/cloud.google.com/go/spanner#ReadWriteTransaction.ReadRowUsingIndex
		row, err := tx.ReadRowUsingIndex(ctx, accountsTable, "ApiTokenIndex", key, []string{"AccountId"})
		if err != nil {
			if spanner.ErrCode(err) == codes.NotFound {
				return err
			}

			return fmt.Errorf("tx.ReadRowUsingIndex: %w", err)
		}

		var id string

		// https://pkg.go.dev/cloud.google.com/go/spanner#Row.ColumnByName
		if err := row.ColumnByName("AccountId", &id); err != nil {
			return fmt.Errorf("row.ColumnByName: %w", err)
		}

		// https://pkg.go.dev/cloud.google.com/go/spanner#Statement
		stmt := spanner.Statement{
			SQL: `UPDATE ` + accountsTable + `
              SET LastAccessed = CURRENT_TIMESTAMP()
              WHERE AccountId = @AccountId`,
			Params: map[string]interface{}{"AccountId": id},
		}

		// https://pkg.go.dev/cloud.google.com/go/spanner#ReadWriteTransaction.Update
		if _, err := tx.Update(ctx, stmt); err != nil {
			return fmt.Errorf("tx.Update: %w", err)
		}

		return nil
	}

	if _, err := ap.spanner.ReadWriteTransaction(ctx, fn); err != nil {
		if spanner.ErrCode(err) == codes.NotFound {
			render.JSON(w, r, &errorResponse{http.StatusUnauthorized, "unauthorized"})
			return
		}

		logger.Err(err).Msg("ap.spanner.Single().ReadRowUsingIndex")
		render.JSON(w, r, errResInternalServerError)
		return
	}

	render.JSON(w, r, &secureEndpointResponse{"ok"})
}
