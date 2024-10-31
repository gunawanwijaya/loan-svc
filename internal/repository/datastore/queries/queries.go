package queries

import "embed"

var (
	_ embed.FS

	//go:embed loan-svc.sqlite3.migration.000.sql
	lss3_migration_000 string
	//go:embed loan-svc.sqlite3.mutation.loan-approved.sql
	lss3_mut_loan_approved string
	//go:embed loan-svc.sqlite3.mutation.loan-disbursed.sql
	lss3_mut_loan_disbursed string
	//go:embed loan-svc.sqlite3.mutation.loan-invested.sql
	lss3_mut_loan_invested string
	//go:embed loan-svc.sqlite3.mutation.loan-proposed.sql
	lss3_mut_loan_proposed string
	//go:embed loan-svc.sqlite3.query.loan.sql
	lss3_qry_loan string

	LoanSvc loan_svc
)

type loan_svc struct{ SQLite3 lss3 }
type lss3 struct{}

func (lss3) Migration000() string          { return lss3_migration_000 }
func (lss3) MutationLoanApproved() string  { return lss3_mut_loan_approved }
func (lss3) MutationLoanDisbursed() string { return lss3_mut_loan_disbursed }
func (lss3) MutationLoanInvested() string  { return lss3_mut_loan_invested }
func (lss3) MutationLoanProposed() string  { return lss3_mut_loan_proposed }
func (lss3) QueryLoan() string             { return lss3_qry_loan }
