package loan

import (
	"context"
	"log/slog"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/pkg"
)

type ViewRequest struct {
	List []ViewRequest `json:"list,omitempty"`

	LoanID []byte `json:"loan_id,omitempty"`
	// LenderID   []byte `json:"lender_id,omitempty"`
	// BorrowerID []byte `json:"borrower_id,omitempty"`
}

type ViewResponse struct {
	List []ViewResponse `json:"list,omitempty"`

	LoanID           []byte       `json:"loan_id,omitempty"`
	LoanState        string       `json:"loan_state,omitempty"`
	BorrowerID       []byte       `json:"borrower_id,omitempty"`
	ExpectedPayments []*pkg.Money `json:"expected_payments,omitempty"`

	Lenders []struct {
		LenderID []byte       `json:"lender_id,omitempty"`
		Payments []*pkg.Money `json:"payments,omitempty"`
	} `json:"lenders,omitempty"`
}

func (x *loan) View(ctx context.Context, req ViewRequest) (res ViewResponse, err error) {
	log := pkg.Context.SlogLogger(ctx)

	var qry datastore.QueryResponse
	qry, err = x.Dependency.Datastore.Query(ctx, datastore.QueryRequest{
		Loans: &datastore.QueryRequestLoans{
			ByLoanID: req.LoanID,
		},
	})
	if err != nil {
		return
	}

	l := len(qry.List)
	res.List = make([]ViewResponse, l, l)
	for _, qry := range qry.List {
		res.LoanID = qry.Loans.Loan.LoanID
		res.LoanState = qry.Loans.Loan.LoanState.String()
		for _, party := range qry.Loans.Loan.Parties {
			switch party.LoanPartyRoleAs {
			case datastore.RoleAsBorrower:
				l := len(party.Payments)
				payments := make([]*pkg.Money, l, l)
				for i, payment := range party.Payments {
					payments[i] = &pkg.Money{
						ISO4217: payment.ISO4217,
						Amount:  payment.Amount,
						Details: payment.Details,
						Time:    time.Unix(payment.Time, 0),
					}
				}
				res.BorrowerID = party.UserID
				res.ExpectedPayments = payments
			case datastore.RoleAsLender:
				l := len(party.Payments)
				payments := make([]*pkg.Money, l, l)
				for i, payment := range party.Payments {
					payments[i] = &pkg.Money{
						ISO4217: payment.ISO4217,
						Amount:  payment.Amount,
						Details: payment.Details,
						Time:    time.Unix(payment.Time, 0),
					}
				}
				res.Lenders = append(res.Lenders, struct {
					LenderID []byte       "json:\"lender_id,omitempty\""
					Payments []*pkg.Money "json:\"payments,omitempty\""
				}{
					LenderID: party.UserID,
					Payments: payments,
				})
			}
		}
	}

	log.DebugContext(ctx, "feature/loan.View",
		slog.Any("req", req),
		slog.Any("res", res),
		slog.Any("err", err),
		slog.Any("qry", qry),
	)

	return
}
