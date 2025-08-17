package postgres

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kxddry/wbf/dbpg"
)

func newTestStorage(t *testing.T) (*Storage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	wrap := &dbpg.DB{Master: db}
	return &Storage{db: wrap}, mock, func() { _ = db.Close() }
}

func TestSaveURL_Success(t *testing.T) {
	s, mock, closeFn := newTestStorage(t)
	defer closeFn()

	insertRe := regexp.MustCompile(`INSERT\s+INTO\s+shortened_urls`)
	mock.ExpectExec(insertRe.String()).
		WithArgs("https://example.com", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	code, err := s.SaveURL(context.Background(), "https://example.com", false, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6-char code, got %q", code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetURL_Found(t *testing.T) {
	s, mock, closeFn := newTestStorage(t)
	defer closeFn()

	queryRe := regexp.MustCompile(`SELECT\s+url\s+FROM\s+shortened_urls`)
	rows := sqlmock.NewRows([]string{"url"}).AddRow("https://example.com")
	mock.ExpectQuery(queryRe.String()).WithArgs("abc123").WillReturnRows(rows)

	url, err := s.GetURL(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://example.com" {
		t.Fatalf("unexpected url: %s", url)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetURL_NotFound(t *testing.T) {
	s, mock, closeFn := newTestStorage(t)
	defer closeFn()

	queryRe := regexp.MustCompile(`SELECT\s+url\s+FROM\s+shortened_urls`)
	rows := sqlmock.NewRows([]string{"url"})
	mock.ExpectQuery(queryRe.String()).WithArgs("missing").WillReturnRows(rows)

	_, err := s.GetURL(context.Background(), "missing")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveURL_WithAlias_Success(t *testing.T) {
	s, mock, closeFn := newTestStorage(t)
	defer closeFn()

	insertRe := regexp.MustCompile(`INSERT\s+INTO\s+shortened_urls`)
	mock.ExpectExec(insertRe.String()).
		WithArgs("https://example.com", "my-alias", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	code, err := s.SaveURL(context.Background(), "https://example.com", true, "my-alias")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != "my-alias" {
		t.Fatalf("expected alias 'my-alias', got %q", code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSaveURL_WithAlias_AlreadyExists(t *testing.T) {
	s, mock, closeFn := newTestStorage(t)
	defer closeFn()

	insertRe := regexp.MustCompile(`INSERT\s+INTO\s+shortened_urls`)
	mock.ExpectExec(insertRe.String()).
		WithArgs("https://example.com", "existing-alias", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected = conflict

	_, err := s.SaveURL(context.Background(), "https://example.com", true, "existing-alias")
	if err == nil {
		t.Fatalf("expected error for existing alias, got nil")
	}
	if err.Error() != "alias already exists" {
		t.Fatalf("unexpected error message: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}