package datastore

type Loan struct {
	LoanID        []byte // ID
	LoanState            //
	Parties       []LoanParty
	ApprovedBy    []byte  // ID of field officer doing the approval
	ApprovedDoc   *string // url of document pointing to the approval
	ApprovedAt    *int64  // Unix timestamp
	ApprovedSign  []byte  // signature of ApprovedAt
	DisbursedBy   []byte  // ID of field officer doing the disbursement
	DisbursedDoc  *string // url of document pointing to the disbursement
	DisbursedAt   *int64  // Unix timestamp
	DisbursedSign []byte  // signature of DisbursedAt
	CreatedAt     int64   // Unix timestamp
	CreatedSign   []byte  // signature of CreatedAt
}

type LoanParty struct {
	LoanPartyID     []byte // ID
	UserID          []byte // ID of said party defined by RoleAs
	LoanPartyRoleAs        // role of said user, either borrower or lender
	Payments        []LoanPartyPayment
	CreatedAt       int64  // Unix timestamp
	CreatedSign     []byte // signature of CreatedAt
}

type LoanPartyPayment struct {
	ISO4217     string
	Amount      float64
	Time        int64  // Unix timestamp
	Details     string // signature of DueTime
	CreatedAt   int64  // Unix timestamp
	CreatedSign []byte // signature of CreatedAt
}

type LoanState int

func (x LoanState) String() string {
	return map[LoanState]string{
		StateProposed:  "proposed",
		StateApproved:  "approved",
		StateInvested:  "invested",
		StateDisbursed: "disbursed",
	}[x]
}

const (
	_ LoanState = iota
	StateProposed
	StateApproved
	StateInvested
	StateDisbursed
)

type LoanPartyRoleAs int

func (x LoanPartyRoleAs) String() string {
	return map[LoanPartyRoleAs]string{
		RoleAsBorrower: "borrower",
		RoleAsLender:   "lender",
	}[x]
}

const (
	_ LoanPartyRoleAs = iota
	RoleAsBorrower
	RoleAsLender
)
