package rest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gunawanwijaya/loan-svc/internal/feature/loan"
	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/rs/cors"
)

type Configuration struct {
	//
}
type Dependency struct {
	loan.Loan
}

type REST interface {
	http.Handler
}

func New(ctx context.Context, cfg Configuration, dep Dependency) (_ REST, err error) {
	cfg, err = pkg.AsValidator(cfg).Validate(ctx)
	if err != nil {
		return nil, err
	}
	dep, err = pkg.AsValidator(dep).Validate(ctx)
	if err != nil {
		return nil, err
	}
	return &rest{cfg, dep}, nil
}

type rest struct {
	Configuration
	Dependency
}

func (x *rest) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := &http.ServeMux{}

	mux.Handle("POST /loan", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := loan.UpsertRequest{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			pkg.Must(json.NewEncoder(w).Encode(obj{
				"errors": []string{err.Error()},
			}))
		}
		res, err := x.Loan.Upsert(ctx, req)
		if err != nil {
			pkg.Must(json.NewEncoder(w).Encode(obj{
				"errors": []string{err.Error()},
			}))
		} else {
			pkg.Must(json.NewEncoder(w).Encode(obj{
				"data": obj{
					"req": req,
					"res": res,
				},
			}))
		}
	}))

	mux.Handle("GET /loan/{id...}", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		req := loan.ViewRequest{}
		id, kind := r.PathValue("id"), "loan"
		if r, ok := strings.CutPrefix(id, "lender/"); ok && r != id {
			id, kind = r, "lender"
		} else if r, ok := strings.CutPrefix(id, "borrower/"); ok && r != id {
			id, kind = r, "borrower"
		}

		if ids := strings.Split(id, ","); len(ids) == 1 {
			switch kind {
			default:
				req.LoanID = pkg.AtoB(ids[0])
				// case "lender":
				// 	req.LenderID = (ids[0])
				// case "borrower":
				// 	req.BorrowerID = (ids[0])
			}
		} else if len(ids) > 1 {
			for _, id := range ids {
				switch kind {
				default:
					req.List = append(req.List, loan.ViewRequest{LoanID: pkg.AtoB(id)})
					// case "lender":
					// 	req.List = append(req.List, loan.ViewRequest{LenderID: (id)})
					// case "borrower":
					// 	req.List = append(req.List, loan.ViewRequest{BorrowerID: (id)})
				}
			}
		}
		res, err := x.Loan.View(ctx, req)
		if err != nil {
			pkg.Must(json.NewEncoder(w).Encode(obj{
				"errors": []string{err.Error()},
			}))
		} else {
			pkg.Must(json.NewEncoder(w).Encode(obj{
				"data": obj{
					"req": req,
					"res": res,
				},
			}))
		}
	}))

	handler := mwcors(mux)
	handler.ServeHTTP(w, r)
}

func (cfg Configuration) Validate(ctx context.Context) (_ Configuration, err error) {
	return cfg, nil
}

func (dep Dependency) Validate(ctx context.Context) (_ Dependency, err error) {
	errFmt := "service/rest: uninitialized %s"
	if dep.Loan == nil {
		return Dependency{}, fmt.Errorf(errFmt, "feature/loan")
	}
	return dep, nil
}

type MW func(next http.Handler) http.Handler

type obj = map[string]any

var (
	mwcors MW = cors.Default().Handler
	// mw1    MW = func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		ctx := r.Context()
	// 		log := pkg.Context.SlogLogger(ctx)
	// 		log.DebugContext(ctx, "1st MW")
	// 		next.ServeHTTP(w, r)
	// 	})
	// }
	// mw2 MW = func(next http.Handler) http.Handler {
	// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 		ctx := r.Context()
	// 		log := pkg.Context.SlogLogger(ctx)
	// 		log.DebugContext(ctx, "2nd MW")
	// 		// next.ServeHTTP(w, r)
	// 		_ = next
	// 	})
	// }
)
