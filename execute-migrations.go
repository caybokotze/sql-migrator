package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

// utf8Bom is the three-byte UTF-8 byte order mark (EF BB BF). Some editors
// and tools write migration files with a leading BOM. Go's os.ReadFile keeps
// it verbatim, and PostgreSQL rejects a statement that starts with one, so
// every migration file read strips it before the content is used.
var utf8Bom = []byte{0xEF, 0xBB, 0xBF}

func stripUtf8Bom(content []byte) []byte {
	return bytes.TrimPrefix(content, utf8Bom)
}

// runMigrations applies all pending up migrations against one target
// (database, or one schema within it). It returns an error instead of
// exiting so that fan-out mode can continue with the remaining schemas.
func runMigrations(options DatabaseOptions) error {
	db, err := createDbConnection(options)
	if err != nil {
		return err
	}
	defer db.Conn.Close()

	if options.SqlSchema != "" {
		if err := db.Dialect.EnsureSchema(db.Conn, options.SqlSchema); err != nil {
			return fmt.Errorf("could not ensure schema '%s' exists: %s", options.SqlSchema, err.Error())
		}
	}

	if err := createSchemaVersionTableIfNotExist(db); err != nil {
		return fmt.Errorf("could not create migration table: %s", err.Error())
	}

	executedMigrations, err := fetchMigrationsFromDb(db)
	if err != nil {
		return err
	}

	fileMigrations, err := getArrayOfMigrationFilesWithoutDuplicates(options)
	if err != nil {
		return err
	}

	pending := findUnexecutedMigrations(fileMigrations, executedMigrations)
	return executeMigrations(db, options, pending)
}

func getArrayOfMigrationFilesWithoutDuplicates(options DatabaseOptions) ([]Schema, error) {
	items, err := os.ReadDir(options.MigrationDirectory)
	if err != nil {
		return nil, fmt.Errorf("could not read migration directory '%s': %s", options.MigrationDirectory, err.Error())
	}
	var schemas []Schema
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		if !strings.HasSuffix(item.Name(), ".sql") {
			continue
		}
		schema, err := parseSchemaFromFileName(item.Name())
		if err != nil {
			return nil, fmt.Errorf("invalid migration file name '%s': %s", item.Name(), err.Error())
		}
		schemas = append(schemas, schema)
	}
	schemas = removeDuplicateSchemas(schemas)
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].id < schemas[j].id
	})
	return schemas, nil
}

// rollbackMigrations executes down scripts from the most recent applied
// migration down to and including the migration with rollbackId. Applied
// migrations below the target are left alone. If nothing at or above the
// target is applied, the rollback is a no-op (which makes re-running a
// partially failed fan-out rollback safe).
func rollbackMigrations(options DatabaseOptions, rollbackId string) error {
	rollbackIdAsInt, err := strconv.ParseInt(rollbackId, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid rollback ID '%s': must be a valid migration timestamp", rollbackId)
	}

	db, err := createDbConnection(options)
	if err != nil {
		return err
	}
	defer db.Conn.Close()

	executedMigrations, err := fetchMigrationsFromDb(db)
	if err != nil {
		return err
	}
	if len(executedMigrations) == 0 {
		color.Cyan.Println("No migrations to rollback.")
		return nil
	}

	// The down scripts live on disk — pair every applied migration with its
	// file so the exact on-disk name (underscores intact) is used.
	fileMigrations, err := getArrayOfMigrationFilesWithoutDuplicates(options)
	if err != nil {
		return err
	}
	filesById := make(map[int64]Schema)
	for _, m := range fileMigrations {
		filesById[m.id] = m
	}

	// Sort in descending order (most recent first)
	sort.Slice(executedMigrations, func(i, j int) bool {
		return executedMigrations[i].id > executedMigrations[j].id
	})

	var migrationsToRollback []Schema
	targetFound := false
	for _, schema := range executedMigrations {
		if schema.id < rollbackIdAsInt {
			break
		}
		if schema.id == rollbackIdAsInt {
			targetFound = true
		}
		if file, ok := filesById[schema.id]; ok {
			schema.fileBase = file.fileBase
		} else {
			return fmt.Errorf("no migration file found for applied migration %s — cannot locate its down script", schema.DisplayName())
		}
		migrationsToRollback = append(migrationsToRollback, schema)
	}

	if len(migrationsToRollback) == 0 {
		color.Cyan.Printf("Nothing to rollback — no applied migration at or above %d.\n", rollbackIdAsInt)
		return nil
	}
	if !targetFound {
		color.Yellow.Printf("Note: migration %d itself is not applied — rolling back the %d applied migration(s) above it.\n",
			rollbackIdAsInt, len(migrationsToRollback))
	}

	color.Cyan.Printf("Rolling back %d migration(s)...\n", len(migrationsToRollback))

	for _, schema := range migrationsToRollback {
		color.Cyan.Printf("Rolling back: %s\n", schema.DisplayName())

		downScript, err := readSchemaContent(schema, false, options.MigrationDirectory)
		if err != nil {
			return err
		}
		if err := command(db, downScript); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %s", schema.DisplayName(), err.Error())
		}

		if err := removeSchemaVersion(db, schema); err != nil {
			return err
		}
		color.Green.Printf("Rolled back: %s\n", schema.DisplayName())
	}

	color.Green.Println("Rollback completed successfully.")
	return nil
}

