package datastore

import (
	"bytes"
	"context"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore/queries"
	"github.com/gunawanwijaya/loan-svc/pkg"
)

type QueryRequest struct {
	List []QueryRequest

	Loans *QueryRequestLoans
}

type QueryResponse struct {
	List []QueryResponse

	Loans *QueryResponseLoans
}

func (x *datastore) Query(ctx context.Context, req QueryRequest) (res QueryResponse, err error) {
	if req.Loans != nil {
		return x.queryLoans(ctx, req)
	}
	return
}

func (x *datastore) queryLoans(ctx context.Context, req QueryRequest) (res QueryResponse, err error) {
	db := x.Dependency.DB.SQLite3
	conn, err := db.Conn(ctx)
	if err != nil {
		return
	}
	rows, err := conn.QueryContext(ctx, queries.LoanSvc.SQLite3.QueryLoan(),
		req.Loans.ByLoanID, req.Loans.ByLoanID,
		req.Loans.ByBorrowerID, req.Loans.ByBorrowerID,
		req.Loans.ByLenderID, req.Loans.ByLenderID,
	)
	if err != nil {
		return
	}
	// if rows.NextResultSet() {
	res.Loans = &QueryResponseLoans{}
	var prevLoanID, prevLoanPartyID []byte
	err = pkg.SQL.Scan(rows, func(i int, rx pkg.SQLRowsX) pkg.SQLScanFlow {
		log := pkg.Context.SlogLogger(ctx)
		_ = log
		// log.DebugContext(ctx, "",
		// 	slog.Int("i", i),
		// 	slog.Any("res", res),
		// )

		var l Loan
		var lp LoanParty
		var lpp LoanPartyPayment
		if err = rx.Err(); err != nil {
			return rx.Flow.Stop(err)
		}
		if err = rx.Scan(
			&l.LoanID,
			&l.LoanState,
			//
			&l.ApprovedBy,
			&l.ApprovedDoc,
			&l.ApprovedAt,
			&l.ApprovedSign,
			//
			&l.DisbursedBy,
			&l.DisbursedDoc,
			&l.DisbursedAt,
			&l.DisbursedSign,
			//
			&l.CreatedAt,
			&l.CreatedSign,
			//
			&lp.LoanPartyID,
			&lp.UserID,
			&lp.LoanPartyRoleAs,
			&lp.CreatedAt,
			&lp.CreatedSign,
			//
			&lpp.ISO4217,
			&lpp.Amount,
			&lpp.Time,
			&lpp.Details,
			&lpp.CreatedAt,
			&lpp.CreatedSign,
		); err != nil {
			return rx.Flow.Stop(err)
		}

		if len(res.List) < 1 {
			lp.Payments = append(lp.Payments, lpp)
			l.Parties = append(l.Parties, lp)
			res.List = append(
				res.List,
				QueryResponse{Loans: &QueryResponseLoans{Loan: l}},
			)

			prevLoanID = l.LoanID
			prevLoanPartyID = lp.LoanPartyID
			return rx.Flow.Next()
		}

		// log.DebugContext(ctx, "",
		// 	slog.Bool("eq l", bytes.Equal(prevLoanID, l.LoanID)),
		// 	slog.Bool("eq lp", bytes.Equal(prevLoanPartyID, lp.LoanPartyID)),
		// )

		if !bytes.Equal(prevLoanID, l.LoanID) {
			res.List = append(
				res.List,
				QueryResponse{Loans: &QueryResponseLoans{Loan: l}},
			)
		}
		var lIdx int = len(res.List) - 1
		if !bytes.Equal(prevLoanPartyID, lp.LoanPartyID) {
			res.List[lIdx].Loans.Loan.Parties = append(
				res.List[lIdx].Loans.Loan.Parties,
				lp,
			)
		}
		var lpIdx int = len(res.List[lIdx].Loans.Loan.Parties) - 1
		res.List[lIdx].Loans.Loan.Parties[lpIdx].Payments = append(
			res.List[lIdx].Loans.Loan.Parties[lpIdx].Payments,
			lpp,
		)

		prevLoanID = l.LoanID
		prevLoanPartyID = lp.LoanPartyID
		return rx.Flow.Next()
	})
	// }
	return
}

type QueryRequestLoans struct {
	ByLoanID     []byte
	ByLenderID   []byte
	ByBorrowerID []byte
}
type QueryResponseLoans struct {
	Loan
}
