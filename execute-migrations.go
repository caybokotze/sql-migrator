package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
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

	createSchemaVersionTable(
		sqlUser,
		sqlPassword,
		sqlHost,
		sqlPort,
		sqlDatabase)

	var excludedMigrations []Schema = findExcludedMigrations()
}

func findExcludedMigrations() []Schema {
	executedMigrations := getAllDbMigrations()
	allFileMigrations := getArrayOfMigrationFiles()


	return allFileMigrations
}

func getArrayOfMigrationFiles() []Schema {
	items, _ := ioutil.ReadDir("./scripts/")
	var schemas []Schema
	for _, item := range items {
		schemas = append(schemas, getSchemaFromFileName(item.Name()))
	}
	return schemas
}

func getSchemaFromFileName(fileName string) Schema {
	return Schema{
		id:           0,
		name:         "",
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

func getAllDbMigrations() []Schema {
	db, err := sql.Open("mysql", "root:pass1@tcp(127.0.0.1:3306)/tuts")

	if err != nil {
		log.Print(err.Error())
	}

	defer db.Close()

	results, err := db.Query("SELECT `id`, `name`, `dateexecuted` FROM `schemaversion`")

	if err != nil {
		panic(err.Error())
	}

	for results.Next() {
		var schema Schema

		err = results.Scan(&schema.id, &schema.name)
		if err != nil {
			panic(err.Error())
		}

		log.Printf(schema.name)
	}

	var schemas = []Schema {
		{
			id:           0,
			name:         "",
			dateexecuted: time.Now(),
		},
	}
	return schemas
}