func removeDuplicateSchemas(schemas []Schema) []Schema {
	var unique []Schema
primaryLoop:
	for _, v := range schemas {
		for i, u := range unique {
			if v.id == u.id {
				unique[i] = v
				continue primaryLoop
			}
		}
		unique = append(unique, v)
	}
	return unique
}

func findUnexecutedMigrations(fileMigrations, dbMigrations []Schema) []Schema {
	executed := make(map[int64]bool)
	for _, m := range dbMigrations {
		executed[m.id] = true
	}

	var pending []Schema
	for _, m := range fileMigrations {
		if !executed[m.id] {
			pending = append(pending, m)
		}
	}
	return pending
}

// parseSchemaFromFileName parses a migration file name of the form
// {timestamp}_{name}_up.sql / {timestamp}_{name}_down.sql (the .sql suffix
// is optional). Underscores inside the name are preserved in fileBase and
// collapsed in the display name.
func parseSchemaFromFileName(fileName string) (Schema, error) {
	base := strings.TrimSuffix(fileName, ".sql")
	switch {
	case strings.HasSuffix(base, "_up"):
		base = strings.TrimSuffix(base, "_up")
	case strings.HasSuffix(base, "_down"):
		base = strings.TrimSuffix(base, "_down")
	default:
		return Schema{}, fmt.Errorf("expected a name ending in _up.sql or _down.sql")
	}

	parts := strings.Split(base, "_")
	if len(parts) < 2 {
		return Schema{}, fmt.Errorf("expected the format {timestamp}_{name}_{up|down}.sql")
	}
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return Schema{}, fmt.Errorf("the leading timestamp '%s' is not numeric", parts[0])
	}
	return Schema{
		id:           id,
		name:         strings.Join(parts[1:], ""),
		fileBase:     base,
		dateExecuted: time.Now(),
	}, nil
}

func createSchemaVersionTableIfNotExist(db ConnectionWithOptions) error {
	// Deliberately not routed through command(): the tracking table is
	// idempotent bookkeeping and must exist even during --dry-run, where
	// command() rolls every transaction back (postgres DDL is transactional,
	// so the rollback would undo the CREATE TABLE).
	_, err := db.Conn.Exec(db.Dialect.CreateMigrationTableSql(db.MigrationTableName))
	return err
}

