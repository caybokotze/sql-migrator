package main

import (
	"database/sql"
	"fmt"
	"net/url"
	"strings"
)

const (
	DriverMysql    = "mysql"
	DriverPostgres = "postgres"
)

// Dialect abstracts the differences between supported database engines:
// how to build a DSN, how the migration tracking table looks, and how
// rows in it are read and written.
type Dialect interface {
	// DriverName is the name registered with database/sql.
	DriverName() string
	// BuildDsn builds a driver-specific connection string from the options.
	BuildDsn(options DatabaseOptions) string
	// SupportsSchemas reports whether the engine has schemas as a namespace
	// within a database (true for postgres, false for mysql where schema == database).
	SupportsSchemas() bool
	// SupportsTransactionalDdl reports whether DDL can be rolled back. When
	// true, --dry-run executes the whole pending batch in one transaction and
	// rolls it back at the end, so migrations that depend on each other can
	// still be validated together.
	SupportsTransactionalDdl() bool
	// EnsureSchema creates the target schema if it does not exist yet.
	EnsureSchema(db *sql.DB, schema string) error
	// CreateMigrationTableSql returns the DDL for the tracking table.
	CreateMigrationTableSql(tableName string) string
	// FetchMigrationsSql returns the query that lists applied migrations.
	FetchMigrationsSql(tableName string) string
	// ScanMigration reads one row produced by FetchMigrationsSql.
	ScanMigration(rows *sql.Rows) (Schema, error)
	// InsertMigrationSql / InsertMigrationArgs record a migration as applied.
	InsertMigrationSql(tableName string) string
	InsertMigrationArgs(schema Schema) []interface{}
	// DeleteMigrationSql / DeleteMigrationArgs remove a migration record on rollback.
	DeleteMigrationSql(tableName string) string
	DeleteMigrationArgs(schema Schema) []interface{}
}

func resolveDialect(driver string) (Dialect, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", DriverMysql, "mariadb":
		return mysqlDialect{}, nil
	case DriverPostgres, "postgresql", "pg":
		return postgresDialect{}, nil
	default:
		return nil, fmt.Errorf("unsupported driver '%s' (supported: mysql, postgres)", driver)
	}
}

// --- MySQL / MariaDB ---------------------------------------------------------
//
// The tracking table keeps its original sql-migrator shape
// (id = migration timestamp, name, date_executed) for backward compatibility
// with databases already managed by this tool.

type mysqlDialect struct{}

func (mysqlDialect) DriverName() string { return "mysql" }

func (mysqlDialect) BuildDsn(options DatabaseOptions) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		options.SqlUser,
		options.SqlPassword,
		options.SqlHost,
		options.SqlPort,
		options.SqlDatabase)
}

func (mysqlDialect) SupportsSchemas() bool { return false }

// MySQL DDL causes an implicit commit, so a dry run there is only an
// approximation — each statement's transaction is rolled back, but DDL
// side effects survive.
func (mysqlDialect) SupportsTransactionalDdl() bool { return false }

func (mysqlDialect) EnsureSchema(_ *sql.DB, schema string) error {
	return fmt.Errorf("the mysql driver does not support schemas — use --database instead of --schema (got schema '%s')", schema)
}

func (mysqlDialect) CreateMigrationTableSql(tableName string) string {
	return fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", tableName) +
		"id BIGINT NOT NULL AUTO_INCREMENT, " +
		"name VARCHAR(255) NULL, " +
		"date_executed DATETIME DEFAULT CURRENT_TIMESTAMP, " +
		"PRIMARY KEY (id));"
}

func (mysqlDialect) FetchMigrationsSql(tableName string) string {
	return fmt.Sprintf("SELECT id, name, date_executed FROM %s;", tableName)
}

func (mysqlDialect) ScanMigration(rows *sql.Rows) (Schema, error) {
	var schema Schema
	var dateExecuted rawTime
	if err := rows.Scan(&schema.id, &schema.name, &dateExecuted); err != nil {
		return Schema{}, err
	}
	parsed, err := dateExecuted.Parse()
	if err != nil {
		return Schema{}, fmt.Errorf("unexpected datetime format (expected YYYY-MM-DD hh:mm:ss): %s", err.Error())
	}
	schema.dateExecuted = parsed
	return schema, nil
}

