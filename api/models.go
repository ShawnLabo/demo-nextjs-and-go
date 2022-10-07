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

import "time"

const accountsTable = "Accounts"

type account struct {
	AccountID    string     `spanner:"AccountId" json:"account_id"`
	APIToken     string     `spanner:"ApiToken" json:"api_token"`
	Email        string     `spanner:"Email" json:"email"`
	Name         string     `spanner:"Name" json:"name"`
	LastAccessed *time.Time `spanner:"LastAccessed" json:"last_accessed"`
}

var allAccountColumns = []string{"AccountId", "ApiToken", "Email", "Name", "LastAccessed"}
