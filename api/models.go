package main

import "time"

const accountsTable = "Accounts"

type account struct {
	AccountID    string     `spanner:"AccountId" json:"account_id"`
	APIToken     string     `spanner:"ApiToken" json:"api_token"`
	Email        string     `spanner:"Email" json:"email"`
	Name         string     `spanner:"Name" json:"name"`
	LastAccessed *time.Time `spanner:"LastAccessed" json:"last_accessed"`
}
