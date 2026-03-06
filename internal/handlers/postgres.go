package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// isValidIdentifier checks if a string is a valid SQL identifier
// (letters, numbers, underscores; must start with letter or underscore)
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return matched
}

// quoteIdentifier quotes a SQL identifier to prevent injection
func quoteIdentifier(name string) string {
	// Replace any double quotes with two double quotes (PostgreSQL escaping)
	escaped := strings.ReplaceAll(name, `"`, `""`)
	return fmt.Sprintf(`"%s"`, escaped)
}

// TestPostgresConnection returns the live connection status.
func (h *Handler) TestPostgresConnection(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}
	if err := h.DB.PostgresDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "disconnected", "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":   "connected",
		"database": h.DB.Config.Postgres.Database,
		"host":     h.DB.Config.Postgres.Host,
	})
}

// ListPostgresDatabases lists all non-template databases.
func (h *Handler) ListPostgresDatabases(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	rows, err := h.DB.PostgresDB.Query(
		"SELECT datname FROM pg_database WHERE datistemplate = false",
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var databases []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		databases = append(databases, name)
	}
	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// ListPostgresTables lists all user tables in the public schema.
func (h *Handler) ListPostgresTables(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close()
	}

	const q = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`

	rows, err := db.Query(q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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
	c.JSON(http.StatusOK, gin.H{"database": database, "tables": tables})
}

// ExecutePostgresQuery runs arbitrary SQL and returns columns + rows for SELECT,
// or rows_affected for DML/DDL statements.
func (h *Handler) ExecutePostgresQuery(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	var req struct {
		Query    string `json:"query" binding:"required"`
		Database string `json:"database"` // Optional: target database
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(req.Database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close() // Close temporary connection after query
	}

	// Trim and determine query type
	query := strings.TrimSpace(req.Query)
	queryUpper := strings.ToUpper(query)

	// Check if this is a SELECT query (returns rows)
	// SELECT, SHOW, WITH (CTE), TABLE, VALUES can return rows
	isSelectLike := strings.HasPrefix(queryUpper, "SELECT") ||
		strings.HasPrefix(queryUpper, "SHOW") ||
		strings.HasPrefix(queryUpper, "WITH") ||
		strings.HasPrefix(queryUpper, "TABLE") ||
		strings.HasPrefix(queryUpper, "VALUES")

	if isSelectLike {
		// Use Query() for SELECT-like statements
		rows, err := db.Query(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		columns, err := rows.Columns()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var results []map[string]interface{}
		for rows.Next() {
			values := make([]interface{}, len(columns))
			ptrs := make([]interface{}, len(columns))
			for i := range columns {
				ptrs[i] = &values[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				continue
			}
			row := make(map[string]interface{})
			for i, col := range columns {
				row[col] = values[i]
			}
			results = append(results, row)
		}

		c.JSON(http.StatusOK, gin.H{
			"columns": columns,
			"rows":    results,
			"count":   len(results),
		})
	} else {
		// Use Exec() for DML/DDL statements (INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, etc.)
		result, err := db.Exec(query)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		rowsAffected, _ := result.RowsAffected()

		// For statements like CREATE DATABASE, DROP TABLE, etc., rowsAffected is typically 0
		// But the command succeeded if no error occurred
		c.JSON(http.StatusOK, gin.H{
			"rows_affected": rowsAffected,
			"message":       "Command executed successfully",
		})
	}
}

// SelectPostgresData returns rows from a table with an optional row limit.
func (h *Handler) SelectPostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")
	limitStr := c.DefaultQuery("limit", "100")

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close()
	}

	// Validate table name to prevent SQL injection (basic check)
	if !isValidIdentifier(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	// Parse and validate limit
	var limit int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit < 1 || limit > 10000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value (must be 1-10000)"})
		return
	}

	// Use parameterized query with identifier quoting for table name
	query := fmt.Sprintf("SELECT * FROM %s LIMIT $1", quoteIdentifier(table))
	rows, err := db.Query(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	columns, _ := rows.Columns()
	var results []map[string]interface{}
	for rows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range columns {
			ptrs[i] = &values[i]
		}
		rows.Scan(ptrs...)
		row := make(map[string]interface{})
		for i, col := range columns {
			row[col] = values[i]
		}
		results = append(results, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"database": database,
		"table":    table,
		"data":     results,
		"count":    len(results),
	})
}

// InsertPostgresData inserts a JSON object as a new row.
// If the table has an 'id' column, it returns the new id.
func (h *Handler) InsertPostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")

	// Validate table name
	if !isValidIdentifier(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close()
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

	// Try to return 'id' if it exists, otherwise just execute
	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		quoteIdentifier(table),
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	var insertedID interface{}
	err = db.QueryRow(query, vals...).Scan(&insertedID)

	if err != nil {
		// If RETURNING id fails (column doesn't exist), try without RETURNING
		query = fmt.Sprintf(
			"INSERT INTO %s (%s) VALUES (%s)",
			quoteIdentifier(table),
			strings.Join(cols, ", "),
			strings.Join(placeholders, ", "),
		)
		_, err = db.Exec(query, vals...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"database": database,
			"table":    table,
			"message":  "Row inserted successfully",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"database":    database,
		"table":       table,
		"inserted_id": insertedID,
	})
}

// UpdatePostgresData updates a row by id using a JSON patch body.
func (h *Handler) UpdatePostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")
	id := c.Param("id")

	// Validate table name
	if !isValidIdentifier(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(data) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No data to update"})
		return
	}

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close()
	}

	setClauses := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data)+1)
	i := 1
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", quoteIdentifier(k), i))
		vals = append(vals, v)
		i++
	}
	vals = append(vals, id) // last placeholder is the WHERE id

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		quoteIdentifier(table),
		strings.Join(setClauses, ", "),
		i,
	)

	result, err := db.Exec(query, vals...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"database":      database,
		"table":         table,
		"id":            id,
		"rows_affected": rowsAffected,
	})
}

// DeletePostgresData removes a row by id.
func (h *Handler) DeletePostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")
	id := c.Param("id")

	// Validate table name
	if !isValidIdentifier(table) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid table name"})
		return
	}

	// Get connection for the specified database
	db, isTemp, err := h.DB.GetPostgresConnection(database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if isTemp {
		defer db.Close()
	}

	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", quoteIdentifier(table))
	result, err := db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"database":      database,
		"table":         table,
		"id":            id,
		"rows_affected": rowsAffected,
	})
}
