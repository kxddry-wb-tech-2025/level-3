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

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT url FROM shortened_urls WHERE short_code = $1`)).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"url"}).AddRow("https://example.com"))

	// ClickCount
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(*) FROM clicks WHERE short_code = $1`)).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(10)))

	// UniqueClickCount
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT COUNT(DISTINCT ip) FROM clicks WHERE short_code = $1`)).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(5)))

	// ByDay
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT TO_CHAR(timestamp, 'YYYY-MM-DD') AS day, COUNT(*) " +
			"FROM clicks WHERE short_code = $1 GROUP BY day ORDER BY day",
	)).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"day", "count"}).AddRow("2025-01-01", int64(1)))

	// ByMonth
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT TO_CHAR(date_trunc('month', timestamp), 'YYYY-MM') AS month, COUNT(*) " +
			"FROM clicks WHERE short_code = $1 GROUP BY month ORDER BY month",
	)).
		WithArgs("abc123").
		WillReturnRows(sqlmock.NewRows([]string{"month", "count"}).AddRow("2025-01", int64(1)))

	// By User-Agent
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT user_agent, COUNT(*) FROM clicks WHERE short_code = $1 GROUP BY user_agent ORDER BY COUNT(*) DESC LIMIT $2",
	)).
		WithArgs("abc123", 10).
		WillReturnRows(sqlmock.NewRows([]string{"user_agent", "count"}).AddRow("ua", int64(1)))

	// By Referer
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT COALESCE(NULLIF(referer, ''), '(direct)') AS ref, COUNT(*) "+
			"FROM clicks WHERE short_code = $1 "+
			"GROUP BY ref ORDER BY COUNT(*) DESC LIMIT $2",
	)).
		WithArgs("abc123", 10).
		WillReturnRows(sqlmock.NewRows([]string{"ref", "count"}).AddRow("(direct)", int64(1)))

	// By IP
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT ip::text, COUNT(*) FROM clicks WHERE short_code = $1 GROUP BY ip::text ORDER BY COUNT(*) DESC LIMIT $2",
	)).
		WithArgs("abc123", 10).
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
