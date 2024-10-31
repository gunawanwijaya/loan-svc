package loan

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/rs/xid"
)

type UpsertRequest struct {
	List []UpsertRequest `json:"list,omitempty"`

	Proposed  *ProposedRequest  `json:"proposed,omitempty"`
	Approved  *ApprovedRequest  `json:"approved,omitempty"`
	Invested  *InvestedRequest  `json:"invested,omitempty"`
	Disbursed *DisbursedRequest `json:"disbursed,omitempty"`
}

type UpsertResponse struct {
	List []UpsertResponse `json:"list,omitempty"`

	LoanID    []byte            `json:"loan_id,omitempty"`
	LoanState string            `json:"loan_state,omitempty"`
	Invested  *InvestedResponse `json:"invested,omitempty"`
}

func (x *loan) Upsert(ctx context.Context, req UpsertRequest) (res UpsertResponse, err error) {
	log := pkg.Context.SlogLogger(ctx)
	switch {
	case req.Proposed != nil:
		log.DebugContext(ctx, "feature/loan.Upsert proposed")
		var p *ProposedRequest
		if p, err = pkg.AsValidator(req.Proposed).Validate(ctx); err == nil {
			return x.upsertProposed(ctx, p)
		}
	case req.Approved != nil:
		log.DebugContext(ctx, "feature/loan.Upsert approved")
		var a *ApprovedRequest
		if a, err = pkg.AsValidator(req.Approved).Validate(ctx); err == nil {
			return x.upsertApproved(ctx, a)
		}
	case req.Invested != nil:
		log.DebugContext(ctx, "feature/loan.Upsert invested")
		var i *InvestedRequest
		if i, err = pkg.AsValidator(req.Invested).Validate(ctx); err == nil {
			return x.upsertInvested(ctx, i)
		}
	case req.Disbursed != nil:
		log.DebugContext(ctx, "feature/loan.Upsert disbursed")
		var d *DisbursedRequest
		if d, err = pkg.AsValidator(req.Disbursed).Validate(ctx); err == nil {
			return x.upsertDisbursed(ctx, d)
		}
	}
	log.DebugContext(ctx, "feature/loan.Upsert",
		slog.Any("req", req),
		slog.Any("res", res),
		slog.Any("err", err),
	)
	return
}

// upsertProposed will assumed the payment is a monthly installment for 12 months consists of
//   - 5% service fee
//   - 10% interest rate
//   - total repayment expected is principal + 5% + 10%
//   - total repayment then converted into 12 times installment evenly spread out
func (x *loan) upsertProposed(ctx context.Context, p *ProposedRequest) (res UpsertResponse, err error) {
	loanID := xid.New()
	loanPartyID := xid.New()
	interest, _, err := p.Principal.Take(x.Configuration.InterestRate)
	service, _, err := p.Principal.Take(x.Configuration.ServiceFee)
	repayment, err := p.Principal.Sum(interest, service)

	split := x.Configuration.NumOfMonthlyInstallment
	installments, i := []*pkg.Money{}, split
	installment, err := repayment.Sum(nil)
	installmentTime := time.Now().AddDate(0, 1, 0)
	payments := []datastore.LoanPartyPayment{{
		ISO4217: p.Principal.ISO4217,
		Amount:  -p.Principal.Amount,
		Time:    p.Principal.Time.Unix(),
		Details: p.Principal.Details,
	}}

	// log := pkg.Context.SlogLogger(ctx)
	// log.DebugContext(ctx, "",
	// 	slog.Any("interest", interest),
	// 	slog.Any("service", service),
	// 	slog.Any("repayment", repayment),
	// 	slog.Any("payments", payments),
	// 	slog.Any("split", split),
	// )

	for range make([]struct{}, split, split) {
		var take *pkg.Money
		take, installment, err = installment.Take(1.0 / float64(i))
		// log.DebugContext(ctx, "",
		// 	slog.Any("take", take),
		// 	slog.Any("installment", installment),
		// 	slog.Any("err", err),
		// )
		take.Time = installmentTime
		installments = append(installments, take)
		payments = append(payments, datastore.LoanPartyPayment{
			ISO4217: take.ISO4217,
			Amount:  take.Amount,
			Time:    take.Time.Unix(),
			Details: fmt.Sprintf("Payment #%d of %d for loan [%s]", split-i+1, split, pkg.BtoA(loanID.Bytes())),
		})
		installmentTime = installmentTime.AddDate(0, 1, 0)
		i--
	}

	var mut datastore.MutationResponse
	mut, err = x.Dependency.Datastore.Mutation(ctx, datastore.MutationRequest{
		Loans: &datastore.MutationRequestLoans{
			Loan: datastore.Loan{
				LoanID:    loanID.Bytes(), // new loanID
				LoanState: datastore.StateProposed,
				Parties: []datastore.LoanParty{
					{
						LoanPartyID:     loanPartyID.Bytes(), // new loanPartyID
						UserID:          p.BorrowerID,
						LoanPartyRoleAs: datastore.RoleAsBorrower,
						Payments:        payments,
					},
				},
			},
		},
	})

	if err == nil {
		res.LoanID = mut.Loans.LoanID
		res.LoanState = mut.Loans.LoanState.String()
	}
	return
}

