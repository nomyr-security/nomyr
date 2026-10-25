package pgx

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func TestBitemporalWriteAndReadBoundaries(t *testing.T) {
	databaseURL := os.Getenv("NOMYR_SPIKE_POSTGRES_URL")
	if databaseURL == "" {
		t.Skip("NOMYR_SPIKE_POSTGRES_URL is not set; PostgreSQL behavior not exercised")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	connection, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		t.Fatalf("connect to explicit local PostgreSQL: %v", err)
	}
	defer connection.Close(context.Background())

	tableName := fmt.Sprintf("spike_assertions_%d", time.Now().UnixNano())
	if _, err := connection.Exec(ctx, "CREATE TABLE "+tableName+` (
		assertion_id text PRIMARY KEY,
		subject_id text NOT NULL,
		valid_from timestamptz NOT NULL,
		valid_to timestamptz,
		recorded_at timestamptz NOT NULL,
		CHECK (valid_to IS NULL OR valid_from < valid_to)
	)`); err != nil {
		t.Fatal(err)
	}
	defer connection.Exec(context.Background(), "DROP TABLE IF EXISTS "+tableName)

	if _, err := connection.Exec(ctx, "INSERT INTO "+tableName+` (assertion_id, subject_id, valid_from, valid_to, recorded_at)
		VALUES ($1, $2, $3, $4, $5)`, "synthetic-invalid", "synthetic-subject", time.Unix(2, 0), time.Unix(1, 0), time.Now()); err == nil {
		t.Fatal("invalid valid-time interval was accepted")
	}
	validFrom := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	validTo := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	recordedAt := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	if _, err := connection.Exec(ctx, "INSERT INTO "+tableName+` (assertion_id, subject_id, valid_from, valid_to, recorded_at)
		VALUES ($1, $2, $3, $4, $5)`, "synthetic-1", "synthetic-subject", validFrom, validTo, recordedAt); err != nil {
		t.Fatal(err)
	}
	query := "SELECT assertion_id FROM " + tableName + `
		WHERE valid_from <= $1 AND (valid_to IS NULL OR $1 < valid_to) AND recorded_at <= $2`
	for _, testCase := range []struct {
		name       string
		validAt    time.Time
		recordedAt time.Time
		want       bool
	}{
		{name: "valid start inclusive", validAt: validFrom, recordedAt: recordedAt, want: true},
		{name: "valid end exclusive", validAt: validTo, recordedAt: recordedAt, want: false},
		{name: "not yet recorded", validAt: validFrom, recordedAt: recordedAt.Add(-time.Nanosecond), want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			rows, err := connection.Query(ctx, query, testCase.validAt, testCase.recordedAt)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			found := rows.Next()
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if found != testCase.want {
				t.Fatalf("found=%t, want %t", found, testCase.want)
			}
		})
	}
}
