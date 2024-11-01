package rest_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/feature/loan"
	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/internal/service/rest"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestRest(t *testing.T) {
	ctx := pkg.Context.PutSlogLogger(context.Background(), slog.Default())
	ctrl := gomock.NewController(t)
	mockLoan := loan.NewMockLoan(ctrl)

	type obj = map[string]any

	svcRest, err := rest.New(ctx, rest.Configuration{
		//
	}, rest.Dependency{
		Loan: mockLoan,
	})
	require.NoError(t, err)

	loanID := xid.New().Bytes()

	buf := &bytes.Buffer{}
	err = json.NewEncoder(buf).Encode(obj{
		"proposed": obj{
			"borrower_id": "MTIz",
			"principal": obj{
				"iso4217": "IDR",
				"amount":  50_000_000.00,
				"details": "yea",
				"time":    "2024-10-30T18:00:00Z",
			},
		},
	})
	require.NoError(t, err)
	w, r := httptest.NewRecorder(), httptest.NewRequestWithContext(ctx, "POST", "/loan", buf)
	{
		mockLoan.EXPECT().
			Upsert(ctx, loan.UpsertRequest{Proposed: &loan.ProposedRequest{
				BorrowerID: []byte("123"),
				Principal: &pkg.Money{
					ISO4217: "IDR",
					Amount:  50_000_000.00,
					Time:    time.Date(2024, 10, 30, 18, 0, 0, 0, time.UTC),
					Details: "yea",
				},
			}}).
			Return(loan.UpsertResponse{
				LoanID:    loanID,
				LoanState: datastore.StateProposed.String(),
			}, nil)
	}
	svcRest.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
    "data": {
        "req": {
            "proposed": {
                "borrower_id": "MTIz",
                "principal": {
                    "iso4217": "IDR",
                    "amount": 50000000.00,
                    "details": "yea",
                    "time": "2024-10-30T18:00:00Z"
                }
            }
        },
        "res": {
            "loan_id": "`+pkg.BtoA(loanID)+`",
            "loan_state": "proposed"
        }
    }
}`, w.Body.String())

	w, r = httptest.NewRecorder(), httptest.NewRequestWithContext(ctx, "GET", "/loan/"+pkg.BtoA(loanID), buf)
	{
		mockLoan.EXPECT().
			View(ctx, loan.ViewRequest{LoanID: loanID}).
			Return(loan.ViewResponse{
				LoanID:     loanID,
				LoanState:  datastore.StateProposed.String(),
				BorrowerID: []byte("123"),
			}, nil)
	}
	svcRest.ServeHTTP(w, r)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{
    "data": {
        "req": {
            "loan_id": "`+pkg.BtoA(loanID)+`"
        },
        "res": {
            "borrower_id": "MTIz",
            "loan_id": "`+pkg.BtoA(loanID)+`",
            "loan_state": "proposed"
        }
    }
}`, w.Body.String())
}
