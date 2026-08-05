package main

import (
	"database/sql"
	"fmt"

	"github.com/gookit/color"
)

// SchemaTarget is one schema produced by the discovery query, optionally
// with its own connection string when the schema lives on a different
// server or database than the one the discovery query ran against.
type SchemaTarget struct {
	Name          string
	ConnectionDsn string
}

// discoverSchemas connects to the configured database (without a schema
// override) and runs the configured schema discovery query.
//
// The first column of every row is taken as a schema name. If the query
// returns a second column, a non-empty value is treated as the full DSN to
// use for that schema (postgres://user:pass@host:port/database); an empty or
// NULL value falls back to the base connection. Further columns are ignored.
func discoverSchemas(options DatabaseOptions) ([]SchemaTarget, error) {
	options.SqlSchema = ""

	db, err := createDbConnection(options)
	if err != nil {
		return nil, err
	}
	defer db.Conn.Close()

	rows, err := db.Conn.Query(options.SchemaQuery)
	if err != nil {
		return nil, fmt.Errorf("schema discovery query failed: %s", err.Error())
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var targets []SchemaTarget
	for rows.Next() {
		values := make([]interface{}, len(columns))
		var schemaName string
		var connectionDsn sql.NullString
		values[0] = &schemaName
		if len(columns) > 1 {
			values[1] = &connectionDsn
		}
		for i := 2; i < len(columns); i++ {
			values[i] = new(interface{})
		}
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("schema discovery query must return a schema name as its first column (and optionally a connection string as its second): %s", err.Error())
		}
		if schemaName == "" {
			continue
		}
		targets = append(targets, SchemaTarget{
			Name:          schemaName,
			ConnectionDsn: connectionDsn.String,
		})
	}
	return targets, rows.Err()
}

// runFanOut discovers all schemas and applies the given action to each one.
// A failing schema does not stop the remaining schemas; the error returned
// at the end reflects whether any schema failed.
func runFanOut(options DatabaseOptions, actionName string, action func(DatabaseOptions) error) error {
	dialect, err := resolveDialect(options.SqlDriver)
	if err != nil {
		return err
	}
	if !dialect.SupportsSchemas() {
		return fmt.Errorf("fan-out requires a driver with schema support (postgres) — the %s driver treats databases as schemas", dialect.DriverName())
	}
	if options.SchemaQuery == "" {
		return fmt.Errorf("fan-out requires a schema discovery query — set 'schemaQuery' in the configuration file or pass --schema-query")
	}

	targets, err := discoverSchemas(options)
	if err != nil {
		return err
	}
	if len(targets) == 0 {
		color.Cyan.Println("Schema discovery returned no schemas — nothing to do.")
		return nil
	}

	color.Cyan.Printf("Fanning out '%s' over %d schema(s)...\n", actionName, len(targets))

	var failed []string
	for _, target := range targets {
		if target.ConnectionDsn != "" {
			color.Cyan.Printf("\n==> %s (own connection)\n", target.Name)
		} else {
			color.Cyan.Printf("\n==> %s\n", target.Name)
		}

		schemaOptions := options
		schemaOptions.SqlSchema = target.Name
		if target.ConnectionDsn != "" {
			if err := dialect.ValidateDsn(target.ConnectionDsn); err != nil {
				failed = append(failed, target.Name)
				color.Red.Printf("%s: FAILED — the connection string from the discovery query is invalid: %s\n", target.Name, err.Error())
				continue
			}
			schemaOptions.SqlDsn = target.ConnectionDsn
		}

		if err := action(schemaOptions); err != nil {
			failed = append(failed, target.Name)
			color.Red.Printf("%s: FAILED — %s\n", target.Name, err.Error())
			continue
		}
	}

	fmt.Println()
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d schema(s) failed: %v", len(failed), len(targets), failed)
	}
	color.Green.Printf("Fan-out '%s' completed across %d schema(s).\n", actionName, len(targets))
	return nil
}
