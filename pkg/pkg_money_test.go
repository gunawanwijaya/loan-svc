package pkg_test

import (
	"context"
	"testing"

	"github.com/gunawanwijaya/loan-svc/pkg"
	"github.com/stretchr/testify/require"
)

func TestMoney(t *testing.T) {
	ctx := context.Background()
	var errValidateMoney *pkg.ValidateMoneyError
	var errTakeMoney *pkg.TakeMoneyError

	m, err := pkg.Validator[*pkg.Money](&pkg.Money{ISO4217: "ABC", Amount: 10000.234}).Validate(ctx)
	_ = err.Error()
	require.ErrorAs(t, err, &errValidateMoney)
	require.Nil(t, m)
	require.Equal(t, "ABC", errValidateMoney.UnknownISO4217)

	m, err = (&pkg.Money{ISO4217: "IDR", Amount: 10000.234}).Validate(ctx)
	_ = err.Error()
	require.ErrorAs(t, err, &errValidateMoney)
	require.Equal(t, 10000.234, errValidateMoney.RawValue)
	require.Equal(t, 10000.23, errValidateMoney.Value)
	require.Equal(t, 10000.23, m.Amount)
	require.Equal(t, "IDR", m.ISO4217)

	m, err = (&pkg.Money{ISO4217: "JPY", Amount: 10000.234}).Validate(ctx)
	_ = err.Error()
	require.ErrorAs(t, err, &errValidateMoney)
	require.Equal(t, 10000.234, errValidateMoney.RawValue)
	require.Equal(t, 10000.0, errValidateMoney.Value)
	require.Equal(t, 10000.0, m.Amount)
	require.Equal(t, "JPY", m.ISO4217)

	tk, rm, err := m.Take(3)
	_ = err.Error()
	require.ErrorAs(t, err, &errTakeMoney)
	require.Equal(t, 3.0, errTakeMoney.Portion)

	tk, rm, err = m.Take(1.0 / 3)
	require.Nil(t, err)
	require.Equal(t, "JPY", tk.ISO4217)
	require.Equal(t, 3333.0, tk.Amount)
	require.Equal(t, "JPY", rm.ISO4217)
	require.Equal(t, 6667.0, rm.Amount)
}
