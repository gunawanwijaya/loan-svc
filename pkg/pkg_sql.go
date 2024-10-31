package pkg

import "database/sql"

var SQL sql_

type sql_ struct{}

type SQLRowsX struct {
	*sql.Rows
	Flow *sqlScanFlow
}
type SQLScanFlow struct{ noCopy }
type sqlScanFlow struct {
	err  error
	stop bool
}

func (ss *sqlScanFlow) Next() (_ SQLScanFlow) { return }

func (ss *sqlScanFlow) Stop(err error) (_ SQLScanFlow) { ss.stop = true; ss.err = err; return }

func (sql_) Scan(rows *sql.Rows, each func(i int, rx SQLRowsX) SQLScanFlow) error {
	var next = &sqlScanFlow{}
	for i := 0; !next.stop && rows.Next(); i++ {
		_ = each(i, SQLRowsX{rows, next})
		if err := next.err; err != nil {
			return err
		}
	}
	return rows.Err()
}
