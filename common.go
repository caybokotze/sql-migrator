package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type DatabaseOptions struct {
	SqlUser     string `json:"sqlUser"`
	SqlPassword string `json:"sqlPassword"`
	SqlHost     string `json:"sqlHost"`
	SqlPort     string `json:"sqlPort"`
	SqlDatabase string `json:"sqlDatabase"`
	DryRun      bool `json:"dryRun"`
	AutoByPass  bool `json:"autoByPass"`
	MigrationTableName string `json:"migrationTableName"`
	Verbose bool
}

type ConnectionWithOptions struct {
	Conn *sql.DB
	DryRun bool
	AutoByPass bool
	MigrationTableName string
}

type Schema struct {
	id int64
	name string
	dateexecuted time.Time
}

type rawTime []byte

type Package struct {
	DatabaseConfiguration DatabaseOptions `json:"sql-migrator-config"`
}

func createDbConnection(options DatabaseOptions) ConnectionWithOptions {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?multiStatements=true",
		options.SqlUser,
		options.SqlPassword,
		options.SqlHost,
		options.SqlPort,
		options.SqlDatabase))
	if err != nil {
		log.Println(err.Error())
		log.Fatal("Could not establish connection to the db.")
	}

	return ConnectionWithOptions{
		Conn: db,
		DryRun: options.DryRun,
		AutoByPass: options.AutoByPass,
		MigrationTableName: options.MigrationTableName,
	}
}


func command(dbConnectionWithOptions ConnectionWithOptions, command string) error {
	transaction, txErr := dbConnectionWithOptions.Conn.Begin()
	if txErr != nil {
		return txErr
	}
	_, execErr := transaction.Exec(command)
	if execErr != nil {
		//_ = transaction.Rollback()
		return execErr
	}
	if dbConnectionWithOptions.DryRun {
		_ = transaction.Rollback()
		return nil
	}
	err := transaction.Commit()
	if err != nil {
		_ = transaction.Rollback()
		return err
	}
	return nil
}

func query(dbConnectionWithOptions ConnectionWithOptions, query string)*sql.Rows {
	result, err := dbConnectionWithOptions.Conn.Query(query)
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
	_ = file.Close()
	err = json.Unmarshal(stringBytes, &jsonPackage)
	if err != nil {
		panic("Could not unmarshal json file")
	}
	if jsonPackage.DatabaseConfiguration.MigrationTableName == "" {
		jsonPackage.DatabaseConfiguration.MigrationTableName = "__db_migrations"
	}
	return jsonPackage.DatabaseConfiguration
}

func openFilesInFileEditor(upScript string, downScript string) {
	var vsCodeIsAvailable = false
	val, present := os.LookupEnv("path")
	if present {
		if strings.Contains(val, "VS Code") {
			vsCodeIsAvailable = true
		}
	}
	if vsCodeIsAvailable {
		executeOSCommand("code", upScript, downScript)
		return
	}
	// todo: refactor this out into separate files
	if !vsCodeIsAvailable {
		if runtime.GOOS == "windows" {
			executeOSCommand("start", upScript, downScript)
		}
		if runtime.GOOS == "osx" {
			executeOSCommand("start", upScript, downScript)
		}
		if runtime.GOOS == "linux" {
			executeOSCommand("start", upScript, downScript)
		}
	}
}

func executeOSCommand(command, upScript, downScript string) {
	cmd := exec.Command(command,
		fmt.Sprintf("scripts/%s.sql", upScript),
		fmt.Sprintf("scripts/%s.sql", downScript))
	_ = cmd.Start()
}