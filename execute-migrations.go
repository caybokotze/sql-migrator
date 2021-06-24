package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

func runMigrations(
	sqlUser string,
	sqlPassword string,
	sqlHost string,
	sqlPort string,
	sqlDatabase string,
	dryRun bool,
	autoByPass bool) {

	databaseDetails := DatabaseOptions{
		sqlUser: sqlUser,
		sqlPassword: sqlPassword,
		sqlHost: sqlHost,
		sqlPort: sqlPort,
		sqlDatabase: sqlDatabase,
	}

	createSchemaVersionTable(
		sqlUser,
		sqlPassword,
		sqlHost,
		sqlPort,
		sqlDatabase)

	migrations := findMigrationToExecute(databaseDetails)

	//var excludedMigrations = findExcludedMigrations()
}

type DatabaseOptions struct {
	sqlUser string
	sqlPassword string
	sqlHost string
	sqlPort string
	sqlDatabase string
}

// this could be more efficient than the findUniqueMigrations method, but still needs work.
func findExcludedMigrations() []Schema {
	executedMigrations := getAllDbMigrations()
	allFileMigrations := getArrayOfMigrationFiles()
	//lastMigrationRun := executedMigrations[len(executedMigrations)-1]

	m := make(map[Schema]int64)
	for _, k := range executedMigrations {
		m[k] |= 1 << 0
	}
	for _, k := range allFileMigrations {
		m[k] |= 1 << 0
	}
	var result []Schema
	for k, v := range m {
		a := v&(1<<0) != 0
		b := v&(1<<1) != 0
		switch {
			case !a && b:
				result = append(result, k)
		}
	}

	return allFileMigrations
}

func findMigrationToExecute(details DatabaseOptions) []Schema {
	executedMigrations := getAllDbMigrations(details)
	allFileMigrations := getArrayOfMigrationFiles()
	allFileMigrations = removeDuplicateSchemas(allFileMigrations)
	allMigrations := append(executedMigrations, allFileMigrations...)
	return findUniqueMigrations(allMigrations)

}

func getArrayOfMigrationFiles() []Schema {
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

// note: here we assume that any migration that only appears once in our list, has not executed and therefore should be.
func findUniqueMigrations(schemas []Schema) []Schema {
	occurred := map[int64]int{}
	var filtered []Schema
	var unique []Schema
	for e := range schemas {
		if occurred[schemas[e].id] == 0 {
			occurred[schemas[e].id] = 1
			filtered = append(filtered, schemas[e])
			continue
		}
		if occurred[schemas[e].id] == 1 {
			occurred[schemas[e].id] = 2
			filtered = append(filtered, schemas[e])
			continue
		}
	}
	for e := range filtered {
		if occurred[filtered[e].id] == 1 {
			unique = append(unique, filtered[e])
		}
	}
	return unique
}

func remove(slice []Schema, s Schema) []Schema {
	return append(slice[:s.id], slice[s.id+1:]...)
}

func getSchemaFromFileName(fileName string) Schema {
	s := strings.Split(fileName, "_")
	id, err := strconv.ParseInt(s[0], 0, 64)
	name := s[1]
	if err != nil {
		panic("id in file string can not be converted to integer")
	}
	return Schema{
		id:           id,
		name:         name,
		dateexecuted: time.Now(),
	}
}

func createSchemaVersionTable(
	dbUser string,
	dbPassword string,
	ipAddress string,
	port string,
	database string,
) {
	createSchemaVersion := fmt.Sprintf("USE %s; " +
		"CREATE TABLE IF NOT EXISTS schemaversion (" +
		"id BIGINT NOT NULL AUTO_INCREMENT, " +
		"name VARCHAR(512) NULL, " +
		"date_executed DATETIME DEFAULT CURRENT_TIMESTAMP, " +
		"PRIMARY KEY (id));", database)

	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)", dbUser, dbPassword, ipAddress, port))

	if err != nil {
		panic(err.Error())
	}

	defer db.Close()

	insert, err := db.Query(createSchemaVersion)

	if err != nil {
		panic(err.Error())
	}

	defer insert.Close()
}

func executeMigrations(options DatabaseOptions, schemas []Schema) {
	db := createDbConnection(options)
	for _, s := range schemas {
		command(db, s.name)
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

func getAllDbMigrations(details DatabaseOptions) []Schema {
	db, err := sql.Open("mysql", "sqltracking:sqltracking@tcp(127.0.0.1:3306)/demodb")

	if err != nil {
		log.Print(err.Error())
	}

	defer db.Close()

	results, err := db.Query("SELECT `id`, `name`, `dateexecuted` FROM `schemaversion`")

	if err != nil {
		panic(err.Error())
	}

	var schemas []Schema

	for results.Next() {
		var schema Schema
		var dateExecuted rawTime
		err = results.Scan(&schema.id, &schema.name, &dateExecuted)
		if err != nil {
			panic(err.Error())
		}
		expectedTime, err := dateExecuted.Parse()
		if err != nil {
			panic("The datetime was in an unexpected format. expected: YYYY-MM-DD hh:mm:ss")
		}
		schema.dateexecuted = expectedTime
		schemas = append(schemas, schema)
	}
	return schemas
}