UPDATE loans
SET loan_state=?, approved_by=?, approved_doc=?, approved_at=?, approved_sign=?
WHERE loan_id=? AND loan_state=?;
