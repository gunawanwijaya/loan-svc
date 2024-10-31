CREATE TABLE IF NOT EXISTS loans (
    loan_id         BLOB    NOT NULL UNIQUE,
    loan_state      INTEGER NOT NULL, -- 1 = borrower; 2 = lender
    -- 
    approved_by     BLOB        NULL, -- ID of field officer doing the approval
    approved_doc    TEXT        NULL, -- url of document pointing to the approval
    approved_at     INTEGER     NULL, -- unix timestamp
    approved_sign   BLOB        NULL, -- signature contains of pk + signature of approved_at
    -- 
    disbursed_by    BLOB        NULL, -- ID of field officer doing the disbursement
    disbursed_doc   TEXT        NULL, -- url of document pointing to the disbursement
    disbursed_at    INTEGER     NULL, -- unix timestamp
    disbursed_sign  BLOB        NULL, -- signature contains of pk + signature of disbursed_at
    -- 
    created_at      INTEGER NOT NULL, -- unix timestamp
    created_sign    BLOB    NOT NULL  -- signature contains of pk + signature of created_at
);

CREATE TABLE IF NOT EXISTS loan_parties (
    loan_party_id   BLOB    NOT NULL UNIQUE,
    loan_id         BLOB    NOT NULL, -- FK to loans.loan_id
    user_id         BLOB    NOT NULL, -- FK to users.user_id
    role_as         INTEGER NOT NULL, -- 1 = borrower; 2 = lender
    created_at      INTEGER NOT NULL, -- unix timestamp
    created_sign    BLOB    NOT NULL  -- signature contains of pk + signature of created_at
);

CREATE TABLE IF NOT EXISTS loan_party_payments (
    loan_party_id   BLOB    NOT NULL, -- FK to loan_parties.loan_party_id
    iso4217         CHAR(3) NOT NULL,
    amount          NUMERIC NOT NULL,
    due_time        INTEGER NOT NULL, -- unix timestamp
    details         TEXT    NOT NULL,
    created_at      INTEGER NOT NULL, -- unix timestamp
    created_sign    BLOB    NOT NULL  -- signature contains of pk + signature of created_at
);