func executeMigrations(db ConnectionWithOptions, options DatabaseOptions, schemas []Schema) error {
	if len(schemas) == 0 {
		color.Cyan.Println("All migrations are up to date.")
		return nil
	}

	if options.DryRun && db.Dialect.SupportsTransactionalDdl() {
		return dryRunMigrations(db, options, schemas)
	}

	color.Cyan.Printf("Executing %d pending migration(s)...\n", len(schemas))

	for _, schema := range schemas {
		color.Cyan.Printf("Executing: %s\n", schema.DisplayName())

		upScript, err := readSchemaContent(schema, true, options.MigrationDirectory)
		if err != nil {
			return err
		}
		if err := command(db, upScript); err != nil {
			if options.AutoByPass {
				color.Yellow.Printf("Bypassing failed migration %s: %s\n", schema.DisplayName(), err.Error())
				if insertErr := insertSchemaVersion(db, schema); insertErr != nil {
					return insertErr
				}
				continue
			}
			return fmt.Errorf("failed to execute migration %s: %s", schema.DisplayName(), err.Error())
		}

		if err := insertSchemaVersion(db, schema); err != nil {
			return err
		}
		color.Green.Printf("Executed: %s\n", schema.DisplayName())
	}

	color.Green.Println("All migrations executed successfully.")
	return nil
}

// dryRunMigrations executes the whole pending batch inside one transaction
// and rolls it back at the end. Later migrations can therefore see the
// effects of earlier ones, exactly as a real run would.
func dryRunMigrations(db ConnectionWithOptions, options DatabaseOptions, schemas []Schema) error {
	color.Cyan.Printf("Dry run: executing %d pending migration(s) in one transaction...\n", len(schemas))

	transaction, err := db.Conn.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = transaction.Rollback() }()

	for _, schema := range schemas {
		color.Cyan.Printf("Executing: %s\n", schema.DisplayName())

		upScript, err := readSchemaContent(schema, true, options.MigrationDirectory)
		if err != nil {
			return err
		}
		if _, err := transaction.Exec(upScript); err != nil {
			return fmt.Errorf("failed to execute migration %s: %s", schema.DisplayName(), err.Error())
		}
		if _, err := transaction.Exec(db.Dialect.InsertMigrationSql(db.MigrationTableName), db.Dialect.InsertMigrationArgs(schema)...); err != nil {
			return fmt.Errorf("could not record migration %s as applied: %s", schema.DisplayName(), err.Error())
		}
		color.Green.Printf("Executed: %s\n", schema.DisplayName())
	}

	if err := transaction.Rollback(); err != nil {
		return err
	}
	color.Green.Printf("Dry run complete — %d migration(s) executed and rolled back.\n", len(schemas))
	return nil
}

func readSchemaContent(schema Schema, isUp bool, migrationDirectory string) (string, error) {
	direction := "down"
	if isUp {
		direction = "up"
	}
	fileName := fmt.Sprintf("%s_%s.sql", schema.fileBase, direction)

	filePath := fmt.Sprintf("%s/%s", migrationDirectory, fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("could not read file '%s': %s", filePath, err.Error())
	}
	content = stripUtf8Bom(content)
	if len(strings.TrimSpace(string(content))) == 0 {
		return "", fmt.Errorf("the migration file '%s' is empty", fileName)
	}
	return string(content), nil
}

func insertSchemaVersion(db ConnectionWithOptions, schema Schema) error {
	err := command(db, db.Dialect.InsertMigrationSql(db.MigrationTableName), db.Dialect.InsertMigrationArgs(schema)...)
	if err != nil {
		return fmt.Errorf("could not record migration %s as applied: %s", schema.DisplayName(), err.Error())
	}
	return nil
}

func removeSchemaVersion(db ConnectionWithOptions, schema Schema) error {
	err := command(db, db.Dialect.DeleteMigrationSql(db.MigrationTableName), db.Dialect.DeleteMigrationArgs(schema)...)
	if err != nil {
		return fmt.Errorf("could not remove migration record %s: %s", schema.DisplayName(), err.Error())
	}
	return nil
}

func fetchMigrationsFromDb(db ConnectionWithOptions) ([]Schema, error) {
	results, err := db.Conn.Query(db.Dialect.FetchMigrationsSql(db.MigrationTableName))
	if err != nil {
		return nil, fmt.Errorf("could not read the migration table '%s': %s", db.MigrationTableName, err.Error())
	}
	defer results.Close()

	var schemas []Schema
	for results.Next() {
		schema, err := db.Dialect.ScanMigration(results)
		if err != nil {
			return nil, err
		}
		schemas = append(schemas, schema)
	}
	if err := results.Err(); err != nil {
		return nil, err
	}

	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].id < schemas[j].id
	})
	return schemas, nil
}
