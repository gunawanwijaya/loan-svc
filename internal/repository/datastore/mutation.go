package datastore

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore/queries"
	"github.com/gunawanwijaya/loan-svc/pkg"
)

type MutationRequest struct {
	List []MutationRequest

	Loans *MutationRequestLoans
}

type MutationResponse struct {
	List []MutationResponse

	Loans *MutationResponseLoans
}

func (x *datastore) Mutation(ctx context.Context, req MutationRequest) (res MutationResponse, err error) {
	if req.Loans != nil {
		return x.mutationLoans(ctx, req)
	}
	return
}

func (x *datastore) mutationLoans(ctx context.Context, req MutationRequest) (res MutationResponse, err error) {
	log := pkg.Context.SlogLogger(ctx)
	var exec sql.Result
	db := x.Dependency.DB.SQLite3
	conn, err := db.Conn(ctx)
	tx, err := conn.BeginTx(ctx, &sql.TxOptions{})

	defer func() {
		if err != nil {
			log.ErrorContext(ctx, "repository/datastore.mutationLoans "+req.Loans.Loan.LoanState.String(),
				slog.Any("err", err),
			)
			_ = tx.Rollback()
			res.Loans = nil
			return
		} else if exec != nil {
			li, liErr := exec.LastInsertId()
			ra, raErr := exec.RowsAffected()
			log.DebugContext(ctx, "repository/datastore.mutationLoans "+req.Loans.Loan.LoanState.String(),
				slog.Int64("li", li), slog.Any("liErr", liErr),
				slog.Int64("ra", ra), slog.Any("raErr", raErr),
			)
		}

		if err = tx.Commit(); err == nil {
			res.Loans = &MutationResponseLoans{Loan: req.Loans.Loan}
		}
	}()

	now := time.Now().Unix()
	msg := []byte(fmt.Sprint(now))
	sig := append(x.Dependency.PublicKey, ed25519.Sign(x.Dependency.PrivateKey, msg)...)

	switch req.Loans.Loan.LoanState {
	default:
		return
	case StateProposed:
		req.Loans.Loan.CreatedAt = now
		req.Loans.Loan.CreatedSign = sig
		for i := range req.Loans.Parties {
			req.Loans.Loan.Parties[i].CreatedAt = now
			req.Loans.Loan.Parties[i].CreatedSign = sig
			for j := range req.Loans.Loan.Parties[i].Payments {
				req.Loans.Loan.Parties[i].Payments[j].CreatedAt = now
				req.Loans.Loan.Parties[i].Payments[j].CreatedSign = sig
				exec, err = tx.ExecContext(ctx, queries.LoanSvc.SQLite3.MutationLoanProposed(),
					req.Loans.Loan.LoanID, req.Loans.Loan.LoanState, req.Loans.Loan.CreatedAt, req.Loans.Loan.CreatedSign,
					req.Loans.Loan.Parties[i].LoanPartyID, req.Loans.Loan.LoanID, req.Loans.Loan.Parties[i].UserID, int(req.Loans.Loan.Parties[i].LoanPartyRoleAs), req.Loans.Loan.Parties[i].CreatedAt, req.Loans.Loan.Parties[i].CreatedSign,
					req.Loans.Loan.Parties[i].LoanPartyID, req.Loans.Loan.Parties[i].Payments[j].ISO4217, req.Loans.Loan.Parties[i].Payments[j].Amount, req.Loans.Loan.Parties[i].Payments[j].Time, req.Loans.Loan.Parties[i].Payments[j].Details, req.Loans.Loan.Parties[i].Payments[j].CreatedAt, req.Loans.Loan.Parties[i].Payments[j].CreatedSign,
				)
				if err != nil {
					return res, err
				}
			}
		}
	case StateApproved:
		req.Loans.Loan.ApprovedAt = &now
		req.Loans.Loan.ApprovedSign = sig
		exec, err = tx.ExecContext(ctx, queries.LoanSvc.SQLite3.MutationLoanApproved(),
			req.Loans.Loan.LoanState, req.Loans.Loan.ApprovedBy, req.Loans.Loan.ApprovedDoc, req.Loans.Loan.ApprovedAt, req.Loans.Loan.ApprovedSign,
			req.Loans.Loan.LoanID, StateProposed, // required StateProposed
		)
	case StateInvested:
		log.DebugContext(ctx, "req.Loans.Parties",
			slog.Any("req.Loans.Parties", req.Loans.Parties),
		)
		for i := range req.Loans.Parties {
			req.Loans.Loan.Parties[i].CreatedAt = now
			req.Loans.Loan.Parties[i].CreatedSign = sig
			for j := range req.Loans.Loan.Parties[i].Payments {
				req.Loans.Loan.Parties[i].Payments[j].CreatedAt = now
				req.Loans.Loan.Parties[i].Payments[j].CreatedSign = sig
				exec, err = tx.ExecContext(ctx, queries.LoanSvc.SQLite3.MutationLoanInvested(),
					req.Loans.Loan.LoanState, req.Loans.Loan.LoanID, StateApproved, // required StateApproved
					req.Loans.Loan.Parties[i].LoanPartyID, req.Loans.Loan.LoanID, req.Loans.Loan.Parties[i].UserID, int(req.Loans.Loan.Parties[i].LoanPartyRoleAs), req.Loans.Loan.Parties[i].CreatedAt, req.Loans.Loan.Parties[i].CreatedSign,
					req.Loans.Loan.Parties[i].LoanPartyID, req.Loans.Loan.Parties[i].Payments[j].ISO4217, req.Loans.Loan.Parties[i].Payments[j].Amount, req.Loans.Loan.Parties[i].Payments[j].Time, req.Loans.Loan.Parties[i].Payments[j].Details, req.Loans.Loan.Parties[i].Payments[j].CreatedAt, req.Loans.Loan.Parties[i].Payments[j].CreatedSign,
				)
				if err != nil {
					return res, err
				}
			}
		}
	case StateDisbursed:
		req.Loans.Loan.DisbursedAt = &now
		req.Loans.Loan.DisbursedSign = sig
		exec, err = tx.ExecContext(ctx, queries.LoanSvc.SQLite3.MutationLoanDisbursed(),
			req.Loans.Loan.LoanState, req.Loans.Loan.DisbursedBy, req.Loans.Loan.DisbursedDoc, req.Loans.Loan.DisbursedAt, req.Loans.Loan.DisbursedSign,
			req.Loans.Loan.LoanID, StateInvested, // required StateInvested
		)
	}
	return
}

type MutationRequestLoans struct {
	Loan
}
type MutationResponseLoans struct {
	Loan
}
