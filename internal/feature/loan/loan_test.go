package loan_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/feature/loan"
	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/rs/xid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestLoan(t *testing.T) {
	ctx := pkg.Context.PutSlogLogger(context.Background(), slog.Default())
	ctrl := gomock.NewController(t)
	mockDatastore := datastore.NewMockDatastore(ctrl)

	featLoan, err := loan.New(ctx, loan.Configuration{
		LenderInterestRate:      .01, // 01%
		InterestRate:            .10, // 10%
		ServiceFee:              .05, // 05%
		MinRateOfInvestment:     .05, // 05%
		NumOfMonthlyInstallment: 12,
	}, loan.Dependency{
		Datastore: mockDatastore,
	})
	require.NoError(t, err)

	loanID := xid.New().Bytes()
	loanPartyID1 := xid.New().Bytes()
	borrowerID := []byte("900")
	lenderID1 := []byte("1111")
	lenderID2 := []byte("1112")
	fieldOfficerID := []byte("777")
	principal := &pkg.Money{
		ISO4217: "IDR",
		Amount:  10_000_000.00,
		Time:    time.Now(),
		Details: "factory expansion",
	}

	{
		mockDatastore.EXPECT().
			Mutation(ctx, gomock.Any()).
			Return(datastore.MutationResponse{Loans: &datastore.MutationResponseLoans{Loan: datastore.Loan{
				LoanID:    loanID,
				LoanState: datastore.StateProposed,
				Parties: []datastore.LoanParty{{
					LoanPartyID:     loanPartyID1,
					UserID:          borrowerID,
					LoanPartyRoleAs: datastore.RoleAsBorrower,
				}},
			}}}, nil)
	}
	resUpsert, err := featLoan.Upsert(ctx, loan.UpsertRequest{Proposed: &loan.ProposedRequest{
		BorrowerID: borrowerID,
		Principal:  principal,
	}})
	require.NoError(t, err)
	require.Equal(t, datastore.StateProposed.String(), resUpsert.LoanState)
	require.Equal(t, loanID, resUpsert.LoanID)

	{
		mockDatastore.EXPECT().
			Mutation(ctx, gomock.Any()).
			Return(datastore.MutationResponse{Loans: &datastore.MutationResponseLoans{Loan: datastore.Loan{
				LoanID:      loanID,
				LoanState:   datastore.StateApproved,
				ApprovedBy:  fieldOfficerID,
				ApprovedDoc: pkg.Ptr("http://google.com"),
			}}}, nil)
	}
	resUpsert, err = featLoan.Upsert(ctx, loan.UpsertRequest{Approved: &loan.ApprovedRequest{
		LoanID:           loanID,
		ApprovedDocument: pkg.Ptr("http://google.com"),
		FieldOfficerID:   fieldOfficerID,
	}})
	require.NoError(t, err)
	require.Equal(t, datastore.StateApproved.String(), resUpsert.LoanState)
	require.Equal(t, loanID, resUpsert.LoanID)

	{
		mockDatastore.EXPECT().
			Query(ctx, gomock.Any()).
			Return(datastore.QueryResponse{Loans: &datastore.QueryResponseLoans{Loan: datastore.Loan{
				LoanID:    loanID,
				LoanState: datastore.StateApproved,
				Parties: []datastore.LoanParty{{
					LoanPartyID:     loanPartyID1,
					UserID:          borrowerID,
					LoanPartyRoleAs: datastore.RoleAsBorrower,
					Payments: []datastore.LoanPartyPayment{{
						ISO4217: principal.ISO4217,
						Amount:  -principal.Amount,
						Time:    principal.Time.Unix(),
						Details: principal.Details,
					}},
				}},
			}}}, nil)
		mockDatastore.EXPECT().
			Mutation(ctx, gomock.Any()).
			Return(datastore.MutationResponse{Loans: &datastore.MutationResponseLoans{Loan: datastore.Loan{
				LoanID:      loanID,
				LoanState:   datastore.StateInvested,
				ApprovedBy:  fieldOfficerID,
				ApprovedDoc: pkg.Ptr("http://google.com"),
			}}}, nil)
	}
	resUpsert, err = featLoan.Upsert(ctx, loan.UpsertRequest{Invested: &loan.InvestedRequest{
		LoanID: loanID,
		Lenders: []loan.LoanLender{{
			LenderID: lenderID1,
			Payment: &pkg.Money{
				ISO4217: "IDR",
				Amount:  5_000_000.00,
				Time:    time.Now(),
				Details: "5mio",
			},
		}, {
			LenderID: lenderID2,
			Payment: &pkg.Money{
				ISO4217: "IDR",
				Amount:  5_000_000.00,
				Time:    time.Now(),
				Details: "5mio",
			},
		}},
	}})
	require.NoError(t, err)
	require.Equal(t, datastore.StateInvested.String(), resUpsert.LoanState)
	require.Equal(t, loanID, resUpsert.LoanID)

	{
		mockDatastore.EXPECT().
			Mutation(ctx, gomock.Any()).
			Return(datastore.MutationResponse{Loans: &datastore.MutationResponseLoans{Loan: datastore.Loan{
				LoanID:       loanID,
				LoanState:    datastore.StateDisbursed,
				ApprovedBy:   fieldOfficerID,
				ApprovedDoc:  pkg.Ptr("http://google.com"),
				DisbursedBy:  fieldOfficerID,
				DisbursedDoc: pkg.Ptr("http://google.com"),
			}}}, nil)
	}
	resUpsert, err = featLoan.Upsert(ctx, loan.UpsertRequest{Disbursed: &loan.DisbursedRequest{
		LoanID:                loanID,
		BorrowerContract:      pkg.Ptr("http://google.com"),
		DisbursementOfficerID: fieldOfficerID,
	}})
	require.NoError(t, err)
	require.Equal(t, datastore.StateDisbursed.String(), resUpsert.LoanState)
	require.Equal(t, loanID, resUpsert.LoanID)

	{
		mockDatastore.EXPECT().
			Query(ctx, gomock.Any()).
			Return(datastore.QueryResponse{Loans: &datastore.QueryResponseLoans{Loan: datastore.Loan{
				LoanID:    loanID,
				LoanState: datastore.StateDisbursed,
			}}}, nil)
	}
	resView, err := featLoan.View(ctx, loan.ViewRequest{LoanID: resUpsert.LoanID})
	require.NoError(t, err)
	require.Equal(t, resUpsert.LoanState, resView.LoanState)
	require.Equal(t, resUpsert.LoanID, resView.LoanID)
}
