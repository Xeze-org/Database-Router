package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"db-router/internal/db"
)

type postgresService struct {
	db *db.Manager
}

// NewPostgresService constructs a PostgresService backed by the given db.Manager.
func NewPostgresService(m *db.Manager) PostgresService {
	return &postgresService{db: m}
}

func (s *postgresService) ListDatabases(ctx context.Context) ([]string, error) {
	if s.db.PostgresDB == nil {
		return nil, ErrNotEnabled("PostgreSQL")
	}

	rows, err := s.db.PostgresDB.QueryContext(ctx,
		"SELECT datname FROM pg_database WHERE datistemplate = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		dbs = append(dbs, name)
	}
	return dbs, nil
}

func (s *postgresService) CreateDatabase(ctx context.Context, name string) error {
	if s.db.PostgresDB == nil {
		return ErrNotEnabled("PostgreSQL")
	}
	if !isValidIdentifier(name) {
		return fmt.Errorf("invalid database name: %q", name)
	}
	// CREATE DATABASE cannot run inside a transaction, so we use a raw
	// connection from the pool with Exec directly.
	query := fmt.Sprintf("CREATE DATABASE %s", quoteIdentifier(name))
	_, err := s.db.PostgresDB.ExecContext(ctx, query)
	return err
}

func (s *postgresService) ListTables(ctx context.Context, database string) ([]string, error) {
	if s.db.PostgresDB == nil {
		return nil, ErrNotEnabled("PostgreSQL")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return nil, err
	}
	if isTemp {
		defer conn.Close()
	}

	const q = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'`

	rows, err := conn.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var t string
		if err := rows.Scan(&t); err != nil {
			continue
		}
		tables = append(tables, t)
	}
	return tables, nil
}

func (s *postgresService) ExecuteQuery(ctx context.Context, query, database string) ([]string, []Row, int64, bool, error) {
	if s.db.PostgresDB == nil {
		return nil, nil, 0, false, ErrNotEnabled("PostgreSQL")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return nil, nil, 0, false, err
	}
	if isTemp {
		defer conn.Close()
	}

	trimmed := strings.TrimSpace(query)
	upper := strings.ToUpper(trimmed)
	isSelect := strings.HasPrefix(upper, "SELECT") ||
		strings.HasPrefix(upper, "SHOW") ||
		strings.HasPrefix(upper, "WITH") ||
		strings.HasPrefix(upper, "TABLE") ||
		strings.HasPrefix(upper, "VALUES")

	if isSelect {
		rows, err := conn.QueryContext(ctx, trimmed)
		if err != nil {
			return nil, nil, 0, true, err
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			return nil, nil, 0, true, err
		}

		var results []Row
		for rows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range columns {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			row := make(Row, len(columns))
			for i, col := range columns {
				row[col] = values[i]
			}
			results = append(results, row)
		}
		return columns, results, int64(len(results)), true, nil
	}

	result, err := conn.ExecContext(ctx, trimmed)
	if err != nil {
		return nil, nil, 0, false, err
	}
	affected, _ := result.RowsAffected()
	return nil, nil, affected, false, nil
}

func (s *postgresService) SelectData(ctx context.Context, database, table string, limit int) ([]Row, error) {
	if s.db.PostgresDB == nil {
		return nil, ErrNotEnabled("PostgreSQL")
	}
	if !isValidIdentifier(table) {
		return nil, fmt.Errorf("invalid table name")
	}
	if limit < 1 || limit > 10000 {
		return nil, fmt.Errorf("limit must be between 1 and 10000")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return nil, err
	}
	if isTemp {
		defer conn.Close()
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1", quoteIdentifier(table))
	rows, err := conn.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []Row
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range columns {
			ptrs[i] = &values[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			continue
		}
		row := make(Row, len(columns))
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}
	return results, nil
}

func (s *postgresService) InsertData(ctx context.Context, database, table string, data Row) (string, error) {
	if s.db.PostgresDB == nil {
		return "", ErrNotEnabled("PostgreSQL")
	}
	if !isValidIdentifier(table) {
		return "", fmt.Errorf("invalid table name")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return "", err
	}
	if isTemp {
		defer conn.Close()
	}

	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data))
	i := 1
	for k, v := range data {
		cols = append(cols, quoteIdentifier(k))
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		vals = append(vals, v)
		i++
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		quoteIdentifier(table), strings.Join(cols, ", "), strings.Join(placeholders, ", "))

	var insertedID interface{}
	err = conn.QueryRowContext(ctx, query, vals...).Scan(&insertedID)
	if err != nil {
		query = fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
			quoteIdentifier(table), strings.Join(cols, ", "), strings.Join(placeholders, ", "))
		_, err = conn.ExecContext(ctx, query, vals...)
		if err != nil {
			return "", err
		}
		return "", nil
	}

	return fmt.Sprintf("%v", insertedID), nil
}

func (s *postgresService) UpdateData(ctx context.Context, database, table, id string, data Row) (int64, error) {
	if s.db.PostgresDB == nil {
		return 0, ErrNotEnabled("PostgreSQL")
	}
	if !isValidIdentifier(table) {
		return 0, fmt.Errorf("invalid table name")
	}
	if len(data) == 0 {
		return 0, fmt.Errorf("no data to update")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return 0, err
	}
	if isTemp {
		defer conn.Close()
	}

	setClauses := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data)+1)
	i := 1
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdentifier(k), i))
		vals = append(vals, v)
		i++
	}
	vals = append(vals, id)

	query := fmt.Sprintf("UPDATE %s SET %s WHERE id = $%d",
		quoteIdentifier(table), strings.Join(setClauses, ", "), i)

	result, err := conn.ExecContext(ctx, query, vals...)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return affected, nil
}

func (s *postgresService) DeleteData(ctx context.Context, database, table, id string) (int64, error) {
	if s.db.PostgresDB == nil {
		return 0, ErrNotEnabled("PostgreSQL")
	}
	if !isValidIdentifier(table) {
		return 0, fmt.Errorf("invalid table name")
	}

	conn, isTemp, err := s.db.GetPostgresConnection(database)
	if err != nil {
		return 0, err
	}
	if isTemp {
		defer conn.Close()
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", quoteIdentifier(table))
	result, err := conn.ExecContext(ctx, query, id)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return affected, nil
}

func (s *postgresService) TestConnection(ctx context.Context) (string, string, error) {
	if s.db.PostgresDB == nil {
		return "", "", ErrNotEnabled("PostgreSQL")
	}
	if err := s.db.PostgresDB.PingContext(ctx); err != nil {
		return "", "", err
	}
	return s.db.Config.Postgres.Host, s.db.Config.Postgres.Database, nil
}

// --- helpers (same logic as original handlers) ---

func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

func quoteIdentifier(name string) string {
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}
