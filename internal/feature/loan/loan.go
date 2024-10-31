// Package loan implements the lifecycle of a loan
//
//	View(ctx, [{ LoanID, BorrowerID, LenderID }]) -> Result, Error
//
// View would return a list of loans on the system
//
//	Upsert(ctx, [{ LoanID }]) -> Result, Error
package loan

import (
	"context"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/pkg"
)

type Configuration struct {
	LenderInterestRate      float64 `json:"lender_interest_rate,omitempty"`
	InterestRate            float64 `json:"interest_rate,omitempty"`
	ServiceFee              float64 `json:"service_fee,omitempty"`
	NumOfMonthlyInstallment int     `json:"num_of_monthly_installment,omitempty"`
	MinRateOfInvestment     float64 `json:"min_rate_of_investment,omitempty"`
}
type Dependency struct {
	datastore.Datastore
}
type Loan interface {
	View(ctx context.Context, req ViewRequest) (res ViewResponse, err error)
	Upsert(ctx context.Context, req UpsertRequest) (res UpsertResponse, err error)
}

func New(ctx context.Context, cfg Configuration, dep Dependency) (_ Loan, err error) {
	cfg, err = pkg.AsValidator(cfg).Validate(ctx)
	if err != nil {
		return nil, err
	}
	dep, err = pkg.AsValidator(dep).Validate(ctx)
	if err != nil {
		return nil, err
	}
	return &loan{cfg, dep}, nil
}

type loan struct {
	Configuration
	Dependency
}

func (cfg Configuration) Validate(ctx context.Context) (_ Configuration, err error) {
	return cfg, nil
}

func (dep Dependency) Validate(ctx context.Context) (_ Dependency, err error) {
	return dep, nil
}
