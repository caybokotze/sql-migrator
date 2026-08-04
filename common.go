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
	_ "github.com/lib/pq"
)

type DatabaseOptions struct {
	SqlDriver          string `json:"sqlDriver"`
	SqlUser            string `json:"sqlUser"`
	SqlPassword        string `json:"sqlPassword"`
	SqlHost            string `json:"sqlHost"`
	SqlPort            string `json:"sqlPort"`
	SqlDatabase        string `json:"sqlDatabase"`
	SqlSchema          string `json:"sqlSchema"`
	SqlSslMode         string `json:"sqlSslMode"`
	SchemaQuery        string `json:"schemaQuery"`
	DryRun             bool   `json:"dryRun"`
	AutoByPass         bool   `json:"autoByPass"`
	MigrationTableName string `json:"migrationTableName"`
	Verbose            bool
	ConfigFileName     string
	MigrationDirectory string
}

type ConnectionWithOptions struct {
	Conn               *sql.DB
	Dialect            Dialect
	DryRun             bool
	AutoByPass         bool
	MigrationTableName string
}

// Schema represents one migration. Depending on where it was read from,
// not every field is populated:
//   - fileBase is set when the migration was found on disk
//     (e.g. "20250319205322_Create_Users" — the exact file name minus the
//     _up.sql/_down.sql suffix, underscores preserved)
//   - scriptName is set when the migration was read from a postgres
//     tracking table (the exact stored script_name)
//   - name is the display name with underscores collapsed (legacy format,
//     also what the mysql tracking table stores)
type Schema struct {
	id           int64
	name         string
	fileBase     string
	scriptName   string
	dateExecuted time.Time
}

// UpScriptName returns the file name of the up script, preferring the exact
// name recorded in the database over one reconstructed from disk.
func (s Schema) UpScriptName() string {
	if s.scriptName != "" {
		return s.scriptName
	}
	return s.fileBase + "_up.sql"
}

// DisplayName renders the migration for log output.
func (s Schema) DisplayName() string {
	if s.fileBase != "" {
		return s.fileBase
	}
	return fmt.Sprintf("%d_%s", s.id, s.name)
}

type rawTime []byte

type Package struct {
	DatabaseConfiguration DatabaseOptions `json:"sql-migrator-config"`
}

func createDbConnection(options DatabaseOptions) (ConnectionWithOptions, error) {
	dialect, err := resolveDialect(options.SqlDriver)
	if err != nil {
		return ConnectionWithOptions{}, err
	}

	db, err := sql.Open(dialect.DriverName(), dialect.BuildDsn(options))
	if err != nil {
		return ConnectionWithOptions{}, fmt.Errorf("could not establish a connection to the db: %s", err.Error())
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return ConnectionWithOptions{}, fmt.Errorf("could not connect to %s:%s/%s: %s",
			options.SqlHost, options.SqlPort, options.SqlDatabase, err.Error())
	}

	return ConnectionWithOptions{
		Conn:               db,
		Dialect:            dialect,
		DryRun:             options.DryRun,
		AutoByPass:         options.AutoByPass,
		MigrationTableName: options.MigrationTableName,
	}, nil
}

func command(dbConnectionWithOptions ConnectionWithOptions, command string, args ...interface{}) error {
	transaction, txErr := dbConnectionWithOptions.Conn.Begin()
	if txErr != nil {
		return txErr
	}
	_, execErr := transaction.Exec(command, args...)
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
