SELECT
    l.loan_id,
    l.loan_state,
    l.approved_by,
    l.approved_doc,
    l.approved_at,
    l.approved_sign,
    l.disbursed_by,
    l.disbursed_doc,
    l.disbursed_at,
    l.disbursed_sign,
    l.created_at,
    l.created_sign,
    lp.loan_party_id,
    lp.user_id,
    lp.role_as,
    lp.created_at,
    lp.created_sign,
    lpp.iso4217,
    lpp.amount,
    lpp.due_time,
    lpp.details,
    lpp.created_at,
    lpp.created_sign
FROM loans l
JOIN loan_parties lp ON lp.loan_id = l.loan_id
JOIN loan_party_payments lpp on lpp.loan_party_id = lp.loan_party_id
WHERE   (l.loan_id = ? AND ? IS NOT NULL)
    OR  (lp.user_id = ? AND lp.role_as = 1 AND ? IS NOT NULL)
    OR  (lp.user_id = ? AND lp.role_as = 2 AND ? IS NOT NULL)
;