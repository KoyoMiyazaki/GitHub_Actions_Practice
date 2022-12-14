package db

import "context"

type Querier interface {
	AddAccountBalance(ctx context.Context, arg AddAccountBalanceParams) (Account, error)
	GetAccount(ctx context.Context, id int64) (Account, error)
	GetAccountForUpdate(ctx context.Context, id int64) (Account, error)
	ListAccounts(ctx context.Context, arg ListAccountsParams) ([]Account, error)
	CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error)
	UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error)
	DeleteAccount(ctx context.Context, id int64) error
	CreateEntry(ctx context.Context, arg CreateEntryParams) (Entry, error)
	GetEntry(ctx context.Context, id int64) (Entry, error)
	ListEntries(ctx context.Context, arg ListEntriesParams) ([]Entry, error)
	CreateTransfer(ctx context.Context, arg CreateTransferParams) (Transfer, error)
	GetTransfer(ctx context.Context, id int64) (Transfer, error)
	ListTransfers(ctx context.Context, arg ListTransfersParams) ([]Transfer, error)
	CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
	GetUser(ctx context.Context, username string) (User, error)
}

var _ Querier = (*Queries)(nil)