func (mysqlDialect) InsertMigrationSql(tableName string) string {
	return fmt.Sprintf("INSERT INTO %s (id, name, date_executed) VALUES (?, ?, ?);", tableName)
}

func (mysqlDialect) InsertMigrationArgs(schema Schema) []interface{} {
	return []interface{}{schema.id, schema.name, schema.dateExecuted.Format("2006-01-02T15:04:05")}
}

func (mysqlDialect) DeleteMigrationSql(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE id = ?;", tableName)
}

func (mysqlDialect) DeleteMigrationArgs(schema Schema) []interface{} {
	return []interface{}{schema.id}
}

// --- PostgreSQL --------------------------------------------------------------
//
// The tracking table is keyed by the migration file name instead of the
// parsed timestamp. This is the shape most migration runners converge on,
// and it lets the tool adopt tables created by other filename-tracking
// runners without a state migration:
//
//	id          BIGSERIAL PRIMARY KEY
//	script_name VARCHAR(255) NOT NULL UNIQUE   -- e.g. 20250319205322_CreateUsers_up.sql
//	applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()

type postgresDialect struct{}

func (postgresDialect) DriverName() string { return "postgres" }

func (postgresDialect) BuildDsn(options DatabaseOptions) string {
	sslMode := options.SqlSslMode
	if sslMode == "" {
		sslMode = "disable"
	}

	query := url.Values{}
	query.Set("sslmode", sslMode)

	// The search_path is part of the connection string so that every pooled
	// connection resolves unqualified names in the target schema. A plain
	// `SET search_path` after connecting would only apply to one connection
	// in the pool.
	if options.SqlSchema != "" {
		query.Set("options", fmt.Sprintf("-csearch_path=%s,public", options.SqlSchema))
	}

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(options.SqlUser, options.SqlPassword),
		Host:     fmt.Sprintf("%s:%s", options.SqlHost, options.SqlPort),
		Path:     "/" + options.SqlDatabase,
		RawQuery: query.Encode(),
	}
	return u.String()
}

func (postgresDialect) SupportsSchemas() bool { return true }

func (postgresDialect) SupportsTransactionalDdl() bool { return true }

func (postgresDialect) EnsureSchema(db *sql.DB, schema string) error {
	// Without this, an unqualified CREATE TABLE lands in the first *existing*
	// schema on the search_path (usually public) instead of the target schema.
	_, err := db.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, quotePgIdentifier(schema)))
	return err
}

func (postgresDialect) CreateMigrationTableSql(tableName string) string {
	return fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
    id          BIGSERIAL PRIMARY KEY,
    script_name VARCHAR(255) NOT NULL UNIQUE,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`, tableName)
}

func (postgresDialect) FetchMigrationsSql(tableName string) string {
	return fmt.Sprintf("SELECT script_name, applied_at FROM %s;", tableName)
}

func (postgresDialect) ScanMigration(rows *sql.Rows) (Schema, error) {
	var scriptName string
	var appliedAt sql.NullTime
	if err := rows.Scan(&scriptName, &appliedAt); err != nil {
		return Schema{}, err
	}
	schema, err := parseSchemaFromFileName(scriptName)
	if err != nil {
		return Schema{}, fmt.Errorf("unparseable script_name '%s' in migration table: %s", scriptName, err.Error())
	}
	schema.scriptName = scriptName
	if appliedAt.Valid {
		schema.dateExecuted = appliedAt.Time
	}
	return schema, nil
}

func (postgresDialect) InsertMigrationSql(tableName string) string {
	return fmt.Sprintf("INSERT INTO %s (script_name) VALUES ($1);", tableName)
}

func (postgresDialect) InsertMigrationArgs(schema Schema) []interface{} {
	return []interface{}{schema.UpScriptName()}
}

func (postgresDialect) DeleteMigrationSql(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s WHERE script_name = $1;", tableName)
}

func (postgresDialect) DeleteMigrationArgs(schema Schema) []interface{} {
	return []interface{}{schema.UpScriptName()}
}

func quotePgIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
