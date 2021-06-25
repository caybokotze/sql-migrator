package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DatabaseOptions struct {
	sqlUser string
	sqlPassword string
	sqlHost string
	sqlPort string
	sqlDatabase string
}

type Schema struct {
	id int64
	name string
	dateexecuted time.Time
}

func createDbConnection(options DatabaseOptions) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)",
		options.sqlUser,
		options.sqlPassword,
		options.sqlHost,
		options.sqlPort))
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Could not establish connection to the db.")
	}
	defer db.Close()
	return db
}

func command(dbInstance *sql.DB, command string)*sql.Rows {
	insert, err := dbInstance.Query(command)
	if err != nil {
		panic(err.Error())
	}
	return insert
}

func query(dbInstance *sql.DB, query string)*sql.Rows {
	result, err := dbInstance.Query(query)
	if err != nil {
		panic(err.Error())
	}
	return result
}

type rawTime []byte

func (t rawTime) Parse() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", string(t))
}