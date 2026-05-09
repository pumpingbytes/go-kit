package postgres

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	rootdberror "github.com/pumpingbytes/go-kit/dberror"
)

func TestNewReturnsClassifier(t *testing.T) {
	t.Parallel()

	classifier := New()
	if classifier == nil {
		t.Fatal("New() = nil, want non-nil")
	}

	var iface rootdberror.Classifier = classifier
	if iface == nil {
		t.Fatal("classifier does not satisfy dberror.Classifier")
	}
}

func TestClassifierIsNoRowsError(t *testing.T) {
	t.Parallel()

	classifier := Classifier{}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "direct pgx no rows", err: pgx.ErrNoRows, want: true},
		{name: "wrapped pgx no rows", err: fmt.Errorf("wrap: %w", pgx.ErrNoRows), want: true},
		{name: "other error", err: stderrors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifier.IsNoRowsError(tc.err); got != tc.want {
				t.Fatalf("IsNoRowsError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClassifierIsForeignKeyViolationError(t *testing.T) {
	t.Parallel()

	classifier := Classifier{}
	foreignKeyErr := &pgconn.PgError{Code: pgErrCodeForeignKeyViolation}
	uniqueErr := &pgconn.PgError{Code: pgErrCodeUniqueViolation}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "direct foreign key", err: foreignKeyErr, want: true},
		{name: "wrapped foreign key", err: fmt.Errorf("wrap: %w", foreignKeyErr), want: true},
		{name: "different pg error code", err: uniqueErr, want: false},
		{name: "other error", err: stderrors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifier.IsForeignKeyViolationError(tc.err); got != tc.want {
				t.Fatalf("IsForeignKeyViolationError() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestClassifierIsUniqueViolationError(t *testing.T) {
	t.Parallel()

	classifier := Classifier{}
	uniqueErr := &pgconn.PgError{Code: pgErrCodeUniqueViolation}
	foreignKeyErr := &pgconn.PgError{Code: pgErrCodeForeignKeyViolation}

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "direct unique", err: uniqueErr, want: true},
		{name: "wrapped unique", err: fmt.Errorf("wrap: %w", uniqueErr), want: true},
		{name: "different pg error code", err: foreignKeyErr, want: false},
		{name: "other error", err: stderrors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifier.IsUniqueViolationError(tc.err); got != tc.want {
				t.Fatalf("IsUniqueViolationError() = %v, want %v", got, tc.want)
			}
		})
	}
}
