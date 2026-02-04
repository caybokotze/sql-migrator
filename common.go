package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type DatabaseOptions struct {
	SqlUser            string `json:"sqlUser"`
	SqlPassword        string `json:"sqlPassword"`
	SqlHost            string `json:"sqlHost"`
	SqlPort            string `json:"sqlPort"`
	SqlDatabase        string `json:"sqlDatabase"`
	DryRun             bool   `json:"dryRun"`
	AutoByPass         bool   `json:"autoByPass"`
	MigrationTableName string `json:"migrationTableName"`
	Verbose            bool
	ConfigFileName     string
	MigrationDirectory string
}

type ConnectionWithOptions struct {
	Conn               *sql.DB
	DryRun             bool
	AutoByPass         bool
	MigrationTableName string
}

type Schema struct {
	id           int64
	name         string
	dateExecuted time.Time
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
		Conn:               db,
		DryRun:             options.DryRun,
		AutoByPass:         options.AutoByPass,
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
		_ = transaction.Rollback()
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

func query(dbConnectionWithOptions ConnectionWithOptions, query string) *sql.Rows {
	result, err := dbConnectionWithOptions.Conn.Query(query)
	if err != nil {
		panic(err.Error())
	}
	return result
}

func (t rawTime) Parse() (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", string(t))
}

func loadConfigFromJsonFile(fileName string) DatabaseOptions {
	if fileName == "" {
		log.Fatal("The specified configuration file name cannot be empty")
	}
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("Could not open configuration file '%s': %s", fileName, err.Error())
	}
	defer file.Close()

	var jsonPackage Package
	content, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("Could not read configuration file '%s': %s", fileName, err.Error())
	}
	if err = json.Unmarshal(content, &jsonPackage); err != nil {
		log.Fatalf("Could not parse configuration file '%s': %s", fileName, err.Error())
	}
	if jsonPackage.DatabaseConfiguration.MigrationTableName == "" {
		jsonPackage.DatabaseConfiguration.MigrationTableName = "__migrations"
	}
	return jsonPackage.DatabaseConfiguration
}

func openFilesInFileEditor(upScript, downScript, migrationDirectory string) {
	upPath := fmt.Sprintf("%s/%s.sql", migrationDirectory, upScript)
	downPath := fmt.Sprintf("%s/%s.sql", migrationDirectory, downScript)

	// Check if VS Code is available in PATH
	if path, present := os.LookupEnv("path"); present && strings.Contains(path, "VS Code") {
		cmd := exec.Command("code", upPath, downPath)
		_ = cmd.Start()
		return
	}

	// Fall back to platform-specific default editor
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", upPath)
	case "darwin":
		cmd = exec.Command("open", upPath, downPath)
	case "linux":
		cmd = exec.Command("xdg-open", upPath)
	default:
		return
	}
	_ = cmd.Start()
}
