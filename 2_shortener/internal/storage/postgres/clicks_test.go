package postgres

import (
	"context"
	"regexp"
	"testing"

	"shortener/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/kxddry/wbf/dbpg"
)

func newClickStorage(t *testing.T) (*Storage, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("mock err: %v", err)
	}
	wrap := &dbpg.DB{Master: db}
	return &Storage{db: wrap}, mock, func() { _ = db.Close() }
}

func TestSaveClick(t *testing.T) {
	s, mock, done := newClickStorage(t)
	defer done()

	insRe := regexp.MustCompile(`INSERT\s+INTO\s+clicks`)
	mock.ExpectExec(insRe.String()).
		WithArgs("abc123", "ua", "127.0.0.1", "", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.SaveClick(context.Background(), domain.Click{ShortCode: "abc123", UserAgent: "ua", IP: "127.0.0.1", Referer: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestClicksByDay(t *testing.T) {
	s, mock, done := newClickStorage(t)
	defer done()

	qRe := regexp.MustCompile(`SELECT\s+TO_CHAR\(timestamp, 'YYYY-MM-DD'\)`) // simplified
	rows := sqlmock.NewRows([]string{"day", "count"}).AddRow("2025-01-01", int64(3))
	mock.ExpectQuery(qRe.String()).WithArgs("abc123").WillReturnRows(rows)

	res, err := s.ClicksByDay(context.Background(), "abc123", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res["2025-01-01"] != 3 {
		t.Fatalf("unexpected count: %v", res)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAnalytics(t *testing.T) {
	s, mock, done := newClickStorage(t)
	defer done()

	// ClickCount
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM clicks WHERE short_code = \$1`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))

	// UniqueClickCount
	mock.ExpectQuery(`SELECT COUNT\(DISTINCT ip\) FROM clicks WHERE short_code = \$1`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	// ByDay
	mock.ExpectQuery(`TO_CHAR\(timestamp, 'YYYY-MM-DD'\)`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"day", "count"}).AddRow("2025-01-01", int64(1)))

	// ByMonth
	mock.ExpectQuery(`date_trunc\('month'`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"month", "count"}).AddRow("2025-01", int64(1)))

	// By UA
	mock.ExpectQuery(`GROUP BY user_agent`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"ua", "count"}).AddRow("ua", int64(1)))

	// By Referer
	mock.ExpectQuery(`GROUP BY ref`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"ref", "count"}).AddRow("(direct)", int64(1)))

	// By IP
	mock.ExpectQuery(`GROUP BY ip::text`).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"ip", "count"}).AddRow("127.0.0.1", int64(1)))

	resp, err := s.Analytics(context.Background(), "abc123", nil, nil, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TotalClicks != 10 || resp.UniqueClicks != 5 {
		t.Fatalf("unexpected analytics numbers: %+v", resp)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
