package datastore

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore/queries"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/mattn/go-sqlite3"
)

type Configuration struct {
	//
}
type Dependency struct {
	DB struct {
		SQLite3 *sql.DB
	}
	ed25519.PublicKey
	ed25519.PrivateKey
}
type Datastore interface {
	Query(ctx context.Context, req QueryRequest) (res QueryResponse, err error)
	Mutation(ctx context.Context, req MutationRequest) (res MutationResponse, err error)
	// View(ctx context.Context, req ViewRequest) (res ViewResponse, err error)
	// Upsert(ctx context.Context, req UpsertRequest) (res UpsertResponse, err error)
}

func New(ctx context.Context, cfg Configuration, dep Dependency) (_ Datastore, err error) {
	cfg, err = pkg.AsValidator(cfg).Validate(ctx)
	if err != nil {
		return nil, err
	}
	dep, err = pkg.AsValidator(dep).Validate(ctx)
	if err != nil {
		return nil, err
	}
	return &datastore{cfg, dep}, nil
}

type datastore struct {
	Configuration
	Dependency
}

func (cfg Configuration) Validate(ctx context.Context) (_ Configuration, err error) {
	return cfg, nil
}

func (dep Dependency) Validate(ctx context.Context) (_ Dependency, err error) {
	{ //////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		db := dep.DB.SQLite3
		if db == nil {
			err = fmt.Errorf("repository/datastore: uninitialized db/sqlite3")
			return dep, err
		}
		var conn *sql.Conn
		if conn, err = db.Conn(ctx); err != nil {
			return dep, err
		}
		defer conn.Close()
		var tx *sql.Tx
		if tx, err = conn.BeginTx(ctx, &sql.TxOptions{}); err != nil {
			return dep, err
		}
		var v0, sid0 string
		if err = tx.QueryRowContext(ctx, "SELECT sqlite_version();").Scan(&v0); err != nil {
			return dep, err
		}
		if err = tx.QueryRowContext(ctx, "SELECT sqlite_source_id();").Scan(&sid0); err != nil {
			return dep, err
		}
		if err = tx.Commit(); err != nil {
			return dep, err
		}
		v, vn, sid := sqlite3.Version()
		log := pkg.Context.SlogLogger(ctx)
		log.DebugContext(ctx, "local",
			slog.Group("sqlite3",
				slog.String("version", v),
				slog.Int("version#", vn),
				slog.String("sourceID", sid),
			),
		)
		log.DebugContext(ctx, "query",
			slog.Group("sqlite3",
				slog.String("version", v0),
				slog.String("sourceID", sid0),
			),
		)
		log.DebugContext(ctx, "stats", slog.Any("stats", db.Stats()))

		now := time.Now().Unix()
		msg := []byte(fmt.Sprint(now))
		sig := ed25519.Sign(dep.PrivateKey, msg)
		ok := ed25519.Verify(dep.PublicKey, msg, sig)
		field := append(dep.PublicKey, sig...)
		ok2 := ed25519.Verify(field[:ed25519.PublicKeySize], msg, field[ed25519.PublicKeySize:])

		_, _ = ok, ok2
		q := queries.LoanSvc.SQLite3.Migration000()
		var res sql.Result
		if res, err = conn.ExecContext(ctx, q); err != nil {
			return dep, err
		}
		var li, ra int64
		if li, err = res.LastInsertId(); err != nil {
			return dep, err
		}
		if ra, err = res.RowsAffected(); err != nil {
			return dep, err
		}
		log.DebugContext(ctx, "migration",
			slog.Int64("LastInsertId", li),
			slog.Int64("RowsAffected", ra),
		)
	} //////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	return dep, nil
}
