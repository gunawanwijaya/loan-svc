package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"database/sql"
	"embed"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/goccy/go-yaml"
	"github.com/gunawanwijaya/loan-svc/internal/feature/loan"
	"github.com/gunawanwijaya/loan-svc/internal/repository/datastore"
	"github.com/gunawanwijaya/loan-svc/internal/service/rest"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/rs/zerolog"
	slogzerolog "github.com/samber/slog-zerolog/v2"
)

var (
	// define the first usable log
	log = func() *slog.Logger {
		w := os.Stdout
		zlog := zerolog.New(zerolog.ConsoleWriter{
			TimeFormat: time.Kitchen,
			Out:        w,
		})
		slog.SetDefault(slog.New(slogzerolog.Option{
			AddSource: false,
			Level:     slog.LevelDebug,
			Logger:    &zlog,
			Converter: slogzerolog.DefaultConverter,
			// AttrFromContext: []func(ctx context.Context) []slog.Attr{},
			// ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {},
		}.NewZerologHandler()))
		return slog.Default()
	}()

	_ embed.FS

	//go:embed main_config.yml
	configBytes []byte

	//go:embed main_secret.yml
	secretBytes []byte
)

type Config struct {
	Feature struct {
		Loan loan.Configuration `json:"loan"`
	} `json:"feature"`
	Repository struct {
		Datastore datastore.Configuration `json:"datastore"`
	} `json:"repository"`
	Service struct {
		REST rest.Configuration `json:"rest"`
	} `json:"service"`
}

type Secret struct {
	Database struct {
		Local struct {
			DSN string `json:"dsn"`
		} `json:"local"`
	} `json:"database"`
}

func main() {
	var (
		ctx, cancel = context.WithCancel(
			pkg.Context.PutSlogLogger(context.Background(), log),
		)
		timeout    = 3 * time.Second
		__graceful = pkg.Callstack()
		gracefully = __graceful.Register
		config     Config
		secret     Secret
	)
	{ //////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		defer func() {
			if r := recover(); r != nil {
				log.ErrorContext(ctx, "graceful", slog.Any("recover", r))
				__graceful.Call(context.WithTimeout(context.Background(), timeout))
			}
			if wait := __graceful.Wait(); wait != nil && !errors.Is(wait, context.Canceled) {
				log.ErrorContext(ctx, "graceful", slog.Any("wait", wait))
			} else {
				log.InfoContext(ctx, "graceful", slog.Any("wait", "done"))
			}
		}()

		chSig := make(chan os.Signal, 1)
		signal.Notify(chSig, os.Interrupt, os.Kill)
		go func() {
			sig := <-chSig
			print("\r") // delete current lint and send the cursor to the first column
			log.WarnContext(ctx, "graceful", slog.String("sig", sig.String()))
			__graceful.Call(context.WithTimeout(context.Background(), timeout))
		}()
		go func() {
			<-ctx.Done()
			log.WarnContext(ctx, "graceful", slog.String("ctx", ctx.Err().Error()))
			__graceful.Call(context.WithTimeout(context.Background(), timeout))
		}()
	} //////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	defer cancel()
	pkg.Must(yaml.Unmarshal(secretBytes, &secret))
	pkg.Must(yaml.Unmarshal(configBytes, &config))
	// log.DebugContext(ctx, "validated", slog.Any("config", config))

	dbSQLite3 := pkg.Must1(sql.Open("sqlite3", secret.Database.Local.DSN))
	gracefully(func() {
		dbSQLite3.Close()
		log.DebugContext(ctx, "gracefully db.Close")
	})

	pub, key := pkg.Must2(ed25519.GenerateKey(rand.Reader))

	repoDatastore := pkg.Must1(datastore.New(ctx, config.Repository.Datastore, datastore.Dependency{
		DB: struct{ SQLite3 *sql.DB }{
			SQLite3: dbSQLite3,
		},
		PublicKey:  pub,
		PrivateKey: key,
	}))

	featLoan := pkg.Must1(loan.New(ctx, config.Feature.Loan, loan.Dependency{
		Datastore: repoDatastore,
	}))

	svcREST := pkg.Must1(rest.New(ctx, config.Service.REST, rest.Dependency{
		Loan: featLoan,
	}))

	scheme := "http://"
	lis := pkg.Must1(net.Listen("tcp4", ":8080"))
	log.InfoContext(ctx, "running", slog.String("addr", scheme+lis.Addr().String()))
	srv := http.Server{
		Handler:           svcREST,
		BaseContext:       func(l net.Listener) context.Context { return ctx },
		ConnContext:       func(ctx context.Context, c net.Conn) context.Context { return ctx },
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      3 * time.Second,
		IdleTimeout:       9 * time.Second,
		ErrorLog:          slog.NewLogLogger(log.Handler(), slog.LevelError),
	}
	gracefully(func() {
		srv.Shutdown(ctx)
		log.DebugContext(ctx, "gracefully srv.Shutdown")
	})

	if err := srv.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
		pkg.Must(err)
	}
}
