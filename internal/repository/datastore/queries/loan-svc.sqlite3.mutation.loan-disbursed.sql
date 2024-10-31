UPDATE loans
SET loan_state=?, disbursed_by=?, disbursed_doc=?, disbursed_at=?, disbursed_sign=?
WHERE loan_id=? AND loan_state=?;
