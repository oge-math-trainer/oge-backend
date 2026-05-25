package migrations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const advisoryLockKey int64 = 2026052501
const legacyBaselineBefore = "011_oauth_identities.sql"

type Result struct {
	Applied   []string
	Baselined []string
	Skipped   []string
}

func Apply(ctx context.Context, pool *pgxpool.Pool, dir string) (Result, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "migrations"
	}

	files, err := migrationFiles(dir)
	if err != nil {
		return Result{}, err
	}

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return Result{}, err
	}
	defer conn.Release()

	if err := ensureSchemaMigrations(ctx, conn); err != nil {
		return Result{}, err
	}

	if _, err := conn.Exec(ctx, `SELECT pg_advisory_lock($1)`, advisoryLockKey); err != nil {
		return Result{}, fmt.Errorf("acquire migration lock: %w", err)
	}
	defer func() {
		_, _ = conn.Exec(context.Background(), `SELECT pg_advisory_unlock($1)`, advisoryLockKey)
	}()

	result := Result{}
	baselined, err := baselineLegacyDatabase(ctx, conn, files)
	if err != nil {
		return result, err
	}
	result.Baselined = baselined

	for _, file := range files {
		name := filepath.Base(file)
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return result, fmt.Errorf("read migration %s: %w", name, err)
		}
		sqlText := strings.TrimSpace(string(sqlBytes))
		if sqlText == "" {
			result.Skipped = append(result.Skipped, name)
			continue
		}

		checksum := checksum(sqlBytes)
		applied, err := isApplied(ctx, conn, name, checksum)
		if err != nil {
			return result, err
		}
		if applied {
			result.Skipped = append(result.Skipped, name)
			continue
		}

		if err := execSQL(ctx, conn, sqlText); err != nil {
			return result, fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, `
			INSERT INTO schema_migrations (filename, checksum, applied_at)
			VALUES ($1, $2, $3)
		`, name, checksum, time.Now().UTC()); err != nil {
			return result, fmt.Errorf("record migration %s: %w", name, err)
		}
		result.Applied = append(result.Applied, name)
	}

	return result, nil
}

func migrationFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read migrations directory %s: %w", dir, err)
	}

	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}
	sort.Strings(files)
	return files, nil
}

func ensureSchemaMigrations(ctx context.Context, conn *pgxpool.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			checksum TEXT NOT NULL,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)
	`)
	if err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}
	return nil
}

func baselineLegacyDatabase(ctx context.Context, conn *pgxpool.Conn, files []string) ([]string, error) {
	hasRecords, err := hasMigrationRecords(ctx, conn)
	if err != nil {
		return nil, err
	}
	if hasRecords {
		return nil, nil
	}

	hasUsers, err := tableExists(ctx, conn, "users")
	if err != nil {
		return nil, err
	}
	if !hasUsers {
		return nil, nil
	}

	baselined := make([]string, 0)
	for _, file := range files {
		name := filepath.Base(file)
		if name >= legacyBaselineBefore {
			continue
		}
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return baselined, fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := conn.Exec(ctx, `
			INSERT INTO schema_migrations (filename, checksum, applied_at)
			VALUES ($1, $2, $3)
			ON CONFLICT (filename) DO NOTHING
		`, name, checksum(sqlBytes), time.Now().UTC()); err != nil {
			return baselined, fmt.Errorf("baseline migration %s: %w", name, err)
		}
		baselined = append(baselined, name)
	}
	return baselined, nil
}

func hasMigrationRecords(ctx context.Context, conn *pgxpool.Conn) (bool, error) {
	var count int
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count); err != nil {
		return false, fmt.Errorf("count schema_migrations: %w", err)
	}
	return count > 0, nil
}

func tableExists(ctx context.Context, conn *pgxpool.Conn, tableName string) (bool, error) {
	var exists bool
	if err := conn.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = $1
		)
	`, tableName).Scan(&exists); err != nil {
		return false, fmt.Errorf("check table %s: %w", tableName, err)
	}
	return exists, nil
}

func isApplied(ctx context.Context, conn *pgxpool.Conn, filename, expectedChecksum string) (bool, error) {
	var existingChecksum string
	err := conn.QueryRow(ctx, `
		SELECT checksum
		FROM schema_migrations
		WHERE filename = $1
	`, filename).Scan(&existingChecksum)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read migration %s state: %w", filename, err)
	}
	if existingChecksum != expectedChecksum {
		return false, fmt.Errorf("migration %s checksum changed after it was applied", filename)
	}
	return true, nil
}

func execSQL(ctx context.Context, conn *pgxpool.Conn, sqlText string) error {
	results := conn.Conn().PgConn().Exec(ctx, sqlText)
	_, err := results.ReadAll()
	return err
}

func checksum(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