// upsertApproved
func (x *loan) upsertApproved(ctx context.Context, a *ApprovedRequest) (res UpsertResponse, err error) {
	var mut datastore.MutationResponse
	mut, err = x.Dependency.Datastore.Mutation(ctx, datastore.MutationRequest{
		Loans: &datastore.MutationRequestLoans{
			Loan: datastore.Loan{
				LoanID:      a.LoanID,
				LoanState:   datastore.StateApproved,
				ApprovedBy:  a.FieldOfficerID,
				ApprovedDoc: a.ApprovedDocument,
			},
		},
	})
	if err == nil {
		res.LoanID = mut.Loans.LoanID
		res.LoanState = mut.Loans.LoanState.String()
	}
	return
}

// upsertInvested will assumed that the sum of all lenders money cover 100% the principal amout or more
// and all lenders gaining profit with the same numbers as the interest rate that we introduce prior as 10%.
//
// less than 100% principal will failed the request, while equal or more than 100% will succeed.
//
// lenders slices will be splitted into `used` & `unused` investment eventually covering all the principal value.
func (x *loan) upsertInvested(ctx context.Context, i *InvestedRequest) (res UpsertResponse, err error) {
	var qry datastore.QueryResponse
	var mut datastore.MutationResponse
	var principal *pkg.Money
	qry, err = x.Dependency.Datastore.Query(ctx, datastore.QueryRequest{
		Loans: &datastore.QueryRequestLoans{
			ByLoanID: i.LoanID,
		},
	})
	if err != nil {
		return
	}

	log := pkg.Context.SlogLogger(ctx)
	log.DebugContext(ctx, "upsertInvested",
		slog.Any("qry", qry),
		slog.Any("err", err),
	)

	if len(qry.List) == 1 {
		qry = qry.List[0]
	}

	if qry.Loans.Loan.LoanState != datastore.StateApproved {
		err = fmt.Errorf("expected state from [approved]")
		return
	}

	for _, party := range qry.Loans.Loan.Parties {
		if party.LoanPartyRoleAs != datastore.RoleAsBorrower {
			continue
		}
		for _, payment := range party.Payments {
			if payment.Amount < 0 {
				principal = &pkg.Money{}
				principal.Amount = -payment.Amount
				principal.Details = payment.Details
				principal.ISO4217 = payment.ISO4217
				principal.Time = time.Unix(payment.Time, 0)
				break
			}
		}
	}

	if principal == nil {
		err = fmt.Errorf("invalid principal value")
		return
	}
	log.DebugContext(ctx, "upsertInvested",
		slog.Any("principal", principal),
	)

	var min, max *pkg.Money
	var covered bool
	min, max, err = principal.Take(x.Configuration.MinRateOfInvestment)
	if err != nil {
		return
	}
	log.DebugContext(ctx, "upsertInvested",
		slog.Any("min", min),
		slog.Any("max", max),
		slog.Any("err", err),
	)

	res.Invested = &InvestedResponse{}
	for _, lender := range i.Lenders {
		if covered {
			break
		}
		if lender.Payment.ISO4217 != principal.ISO4217 {
			res.Invested = nil
			err = fmt.Errorf("different currency")
			return
		}

		if lender.Payment.Amount >= min.Amount && lender.Payment.Amount <= max.Amount && lender.Payment.Amount < principal.Amount {
			// between 5% - 95% principal / remaining covered
			principal.Amount -= lender.Payment.Amount

			var interest *pkg.Money
			interest, _, err = lender.Payment.Take(x.Configuration.LenderInterestRate)
			lender.Repayment, err = lender.Payment.Sum(interest)
			lender.Repayment.Amount *= -1 // repayment to lenders is negative
			lender.Repayment.Time = lender.Repayment.Time.AddDate(0, x.Configuration.NumOfMonthlyInstallment, 0)
			lender.Repayment.Details = fmt.Sprintf("Repayment for loan [%s]", pkg.BtoA(i.LoanID))

			res.Invested.Used = append(res.Invested.Used, lender)
			covered = (principal.Amount == 0)
		} else if lender.Payment.Amount > principal.Amount {
			// more than principal amount
			usedLender, unusedLender := lender, lender
			var interest *pkg.Money
			usedLender.Payment, unusedLender.Payment, err = lender.Payment.Take(principal.Amount / lender.Payment.Amount)

			interest, _, err = usedLender.Payment.Take(x.Configuration.LenderInterestRate)
			usedLender.Repayment, err = usedLender.Payment.Sum(interest)
			usedLender.Repayment.Amount *= -1 // repayment to lenders is negative
			usedLender.Repayment.Time = usedLender.Repayment.Time.AddDate(0, x.Configuration.NumOfMonthlyInstallment, 0)
			usedLender.Repayment.Details = fmt.Sprintf("Repayment for loan [%s]", pkg.BtoA(i.LoanID))

			res.Invested.Used = append(res.Invested.Used, usedLender)
			res.Invested.Unused = append(res.Invested.Unused, unusedLender)
			covered = true
		} else if lender.Payment.Amount == principal.Amount {
			// exact amount of principal amount
			principal.Amount -= lender.Payment.Amount

			var interest *pkg.Money
			interest, _, err = lender.Payment.Take(x.Configuration.LenderInterestRate)
			lender.Repayment, err = lender.Payment.Sum(interest)
			lender.Repayment.Amount *= -1 // repayment to lenders is negative
			lender.Repayment.Time = lender.Repayment.Time.AddDate(0, x.Configuration.NumOfMonthlyInstallment, 0)
			lender.Repayment.Details = fmt.Sprintf("Repayment for loan [%s]", pkg.BtoA(i.LoanID))

			res.Invested.Used = append(res.Invested.Used, lender)
			covered = true
		} else {
			res.Invested.Unused = append(res.Invested.Unused, lender)
		}
	}
	if !covered {
		err = fmt.Errorf("principal is not fully covered, missing %s", principal)
		return
	}

	l := len(res.Invested.Used)
	parties := make([]datastore.LoanParty, l, l)
	for i, used := range res.Invested.Used {
		loanPartyID := xid.New()
		parties[i] = datastore.LoanParty{
			LoanPartyID:     loanPartyID.Bytes(),
			UserID:          used.LenderID,
			LoanPartyRoleAs: datastore.RoleAsLender,
			Payments: []datastore.LoanPartyPayment{{
				ISO4217: used.Payment.ISO4217,
				Amount:  used.Payment.Amount,
				Time:    used.Payment.Time.Unix(),
				Details: used.Payment.Details,
			}, {
				ISO4217: used.Repayment.ISO4217,
				Amount:  used.Repayment.Amount,
				Time:    used.Repayment.Time.Unix(),
				Details: used.Repayment.Details,
			}},
		}
	}

	_ = mut
	mut, err = x.Dependency.Datastore.Mutation(ctx, datastore.MutationRequest{
		Loans: &datastore.MutationRequestLoans{
			Loan: datastore.Loan{
				LoanID:    i.LoanID,
				LoanState: datastore.StateInvested,
				Parties:   parties,
			},
		},
	})
	if err == nil {
		res.LoanID = mut.Loans.LoanID
		res.LoanState = mut.Loans.LoanState.String()
	}
	return
}

