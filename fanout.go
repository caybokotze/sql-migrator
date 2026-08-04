package main

import (
	"fmt"

	"github.com/gookit/color"
)

// discoverSchemas connects to the configured database (without a schema
// override) and runs the configured schema discovery query. The first
// column of every row is taken as a schema name; other columns are ignored.
func discoverSchemas(options DatabaseOptions) ([]string, error) {
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

	var schemas []string
	for rows.Next() {
		values := make([]interface{}, len(columns))
		var schemaName string
		values[0] = &schemaName
		for i := 1; i < len(columns); i++ {
			values[i] = new(interface{})
		}
		if err := rows.Scan(values...); err != nil {
			return nil, fmt.Errorf("schema discovery query must return a schema name as its first column: %s", err.Error())
		}
		if schemaName != "" {
			schemas = append(schemas, schemaName)
		}
	}
	return schemas, rows.Err()
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

	schemas, err := discoverSchemas(options)
	if err != nil {
		return err
	}
	if len(schemas) == 0 {
		color.Cyan.Println("Schema discovery returned no schemas — nothing to do.")
		return nil
	}

	color.Cyan.Printf("Fanning out '%s' over %d schema(s)...\n", actionName, len(schemas))

	var failed []string
	for _, schemaName := range schemas {
		color.Cyan.Printf("\n==> %s\n", schemaName)

		schemaOptions := options
		schemaOptions.SqlSchema = schemaName

		if err := action(schemaOptions); err != nil {
			failed = append(failed, schemaName)
			color.Red.Printf("%s: FAILED — %s\n", schemaName, err.Error())
			continue
		}
	}

	fmt.Println()
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d schema(s) failed: %v", len(failed), len(schemas), failed)
	}
	color.Green.Printf("Fan-out '%s' completed across %d schema(s).\n", actionName, len(schemas))
	return nil
}
