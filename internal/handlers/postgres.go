package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

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
	const q = `
		SELECT table_name
		FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE'
	`

	rows, err := h.DB.PostgresDB.Query(q)
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

// ExecutePostgresQuery runs arbitrary SQL and returns columns + rows.
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

	rows, err := db.Query(req.Query)
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
}

// SelectPostgresData returns rows from a table with an optional row limit.
func (h *Handler) SelectPostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")
	limit := c.DefaultQuery("limit", "100")

	rows, err := h.DB.PostgresDB.Query("SELECT * FROM " + table + " LIMIT " + limit)
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

// InsertPostgresData inserts a JSON object as a new row, returning the new id.
func (h *Handler) InsertPostgresData(c *gin.Context) {
	if h.DB.PostgresDB == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "PostgreSQL not enabled"})
		return
	}

	database := c.Param("database")
	table := c.Param("table")

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cols := make([]string, 0, len(data))
	placeholders := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data))
	i := 1
	for k, v := range data {
		cols = append(cols, k)
		placeholders = append(placeholders, fmt.Sprintf("$%d", i))
		vals = append(vals, v)
		i++
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s) RETURNING id",
		table,
		strings.Join(cols, ", "),
		strings.Join(placeholders, ", "),
	)

	var insertedID interface{}
	if err := h.DB.PostgresDB.QueryRow(query, vals...).Scan(&insertedID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	setClauses := make([]string, 0, len(data))
	vals := make([]interface{}, 0, len(data)+1)
	i := 1
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", k, i))
		vals = append(vals, v)
		i++
	}
	vals = append(vals, id) // last placeholder is the WHERE id

	query := fmt.Sprintf(
		"UPDATE %s SET %s WHERE id = $%d",
		table,
		strings.Join(setClauses, ", "),
		i,
	)

	result, err := h.DB.PostgresDB.Exec(query, vals...)
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

	result, err := h.DB.PostgresDB.Exec("DELETE FROM "+table+" WHERE id = $1", id)
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