// upsertDisbursed
func (x *loan) upsertDisbursed(ctx context.Context, d *DisbursedRequest) (res UpsertResponse, err error) {
	var mut datastore.MutationResponse
	mut, err = x.Dependency.Datastore.Mutation(ctx, datastore.MutationRequest{
		Loans: &datastore.MutationRequestLoans{
			Loan: datastore.Loan{
				LoanID:       d.LoanID,
				LoanState:    datastore.StateDisbursed,
				DisbursedBy:  d.DisbursementOfficerID,
				DisbursedDoc: d.BorrowerContract,
			},
		},
	})
	if err == nil {
		res.LoanID = mut.Loans.LoanID
		res.LoanState = mut.Loans.LoanState.String()
	}
	return
}

type ProposedRequest struct {
	BorrowerID []byte     `json:"borrower_id,omitempty"`
	Principal  *pkg.Money `json:"principal,omitempty"`
}

func (x *ProposedRequest) Validate(ctx context.Context) (_ *ProposedRequest, err error) {
	if len(x.BorrowerID) < 1 {
		return nil, fmt.Errorf("invalid borrower_id")
	}
	if x.Principal, err = x.Principal.Validate(ctx); err != nil {
		return nil, err
	}
	return x, nil
}

type ProposedResponse struct {
	LoanID []byte `json:"loan_id,omitempty"`
}

