UPDATE loans SET loan_state=? WHERE loan_id=? AND loan_state=?;
INSERT OR IGNORE INTO loan_parties (loan_party_id, loan_id, user_id, role_as, created_at, created_sign) VALUES (?,?,?,?,?,?);
INSERT OR IGNORE INTO loan_party_payments (loan_party_id, iso4217, amount, due_time, details, created_at, created_sign) VALUES (?,?,?,?,?,?,?);