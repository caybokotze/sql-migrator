package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gookit/color"
)

func runMigrations(options DatabaseOptions) {
	createSchemaVersionTableIfNotExist(options)
	migrations := findMigrationToExecute(options)
	executeMigrations(options, migrations)
}

func findMigrationToExecute(options DatabaseOptions) []Schema {
	executedMigrations := fetchMigrationsFromDb(options)
	allFileMigrations := getArrayOfMigrationFilesWithoutDuplicates(options)
	return findUnexecutedMigrations(allFileMigrations, executedMigrations)
}

func getArrayOfMigrationFilesWithoutDuplicates(options DatabaseOptions) []Schema {
	items, err := os.ReadDir(options.MigrationDirectory)
	if err != nil {
		log.Fatalf("Could not read migration directory '%s': %s", options.MigrationDirectory, err.Error())
	}
	var schemas []Schema
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		schemas = append(schemas, generateSchemaFromFileName(item.Name()))
	}
	schemas = removeDuplicateSchemas(schemas)
	return schemas
}

// rollbackMigrations fetches executed migrations from the database,
// reverses their order, and executes down scripts from the most recent
// migration down to and including the specified rollbackId.
func rollbackMigrations(options DatabaseOptions, rollbackId string) {
	executedMigrations := fetchMigrationsFromDb(options)
	if len(executedMigrations) == 0 {
		color.Cyan.Println("No migrations to rollback.")
		return
	}

	rollbackIdAsInt, err := strconv.ParseInt(rollbackId, 10, 64)
	if err != nil {
		log.Fatalf("Invalid rollback ID '%s': must be a valid migration timestamp.", rollbackId)
	}

	// Sort in descending order (most recent first)
	sort.Slice(executedMigrations, func(i, j int) bool {
		return executedMigrations[i].id > executedMigrations[j].id
	})

	// Collect migrations to rollback (from most recent to target, inclusive)
	var migrationsToRollback []Schema
	for _, schema := range executedMigrations {
		migrationsToRollback = append(migrationsToRollback, schema)
		if schema.id == rollbackIdAsInt {
			break
		}
	}

	// Verify target migration was found
	if len(migrationsToRollback) == 0 || migrationsToRollback[len(migrationsToRollback)-1].id != rollbackIdAsInt {
		color.Red.Printf("Migration ID %d not found in executed migrations.\n", rollbackIdAsInt)
		color.Cyan.Println("Available migrations:")
		for _, m := range executedMigrations {
			fmt.Printf("  %d_%s\n", m.id, m.name)
		}
		os.Exit(1)
	}

	// Create single database connection for all rollback operations
	db := createDbConnection(options)
	defer db.Conn.Close()

	color.Cyan.Printf("Rolling back %d migration(s)...\n", len(migrationsToRollback))

	for _, schema := range migrationsToRollback {
		color.Cyan.Printf("Rolling back: %d_%s\n", schema.id, schema.name)

		downScript := readSchemaContent(schema, false, options.MigrationDirectory)
		if err := command(db, downScript); err != nil {
			color.Red.Printf("Failed to rollback migration %d_%s: %s\n", schema.id, schema.name, err.Error())
			os.Exit(1)
		}

		removeSchemaVersion(db, schema)
		color.Green.Printf("Rolled back: %d_%s\n", schema.id, schema.name)
	}

	color.Green.Println("Rollback completed successfully.")
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

func generateSchemaFromFileName(fileName string) Schema {
	s := strings.Split(fileName, "_")
	arrayLength := len(s)
	id, err := strconv.ParseInt(s[0], 10, 64)
	if err != nil {
		panic("id in file string can not be converted to integer")
	}
	return Schema{
		id:           id,
		name:         strings.Join(s[1:arrayLength-1], ""),
		dateExecuted: time.Now(),
	}
}

func createSchemaVersionTableIfNotExist(options DatabaseOptions) {
	db := createDbConnection(options)
	defer db.Conn.Close()
	err := command(db, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (", options.MigrationTableName)+
		"id BIGINT NOT NULL AUTO_INCREMENT, "+
		"name VARCHAR(255) NULL, "+
		"date_executed DATETIME DEFAULT CURRENT_TIMESTAMP, "+
		"PRIMARY KEY (id));")
	if err != nil {
		panic(err.Error())
	}
}

func executeMigrations(options DatabaseOptions, schemas []Schema) {
	if len(schemas) == 0 {
		color.Cyan.Println("All migrations are up to date.")
		return
	}

	db := createDbConnection(options)
	defer db.Conn.Close()

	color.Cyan.Printf("Executing %d pending migration(s)...\n", len(schemas))

	for _, schema := range schemas {
		color.Cyan.Printf("Executing: %s\n", getSchemaFileName(schema))

		upScript := readSchemaContent(schema, true, options.MigrationDirectory)
		if err := command(db, upScript); err != nil {
			if options.AutoByPass {
				color.Yellow.Printf("Bypassing failed migration %s: %s\n", getSchemaFileName(schema), err.Error())
				insertSchemaVersion(db, schema)
				continue
			}
			color.Red.Printf("Failed to execute migration %s: %s\n", getSchemaFileName(schema), err.Error())
			os.Exit(1)
		}

		insertSchemaVersion(db, schema)
		color.Green.Printf("Executed: %s\n", getSchemaFileName(schema))
	}

	color.Green.Println("All migrations executed successfully.")
}

func readSchemaContent(schema Schema, isUp bool, migrationDirectory string) string {
	var fileName string
	if isUp {
		fileName = getSchemaUpScript(getSchemaFileName(schema))
	} else {
		fileName = getSchemaDownScript(getSchemaFileName(schema))
	}

	filePath := fmt.Sprintf("%s/%s", migrationDirectory, fileName)
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Could not read file '%s': %s", filePath, err.Error())
	}
	if len(strings.TrimSpace(string(content))) == 0 {
		log.Fatalf("The migration file '%s' is empty", fileName)
	}
	return string(content)
}

func getSchemaFileName(schema Schema) string {
	return fmt.Sprintf("%s_%s", strconv.FormatInt(schema.id, 10), schema.name)
}

func getSchemaUpScript(fileName string) string {
	return fmt.Sprintf("%s_%s", fileName, "up.sql")
}

func getSchemaDownScript(fileName string) string {
	return fmt.Sprintf("%s_%s", fileName, "down.sql")
}

func insertSchemaVersion(dbConnectionWithOptions ConnectionWithOptions, schema Schema) {
	sqlText := fmt.Sprintf("INSERT INTO %s VALUES (%d, '%s', '%s');",
		dbConnectionWithOptions.MigrationTableName,
		schema.id,
		schema.name,
		schema.dateExecuted.Format("2006-01-02T15:04:05"))

	err := command(dbConnectionWithOptions, sqlText)
	if err != nil {
		panic(err.Error())
	}
}

func removeSchemaVersion(dbConnectionWithOptions ConnectionWithOptions, schema Schema) {
	sqlText := fmt.Sprintf("DELETE FROM %s WHERE id = %d;",
		dbConnectionWithOptions.MigrationTableName,
		schema.id)
	err := command(dbConnectionWithOptions, sqlText)
	if err != nil {
		panic(err.Error())
	}
}

func fetchMigrationsFromDb(details DatabaseOptions) []Schema {
	db := createDbConnection(details)
	defer db.Conn.Close()
	results := query(db, fmt.Sprintf("SELECT id, name, date_executed FROM %s;", details.MigrationTableName))

	var schemas []Schema

	for results.Next() {
		var schema Schema
		var dateExecuted rawTime
		err := results.Scan(&schema.id, &schema.name, &dateExecuted)
		if err != nil {
			panic(err.Error())
		}
		expectedTime, dateErr := dateExecuted.Parse()
		if dateErr != nil {
			panic("The datetime was in an unexpected format. expected: YYYY-MM-DD hh:mm:ss")
		}
		schema.dateExecuted = expectedTime
		schemas = append(schemas, schema)
	}
	if len(schemas) > 1 {
		sort.Slice(schemas, func(i, j int) bool {
			return schemas[i].id < schemas[j].id
		})
	}

	return schemas
}