type ApprovedRequest struct {
	LoanID           []byte  `json:"loan_id,omitempty"`
	ApprovedDocument *string `json:"approved_document,omitempty"`
	FieldOfficerID   []byte  `json:"field_officer_id,omitempty"`
}

func (x *ApprovedRequest) Validate(ctx context.Context) (_ *ApprovedRequest, err error) {
	if len(x.LoanID) < 1 {
		return nil, fmt.Errorf("invalid loan_id")
	}
	if x.ApprovedDocument == nil || len(*x.ApprovedDocument) < 1 {
		return nil, fmt.Errorf("invalid approved_document")
	}
	if len(x.FieldOfficerID) < 1 {
		return nil, fmt.Errorf("invalid field_officer_id")
	}
	return x, nil
}

type InvestedRequest struct {
	LoanID  []byte       `json:"loan_id,omitempty"`
	Lenders []LoanLender `json:"lenders,omitempty"`
}

type InvestedResponse struct {
	Used   []LoanLender `json:"used,omitempty"`
	Unused []LoanLender `json:"unused,omitempty"`
}

func (x *InvestedRequest) Validate(ctx context.Context) (_ *InvestedRequest, err error) {
	if len(x.LoanID) < 1 {
		return nil, fmt.Errorf("invalid loan_id")
	}
	currency := ""
	for i, lender := range x.Lenders {
		if lender.Payment.Amount <= 0 {
			return nil, fmt.Errorf("payment amount should be more than 0")
		}
		if currency != "" && currency != lender.Payment.ISO4217 {
			return nil, fmt.Errorf("different currency")
		}
		lender.Payment, err = lender.Payment.Validate(ctx)
		if err != nil || lender.LenderID == nil || lender.Payment == nil {
			x.Lenders = slices.Delete(x.Lenders, i, i+1)
		}
		currency = lender.Payment.ISO4217
	}

	if len(x.Lenders) < 1 {
		return nil, fmt.Errorf("empty lenders")
	}
	return x, nil
}

type DisbursedRequest struct {
	LoanID                []byte  `json:"loan_id,omitempty"`
	BorrowerContract      *string `json:"borrower_contract,omitempty"`
	DisbursementOfficerID []byte  `json:"disbursement_officer_id,omitempty"`
}

func (x *DisbursedRequest) Validate(ctx context.Context) (_ *DisbursedRequest, err error) {
	if len(x.LoanID) < 1 {
		return nil, fmt.Errorf("invalid loan_id")
	}
	if x.BorrowerContract == nil || len(*x.BorrowerContract) < 1 {
		return nil, fmt.Errorf("invalid borrower_contract")
	}
	if len(x.DisbursementOfficerID) < 1 {
		return nil, fmt.Errorf("invalid disbursement_officer_id")
	}
	return x, nil
}

type LoanLender struct {
	LenderID       []byte     `json:"lender_id,omitempty"`
	LenderContract string     `json:"lender_contract,omitempty"`
	Payment        *pkg.Money `json:"payment,omitempty"`
	Repayment      *pkg.Money `json:"repayment,omitempty"`
}
