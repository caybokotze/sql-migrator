package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type DatabaseOptions struct {
	sqlUser string `json:"dbUser"`
	sqlPassword string `json:"dbPassword"`
	sqlHost string `json:"sqlHost"`
	sqlPort string `json:"sqlPort"`
	sqlDatabase string `json:"sqlDatabase"`
}

type Schema struct {
	id int64
	name string
	dateexecuted time.Time
}

type rawTime []byte

type Package struct {
	config DatabaseOptions `json:"config"`
}

func createDbConnection(options DatabaseOptions) *sql.DB {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		options.sqlUser,
		options.sqlPassword,
		options.sqlHost,
		options.sqlPort,
		options.sqlDatabase))
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Could not establish connection to the db.")
	}
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

func (t rawTime) Parse() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", string(t))
}

func loadConfigFromJsonFile() DatabaseOptions {
	file, err := os.Open("package.json")
	if err != nil {
		panic("Could not open package.json file")
	}
	var jsonPackage Package
	stringBytes, _ := ioutil.ReadAll(file)
	defer file.Close()
	err = json.Unmarshal(stringBytes, &jsonPackage)
	if err != nil {
		panic("Could not unmarshal json file")
	}
	return jsonPackage.config
}

func removeFromSlice(slice []Schema, s Schema) []Schema {
	return append(slice[:s.id], slice[s.id+1:]...)
}

func parseIdDateToDate(str string) time.Time {
	t, err := time.Parse("20060102150405", str)
	if err != nil {
		panic(err.Error())
	}
	return t
}