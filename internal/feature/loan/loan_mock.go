// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gunawanwijaya/loan-svc/internal/feature/loan (interfaces: Loan)
//
// Generated by this command:
//
//	mockgen -destination loan_mock.go -package loan . Loan
//

// Package loan is a generated GoMock package.
package loan

import (
	context "context"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLoan is a mock of Loan interface.
type MockLoan struct {
	ctrl     *gomock.Controller
	recorder *MockLoanMockRecorder
	isgomock struct{}
}

// MockLoanMockRecorder is the mock recorder for MockLoan.
type MockLoanMockRecorder struct {
	mock *MockLoan
}

// NewMockLoan creates a new mock instance.
func NewMockLoan(ctrl *gomock.Controller) *MockLoan {
	mock := &MockLoan{ctrl: ctrl}
	mock.recorder = &MockLoanMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLoan) EXPECT() *MockLoanMockRecorder {
	return m.recorder
}

// Upsert mocks base method.
func (m *MockLoan) Upsert(ctx context.Context, req UpsertRequest) (UpsertResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Upsert", ctx, req)
	ret0, _ := ret[0].(UpsertResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Upsert indicates an expected call of Upsert.
func (mr *MockLoanMockRecorder) Upsert(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Upsert", reflect.TypeOf((*MockLoan)(nil).Upsert), ctx, req)
}

// View mocks base method.
func (m *MockLoan) View(ctx context.Context, req ViewRequest) (ViewResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "View", ctx, req)
	ret0, _ := ret[0].(ViewResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// View indicates an expected call of View.
func (mr *MockLoanMockRecorder) View(ctx, req any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "View", reflect.TypeOf((*MockLoan)(nil).View), ctx, req)
}
