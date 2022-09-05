package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	db "simple_bank/db/models"
	"simple_bank/mocks"
	"simple_bank/token"
	"simple_bank/util"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCreateTransferAPI(t *testing.T) {
	user1, _ := randomUser(t)
	user2, _ := randomUser(t)
	user3, _ := randomUser(t)

	account1 := randomAccount(user1.Username)
	account2 := randomAccount(user2.Username)
	account3 := randomAccount(user3.Username)

	account1.Currency = util.USD
	account2.Currency = util.USD
	account3.Currency = util.EUR

	amount := int64(10)

	testCases := []struct {
		name          string
		requestBody   gin.H
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(storeMock *mocks.Store)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				arg := db.TransferTxParams{
					FromAccountID: account1.ID,
					ToAccountID:   account2.ID,
					Amount:        amount,
				}

				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(account1, nil)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Once().
					Return(account2, nil)
				storeMock.
					On("TransferTx", mock.Anything, arg).
					Return(db.TransferTxResult{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "UnauthorizedUser",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, "unauthorized_user", time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(account1, nil)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, mock.Anything).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "FromAccountNotFound",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(db.Account{}, sql.ErrNoRows)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "ToAccountNotFound",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(account1, nil)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Once().
					Return(db.Account{}, sql.ErrNoRows)
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "FromAccountCurrencyMismatch",
			requestBody: gin.H{
				"from_account_id": account3.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user3.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account3.ID).
					Once().
					Return(db.Account{}, nil)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "ToAccountCurrencyMismatch",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account3.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(account1, nil)
				storeMock.
					On("GetAccount", mock.Anything, account3.ID).
					Once().
					Return(db.Account{}, nil)
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidAmount",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          -1,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, mock.Anything).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "InvalidCurrency",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        "Invalid Currency",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, mock.Anything).
					Maybe()
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "GetAccountError",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, mock.Anything).
					Return(db.Account{}, sql.ErrConnDone)
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Maybe()
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "TransferTxError",
			requestBody: gin.H{
				"from_account_id": account1.ID,
				"to_account_id":   account2.ID,
				"amount":          amount,
				"currency":        util.USD,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user1.Username, time.Minute)
			},
			buildStubs: func(storeMock *mocks.Store) {
				storeMock.
					On("GetAccount", mock.Anything, account1.ID).
					Once().
					Return(account1, nil)
				storeMock.
					On("GetAccount", mock.Anything, account2.ID).
					Once().
					Return(account2, nil)
				storeMock.
					On("TransferTx", mock.Anything, mock.Anything).
					Return(db.TransferTxResult{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			storeMock := mocks.NewStore(t)
			tc.buildStubs(storeMock)

			server := newTestServer(t, storeMock)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.requestBody)
			require.NoError(t, err)

			url := "/transfers"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}
