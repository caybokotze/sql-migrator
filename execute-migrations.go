package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gookit/color"
	"io/ioutil"
	"log"
	"os"
	_ "reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func runMigrations(
	details DatabaseOptions) {
	createSchemaVersionTableIfNotExist(details)
	migrations := findMigrationToExecute(details)
	executeMigrations(details, migrations)
}

func findMigrationToExecute(options DatabaseOptions) []Schema {
	executedMigrations := fetchMigrationsFromDb(options)
	allFileMigrations := getArrayOfMigrationFilesWithoutDuplicates()
	return findUnexecutedMigrations(allFileMigrations, executedMigrations)
}

func getArrayOfMigrationFilesWithoutDuplicates() []Schema {
	items, _ := ioutil.ReadDir("./scripts/")
	var schemas []Schema
	for _, item := range items {
		schemas = append(schemas, generateSchemaFromFileName(item.Name()))
	}
	schemas = removeDuplicateSchemas(schemas)
	return schemas
}

/*
	Summary: Fetches the migrations in the db which have already executed.
	Reverse the order of those migrations.
	Execute the down scripts for those migrations until the rollbackId is matched.
 */
func rollbackMigrations(options DatabaseOptions, rollbackId string) {
	executedMigrations := fetchMigrationsFromDb(options)
	var rollbackIdAsInt, err = strconv.ParseInt(rollbackId, 10, 64)
	if err != nil {
		log.Fatal("Could not convert rollback id to a int 64.")
	}
	// reverse order...
	if len(executedMigrations) > 1 {
		sort.Slice(executedMigrations, func(i, j int) bool {
			return executedMigrations[i].id > executedMigrations[j].id
		})
	}

	var rollbackMigrationArray []Schema
	var foundMigrationToRollbackTo = false
	for _, schema := range executedMigrations {
		rollbackMigrationArray = append(rollbackMigrationArray, schema)
		if schema.id == rollbackIdAsInt {
			foundMigrationToRollbackTo = true
			break
		}
	}
	if foundMigrationToRollbackTo {
		for _, schema := range rollbackMigrationArray {
			db := createDbConnection(options)
			err := command(db, readSchemaContent(schema, false))
			if err != nil {
				panic(err.Error())
			}
			removeSchemaVersion(db, schema)
			color.Green.Println(fmt.Sprintf("Rolled back migration %d_%s successfully.", schema.id, schema.name))
		}
	}
	if !foundMigrationToRollbackTo {
		color.Cyan.Println("Cannot rollback to this migration.")
		os.Exit(0)
	}
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

func findUnexecutedMigrations(fileMigrations []Schema, dbMigrations []Schema) []Schema {
	occurred := map[int64]bool{}
	var unique []Schema
	for _, v := range dbMigrations {
		occurred[v.id] = true
	}
	for _, v := range fileMigrations {
		if occurred[v.id] == true {
			continue
		}
		unique = append(unique, v)
	}
	return unique
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
	db := createDbConnection(options)
	defer db.Conn.Close()
	if len(schemas) == 0 {
		color.Cyan.Println("All up to date...")
		os.Exit(0)
	}
	for _, s := range schemas {
		color.Cyan.Println(fmt.Sprintf("Preparing to execute migration: %s", getSchemaFileName(s)))
		err := command(db, readSchemaContent(s, true))
		// todo: Code to handle autoByPass...
		if err != nil {
			panic(err)
		}
		insertSchemaVersion(db, s)
		color.Green.Println(getSchemaFileName(s), "executed successfully...")
	}
}

func readSchemaContent(schema Schema, isUp bool) string {
	var fileName = ""
	if isUp {
		fileName = getSchemaUpScript(getSchemaFileName(schema))
	}
	if !isUp {
		fileName = getSchemaDownScript(getSchemaFileName(schema))
	}
	content, err := ioutil.ReadFile(fmt.Sprintf("./scripts/%s", fileName))
	if len(strings.TrimSpace(string(content))) == 0 {
		log.Fatalf("The migration file %s is empty", fileName)
	}
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not open file, %s", fileName))
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