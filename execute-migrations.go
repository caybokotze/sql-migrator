package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	_ "reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

func runMigrations(
	details DatabaseOptions) {
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
		schemas = append(schemas, getSchemaFromFileName(item.Name()))
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

func getSchemaFromFileName(fileName string) Schema {
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
	command(db, "CREATE TABLE IF NOT EXISTS schemaversion (" +
		"id BIGINT NOT NULL AUTO_INCREMENT, " +
		"name VARCHAR(512) NULL, " +
		"date_executed DATETIME DEFAULT CURRENT_TIMESTAMP, " +
		"PRIMARY KEY (id));")
}

func executeMigrations(options DatabaseOptions, schemas []Schema) {
	db := createDbConnection(options)
	defer db.Close()
	sort.Slice(schemas, func(i, j int) bool {
		return schemas[i].id < schemas[j].id
	})
	for _, s := range schemas {
		command(db, readSchemaContent(s))
	}
}

func readSchemaContent(schema Schema) string {
	fileName := fmt.Sprintf("%s_%s_%s", schema.id, schema.name, "up")
	content, err := ioutil.ReadFile(fmt.Sprintf("./scripts/%s", fileName))
	if err != nil {
		log.Fatal(fmt.Sprintf("Could not open file, %s", fileName))
	}
	return string(content)
}

func fetchMigrationsFromDb(details DatabaseOptions) []Schema {
	db := createDbConnection(details)
	defer db.Close()
	results := query(db, "SELECT id, name, dateexecuted FROM schemaversion")

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