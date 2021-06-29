package main

import (
	"database/sql"
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

var databaseOptions DatabaseOptions

func runMigrations(
	details DatabaseOptions) {
	databaseOptions = details
	createSchemaVersionTable(details)
	migrations := findMigrationToExecute(details)
	executeMigrations(details, migrations)
}

func findMigrationToExecute(details DatabaseOptions) []Schema {
	executedMigrations := fetchMigrationsFromDb(details)
	allFileMigrations := getArrayOfMigrationFilesWithoutDuplicates()
	return findUniqueMigrations(allFileMigrations, executedMigrations)
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

func findUniqueMigrations(fileMigrations []Schema, dbMigrations []Schema) []Schema {
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
	id, err := strconv.ParseInt(s[0], 0, 64)
	if err != nil {
		panic("id in file string can not be converted to integer")
	}
	return Schema{
		id:           id,
		name:         s[1],
		dateexecuted: time.Now(),
	}
}

func createSchemaVersionTable(options DatabaseOptions) {
	db := createDbConnection(options)
	defer db.Close()
	_, err := command(db, "CREATE TABLE IF NOT EXISTS __schema_versioning ("+
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
	defer db.Close()
	if len(schemas) == 0 {
		color.Cyan.Println("All up to date...")
		os.Exit(0)
	}
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].id < schemas[j].id
	})
	for _, s := range schemas {
		_, err := command(db, readSchemaContent(s))
		// todo: Code to handle autoByPass...
		if err != nil {
			panic(err)
		}
		insertSchemaVersion(db, s)
		color.Green.Println(getSchemaFileName(s), "executed successfully...")
	}
}

func readSchemaContent(schema Schema) string {
	fileName := getSchemaUpScript(getSchemaFileName(schema))
	content, err := ioutil.ReadFile(fmt.Sprintf("./scripts/%s", fileName))
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not open file, %s", fileName))
	}
	return string(content)
}

func backtrackMigrations() {
	// todo: Code to roll back migrations
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

func insertSchemaVersion(db *sql.DB, schema Schema) {
	// todo: make the table renaming possible for the user.
	sqlText := fmt.Sprintf("INSERT INTO __schema_versioning VALUES (%d, '%s', '%s');",
		schema.id,
		schema.name,
		schema.dateexecuted.Format("2006-01-02T15:04:05"))

	_, err := command(db, sqlText)
	if err != nil {
		panic(err.Error())
	}
}

func fetchMigrationsFromDb(details DatabaseOptions) []Schema {
	db := createDbConnection(details)
	defer db.Close()
	results := query(db, "SELECT id, name, date_executed FROM __schema_versioning")

	var schemas []Schema

	for results.Next() {
		var schema Schema
		var dateExecuted rawTime
		err := results.Scan(&schema.id, &schema.name, &dateExecuted)
		if err != nil {
			panic(err.Error())
		}
		expectedTime, error := dateExecuted.Parse()
		if error != nil {
			panic("The datetime was in an unexpected format. expected: YYYY-MM-DD hh:mm:ss")
		}
		schema.dateexecuted = expectedTime
		schemas = append(schemas, schema)
	}
	return schemas
}